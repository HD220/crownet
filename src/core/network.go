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
	SpaceSize             float64
	PulsePropagationSpeed float64
	MaxCycles             int

	NeuronDistribution map[NeuronType]float64

	CortisolProductionRate    float64
	CortisolDecayRate         float64
	CortisolEffectOnThreshold float64
	CortisolEffectOnSynapto   float64

	DopamineProductionByNeuron float64
	DopamineDecayRate          float64
	DopamineEffectOnThreshold  float64
	DopamineEffectOnSynapto    float64

	SynaptoMovementRateAttract float64
	SynaptoMovementRateRepel   float64

	NeuronBaseFiringThreshold      float64
	NeuronRefractoryPeriodAbsolute int
	NeuronRefractoryPeriodRelative int
	NeuronPotentialDecayRate       float64

	RandomSeed int64
}

// DefaultNetworkConfig retorna uma configuração padrão para a rede.
func DefaultNetworkConfig() NetworkConfig {
	dist := make(map[NeuronType]float64)
	dist[InputNeuron] = 0.05
	dist[OutputNeuron] = 0.05
	dist[DopaminergicNeuron] = 0.01
	// Ajuste para garantir que as proporções somem 1.0
	// (0.05+0.05+0.01 = 0.11. Restantes 0.89 para Excitatório e Inibitório)
	// Proporção Inib/Excit = 30/69
	totalRemainingPct := 1.0 - (dist[InputNeuron] + dist[OutputNeuron] + dist[DopaminergicNeuron])
	propInhib := 30.0 / (30.0 + 69.0) // 30/99
	propExcit := 69.0 / (30.0 + 69.0) // 69/99
	dist[InhibitoryNeuron] = totalRemainingPct * propInhib
	dist[ExcitatoryNeuron] = totalRemainingPct * propExcit

	// Recalcular a soma para verificar se precisa de ajuste fino devido a float precision
	currentSum := dist[InputNeuron] + dist[OutputNeuron] + dist[DopaminergicNeuron] + dist[InhibitoryNeuron] + dist[ExcitatoryNeuron]
	if currentSum != 1.0 {
	    dist[ExcitatoryNeuron] += (1.0 - currentSum) // Ajusta no mais numeroso
	}

	return NetworkConfig{
		NumNeurons:            1000,
		SpaceSize:             100.0,
		PulsePropagationSpeed: 0.6,
		MaxCycles:             200,
		NeuronDistribution:    dist,
		CortisolProductionRate:    0.05,
		CortisolDecayRate:         0.01,
		CortisolEffectOnThreshold: 0.1,
		CortisolEffectOnSynapto:   0.1,
		DopamineProductionByNeuron: 0.1,
		DopamineDecayRate:          0.05,
		DopamineEffectOnThreshold:  0.1,
		DopamineEffectOnSynapto:    0.2,
		SynaptoMovementRateAttract: 0.05,
		SynaptoMovementRateRepel:   0.03,
		NeuronBaseFiringThreshold:      1.0,
		NeuronRefractoryPeriodAbsolute: 2,
		NeuronRefractoryPeriodRelative: 3,
		NeuronPotentialDecayRate:       0.1, // Este não é usado diretamente se o decaimento está em Neuron.UpdateState
		RandomSeed:                     time.Now().UnixNano(),
	}
}

// Network representa a rede neural completa.
type Network struct {
	Config    NetworkConfig
	Neurons   []*Neuron
	Gland     *Gland
	Pulses    []*Pulse
	rng       *rand.Rand
	CurrentCycle int
	CortisolLevel float64
	DopamineLevel float64
	mu sync.RWMutex
}

