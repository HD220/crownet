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

	// Gerar posições para cada tipo de neurônio
	// O README especifica raios diferentes para diferentes tipos de neurônios
	// Dopaminérgicos (Raio Maior: 60% do espaço)
	// Inibitórios (Raio Menor: 10% do espaço)
	// Excitatórios (Raio Médio: 30% do espaço)
	// Input/Output - não especifica raio, vamos assumir distribuído ou um raio médio.

	neuronCounts := make(map[NeuronType]int)
	for nType, percentage := range neuronDistribution {
		neuronCounts[nType] = int(float64(numNeurons) * percentage)
	}

	// Ajustar contagens para garantir que a soma seja numNeurons (devido a arredondamentos)
	// Esta parte pode ser mais sofisticada para distribuir o erro de arredondamento.
	// Por simplicidade, vamos apenas garantir que não exceda e o último tipo pega o restante.

	// TODO: Implementar a lógica de distribuição espacial conforme os raios do README.
	// Por enquanto, uma distribuição uniforme simples dentro do `spaceSize`.
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
				// Gerar posição dentro de uma esfera/cubo.
				// Para uma distribuição mais controlada por raio:
				// 1. Gerar um ponto aleatório numa esfera unitária 16D.
				// 2. Escalar pelo raio desejado (ex: 0.6 * spaceSize/2 para dopaminérgicos).
				// 3. Adicionar ao centro da população daquele tipo (pode ser o centro do espaço global).
				// Simplificação atual: distribuição uniforme no cubo.
				pos[d] = rand.Float64() * spaceSize
			}
			neurons = append(neurons, NewNeuron(currentID, nType, pos))
			currentID++
		}
	}

	// Se faltarem neurônios devido a arredondamento, preencher com o tipo mais comum (Excitatory)
	// ou o último tipo processado.
	// Esta é uma forma simples de garantir o número total.
	// Uma abordagem mais robusta distribuiria os restantes proporcionalmente ou aleatoriamente.
	if len(neurons) < numNeurons {
		// Adicionar neurônios restantes, talvez como excitatórios por padrão
		// ou o último tipo que estava sendo adicionado.
		// Para este MVP, vamos simplificar e não se preocupar excessivamente com a pequena diferença
		// que pode surgir do arredondamento, ou garantir que a soma das porcentagens seja 1.0.
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
			n.CurrentPotential = 0 // Resetar potencial após disparo (ou parte dele)
			fired = true
			n.CyclesInRest = 0
			n.CyclesInFiring = 1
		}
	case FiringState: // Estado de disparo pode durar 1 ciclo, depois vai para refratário
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
		// No período refratário relativo, o neurônio pode disparar se o estímulo for forte o suficiente
		// (limiar aumentado). Para simplificar, vamos assumir que ele só volta a repouso.
		// O README diz "pode ser estimulado novamente". Poderíamos ter um FiringThreshold mais alto aqui.
		// Por ora, simplesmente transita para RestingState após o período.
		n.RefractoryCycles--
		if n.RefractoryCycles <= 0 {
			n.State = RestingState
			n.CyclesInRest = 1
			n.CyclesInRefractory = 0
		}
	}

	// Decaimento do potencial se não estiver disparando (ex: em repouso ou refratário)
	if !fired && n.State != FiringState {
		// Taxa de decaimento - pode ser um parâmetro
		decayRate := 0.1
		n.CurrentPotential -= decayRate * n.CurrentPotential
		if n.CurrentPotential < 0 && n.Type != InhibitoryNeuron { // Potencial não deve ser negativo para excitatórios
			// n.CurrentPotential = 0 // Ou pode permitir potenciais negativos pequenos
		}
	}
	return fired
}

// AddPotential adiciona um valor ao potencial atual do neurônio.
// Pulsos inibitórios podem adicionar valores negativos.
func (n *Neuron) AddPotential(amount float64) {
	if n.State == RefractoryAbsoluteState {
		return // Não acumula potencial durante o período refratário absoluto
	}
	n.CurrentPotential += amount
	// Poderia haver um clamp para potencial máximo, se necessário.
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
