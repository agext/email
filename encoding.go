package email

const (
	hextable    = "0123456789ABCDEF"
	base64table = "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/"
)

// QuotedPrintableEncode encodes the src data using the quoted-printable content
// transfer encoding specified by RFC 2045. Although RFC 2045 does not require that
// UTF multi-byte characters be kept on the same line of encoded text, this function
// does so.
func QuotedPrintableEncode(src []byte) []byte {
	srcLen := len(src)
	if srcLen == 0 {
		return []byte{}
	}
	// guestimate max size of dst, trying to avoid reallocation on append
	dst := make([]byte, 0, 2*srcLen)
	pos := 0

	var (
		c  byte
		le int
		// 'ending in space'; does the encoded text end in a whitespace ?
		eis bool
	)
	enc := make([]byte, 0, 12) // enough for encoding a 4-byte utf symbol
	for i := 0; i < srcLen; i++ {
		enc, eis = enc[:0], false
		switch c = src[i]; {
		case c == '\t', c == ' ':
			enc = append(enc, c)
			eis = true
		case '!' <= c && c <= '~' && c != '=':
			enc = append(enc, c)
		case c&0xC0 == 0xC0:
			// start of utf-8 rune; subsequent bytes always have the top two bits set to 10.
			enc = append(make([]byte, 0, 12), '=', hextable[c>>4], hextable[c&0x0f])
			for i++; i < srcLen; i++ {
				c = src[i]
				if c&0xC0 != 0x80 {
					// stepped past the end of the rune; step back and break out
					i--
					break
				}
				enc = append(enc, '=', hextable[c>>4], hextable[c&0x0f])
			}
		default:
			enc = append(enc, '=', hextable[c>>4], hextable[c&0x0f])
		}
		le = len(enc)
		if pos += le; pos > 75 { // max 76; need room for '='
			dst = append(dst, []byte("=\r\n")...)
			pos = le
		}
		dst = append(dst, enc...)
	}
	if eis {
		dst = append(dst, '=')
	}
	return dst
}

// QEncode encodes the src data using the q-encoding encoded-word syntax specified
// by RFC 2047. Since RFC 2047 requires that each line of a header that includes
// encoded-word text be no longer than 76, this function takes an offset argument
// for the length of the current header line already used up, e.g. by the header
// name, colon and space.
func QEncode(src []byte, offset int) (dst []byte, pos int) {
	srcLen := len(src)
	if srcLen == 0 {
		return []byte{}, offset
	}

	// guestimate max size of dst, trying to avoid reallocation on append
	dst = make([]byte, 0, 12+2*srcLen)

	if offset < 1 {
		// header line can be max 76, but encoded-words can only be max 75;
		// on subsequent lines, if any, the leading space evens things out,
		// but if the first line is empty, we need to pretend it has one char.
		offset = 1
	}
	// count in the 10 chars of "=?utf-8?q?", but do not add them yet! There is
	// a chance that we cannot fit even one encoded character on the first line,
	// but we won't know its length until we encoded it.
	pos = 10 + offset

	var (
		c  byte
		le int
	)
	enc := make([]byte, 0, 12) // enough for encoding a 4-byte utf symbol
	for i := 0; i < srcLen; i++ {
		enc = enc[:0]
		switch c = src[i]; {
		case c == ' ':
			enc = append(enc, '_')
		case '!' <= c && c <= '~' && c != '=' && c != '?' && c != '_':
			enc = append(enc, c)
		case c&0xC0 == 0xC0:
			// start of utf-8 rune; subsequent bytes always have the top two bits set to 10.
			enc = append(make([]byte, 0, 12), '=', hextable[c>>4], hextable[c&0x0f])
			for i++; i < srcLen; i++ {
				c = src[i]
				if c&0xC0 != 0x80 {
					// stepped past the end of the rune; step back and break out
					i--
					break
				}
				enc = append(enc, '=', hextable[c>>4], hextable[c&0x0f])
			}
		default:
			enc = append(enc, '=', hextable[c>>4], hextable[c&0x0f])
		}
		le = len(enc)
		if pos += le; pos > 74 { // max 76; need room for '?='
			if len(dst) > 0 {
				dst = append(dst, []byte("?=\r\n =?utf-8?q?")...)
			} else {
				// the first encoded char doesn't fit on the first line, so
				// start a new line and the encoded-word
				dst = append(dst, []byte("\r\n =?utf-8?q?")...)
			}
			pos = le + 11
		} else {
			if len(dst) == 0 {
				// the first encoded char fits on the first line, so start the encoded-word
				dst = append(dst, []byte("=?utf-8?q?")...)
			}
		}
		dst = append(dst, enc...)
	}
	dst = append(dst, '?', '=')
	pos += 2
	return
}

