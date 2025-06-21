package core

import (
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// NetworkConfig contém os parâmetros de configuração para a rede neural.
type NetworkConfig struct {
	NumNeurons            int
	SpaceSize             float64 // Tamanho do espaço (ex: 100.0 unidades)
	PulsePropagationSpeed float64 // Unidades de distância por ciclo (ex: 0.6)
	MaxCycles             int     // Número máximo de ciclos para uma simulação/processamento de input

	// Distribuição dos tipos de neurônios
	NeuronDistribution map[NeuronType]float64

	// Parâmetros de Cortisol
	CortisolProductionRate    float64 // Quanto cortisol é produzido por pulso excitatório na glândula
	CortisolDecayRate         float64 // Taxa de decaimento do cortisol por ciclo
	CortisolEffectOnThreshold float64 // Fator de modulação do limiar (pode ser mais complexo, com pico)
	CortisolEffectOnSynapto   float64 // Fator de modulação da sinaptogênese

	// Parâmetros de Dopamina
	DopamineProductionByNeuron float64 // Quanto dopamina é liberada por disparo de neurônio dopaminérgico
	DopamineDecayRate          float64 // Taxa de decaimento da dopamina por ciclo
	DopamineEffectOnThreshold  float64 // Fator de modulação do limiar
	DopamineEffectOnSynapto    float64 // Fator de modulação da sinaptogênese

	// Parâmetros de Sinaptogênese
	SynaptoMovementRateAttract float64 // Taxa base de movimento de aproximação
	SynaptoMovementRateRepel   float64 // Taxa base de movimento de repulsão

	// Parâmetros de Neurônios (padrões, podem ser sobrescritos por tipo)
	NeuronBaseFiringThreshold      float64
	NeuronRefractoryPeriodAbsolute int
	NeuronRefractoryPeriodRelative int
	NeuronPotentialDecayRate       float64

	// Semente para RNG
	RandomSeed int64
}

// DefaultNetworkConfig retorna uma configuração padrão para a rede.
func DefaultNetworkConfig() NetworkConfig {
	dist := make(map[NeuronType]float64)
	// Conforme RF-CORE-003 e README, ajustando para somar 100%
	// 5% Input, 5% Output = 10%
	// Restantes 90% distribuídos:
	// 1% Dopaminérgicos (do total, então 0.01 / 0.90 dos restantes ~1.11%)
	// 30% Inibitórios (do total, então 0.30 / 0.90 dos restantes ~33.33%)
	// 69% Excitatórios (do total, então 0.69 / 0.90 dos restantes ~76.66%)
	// Essas proporções são do total de neurônios.
	dist[InputNeuron] = 0.05
	dist[OutputNeuron] = 0.05
	dist[DopaminergicNeuron] = 0.01
	// Os restantes 89% são Inibitórios e Excitatórios
	// Proporção Inibitórios / (Inibitórios + Excitatórios) = 30 / (30+69) = 30/99
	// Proporção Excitatórios / (Inibitórios + Excitatórios) = 69 / (30+69) = 69/99
	dist[InhibitoryNeuron] = 0.89 * (30.0 / 99.0)
	dist[ExcitatoryNeuron] = 0.89 * (69.0 / 99.0)

	// Verificar soma e ajustar o maior para garantir 1.0
	// (Esta é uma forma simples, pode ser mais elegante)
	sum := dist[InputNeuron] + dist[OutputNeuron] + dist[DopaminergicNeuron] + dist[InhibitoryNeuron] + dist[ExcitatoryNeuron]
	if sum != 1.0 {
		dist[ExcitatoryNeuron] += (1.0 - sum)
	}


	return NetworkConfig{
		NumNeurons:            1000, // Um número razoável para começar
		SpaceSize:             100.0,
		PulsePropagationSpeed: 0.6,
		MaxCycles:             200, // Ex: ~20 segundos se 10 ciclos/segundo
		NeuronDistribution:    dist,

		CortisolProductionRate:    0.05,
		CortisolDecayRate:         0.01,
		CortisolEffectOnThreshold: 0.1, // Este valor pode ser positivo ou negativo dependendo do nível de cortisol
		CortisolEffectOnSynapto:   0.1, // Níveis altos diminuem sinapto

		DopamineProductionByNeuron: 0.1,
		DopamineDecayRate:          0.05, // Mais acentuada que cortisol
		DopamineEffectOnThreshold:  0.1, // Aumenta limiar
		DopamineEffectOnSynapto:    0.2, // Aumenta sinapto

		SynaptoMovementRateAttract: 0.05, // Pequenos ajustes por ciclo
		SynaptoMovementRateRepel:   0.03,

		NeuronBaseFiringThreshold:      1.0,
		NeuronRefractoryPeriodAbsolute: 2,
		NeuronRefractoryPeriodRelative: 3,
		NeuronPotentialDecayRate:       0.1,
		RandomSeed:                     time.Now().UnixNano(),
	}
}

// Network representa a rede neural completa.
type Network struct {
	Config    NetworkConfig
	Neurons   []*Neuron
	Gland     *Gland          // Glândula de cortisol
	Pulses    []*Pulse        // Lista de pulsos ativos na rede
	rng       *rand.Rand      // Gerador de números aleatórios para reprodutibilidade
	CurrentCycle int

	// Níveis globais de neuroquímicos
	CortisolLevel float64
	DopamineLevel float64

	// Mutex para proteger acesso concorrente se usarmos goroutines extensivamente
	mu sync.RWMutex
}

// NewNetwork cria uma nova rede neural com base na configuração.
func NewNetwork(config NetworkConfig) *Network {
	rng := rand.New(rand.NewSource(config.RandomSeed))

	// Inicializa a glândula de cortisol no centro do espaço
	// Assumindo que o espaço vai de 0 a SpaceSize em cada dimensão, o centro é SpaceSize/2.
	var glandPos [SpaceDimensions]float64
	for i := range glandPos {
		glandPos[i] = config.SpaceSize / 2.0
	}
	gland := &Gland{Position: glandPos}

	net := &Network{
		Config:    config,
		Gland:     gland,
		rng:       rng,
		CurrentCycle: 0,
		CortisolLevel: 0.0, // Nível inicial
		DopamineLevel: 0.0, // Nível inicial
	}

	// Inicializa neurônios
	net.Neurons = make([]*Neuron, 0, config.NumNeurons)
	neuronCounts := make(map[NeuronType]int)
	totalCount := 0
	for nType, percentage := range config.NeuronDistribution {
		count := int(float64(config.NumNeurons) * percentage)
		neuronCounts[nType] = count
		totalCount += count
	}

	// Ajustar a contagem do último tipo (ou o mais numeroso) para garantir NumNeurons
	if totalCount != config.NumNeurons && len(config.NeuronDistribution) > 0 {
		// Encontrar um tipo para ajustar, ex: ExcitatoryNeuron se existir
		// Esta é uma forma simples de garantir a contagem exata.
		// Poderia ser mais sofisticado na distribuição do erro.
		diff := config.NumNeurons - totalCount
		adjustType := ExcitatoryNeuron // Default
		// Seleciona o primeiro tipo do mapa como fallback se Excitatory não estiver lá.
		for nType := range config.NeuronDistribution {
			adjustType = nType
			break
		}
		if _, ok := config.NeuronDistribution[ExcitatoryNeuron]; ok {
			neuronCounts[ExcitatoryNeuron] += diff
		} else {
			neuronCounts[adjustType] += diff
		}
	}


	currentID := 0
	for nType, count := range neuronCounts {
		// TODO: Implementar distribuição espacial com raios diferentes conforme README.
		// Por agora, distribuição uniforme dentro de Config.SpaceSize.
		// Raio para Dopaminérgicos: 0.6 * Config.SpaceSize / 2 (se centro é 0,0) ou 0.6 * Config.SpaceSize (se de 0 a S)
		// Raio para Inibitórios: 0.1 * Config.SpaceSize / 2
		// Raio para Excitatórios: 0.3 * Config.SpaceSize / 2
		// Input/Output: Sem especificação de raio, distribuir uniformemente ou em região específica.

		// Simplificação para MVP: todos distribuídos uniformemente no cubo [0, SpaceSize]^16
		for i := 0; i < count; i++ {
			if currentID >= config.NumNeurons {
				break
			}
			var pos [SpaceDimensions]float64
			for d := 0; d < SpaceDimensions; d++ {
				pos[d] = net.rng.Float64() * config.SpaceSize
			}

			neuron := NewNeuron(currentID, nType, pos)
			// Aplicar configurações globais de neurônio
			neuron.BaseFiringThreshold = config.NeuronBaseFiringThreshold
			neuron.FiringThreshold = config.NeuronBaseFiringThreshold
			neuron.RefractoryPeriodAbsolute = config.NeuronRefractoryPeriodAbsolute
			neuron.RefractoryPeriodRelative = config.NeuronRefractoryPeriodRelative
			// Aqui poderiam ter mais customizações por tipo de neurônio se necessário

			net.Neurons = append(net.Neurons, neuron)
			currentID++
		}
	}
	if len(net.Neurons) != config.NumNeurons {
		// Isso pode acontecer se a soma das porcentagens não for exatamente 1.0
		// ou por erros de arredondamento. Adicionar um log ou tratamento.
		fmt.Printf("Warning: Initialized %d neurons, expected %d\n", len(net.Neurons), config.NumNeurons)
	}

	return net
}

// GetNeuronsByType retorna uma lista de neurônios de um tipo específico.
func (n *Network) GetNeuronsByType(neuronType NeuronType) []*Neuron {
	var result []*Neuron
	for _, neuron := range n.Neurons {
		if neuron.Type == neuronType {
			result = append(result, neuron)
		}
	}
	return result
}


// SimulateCycle executa um ciclo completo de simulação da rede.
func (n *Network) SimulateCycle() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.CurrentCycle++
	var newPulses []*Pulse

	// 1. Propagação de Pulsos e Efeitos nos Neurônios
	activePulses := make([]*Pulse, 0, len(n.Pulses))
	for _, pulse := range n.Pulses {
		oldPulseRadius := pulse.CurrentRadius
		pulseStillActiveAfterUpdate := pulse.UpdatePropagation(n.Config.PulsePropagationSpeed) // Avança o raio do pulso
		newPulseRadius := pulse.CurrentRadius

		for _, neuron := range n.Neurons {
			if neuron.ID == pulse.OriginNeuronID { // Pulso não afeta o neurônio de origem
				continue
			}
			dist := DistanceEuclidean(pulse.OriginPosition, neuron.Position)

			// Afeta se estiver na nova casca esférica (entre o raio antigo e o novo, incluindo a borda do novo)
			if dist > oldPulseRadius && dist <= newPulseRadius {
				// Neurônio está na "casca" de propagação do pulso neste ciclo
				if pulse.IsDopaminePulse() {
					// Efeito especial para pulsos de dopamina: aumentar dopamina (global ou local)
					// README: "se for dopamina soma a quantidade de dopamina do neuronio" - Ambíguo.
					// Opção 1: Aumenta o nível de DopamineLevel global da rede.
					// Opção 2: O neurônio alvo tem um atributo "dopamineReceived" que é aumentado.
					// Opção 3: O pulso tem um Strength que afeta o potencial E também aumenta dopamina.
					// Para MVP, vamos usar DopamineLevel global.
					n.DopamineLevel += n.Config.DopamineProductionByNeuron // Ou um valor do pulso
					// E também pode ter um efeito de potencial, se Strength não for 0.
					// Se o Strength for, por exemplo, 0.1, ele também excita.
					neuron.AddPotential(pulse.GetStrength())
				} else {
					neuron.AddPotential(pulse.GetStrength())
				}

				// A lógica de produção de cortisol pela glândula foi movida para uma checagem separada
				// após este loop, iterando sobre os pulsos e verificando sua proximidade com a glândula.
			}
		}
		// Atualizar propagação do pulso e manter se ainda ativo
		// A atualização da propagação (pulse.UpdatePropagation) já foi feita no início do loop do pulso.
		// Agora apenas verificamos se ele ainda está ativo para mantê-lo.
		if pulseStillActiveAfterUpdate {
			activePulses = append(activePulses, pulse)
		}
	}
	n.Pulses = activePulses

	// Checagem de pulsos excitatórios atingindo a glândula de cortisol
	// Esta é uma forma mais correta de lidar com a glândula.
	// Esta lógica deve usar os raios old/new do pulso, similar à afetação de neurônios.
	for _, pulse := range n.Pulses { // Iterar sobre os pulsos que ainda estão ativos (n.Pulses já é activePulses)
		if pulse.Strength > 0 { // Pulso excitatório
			distOriginToGland := DistanceEuclidean(pulse.OriginPosition, n.Gland.Position)
			// A glândula é afetada se estiver na "casca" de propagação do pulso neste ciclo.
			// O pulse.CurrentRadius já reflete o raio *após* a propagação deste ciclo.
			// O raio *antes* da propagação neste ciclo seria pulse.CurrentRadius - n.Config.PulsePropagationSpeed.
			oldRadiusForGlandCheck := pulse.CurrentRadius - n.Config.PulsePropagationSpeed
			if oldRadiusForGlandCheck < 0 { oldRadiusForGlandCheck = 0 }

			if distOriginToGland > oldRadiusForGlandCheck && distOriginToGland <= pulse.CurrentRadius {
				n.CortisolLevel += n.Config.CortisolProductionRate
			}
		}
	}


	// 2. Atualização do Estado dos Neurônios e Geração de Novos Pulsos
	for _, neuron := range n.Neurons {
		fired := neuron.UpdateState(n.CurrentCycle) // Passar o potencial de decaimento da config
		if fired {
			var strength float64
			switch neuron.Type {
			case ExcitatoryNeuron, InputNeuron, OutputNeuron: // Input/Output são considerados excitatórios por padrão de pulso
				strength = 0.3 // Valor exemplo do README
			case InhibitoryNeuron:
				strength = -0.3 // Valor exemplo
			case DopaminergicNeuron:
				strength = 0.1 // Pulsos dopaminérgicos podem ter um pequeno efeito excitatório direto
				// A maior parte do seu efeito é o aumento da dopamina na rede (tratado acima ou abaixo)
			}
			newPulses = append(newPulses, NewPulse(neuron.ID, neuron.Position, strength, n.CurrentCycle, neuron.Type))

			// Se um neurônio dopaminérgico disparou, aumenta o nível de dopamina global
			// Esta é outra forma de lidar com a produção de dopamina, mais direta.
			// O README: "Dopamina: A dopamina é gerada pelos neurônios dopaminérgicos"
			if neuron.Type == DopaminergicNeuron {
				n.DopamineLevel += n.Config.DopamineProductionByNeuron
			}
		}
		// Aplicar decaimento do potencial (já feito em neuron.UpdateState se não disparou)
		// neuron.CurrentPotential *= (1.0 - n.Config.NeuronPotentialDecayRate) // Se não feito em UpdateState
	}
	n.Pulses = append(n.Pulses, newPulses...)

	// 3. Atualização dos Níveis de Neuroquímicos (Decaimento)
	n.CortisolLevel *= (1.0 - n.Config.CortisolDecayRate)
	if n.CortisolLevel < 0 { n.CortisolLevel = 0 }

	n.DopamineLevel *= (1.0 - n.Config.DopamineDecayRate)
	if n.DopamineLevel < 0 { n.DopamineLevel = 0 }

	// 4. Efeitos dos Neuroquímicos nos Neurônios
	// Cortisol: "diminui o limiar de disparo dos neurônios inicialmente, mas ao atingir um pico, começa a reduzir o limiar" (ambíguo, parece "aumentar")
	// Assumindo: baixo cortisol -> reduz limiar. Alto cortisol -> aumenta limiar. Efeito em sinaptogênese.
	// Dopamina: "aumentar o limiar de disparo dos neurônios e também aumentar a sinapogênese."
	for _, neuron := range n.Neurons {
		// cortisolEffect := 0.0 // Removido, pois a lógica foi integrada diretamente abaixo
		// Modelo simples de efeito do cortisol:
		// Se cortisol baixo (ex: < 0.5), reduz limiar. Se alto (ex: > 1.0), aumenta limiar.
		// Esta é uma interpretação. O README é "diminui o limiar ... mas ao atingir um pico, começa a reduzir o limiar"
		// Isso parece dizer que sempre reduz, mas mais ou menos intensamente. Ou "reduzir" é um typo para "aumentar".
		// Vamos com: cortisol baixo -> reduz limiar, cortisol alto -> aumenta limiar.
		// Efeito de pico: se cortisol entre 0.5 e 1.5, maior redução. Fora disso, menor redução ou aumento.
		// Simplificação para MVP: Efeito linear ou quadrático.
		// Cortisol aumenta limiar (efeito de "estresse") e dopamina também aumenta (efeito de "foco"?)
		// Isso contradiz um pouco o "cortisol diminui o limiar inicialmente".
		// Vamos seguir o README mais literalmente para cortisol:
		// Cortisol diminui limiar. Pico -> começa a *reduzir* o limiar (esta parte é confusa, talvez quis dizer "aumentar" o limiar ou "reduzir o efeito de diminuição").
		// "diminuindo também a sinapogênese" (em pico).
		// Para MVP: Cortisol tem um efeito complexo. Dopamina aumenta limiar.
		// Vamos simplificar:
		// Cortisol: Níveis moderados (e.g. 0.2-0.8) reduzem o limiar. Níveis altos (>1.0) aumentam o limiar e reduzem sinapto.
		// Dopamina: Aumenta limiar e aumenta sinapto.

		// Efeito no limiar:
	// A variável cortisolEffect foi removida pois o cálculo é feito diretamente no thresholdAdjustment.
		thresholdAdjustment := 0.0
		// Dopamina aumenta limiar:
		thresholdAdjustment += n.DopamineLevel * n.Config.DopamineEffectOnThreshold

		// Cortisol:
		if n.CortisolLevel > 1.0 { // Nível alto de cortisol
			thresholdAdjustment += (n.CortisolLevel - 1.0) * n.Config.CortisolEffectOnThreshold // Aumenta limiar
		} else if n.CortisolLevel > 0.1 { // Nível moderado
		// O README diz "diminui o limiar de disparo dos neurônios inicialmente"
		// "mas ao atingir um pico, começa a reduzir o limiar" - interpretado como "reduzir o *efeito de diminuição* do limiar" ou "aumentar o limiar"
		// A implementação atual é: cortisol moderado (0.1 a 1.0) diminui o limiar.
		// A força da diminuição é -(CortisolLevel * Config.CortisolEffectOnThreshold * 0.5)
		// Ex: C=0.5, Effect=0.1 => -(0.5 * 0.1 * 0.5) = -0.025
		thresholdAdjustment -= (n.CortisolLevel) * n.Config.CortisolEffectOnThreshold * 0.5 // Diminui limiar
		}
		neuron.AdjustFiringThreshold(thresholdAdjustment)
	}

	// 5. Sinaptogênese (movimentação dos neurônios)
	// Esta parte precisa de uma implementação cuidadosa (em synaptogenesis.go)
	// e será chamada aqui.
	n.applySynaptogenesis()

}

