package contracts

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"reflect"
	"strings"
	"unicode/utf8"
)

const AssuranceCanonicalizationToken = "runecontext-canonical-json-v1"

// AssuranceEnvelope is a minimal portable assurance artifact envelope.
type AssuranceEnvelope struct {
	SchemaVersion    int    `json:"schema_version" yaml:"schema_version"`
	Kind             string `json:"kind" yaml:"kind"`
	SubjectID        string `json:"subject_id" yaml:"subject_id"`
	CreatedAt        int64  `json:"created_at" yaml:"created_at"`
	Canonicalization string `json:"canonicalization" yaml:"canonicalization"`
	Value            any    `json:"value,omitempty" yaml:"value,omitempty"`
}

// ReceiptArtifact extends the envelope with receipt-specific fields.
type ReceiptArtifact struct {
	AssuranceEnvelope
	ReceiptID   string `json:"receipt_id" yaml:"receipt_id"`
	ReceiptHash string `json:"receipt_hash" yaml:"receipt_hash"`
	Provenance  string `json:"provenance,omitempty" yaml:"provenance,omitempty"`
}

// ComputeArtifactCanonicalJSON serializes the provided object into a deterministic
// compact JSON string with sorted object keys. It checks for UTF-8 validity in
// all string values and returns an error if invalid string data is encountered.
func ComputeArtifactCanonicalJSON(obj any) (string, error) {
	c := canonicalize(obj)
	if err := ensureUTF8Strings(c); err != nil {
		return "", err
	}
	b, err := marshalCanonicalJSON(c)
	if err != nil {
		return "", fmt.Errorf("canonical json marshal: %w", err)
	}
	return string(b), nil
}

func ensureUTF8Strings(v any) error {
	switch t := v.(type) {
	case map[string]any:
		return ensureUTF8InMap(t)
	case []any:
		return ensureUTF8InSlice(t)
	case string:
		return ensureUTF8String(t)
	default:
		rv := indirect(reflect.ValueOf(v))
		switch rv.Kind() {
		case reflect.Struct:
			return ensureUTF8InStruct(rv)
		case reflect.Map:
			if rv.Type().Key().Kind() == reflect.String {
				return ensureUTF8InMapValue(rv)
			}
		case reflect.Slice, reflect.Array:
			return ensureUTF8InSliceValue(rv)
		case reflect.Interface, reflect.Pointer:
			if !rv.IsNil() {
				return ensureUTF8Strings(rv.Interface())
			}
		}
		return nil
	}
}

func ensureUTF8InStruct(rv reflect.Value) error {
	rt := rv.Type()
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if field.PkgPath != "" && !field.Anonymous {
			continue
		}
		value := rv.Field(i)
		if err := ensureUTF8Strings(value.Interface()); err != nil {
			return fmt.Errorf("field %s: %w", field.Name, err)
		}
	}
	return nil
}

func ensureUTF8InMapValue(rv reflect.Value) error {
	m := make(map[string]any, rv.Len())
	for _, key := range rv.MapKeys() {
		val := rv.MapIndex(key)
		m[key.String()] = val.Interface()
	}
	return ensureUTF8InMap(m)
}

func ensureUTF8InSliceValue(rv reflect.Value) error {
	length := rv.Len()
	items := make([]any, length)
	for i := 0; i < length; i++ {
		items[i] = rv.Index(i).Interface()
	}
	return ensureUTF8InSlice(items)
}

func ensureUTF8String(s string) error {
	if !utf8.ValidString(s) {
		return fmt.Errorf("invalid UTF-8 string")
	}
	return nil
}

func ensureUTF8InMap(m map[string]any) error {
	for k, val := range m {
		if err := ensureUTF8String(k); err != nil {
			return fmt.Errorf("key %q: %w", k, err)
		}
		if err := ensureUTF8Strings(val); err != nil {
			return fmt.Errorf("key %q: %w", k, err)
		}
	}
	return nil
}

