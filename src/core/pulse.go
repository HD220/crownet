package core

// Pulse representa um sinal elétrico propagando pela rede.
type Pulse struct {
	OriginNeuronID int        // ID do neurônio que disparou o pulso
	TargetNeuronID int        // ID do neurônio alvo do pulso (em uma implementação ponto a ponto)
	Position       [SpaceDimensions]float64 // Posição atual do pulso (se modelarmos a viagem)
	Strength       float64    // Força do pulso (ex: 0.3 para excitatório, -0.3 para inibitório)
	EmittedCycle   int        // Ciclo em que o pulso foi emitido

	// Para propagação baseada em área/distância, não em alvo específico inicialmente:
	CurrentRadius float64 // Raio atual de propagação do pulso desde a origem
	MaxRadius     float64 // Raio máximo que este pulso pode alcançar (pode ser global ou por tipo de neurônio)
	OriginPosition [SpaceDimensions]float64 // Posição de origem do pulso
	SourceNeuronType NeuronType // Tipo do neurônio que emitiu o pulso
}

// NewPulse cria um novo pulso.
func NewPulse(originID int, originPos [SpaceDimensions]float64, strength float64, emittedCycle int, sourceNeuronType NeuronType) *Pulse {
	return &Pulse{
		OriginNeuronID: originID,
		OriginPosition: originPos,
		Strength:       strength,
		EmittedCycle:   emittedCycle,
		CurrentRadius:  0.0,
		MaxRadius:      8.0, // Distância máxima do espaço, conforme README (ou um valor configurável)
		SourceNeuronType: sourceNeuronType,
	}
}

// UpdatePropagation atualiza o raio de propagação do pulso.
// Retorna true se o pulso ainda estiver ativo/propagando.
func (p *Pulse) UpdatePropagation(pulsePropagationSpeed float64) bool {
	p.CurrentRadius += pulsePropagationSpeed
	return p.CurrentRadius <= p.MaxRadius
}

// GetEffectiveRange foi removido pois a lógica de propagação foi simplificada
// e movida para Network.SimulateCycle, usando oldPulseRadius e newPulseRadius.

// IsDopaminePulse verifica se o pulso é de um neurônio dopaminérgico.
func (p *Pulse) IsDopaminePulse() bool {
	return p.SourceNeuronType == DopaminergicNeuron
}

// GetStrength retorna a força nominal do pulso.
// Efeitos específicos (como os de pulsos dopaminérgicos) são tratados
// na lógica da rede que recebe o pulso.
func (p *Pulse) GetStrength() float64 {
	return p.Strength
}