// SetInput ativa um conjunto de neurônios de input com uma certa frequência/potencial.
// `inputPattern` é um slice de float64, onde cada valor corresponde a um neurônio de input.
// O valor pode ser interpretado como potencial a ser adicionado ou como uma frequência desejada.
// Para MVP, vamos adicionar o valor diretamente ao potencial dos neurônios de input.
func (n *Network) SetInput(inputPattern []float64) {
	n.mu.Lock()
	defer n.mu.Unlock()

	inputNeurons := n.GetNeuronsByType(InputNeuron)
	if len(inputNeurons) == 0 {
		return // Nenhum neurônio de input
	}

	for i, val := range inputPattern {
		if i < len(inputNeurons) {
			// Adicionar potencial. Se o valor for alto, pode fazer o neurônio disparar
			// múltiplos ciclos se não for resetado, simulando frequência.
			// Ou, podemos ter uma lógica mais explícita para forçar disparos por alguns ciclos.
			inputNeurons[i].AddPotential(val)
		}
	}
}

// GetOutput lê a atividade dos neurônios de output.
// Retorna um slice de float64 representando a "ativação" de cada neurônio de output.
// Pode ser a CurrentPotential, ou a frequência de disparos recentes.
// Para MVP, vamos usar CurrentPotential.
func (n *Network) GetOutput() []float64 {
	n.mu.RLock()
	defer n.mu.RUnlock()

	outputNeurons := n.GetNeuronsByType(OutputNeuron)
	outputValues := make([]float64, len(outputNeurons))

	for i, neuron := range outputNeurons {
		// Opção 1: Potencial atual
		outputValues[i] = neuron.CurrentPotential
		// Opção 2: Frequência de disparos (ex: contagem de disparos nos últimos X ciclos)
		// Isso exigiria armazenar histórico de disparos ou um contador no neurônio.
		// neuron.CyclesInFiring poderia ser usado se o estado FiringState durasse múltiplos ciclos
		// ou se tivéssemos um contador de disparos recentes.
	}
	return outputValues
}