// QEncodeIfNeeded q-encodes the src data only if it contains 'unsafe' characters.
func QEncodeIfNeeded(src []byte, offset int) (dst []byte) {
	safe := true
	for i, sl := 0, len(src); i < sl && safe; i++ {
		safe = ' ' <= src[i] && src[i] <= '~'
	}
	if safe {
		return src
	}
	dst, _ = QEncode(src, offset)
	return dst
}

// Base64Encode encodes the src data using the base64 content transfer encoding
// specified by RFC 2045. The result is the equivalent of base64-encoding src using
// StdEncoding from the standard package encoding/base64, then breaking it into
// lines of maximum 76 characters, separated by CRLF. Besides convenience, this
// function also has the advantage of combining the encoding and line-breaking
// steps into a single pass, with a single buffer allocation.
func Base64Encode(src []byte) []byte {
	if len(src) == 0 {
		return []byte{}
	}
	dstLen := ((len(src) + 2) / 3 * 4) // base64 encoded length
	dstLen += (dstLen - 1) / 76 * 2    // add 2 bytes for each full 76-char line
	dst := make([]byte, dstLen)
	// fmt.Println(len(src), dstLen)

	var (
		p [4]int
	)

	for pos, lpos := 0, 0; len(src) > 0; {
		// fmt.Println("step", pos, len(src), len(dst))
		switch 76 - lpos {
		case 0:
			dst[pos], dst[pos+1] = '\r', '\n'
			p[0], p[1], p[2], p[3] = pos+2, pos+3, pos+4, pos+5
			pos += 6
			lpos = 4
		case 1:
			dst[pos+1], dst[pos+2] = '\r', '\n'
			p[0], p[1], p[2], p[3] = pos, pos+3, pos+4, pos+5
			pos += 6
			lpos = 3
		case 2:
			dst[pos+2], dst[pos+3] = '\r', '\n'
			p[0], p[1], p[2], p[3] = pos, pos+1, pos+4, pos+5
			pos += 6
			lpos = 2
		case 3:
			dst[pos+3], dst[pos+4] = '\r', '\n'
			p[0], p[1], p[2], p[3] = pos, pos+1, pos+2, pos+5
			pos += 6
			lpos = 1
		default:
			p[0], p[1], p[2], p[3] = pos, pos+1, pos+2, pos+3
			pos += 4
			lpos += 4
		}

		switch len(src) {
		case 1:
			dst[p[3]], dst[p[2]] = '=', '='
			dst[p[1]] = base64table[(src[0]<<4)&0x3F]
			dst[p[0]] = base64table[src[0]>>2]
			return dst
		case 2:
			dst[p[3]] = '='
			dst[p[2]] = base64table[(src[1]<<2)&0x3F]
			dst[p[1]] = base64table[(src[1]>>4)|(src[0]<<4)&0x3F]
			dst[p[0]] = base64table[src[0]>>2]
			return dst
		default:
			dst[p[3]] = base64table[src[2]&0x3F]
			dst[p[2]] = base64table[(src[2]>>6)|(src[1]<<2)&0x3F]
			dst[p[1]] = base64table[(src[1]>>4)|(src[0]<<4)&0x3F]
			dst[p[0]] = base64table[src[0]>>2]
			src = src[3:]
		}
	}

	return dst
}
