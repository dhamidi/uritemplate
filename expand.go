package uritemplate

import (
	"io"
	"strings"
)

// Expand applies the given variable values to the template, producing a URI string.
func (t *Template) Expand(vars Values) (string, error) {
	var buf strings.Builder
	for _, p := range t.parts {
		if err := p.expand(&buf, vars); err != nil {
			return "", err
		}
	}
	return buf.String(), nil
}

func (l literal) expand(w io.Writer, vars Values) error {
	_, err := io.WriteString(w, string(l))
	return err
}

func (e *expression) expand(w io.Writer, vars Values) error {
	op := e.operator
	encode := encodeUnreserved
	if op.allow == encUR {
		encode = encodeReservedAndUnreserved
	}

	var parts []string
	for _, vs := range e.vars {
		val, ok := vars[vs.name]
		if !ok || !val.defined {
			continue
		}
		expanded := expandVar(op, vs, val, encode)
		parts = append(parts, expanded...)
	}

	if len(parts) == 0 {
		return nil
	}

	result := op.prefix + strings.Join(parts, op.sep)
	_, err := io.WriteString(w, result)
	return err
}

func expandVar(op operator, vs varSpec, val Value, encode func(string) string) []string {
	switch val.kind {
	case kindString:
		return expandString(op, vs, val.str, encode)
	case kindList:
		return expandList(op, vs, val.list, encode)
	case kindKeys:
		return expandKeys(op, vs, val.keys, encode)
	}
	return nil
}

func expandString(op operator, vs varSpec, s string, encode func(string) string) []string {
	v := s
	if vs.modifier == modPrefix && vs.maxLen < len(v) {
		v = truncateUTF8(v, vs.maxLen)
	}
	v = encode(v)

	if op.named {
		if v == "" {
			return []string{encode(vs.name) + op.ifemp}
		}
		return []string{encode(vs.name) + "=" + v}
	}
	return []string{v}
}

func expandList(op operator, vs varSpec, items []string, encode func(string) string) []string {
	if len(items) == 0 {
		return nil
	}

	if vs.modifier == modExplode {
		var parts []string
		for _, item := range items {
			v := encode(item)
			if op.named {
				if v == "" {
					parts = append(parts, encode(vs.name)+op.ifemp)
				} else {
					parts = append(parts, encode(vs.name)+"="+v)
				}
			} else {
				parts = append(parts, v)
			}
		}
		return parts
	}

	// Without explode: join with comma
	encoded := make([]string, len(items))
	for i, item := range items {
		encoded[i] = encode(item)
	}
	v := strings.Join(encoded, ",")

	if op.named {
		if v == "" {
			return []string{encode(vs.name) + op.ifemp}
		}
		return []string{encode(vs.name) + "=" + v}
	}
	return []string{v}
}

func expandKeys(op operator, vs varSpec, pairs []KeyValue, encode func(string) string) []string {
	if len(pairs) == 0 {
		return nil
	}

	if vs.modifier == modExplode {
		var parts []string
		for _, kv := range pairs {
			k := encode(kv.Key)
			v := encode(kv.Value)
			if op.named {
				if v == "" {
					parts = append(parts, k+op.ifemp)
				} else {
					parts = append(parts, k+"="+v)
				}
			} else {
				parts = append(parts, k+"="+v)
			}
		}
		return parts
	}

	// Without explode: key,value,key,value joined with comma
	var encoded []string
	for _, kv := range pairs {
		encoded = append(encoded, encode(kv.Key), encode(kv.Value))
	}
	v := strings.Join(encoded, ",")

	if op.named {
		if v == "" {
			return []string{encode(vs.name) + op.ifemp}
		}
		return []string{encode(vs.name) + "=" + v}
	}
	return []string{v}
}

// truncateUTF8 truncates a string to at most maxLen Unicode code points.
func truncateUTF8(s string, maxLen int) string {
	count := 0
	for i := range s {
		if count == maxLen {
			return s[:i]
		}
		count++
	}
	return s
}
