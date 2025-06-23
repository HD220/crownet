package neuron

// Type define o tipo funcional de um neurônio na rede.
type Type int

const (
	Excitatory Type = iota
	Inhibitory
	Dopaminergic
	Input
	Output
)

// String para representação textual do Type.
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

// State define o estado operacional atual de um neurônio.
type State int

const (
	Resting State = iota
	Firing
	AbsoluteRefractory
	RelativeRefractory
)

// String para representação textual do State.
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
