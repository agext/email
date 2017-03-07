package email

type buffer []byte

func newBuffer(size int) *buffer {
	b := buffer(make([]byte, 0, size))
	return &b
}

func (b *buffer) Write(data ...interface{}) {
	for _, value := range data {
		switch v := value.(type) {
		case string:
			*b = append(*b, v...)
		case []byte:
			*b = append(*b, v...)
		case byte:
			*b = append(*b, v)
		case rune:
			*b = append(*b, string(v)...)
		}
	}
}

func (b *buffer) Bytes() []byte {
	return *b
}