// ResetNetworkState reseta o estado volátil da rede (potenciais, pulsos),
// mas mantém as posições dos neurônios e os parâmetros aprendidos (implícitos nas posições/limiares).
func (n *Network) ResetNetworkState() {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.Pulses = nil
	n.CurrentCycle = 0
	// Não resetar CortisolLevel e DopamineLevel para zero necessariamente,
	// pois podem carregar um "estado emocional" entre inputs.
	// Mas para uma avaliação limpa, pode ser desejável.
	// n.CortisolLevel = 0
	// n.DopamineLevel = 0

	for _, neuron := range n.Neurons {
		neuron.CurrentPotential = 0
		neuron.State = RestingState
		neuron.RefractoryCycles = 0
		neuron.LastFiringCycle = -1
		// Não resetar FiringThreshold base, pois pode ter sido ajustado pelo aprendizado.
		// Mas resetar o FiringThreshold atual para o Base + Efeitos Químicos Atuais.
		// Ou, se o aprendizado está nos pesos/posições, então o BaseThreshold é fixo.
		// O README sugere que cortisol/dopamina afetam o limiar dinamicamente.
		// O BaseFiringThreshold seria o valor "genético".
		neuron.FiringThreshold = neuron.BaseFiringThreshold // Resetar para o base antes de aplicar neuroquímicos do próximo ciclo.
	}
	// Após resetar, aplicar efeitos de neuroquímicos atuais (se não foram resetados)
	// Esta lógica já está no SimulateCycle, então a primeira chamada a SimulateCycle após Reset
	// irá restabelecer os limiares corretos.
}

// applySynaptogenesis será implementado em synaptogenesis.go e chamado aqui.
// Por enquanto, é um placeholder.
func (n *Network) applySynaptogenesis() {
	// Lógica de movimentação dos neurônios
	// Precisa iterar sobre os neurônios e ajustar suas posições
	// com base na atividade recente de outros neurônios e nos níveis de neuroquímicos.
	// Exemplo muito simplificado:
	// for _, neuron := range n.Neurons {
	//     // Calcular vetor de movimento
	//     // Aplicar movimento: neuron.SetPosition(...)
	//     // Garantir que a posição fique dentro dos limites do espaço: ClampPosition
	// }
	 ApplySynaptogenesis(n) // Chamar a função de synaptogenesis.go
}


// GetNeuronByID retorna um neurônio pelo seu ID.
func (n *Network) GetNeuronByID(id int) *Neuron {
    if id < 0 || id >= len(n.Neurons) {
        return nil
    }
    // Assumindo que os IDs são os índices no slice n.Neurons
    // Se não for o caso, precisa de um mapa ou busca.
    return n.Neurons[id]
}
