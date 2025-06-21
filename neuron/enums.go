package neuron

// Type define o tipo funcional de um neurônio na rede.
type Type int

const (
	// Excitatory transmite sinais que aumentam a probabilidade de disparo do neurônio pós-sináptico.
	Excitatory Type = iota
	// Inhibitory transmite sinais que diminuem a probabilidade de disparo do neurônio pós-sináptico.
	Inhibitory
	// Dopaminergic libera dopamina, modulando a atividade da rede e o aprendizado.
	// Seu "disparo" principal é a liberação química, não necessariamente um pulso propagado da mesma forma.
	Dopaminergic
	// Input recebe estímulos externos e os introduz na rede.
	Input
	// Output representa a saída da rede, seu padrão de atividade é o resultado do processamento.
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
	// Resting é o estado padrão, onde o neurônio pode integrar inputs e disparar.
	Resting State = iota
	// Firing indica que o neurônio acabou de exceder seu limiar e está emitindo um pulso.
	// Este estado é tipicamente transiente dentro de um ciclo.
	Firing
	// AbsoluteRefractory é o período imediatamente após o disparo, durante o qual o neurônio não pode disparar novamente,
	// independentemente da intensidade do estímulo.
	AbsoluteRefractory
	// RelativeRefractory é o período após o refratário absoluto, onde o neurônio pode disparar,
	// mas requer um estímulo mais forte que o normal (limiar de disparo aumentado).
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
```
