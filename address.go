package email

import (
	"errors"
)

// Address represents a human-friendly email address: a name plus the actual address.
type Address struct {
	Name string
	Addr string
}

// NewAddress creates a new Address enforcing a very basic validity check - see `SeemsValidAddr`.
func NewAddress(name, addr string) (*Address, error) {
	if !SeemsValidAddr(addr) {
		return nil, errors.New("invalid address: " + addr)
	}
	return &Address{name, addr}, nil
}

// SeemsValidAddr does a very loose check on addr, to weed out obviously invalid addresses.
// This function only checks that addr contains one and only one '@', followed by a domain name
// that has a TLD part.
func SeemsValidAddr(addr string) bool {
	var seenAt, seenDom, seenDot, seenTld bool

	for _, char := range addr {
		switch char {
		case '@':
			if seenAt { // more than one '@'
				return false
			}
			seenAt = true
		case '.':
			seenDot = seenAt && seenDom // only care about '.' after '@' and domain name
		default:
			if '!' > char || char > '~' {
				return false
			}
			if seenAt {
				// https://tools.ietf.org/html/rfc5322#section-3.4.1
				if char == '[' || char == ']' || char == '\\' {
					return false
				}
				seenDom = !seenDot
				seenTld = seenDot
			}
		}
	}
	return seenTld
}

// Clone creates a new Address with the same contents as the receiver.
func (a *Address) Clone() *Address {
	if a == nil {
		return nil
	}
	return &Address{a.Name, a.Addr}
}

// Domain extracts the domain portion of the email address in the receiver.
func (a *Address) Domain() string {
	for i := len(a.Addr) - 1; i > -1; i-- {
		if a.Addr[i] == '@' {
			return a.Addr[i+1:]
		}
	}
	return ""
}

func (a *Address) encode(offset int) (dst []byte, pos int) {
	la := len(a.Addr)
	if ln := len(a.Name); ln > 0 {
		nq, safe := 0, true
		for i := 0; i < ln && safe; i++ {
			c := a.Name[i]
			safe = ' ' <= c && c <= '~'
			if c == '\\' || c == '"' {
				nq++
			}
		}
		if safe {
			dst = make([]byte, 0, ln+nq+la+7) // 2*'"'+(' ' or "\r\n ")+'<'+'>'
			dst = append(dst, '"')
			for i := 0; i < ln; i++ {
				c := a.Name[i]
				if c == '\\' || c == '"' {
					dst = append(dst, '\\')
				}
				dst = append(dst, c)
			}

			offset += ln + nq + 3 // 2*'"'+' '
			if offset+la <= 74 {  // max 76; need room for '<' and '>'
				dst = append(dst, '"', ' ')
			} else {
				dst = append(dst, '"', '\r', '\n', ' ')
				offset = 1
			}
		} else {
			var buf []byte
			buf, offset = QEncode([]byte(a.Name), offset)
			dst = make([]byte, len(buf), len(buf)+la+5) // (' ' or "\r\n ")+'<'+'>'
			copy(dst, buf)
			offset++
			if offset+la <= 74 { // max 76; need room for '<' and '>'
				dst = append(dst, ' ')
			} else {
				dst = append(dst, '\r', '\n', ' ')
				offset = 1
			}
		}
	} else {
		if offset+la <= 74 { // max 76; need room for '<' and '>'
			dst = make([]byte, 0, la+2)
		} else {
			dst = make([]byte, 0, la+5)
			dst = append(dst, '\r', '\n', ' ')
			offset = 1
		}
	}
	dst = append(dst, '<')
	dst = append(dst, []byte(a.Addr)...)
	dst = append(dst, '>')
	offset += la + 2
	return dst, offset
}

type addrList []*Address

func (al addrList) Clone() addrList {
	if len(al) == 0 {
		return nil
	}
	cl := make(addrList, len(al))
	for i, a := range al {
		cl[i] = a.Clone()
	}
	return cl
}
