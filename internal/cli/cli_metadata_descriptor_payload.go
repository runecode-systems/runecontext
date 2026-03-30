package cli

import (
	"encoding/json"
	"fmt"
)

type descriptorMapCodec struct {
	marshal   func(any) ([]byte, error)
	unmarshal func([]byte, any) error
}

func defaultDescriptorMapCodec() descriptorMapCodec {
	return descriptorMapCodec{
		marshal:   json.Marshal,
		unmarshal: json.Unmarshal,
	}
}

func descriptorMap(descriptor capabilityDescriptor) (map[string]any, error) {
	return descriptorMapWithCodec(descriptor, defaultDescriptorMapCodec())
}

func descriptorMapWithCodec(descriptor capabilityDescriptor, codec descriptorMapCodec) (map[string]any, error) {
	data, err := codec.marshal(descriptor)
	if err != nil {
		return nil, fmt.Errorf("marshal capability descriptor payload: %w", err)
	}
	var value map[string]any
	if err := codec.unmarshal(data, &value); err != nil {
		return nil, fmt.Errorf("unmarshal capability descriptor payload: %w", err)
	}
	return value, nil
}