func ensureUTF8InSlice(s []any) error {
	for i, item := range s {
		if err := ensureUTF8Strings(item); err != nil {
			return fmt.Errorf("index %d: %w", i, err)
		}
	}
	return nil
}

// ComputeSHA256Hex returns the lowercase hex SHA256 digest for the given data.
func ComputeSHA256Hex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// GenerateReceiptID creates a collision-resistant receipt id using the kind,
// a subject-derived prefix, the createdAt timestamp, and a random suffix.
func GenerateReceiptID(kind string, subjectID string, createdAt int64) (string, error) {
	sub := sha256.Sum256([]byte(subjectID))
	subPrefix := hex.EncodeToString(sub[:])[:8]
	randBytes := make([]byte, 6)
	if _, err := rand.Read(randBytes); err != nil {
		return "", fmt.Errorf("generate receipt id random bytes: %w", err)
	}
	randHex := hex.EncodeToString(randBytes)
	return fmt.Sprintf("%s-%s-%d-%s", sanitizeForID(kind), subPrefix, createdAt, randHex), nil
}

// GenerateReceiptFilename produces a safe filename for a receipt using the
// receipt id, a short prefix of the receipt hash, and the kind. The result is
// safe to store in a typical filesystem and avoids path separators.
func GenerateReceiptFilename(receiptID string, receiptHash string, kind string) string {
	hashPrefix := receiptHash
	if len(hashPrefix) > 12 {
		hashPrefix = hashPrefix[:12]
	}
	return fmt.Sprintf("%s--%s-%s.json", sanitizeForID(kind), sanitizeForID(receiptID), sanitizeForID(hashPrefix))
}

func BuildCapturedVerifiedReceipt(family, subjectID string, value map[string]any, createdAt int64) (ReceiptArtifact, string, error) {
	family = strings.TrimSpace(family)
	if family == "" {
		return ReceiptArtifact{}, "", fmt.Errorf("receipt family is required")
	}
	subjectID = strings.TrimSpace(subjectID)
	if subjectID == "" {
		return ReceiptArtifact{}, "", fmt.Errorf("receipt subject_id is required")
	}
	if value == nil {
		value = map[string]any{}
	}
	if strings.TrimSpace(fmt.Sprint(value["receipt_family"])) == "" {
		value["receipt_family"] = family
	}
	receiptID, err := GenerateReceiptID(family, subjectID, createdAt)
	if err != nil {
		return ReceiptArtifact{}, "", err
	}
	artifact := ReceiptArtifact{
		AssuranceEnvelope: AssuranceEnvelope{
			SchemaVersion:    1,
			Kind:             "receipt",
			SubjectID:        subjectID,
			CreatedAt:        createdAt,
			Canonicalization: AssuranceCanonicalizationToken,
			Value:            value,
		},
		ReceiptID:  receiptID,
		Provenance: "captured_verified",
	}
	hash, err := ComputeReceiptHash(artifact)
	if err != nil {
		return ReceiptArtifact{}, "", err
	}
	artifact.ReceiptHash = hash
	filename := GenerateReceiptFilename(artifact.ReceiptID, artifact.ReceiptHash, family)
	return artifact, filename, nil
}

func ComputeReceiptHash(artifact ReceiptArtifact) (string, error) {
	input := map[string]any{
		"schema_version":   artifact.SchemaVersion,
		"kind":             artifact.Kind,
		"subject_id":       artifact.SubjectID,
		"created_at":       artifact.CreatedAt,
		"canonicalization": artifact.Canonicalization,
		"provenance":       artifact.Provenance,
		"receipt_id":       artifact.ReceiptID,
		"value":            artifact.Value,
	}
	canonical, err := ComputeArtifactCanonicalJSON(input)
	if err != nil {
		return "", err
	}
	return ComputeSHA256Hex([]byte(canonical)), nil
}

func sanitizeForID(s string) string {
	// Allow a safe subset: alphanumerics, hyphen, underscore. Replace others with '-'.
	out := make([]rune, 0, len(s))
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			out = append(out, r)
		} else {
			out = append(out, '-')
		}
	}
	return string(out)
}
