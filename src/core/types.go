package core

// Vector16D representa um vetor no espaço de 16 dimensões.
type Vector16D [16]float64

// NeuronType define o tipo de um neurônio.
type NeuronType int

const (
	Dopaminergic NeuronType = iota // Neurônio Dopaminérgico
	Inhibitory                     // Neurônio Inibitório
	Excitatory                     // Neurônio Excitatório
	Input                          // Neurônio de Input
	Output                         // Neurônio de Output
)

// NeuronState define o ciclo de estado de um neurônio.
type NeuronState int

const (
	Resting          NeuronState = iota // Repouso
	Firing                              // Disparando
	AbsoluteRefractory                  // Refratário Absoluto
	RelativeRefractory                  // Refratário Relativo (apenas "Refratário" no README)
)

// Neuron representa um único neurônio na rede.
type Neuron struct {
	ID                int
	Position          Vector16D
	Type              NeuronType
	State             NeuronState
	CurrentPotential  float64   // Potencial acumulado dos pulsos recebidos
	FiringThreshold   float64   // Limiar de disparo
	LastFiringCycle   uint64    // Último ciclo em que o neurônio disparou
	CyclesInState     uint64    // Número de ciclos que o neurônio está no estado atual
	TargetConnections []int     // IDs dos neurônios para os quais este neurônio tem conexões (simplificação inicial)
	// Outros campos podem ser adicionados conforme necessário, ex: receptores específicos, etc.
}

// Pulse representa um pulso propagando pela rede.
type Pulse struct {
	SourceNeuronID int
	TargetNeuronID int // Pode ser removido se a propagação for baseada em área
	Strength       float64   // Força do pulso (pode ser positivo ou negativo)
	EmittedCycle   uint64    // Ciclo em que o pulso foi emitido
	CurrentPosition Vector16D // Posição atual do pulso no espaço (para simular propagação)
	ArrivalTime    uint64    // Ciclo em que o pulso deve chegar ao(s) alvo(s)
	Processed      bool      // Indica se o pulso já foi processado no ciclo de chegada
}

// Gland representa a glândula de cortisol.
type Gland struct {
	Position      Vector16D
	CortisolLevel float64
	// Outros parâmetros relevantes para a dinâmica do cortisol
}

// NeuralNetwork representa a rede neural como um todo.
type NeuralNetwork struct {
	Neurons        map[int]*Neuron // Mapa de IDs de neurônios para suas estruturas
	Pulses         []*Pulse        // Lista de pulsos ativos na rede
	CortisolGland  *Gland
	DopamineLevels map[int]float64 // Nível de dopamina por neurônio (simplificação, pode ser zonal)
	CurrentCycle   uint64          // Ciclo atual da simulação

	// Parâmetros da simulação (podem vir de um arquivo de configuração)
	NumNeurons            int
	PulsePropagationSpeed float64
	MaxSpaceDistance      float64 // Distância máxima no espaço 16D (ex: 8 unidades)
	// ... outros parâmetros globais
}

// Config contém os parâmetros de configuração para inicializar a rede.
type Config struct {
	NumNeurons            int
	DopaminergicRatio     float64
	InhibitoryRatio       float64
	ExcitatoryRatio       float64
	InputRatio            float64
	OutputRatio           float64
	PulsePropagationSpeed float64
	DefaultFiringThreshold float64
	CortisolGlandPosition Vector16D
	// Outros parâmetros de configuração
}

// PointReference define um ponto de referência usado no cálculo de vizinhança.
// No CrowNet, são os vértices de um hipercubo e o centro. Para 16D, são 2^16 vértices + 1 centro.
// Para o MVP, podemos simplificar ou adiar a implementação exata dessa busca de vizinhos.
type PointReference struct {
	Position Vector16D
	// Radius pode ser associado aqui se for fixo por ponto de referência,
	// ou pode ser o raio de alcance do pulso emissor.
}
