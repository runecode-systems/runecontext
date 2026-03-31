package contracts

const capabilityDescriptorSchemaVersion = 2

// CapabilityDescriptorSchemaVersionForCLI exposes the canonical metadata
// descriptor schema version used by machine-facing capability descriptor
// surfaces.
func CapabilityDescriptorSchemaVersionForCLI() int {
	return capabilityDescriptorSchemaVersion
}
