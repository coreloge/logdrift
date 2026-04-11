package replay

import (
	"fmt"
	"time"
)

// Preset names for common replay speeds.
const (
	SpeedInstant = "instant"
	SpeedSlow    = "slow"
	SpeedNormal  = "normal"
	SpeedFast    = "fast"
)

// PresetDelay maps a named speed preset to a per-line delay duration.
// It returns an error for unrecognised preset names.
func PresetDelay(name string) (time.Duration, error) {
	switch name {
	case SpeedInstant:
		return 0, nil
	case SpeedSlow:
		return 200 * time.Millisecond, nil
	case SpeedNormal:
		return 50 * time.Millisecond, nil
	case SpeedFast:
		return 5 * time.Millisecond, nil
	default:
		return 0, fmt.Errorf("replay: unknown speed preset %q (choose: instant, slow, normal, fast)", name)
	}
}

// OptionsFromPreset builds Options using a named speed preset and service label.
func OptionsFromPreset(preset, service string) (Options, error) {
	delay, err := PresetDelay(preset)
	if err != nil {
		return Options{}, err
	}
	return Options{
		DelayPerLine: delay,
		Service:      service,
	}, nil
}
