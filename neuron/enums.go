package neuron

// Type defines the functional role of a neuron within the network.
type Type int

const (
	// Excitatory neurons increase the likelihood of firing in postsynaptic neurons.
	Excitatory Type = iota
	// Inhibitory neurons decrease the likelihood of firing in postsynaptic neurons.
	Inhibitory
	// Dopaminergic neurons modulate network activity, often related to learning and reward.
	Dopaminergic
	// Input neurons receive external stimuli and introduce information into the network.
	Input
	// Output neurons represent the network's response or results.
	Output
)

// String returns the textual representation of the neuron Type.
func (nt Type) String() string {
	switch nt {
	case Excitatory:
		return "Excitatory"
	case Inhibitory:
		return "Inhibitory"
	case Dopaminergic:
		return "Dopaminergic"
	case Input:
		return "Input"
	case Output:
		return "Output"
	default:
		return "Unknown"
	}
}

// State defines the operational condition of a neuron at a given time.
type State int

const (
	// Resting state indicates the neuron is stable and not actively firing, but can integrate potentials.
	Resting State = iota
	// Firing state indicates the neuron has reached its threshold and is emitting a pulse.
	Firing
	// AbsoluteRefractory state is a brief period after firing
	// during which the neuron cannot fire again, regardless of input.
	AbsoluteRefractory
	// RelativeRefractory state follows the absolute refractory period;
	// the neuron can fire, but its threshold is elevated.
	RelativeRefractory
)

// String returns the textual representation of the neuron State.
func (ns State) String() string {
	switch ns {
	case Resting:
		return "Resting"
	case Firing:
		return "Firing"
	case AbsoluteRefractory:
		return "AbsoluteRefractory"
	case RelativeRefractory:
		return "RelativeRefractory"
	default:
		return "Unknown"
	}
}
