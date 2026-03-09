package uritemplate

// Value represents a template variable value, which can be a string, a list, or a key-value map.
type Value struct {
	str     string
	list    []string
	keys    []KeyValue
	kind    valueKind
	defined bool
}

type valueKind int

const (
	kindString valueKind = iota
	kindList
	kindKeys
)

// KeyValue represents a key-value pair for associative array values.
type KeyValue struct {
	Key   string
	Value string
}

// String creates a string Value.
func String(s string) Value {
	return Value{str: s, kind: kindString, defined: true}
}

// List creates a list Value.
func List(items ...string) Value {
	return Value{list: items, kind: kindList, defined: true}
}

// Keys creates an associative array Value.
func Keys(pairs ...KeyValue) Value {
	return Value{keys: pairs, kind: kindKeys, defined: true}
}

// Values maps variable names to their values.
type Values map[string]Value
