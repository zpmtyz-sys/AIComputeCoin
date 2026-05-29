package hardware

import (
	"fmt"
	"strings"
)

// AntiSpoofChecker validates GPU hardware information for consistency.
type AntiSpoofChecker struct{}

// NewAntiSpoofChecker creates a new AntiSpoofChecker.
func NewAntiSpoofChecker() *AntiSpoofChecker {
	return &AntiSpoofChecker{}
}

// Validate checks GPU info for spoofing indicators.
func (c *AntiSpoofChecker) Validate(info *GPUInfo) error {
	if err := c.validatePCISlot(info); err != nil {
		return err
	}
	if err := c.validateDriverVersion(info); err != nil {
		return err
	}
	if err := c.validateVRAM(info); err != nil {
		return err
	}
	return nil
}

// validatePCISlot checks PCI device tree consistency.
func (c *AntiSpoofChecker) validatePCISlot(info *GPUInfo) error {
	// PCI slot should follow standard format: DDDD:BB:DD.F
	parts := strings.Split(info.PCISlot, ":")
	if len(parts) != 3 {
		return fmt.Errorf("invalid PCI slot format: %s", info.PCISlot)
	}
	// Domain should be 4 hex chars
	if len(parts[0]) != 4 {
		return fmt.Errorf("invalid PCI domain in slot: %s", info.PCISlot)
	}
	// Bus should be 2 hex chars
	if len(parts[1]) != 2 {
		return fmt.Errorf("invalid PCI bus in slot: %s", info.PCISlot)
	}
	// Device.Function should contain a dot
	if !strings.Contains(parts[2], ".") {
		return fmt.Errorf("invalid PCI device.function in slot: %s", info.PCISlot)
	}
	return nil
}

// validateDriverVersion checks driver version matches GPU model expectations.
func (c *AntiSpoofChecker) validateDriverVersion(info *GPUInfo) error {
	profile, ok := KnownGPUProfiles[info.Model]
	if !ok {
		// Unknown model, cannot validate
		return nil
	}
	if !strings.HasPrefix(info.DriverVersion, profile.DriverPrefix) {
		return fmt.Errorf("driver version %s does not match expected prefix %s for %s",
			info.DriverVersion, profile.DriverPrefix, info.Model)
	}
	return nil
}

// validateVRAM checks VRAM capacity matches known model specs.
func (c *AntiSpoofChecker) validateVRAM(info *GPUInfo) error {
	profile, ok := KnownGPUProfiles[info.Model]
	if !ok {
		// Unknown model, cannot validate
		return nil
	}
	if info.VRAM < profile.MinVRAM || info.VRAM > profile.MaxVRAM {
		return fmt.Errorf("VRAM %d MB does not match expected range [%d, %d] for %s",
			info.VRAM, profile.MinVRAM, profile.MaxVRAM, info.Model)
	}
	return nil
}
