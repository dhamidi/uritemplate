package uritemplate

const upperHex = "0123456789ABCDEF"

func isUnreserved(c byte) bool {
	return (c >= 'A' && c <= 'Z') ||
		(c >= 'a' && c <= 'z') ||
		(c >= '0' && c <= '9') ||
		c == '-' || c == '.' || c == '_' || c == '~'
}

func isReserved(c byte) bool {
	switch c {
	case ':', '/', '?', '#', '[', ']', '@', '!', '$', '&', '\'', '(', ')', '*', '+', ',', ';', '=':
		return true
	}
	return false
}

// encodeUnreserved percent-encodes all characters except unreserved (ALPHA, DIGIT, -, ., _, ~).
func encodeUnreserved(s string) string {
	needed := false
	for i := 0; i < len(s); i++ {
		if !isUnreserved(s[i]) {
			needed = true
			break
		}
	}
	if !needed {
		return s
	}

	buf := make([]byte, 0, len(s)*3)
	for i := 0; i < len(s); {
		c := s[i]
		if isUnreserved(c) {
			buf = append(buf, c)
			i++
		} else {
			// encode UTF-8 bytes
			buf = pctEncodeByte(buf, c)
			i++
		}
	}
	return string(buf)
}

// encodeReservedAndUnreserved percent-encodes all characters except unreserved + reserved.
// It also preserves existing valid pct-encoded triplets.
func encodeReservedAndUnreserved(s string) string {
	needed := false
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !isUnreserved(c) && !isReserved(c) && c != '%' {
			needed = true
			break
		}
	}
	if !needed && !containsPercent(s) {
		return s
	}

	buf := make([]byte, 0, len(s)*3)
	for i := 0; i < len(s); {
		c := s[i]
		if isUnreserved(c) || isReserved(c) {
			buf = append(buf, c)
			i++
		} else if c == '%' && i+2 < len(s) && isHex(s[i+1]) && isHex(s[i+2]) {
			// preserve valid pct-encoded triplet
			buf = append(buf, s[i], s[i+1], s[i+2])
			i += 3
		} else {
			buf = pctEncodeByte(buf, c)
			i++
		}
	}
	return string(buf)
}

func pctEncodeByte(buf []byte, c byte) []byte {
	buf = append(buf, '%', upperHex[c>>4], upperHex[c&0x0f])
	return buf
}

func containsPercent(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == '%' {
			return true
		}
	}
	return false
}
