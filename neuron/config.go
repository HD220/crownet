package neuron

const (
	// PulsePropagationSpeed is units per cycle
	PulsePropagationSpeed = 0.6

	// PulseExcitatoryValue is the value added by an excitatory pulse
	PulseExcitatoryValue = 0.3
	// PulseInhibitoryValue is the value subtracted by an inhibitory pulse (note: it's positive, subtraction is handled by logic)
	PulseInhibitoryValue = 0.3

	// AccumulatedPulseDecayRate is the factor by which accumulated pulse decays per cycle if no new pulses
	AccumulatedPulseDecayRate = 0.1 // e.g., decays by 10% of current value towards 0

	// Refractory periods in cycles
	AbsoluteRefractoryCycles = 2 // e.g., 2 cycles after firing
	RelativeRefractoryCycles = 3 // e.g., 3 cycles after absolute refractory period ends

	// DefaultFiringThreshold is a common threshold, can be modulated by dopamine/cortisol
	DefaultFiringThreshold = 1.0
)
