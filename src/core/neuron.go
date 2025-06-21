package core

import "math/rand"

const (
	SpaceDimensions = 16 // Dimensões do espaço vetorial
)

// NeuronType define o tipo do neurônio.
type NeuronType int

const (
	ExcitatoryNeuron NeuronType = iota
	InhibitoryNeuron
	DopaminergicNeuron
	InputNeuron // Para entrada de dados na rede
	OutputNeuron // Para saída de dados da rede
)

// NeuronState define o estado atual do neurônio.
type NeuronState int

const (
	RestingState NeuronState = iota // Repouso
	FiringState                   // Disparo
	RefractoryAbsoluteState       // Refratário Absoluto
	RefractoryRelativeState       // Refratário Relativo (no README apenas "Refratário")
)

// Neuron representa um único neurônio na rede.
type Neuron struct {
	ID       int
	Position [SpaceDimensions]float64
	Type     NeuronType
	State    NeuronState

	CurrentPotential float64 // Potencial acumulado dos pulsos recebidos
	FiringThreshold  float64 // Limiar de disparo

	LastFiringCycle    int // Ciclo em que o neurônio disparou pela última vez
	RefractoryCycles   int // Contador para o período refratário
	CyclesInRest       int // Contador de ciclos em repouso
	CyclesInFiring     int // Contador de ciclos em disparo (pode ser útil para frequência)
	CyclesInRefractory int // Contador de ciclos em período refratário

	// Parâmetros de configuração (podem ser movidos para um struct de config)
	BaseFiringThreshold      float64
	RefractoryPeriodAbsolute int // Número de ciclos para o período refratário absoluto
	RefractoryPeriodRelative int // Número de ciclos para o período refratário relativo
}

// Gland representa a glândula de cortisol.
type Gland struct {
	Position [SpaceDimensions]float64
	// Outros atributos relevantes para a glândula, como sensibilidade, etc.
}

// NewNeuron cria uma nova instância de Neuron.
// A posição inicial pode ser gerada aqui ou passada como argumento.
func NewNeuron(id int, nType NeuronType, initialPosition [SpaceDimensions]float64) *Neuron {
	// Valores padrão, podem ser ajustados/configurados
	baseThreshold := 1.0
	if nType == InhibitoryNeuron {
		baseThreshold = 0.8 // Exemplo: Inibitórios podem ter limiar diferente
	}

	return &Neuron{
		ID:                       id,
		Position:                 initialPosition,
		Type:                     nType,
		State:                    RestingState,
		CurrentPotential:         0.0,
		FiringThreshold:          baseThreshold,
		BaseFiringThreshold:      baseThreshold,
		LastFiringCycle:          -1, // Nenhum disparo ainda
		RefractoryPeriodAbsolute: 2,  // Ex: 2 ciclos
		RefractoryPeriodRelative: 3,  // Ex: 3 ciclos após o absoluto
	}
}

// InitializeNeurons cria um slice de neurônios com posições procedurais.
// Esta é uma implementação simples; OpenNoise seria mais complexo.
func InitializeNeurons(numNeurons int, neuronDistribution map[NeuronType]float64, spaceSize float64) []*Neuron {
	neurons := make([]*Neuron, 0, numNeurons)
	currentID := 0

	neuronCounts := make(map[NeuronType]int)
	for nType, percentage := range neuronDistribution {
		neuronCounts[nType] = int(float64(numNeurons) * percentage)
	}

	for nType, count := range neuronCounts {
		// radiusFactor := 0.5 // Removido pois não estava sendo usado.
		// TODO: Usar os raios específicos do README (60%, 10%, 30%)
		// e decidir como tratar Input/Output.
		// Por simplicidade, todos os neurônios são distribuídos no mesmo volume por agora.
		// O centro do espaço é [0,0,...,0] se spaceSize for o limite em cada dimensão (-spaceSize/2 a +spaceSize/2)
		// Ou [spaceSize/2, ..., spaceSize/2] se for de 0 a spaceSize.
		// Assumindo 0 a spaceSize para simplificar a geração randômica.

		for i := 0; i < count; i++ {
			if currentID >= numNeurons {
				break
			}
			var pos [SpaceDimensions]float64
			for d := 0; d < SpaceDimensions; d++ {
				pos[d] = rand.Float64() * spaceSize
			}
			neurons = append(neurons, NewNeuron(currentID, nType, pos))
			currentID++
		}
	}

	if len(neurons) < numNeurons {
		// Lógica de preenchimento simples se necessário
	}


	return neurons
}

// UpdateState atualiza o estado do neurônio com base em seu potencial e ciclo refratário.
func (n *Neuron) UpdateState(currentCycle int) (fired bool) {
	fired = false

	switch n.State {
	case RestingState:
		n.CyclesInRest++
		n.CyclesInFiring = 0
		n.CyclesInRefractory = 0
		if n.CurrentPotential >= n.FiringThreshold {
			n.State = FiringState
			n.LastFiringCycle = currentCycle
			n.CurrentPotential = 0
			fired = true
			n.CyclesInRest = 0
			n.CyclesInFiring = 1
		}
	case FiringState:
		n.CyclesInFiring++
		n.State = RefractoryAbsoluteState
		n.RefractoryCycles = n.RefractoryPeriodAbsolute
		n.CyclesInRefractory = 1
	case RefractoryAbsoluteState:
		n.CyclesInRefractory++
		n.RefractoryCycles--
		if n.RefractoryCycles <= 0 {
			n.State = RefractoryRelativeState
			n.RefractoryCycles = n.RefractoryPeriodRelative
		}
	case RefractoryRelativeState:
		n.CyclesInRefractory++
		n.RefractoryCycles--
		if n.RefractoryCycles <= 0 {
			n.State = RestingState
			n.CyclesInRest = 1
			n.CyclesInRefractory = 0
		}
	}

	if !fired && n.State != FiringState {
		decayRate := 0.1
		n.CurrentPotential -= decayRate * n.CurrentPotential
		if n.CurrentPotential < 0 && n.Type != InhibitoryNeuron {
			// n.CurrentPotential = 0
		}
	}
	return fired
}

// AddPotential adiciona um valor ao potencial atual do neurônio.
func (n *Neuron) AddPotential(amount float64) {
	if n.State == RefractoryAbsoluteState {
		return
	}
	n.CurrentPotential += amount
}

// GetPosition retorna a posição do neurônio.
func (n *Neuron) GetPosition() [SpaceDimensions]float64 {
	return n.Position
}

// SetPosition define a posição do neurônio (usado pela sinaptogênese).
func (n *Neuron) SetPosition(newPosition [SpaceDimensions]float64) {
	n.Position = newPosition
}

// AdjustFiringThreshold ajusta o limiar de disparo (usado por cortisol/dopamina).
func (n *Neuron) AdjustFiringThreshold(adjustment float64) {
	n.FiringThreshold = n.BaseFiringThreshold + adjustment
	if n.FiringThreshold < 0.1 { // Limiar mínimo
		n.FiringThreshold = 0.1
	}
}