// NewNetwork cria uma nova rede neural com base na configuração.
func NewNetwork(config NetworkConfig) *Network {
	rng := rand.New(rand.NewSource(config.RandomSeed))
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
		CortisolLevel: 0.0,
		DopamineLevel: 0.0,
	}

	net.Neurons = make([]*Neuron, 0, config.NumNeurons)
	neuronCounts := make(map[NeuronType]int)
	totalAllocated := 0
	for nType, percentage := range config.NeuronDistribution {
		count := int(float64(config.NumNeurons) * percentage)
		neuronCounts[nType] = count
		totalAllocated +=count
	}
    // Ajustar a contagem do tipo mais numeroso (Excitatory) se a soma não bater NumNeurons
    // devido a arredondamentos de float para int.
    if totalAllocated != config.NumNeurons {
        if diff := config.NumNeurons - totalAllocated; diff != 0 {
            // Tenta adicionar/remover do excitatório, se existir. Senão, do primeiro tipo que encontrar.
            targetType := ExcitatoryNeuron
            _, hasExcitatory := neuronCounts[ExcitatoryNeuron]
            if !hasExcitatory && len(neuronCounts) > 0 { // Fallback se não houver excitatórios
                for t := range neuronCounts { targetType = t; break }
            }
            neuronCounts[targetType] += diff
        }
    }

	currentID := 0
	for nType, count := range neuronCounts {
		for i := 0; i < count; i++ {
			if currentID >= config.NumNeurons { break } // Segurança extra
			var pos [SpaceDimensions]float64
			for d := 0; d < SpaceDimensions; d++ {
				pos[d] = net.rng.Float64() * config.SpaceSize
			}
			neuron := NewNeuron(currentID, nType, pos)
			neuron.BaseFiringThreshold = config.NeuronBaseFiringThreshold
			neuron.FiringThreshold = config.NeuronBaseFiringThreshold
			neuron.RefractoryPeriodAbsolute = config.NeuronRefractoryPeriodAbsolute
			neuron.RefractoryPeriodRelative = config.NeuronRefractoryPeriodRelative
			net.Neurons = append(net.Neurons, neuron)
			currentID++
		}
	}
	if len(net.Neurons) != config.NumNeurons {
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
		pulseStillActiveAfterUpdate := pulse.UpdatePropagation(n.Config.PulsePropagationSpeed)
		newPulseRadius := pulse.CurrentRadius

		for _, neuron := range n.Neurons {
			if neuron.ID == pulse.OriginNeuronID { continue }
			dist := DistanceEuclidean(pulse.OriginPosition, neuron.Position)
			if dist > oldPulseRadius && dist <= newPulseRadius { // Neurônio na nova casca esférica
				if pulse.IsDopaminePulse() {
					n.DopamineLevel += n.Config.DopamineProductionByNeuron
					neuron.AddPotential(pulse.GetStrength())
				} else {
					neuron.AddPotential(pulse.GetStrength())
				}
			}
		}
		if pulseStillActiveAfterUpdate {
			activePulses = append(activePulses, pulse)
		}
	}
	n.Pulses = activePulses

	// Checagem de pulsos excitatórios atingindo a glândula de cortisol
	for _, pulse := range n.Pulses {
		if pulse.Strength > 0 {
			distOriginToGland := DistanceEuclidean(pulse.OriginPosition, n.Gland.Position)
			oldRadiusForGlandCheck := pulse.CurrentRadius - n.Config.PulsePropagationSpeed
			if oldRadiusForGlandCheck < 0 { oldRadiusForGlandCheck = 0 }
			if distOriginToGland > oldRadiusForGlandCheck && distOriginToGland <= pulse.CurrentRadius {
				n.CortisolLevel += n.Config.CortisolProductionRate
			}
		}
	}

	// 2. Atualização do Estado dos Neurônios e Geração de Novos Pulsos
	for _, neuron := range n.Neurons {
		fired := neuron.UpdateState(n.CurrentCycle)
		if fired {
			var strength float64
			switch neuron.Type {
			case ExcitatoryNeuron, InputNeuron, OutputNeuron:
				strength = 0.3
			case InhibitoryNeuron:
				strength = -0.3
			case DopaminergicNeuron:
				strength = 0.1
			}
			newPulses = append(newPulses, NewPulse(neuron.ID, neuron.Position, strength, n.CurrentCycle, neuron.Type))
			if neuron.Type == DopaminergicNeuron {
				n.DopamineLevel += n.Config.DopamineProductionByNeuron
			}
		}
	}
	n.Pulses = append(n.Pulses, newPulses...)

	// 3. Atualização dos Níveis de Neuroquímicos (Decaimento)
	n.CortisolLevel *= (1.0 - n.Config.CortisolDecayRate)
	if n.CortisolLevel < 0 { n.CortisolLevel = 0 }
	n.DopamineLevel *= (1.0 - n.Config.DopamineDecayRate)
	if n.DopamineLevel < 0 { n.DopamineLevel = 0 }

	// 4. Efeitos dos Neuroquímicos nos Neurônios
	for _, neuron := range n.Neurons {
		thresholdAdjustment := 0.0
		thresholdAdjustment += n.DopamineLevel * n.Config.DopamineEffectOnThreshold
		if n.CortisolLevel > 1.0 {
			thresholdAdjustment += (n.CortisolLevel - 1.0) * n.Config.CortisolEffectOnThreshold
		} else if n.CortisolLevel > 0.1 {
			thresholdAdjustment -= (n.CortisolLevel) * n.Config.CortisolEffectOnThreshold * 0.5
		}
		neuron.AdjustFiringThreshold(thresholdAdjustment)
	}

	// 5. Sinaptogênese
	n.applySynaptogenesis()
}

// SetInput ativa neurônios de input.
func (n *Network) SetInput(inputPattern []float64) {
	n.mu.Lock()
	defer n.mu.Unlock()
	inputNeurons := n.GetNeuronsByType(InputNeuron)
	if len(inputNeurons) == 0 { return }
	for i, val := range inputPattern {
		if i < len(inputNeurons) {
			inputNeurons[i].AddPotential(val)
		}
	}
}

// GetOutput lê neurônios de output.
func (n *Network) GetOutput() []float64 {
	n.mu.RLock()
	defer n.mu.RUnlock()
	outputNeurons := n.GetNeuronsByType(OutputNeuron)
	outputValues := make([]float64, len(outputNeurons))
	for i, neuron := range outputNeurons {
		outputValues[i] = neuron.CurrentPotential
	}
	return outputValues
}

// ResetNetworkState reseta o estado volátil da rede.
func (n *Network) ResetNetworkState() {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.Pulses = nil
	n.CurrentCycle = 0
	for _, neuron := range n.Neurons {
		neuron.CurrentPotential = 0
		neuron.State = RestingState
		neuron.RefractoryCycles = 0
		neuron.LastFiringCycle = -1
		neuron.FiringThreshold = neuron.BaseFiringThreshold
	}
}

// applySynaptogenesis (chamada placeholder)
func (n *Network) applySynaptogenesis() {
	 ApplySynaptogenesis(n)
}

// GetNeuronByID retorna um neurônio pelo seu ID.
func (n *Network) GetNeuronByID(id int) *Neuron {
    if id < 0 || id >= len(n.Neurons) {
        return nil
    }
    return n.Neurons[id]
}
