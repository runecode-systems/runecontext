package contracts

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"unicode/utf8"
)

func (p *ContextPack) computePackHash() (string, error) {
	canonical, err := canonicalContextPackHashInput(p)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(canonical)
	return fmt.Sprintf("%x", sum[:]), nil
}

func canonicalContextPackHashInput(pack *ContextPack) ([]byte, error) {
	if pack == nil {
		return nil, fmt.Errorf("context pack is required")
	}
	return marshalCanonicalJSON(contextPackCanonicalValue(pack))
}

func marshalCanonicalJSON(value any) ([]byte, error) {
	var buf bytes.Buffer
	if err := writeCanonicalJSON(&buf, value); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// writeCanonicalJSON intentionally supports only the concrete value shapes that
// context-pack canonical hash inputs emit: maps with string keys, arrays,
// strings, booleans, nil, and integral numbers. This keeps the implementation
// narrowly correct for emitted pack data while still following RFC 8785-style
// ordering and string escaping rules for those values.
func writeCanonicalJSON(buf *bytes.Buffer, value any) error {
	switch typed := value.(type) {
	case nil:
		return writeCanonicalNull(buf)
	case string:
		return writeCanonicalStringValue(buf, typed)
	case bool:
		return writeCanonicalBool(buf, typed)
	case []string:
		return writeCanonicalStrings(buf, typed)
	case []any:
		return writeCanonicalArray(buf, typed)
	case map[string]any:
		return writeCanonicalObject(buf, typed)
	default:
		return writeCanonicalScalar(buf, value)
	}
}

func writeCanonicalNull(buf *bytes.Buffer) error {
	buf.WriteString("null")
	return nil
}

func writeCanonicalStringValue(buf *bytes.Buffer, value string) error {
	if !utf8.ValidString(value) {
		return fmt.Errorf("canonical JSON strings must be valid UTF-8")
	}
	writeCanonicalJSONString(buf, value)
	return nil
}

func writeCanonicalBool(buf *bytes.Buffer, value bool) error {
	if value {
		buf.WriteString("true")
	} else {
		buf.WriteString("false")
	}
	return nil
}

func writeCanonicalStrings(buf *bytes.Buffer, items []string) error {
	return writeCanonicalSequence(buf, len(items), func(index int) error {
		return writeCanonicalJSON(buf, items[index])
	})
}

func writeCanonicalArray(buf *bytes.Buffer, items []any) error {
	return writeCanonicalSequence(buf, len(items), func(index int) error {
		return writeCanonicalJSON(buf, items[index])
	})
}

func writeCanonicalSequence(buf *bytes.Buffer, length int, writeItem func(index int) error) error {
	buf.WriteByte('[')
	for i := range length {
		if i > 0 {
			buf.WriteByte(',')
		}
		if err := writeItem(i); err != nil {
			return err
		}
	}
	buf.WriteByte(']')
	return nil
}

func writeCanonicalObject(buf *bytes.Buffer, value map[string]any) error {
	keys := sortedCanonicalObjectKeys(value)
	buf.WriteByte('{')
	for i, key := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		if err := writeCanonicalJSON(buf, key); err != nil {
			return err
		}
		buf.WriteByte(':')
		if err := writeCanonicalJSON(buf, value[key]); err != nil {
			return err
		}
	}
	buf.WriteByte('}')
	return nil
}

func sortedCanonicalObjectKeys(value map[string]any) []string {
	keys := make([]string, 0, len(value))
	for key := range value {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func writeCanonicalScalar(buf *bytes.Buffer, value any) error {
	if writeCanonicalInteger(buf, value) {
		return nil
	}
	return fmt.Errorf("unsupported canonical JSON value %T", value)
}

func writeCanonicalJSONString(buf *bytes.Buffer, value string) {
	buf.WriteByte('"')
	for _, r := range value {
		switch r {
		case '\\', '"':
			buf.WriteByte('\\')
			buf.WriteRune(r)
		case '\b':
			buf.WriteString(`\b`)
		case '\f':
			buf.WriteString(`\f`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		default:
			if r >= 0 && r < 0x20 {
				buf.WriteString(fmt.Sprintf(`\u%04x`, r))
				continue
			}
			buf.WriteRune(r)
		}
	}
	buf.WriteByte('"')
}

func writeCanonicalInteger(buf *bytes.Buffer, value any) bool {
	rv := reflect.ValueOf(value)
	if !rv.IsValid() {
		return false
	}
	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		buf.WriteString(strconv.FormatInt(rv.Int(), 10))
		return true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		buf.WriteString(strconv.FormatUint(rv.Uint(), 10))
		return true
	default:
		return false
	}
}
