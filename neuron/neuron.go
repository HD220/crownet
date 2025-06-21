package neuron

import "math"

// Point represents a point in 16-dimensional space.
type Point [16]float64

// NeuronType defines the type of a neuron.
type NeuronType int

const (
	ExcitatoryNeuron NeuronType = iota
	InhibitoryNeuron
	DopaminergicNeuron
	InputNeuron  // Typically excitatory, but with a distinct role
	OutputNeuron // Typically excitatory, but with a distinct role
)

// NeuronState defines the current state of a neuron.
type NeuronState int

const (
	RestingState            NeuronState = iota // Can receive pulses and fire
	FiringState                                // Neuron has just fired in the current cycle
	AbsoluteRefractoryState                    // Cannot fire, regardless of stimulus
	RelativeRefractoryState                    // Can fire, but requires stronger stimulus (higher threshold)
)

// Neuron represents a single neuron in the CrowNet model.
type Neuron struct {
	ID                     int
	Position               Point
	Type                   NeuronType
	State                  NeuronState
	AccumulatedPulse       float64 // Current sum of received pulse values
	BaseFiringThreshold    float64 // Base threshold, can be modulated
	CurrentFiringThreshold float64 // Actual threshold for firing in current state
	LastFiredCycle         int
	CyclesInCurrentState   int   // Tracks cycles spent in refractory states
	Velocity               Point // Current velocity vector for movement (synaptogenesis)
	// Connections are implicit via proximity and pulse propagation
}

// NewNeuron creates a new neuron.
func NewNeuron(id int, pos Point, nType NeuronType, baseThreshold float64) *Neuron {
	return &Neuron{
		ID:                     id,
		Position:               pos,
		Type:                   nType,
		State:                  RestingState,
		AccumulatedPulse:       0.0,
		BaseFiringThreshold:    baseThreshold,
		CurrentFiringThreshold: baseThreshold, // Initially same as base
		LastFiredCycle:         -1,            // -1 indicates never fired
		CyclesInCurrentState:   0,
		Velocity:               Point{}, // Initialize velocity to zero vector
	}
}

// UpdateState progresses the neuron's state machine based on its current state and refractory periods.
// This should be called once per cycle for each neuron.
func (n *Neuron) UpdateState(currentCycle int) {
	n.CyclesInCurrentState++

	switch n.State {
	case FiringState:
		// After firing, neuron enters absolute refractory state.
		n.State = AbsoluteRefractoryState
		n.CyclesInCurrentState = 0
		n.LastFiredCycle = currentCycle
		// AccumulatedPulse is typically reset after firing, or it quickly decays.
		// The problem states "Quando não recebem pulsos, a soma dos pulsos diminui gradativamente".
		// This implies it doesn't necessarily reset to 0 immediately after firing, but will decay.
		// Let's assume it does reset or significantly reduce to prevent immediate re-firing from its own high value.
		// For now, let's reset it. This can be tuned.
		n.AccumulatedPulse = 0 // Reset after firing to prevent re-triggering from the same sum.

	case AbsoluteRefractoryState:
		if n.CyclesInCurrentState >= AbsoluteRefractoryCycles {
			n.State = RelativeRefractoryState
			n.CyclesInCurrentState = 0
			// During relative refractory, threshold might be higher.
			// This is not explicitly in README, but common. For now, keep threshold same.
			// n.CurrentFiringThreshold = n.BaseFiringThreshold * 1.5 // Example
		}

	case RelativeRefractoryState:
		if n.CyclesInCurrentState >= RelativeRefractoryCycles {
			n.State = RestingState
			n.CyclesInCurrentState = 0
			n.CurrentFiringThreshold = n.BaseFiringThreshold // Reset threshold
		}

	case RestingState:
		// Stays in resting state unless it fires.
		// If it fires, its state will be changed to FiringState by the firing logic.
		// Decay accumulated pulse if no new pulses are received (or even with them).
		// This decay should ideally happen *before* checking for firing.
		break // Decay is handled separately for now.
	}
}

// DecayPulseAccumulation handles the gradual decrease of accumulated pulses.
// This should be called each cycle, likely before processing new incoming pulses.
func (n *Neuron) DecayPulseAccumulation() {
	if n.AccumulatedPulse > 0 {
		n.AccumulatedPulse -= AccumulatedPulseDecayRate * n.AccumulatedPulse
		if n.AccumulatedPulse < 0.001 { // Threshold to prevent tiny values
			n.AccumulatedPulse = 0
		}
	} else if n.AccumulatedPulse < 0 { // For inhibitory accumulation, decay towards 0
		n.AccumulatedPulse += AccumulatedPulseDecayRate * math.Abs(n.AccumulatedPulse)
		if math.Abs(n.AccumulatedPulse) < 0.001 {
			n.AccumulatedPulse = 0
		}
	}
}

// ReceivePulse adjusts the neuron's accumulated pulse value.
// Returns true if the neuron fires as a result.
func (n *Neuron) ReceivePulse(value float64, currentCycle int) (fired bool) {
	if n.State == AbsoluteRefractoryState {
		return false // Cannot respond to pulses
	}

	n.AccumulatedPulse += value

	// Check for firing
	if n.State != FiringState && n.AccumulatedPulse >= n.CurrentFiringThreshold {
		// Dopaminergic neurons generate dopamine, they don't "fire" in the same way to propagate typical pulses.
		// Their "firing" is effectively releasing dopamine.
		// Input neurons also fire to signal external input.
		// Output neurons fire to signal network output.
		if n.Type == DopaminergicNeuron {
			// Handle dopamine release separately, not as a standard "fire" for pulse propagation.
			// This might involve setting a flag or directly influencing dopamine levels.
			// For now, let's assume they also go through a "firing" cycle for state management,
			// but their effect (dopamine release) is handled elsewhere.
		}

		n.State = FiringState      // Transition to firing state
		n.CyclesInCurrentState = 0 // Reset counter for the new state
		// n.LastFiredCycle = currentCycle // This is set in UpdateState after FiringState
		return true
	}
	return false
}

// GetBasePulseSign returns the base sign/type of pulse this neuron would emit if it fires.
// This will be modulated by synaptic weights.
// Dopaminergic neurons have a special role and don't emit standard weighted pulses.
func (n *Neuron) GetBasePulseSign() float64 {
	switch n.Type {
	case ExcitatoryNeuron, InputNeuron, OutputNeuron: // Standard excitatory effect
		return 1.0
	case InhibitoryNeuron:
		return -1.0 // Standard inhibitory effect
	case DopaminergicNeuron:
		return 0.0 // Dopaminergic neurons' "pulse" is dopamine release, not a weighted signal in this context.
	default:
		return 0.0
	}
}

// GetEffectivePulseValue is now DEPRECATED for weighted connections.
// Use GetBasePulseSign and multiply by synaptic weight in network logic.
// Keeping it here for now to avoid breaking older test code immediately, but it should be removed.
func (n *Neuron) GetEffectivePulseValue() float64 {
	switch n.Type {
	case ExcitatoryNeuron, InputNeuron, OutputNeuron:
		return PulseExcitatoryValue
	case InhibitoryNeuron:
		return -PulseInhibitoryValue
	case DopaminergicNeuron:
		return 0
	default:
		return 0
	}
}
