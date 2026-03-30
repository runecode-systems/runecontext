package cli

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateCapabilityDescriptorSchemaUsesMetadataOutputInstancePath(t *testing.T) {
	root, err := repoRootForTests()
	if err != nil {
		t.Fatalf("repo root: %v", err)
	}
	descriptor := buildCapabilityDescriptor()
	descriptor.Binary = "invalid-binary"
	err = validateCapabilityDescriptorSchemaAtRoot(schemaRootForTests(root), descriptor)
	if err == nil {
		t.Fatal("expected schema validation failure")
	}
	if !strings.Contains(err.Error(), metadataOutputInstancePath) {
		t.Fatalf("expected validation error to reference %q, got %v", metadataOutputInstancePath, err)
	}
}

func TestDescriptorMapReturnsMarshalError(t *testing.T) {
	_, err := descriptorMapWithCodec(buildCapabilityDescriptor(), descriptorMapCodec{
		marshal: func(any) ([]byte, error) {
			return nil, errors.New("marshal boom")
		},
		unmarshal: defaultDescriptorMapCodec().unmarshal,
	})
	if err == nil || !strings.Contains(err.Error(), "marshal capability descriptor payload") {
		t.Fatalf("expected marshal payload error, got %v", err)
	}
}

func TestDescriptorMapReturnsUnmarshalError(t *testing.T) {
	_, err := descriptorMapWithCodec(buildCapabilityDescriptor(), descriptorMapCodec{
		marshal: defaultDescriptorMapCodec().marshal,
		unmarshal: func([]byte, any) error {
			return errors.New("unmarshal boom")
		},
	})
	if err == nil || !strings.Contains(err.Error(), "unmarshal capability descriptor payload") {
		t.Fatalf("expected unmarshal payload error, got %v", err)
	}
}

func schemaRootForTests(root string) string {
	return filepath.Join(root, "schemas")
}
