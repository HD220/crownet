package pulse

import (
	"crownet/common"
	"crownet/config"
	"crownet/neuron"
	"crownet/space"
	"crownet/synaptic"
)

// Pulse representa um sinal elétrico ou químico propagando-se pela rede.
type Pulse struct {
	EmittingNeuronID common.NeuronID
	OriginPosition   common.Point
	BaseSignalValue  common.PulseValue
	CreationCycle    common.CycleCount
	CurrentDistance  float64
	MaxTravelRadius  float64
}

// New cria um novo Pulso.
func New(emitterID common.NeuronID, origin common.Point, signal common.PulseValue, creationCycle common.CycleCount, maxRadius float64) *Pulse {
	return &Pulse{
		EmittingNeuronID: emitterID,
		OriginPosition:   origin,
		BaseSignalValue:  signal,
		CreationCycle:    creationCycle,
		CurrentDistance:  0.0,
		MaxTravelRadius:  maxRadius,
	}
}

// Propagate avança a distância do pulso e verifica se ele ainda está ativo.
func (p *Pulse) Propagate(simParams *config.SimulationParameters) (isActive bool) {
	p.CurrentDistance += simParams.PulsePropagationSpeed
	return p.CurrentDistance < p.MaxTravelRadius
}

// GetEffectShellForCycle calcula o raio interno e externo da "casca" esférica
func (p *Pulse) GetEffectShellForCycle(simParams *config.SimulationParameters) (shellStartRadius, shellEndRadius float64) {
	shellEndRadius = p.CurrentDistance
	shellStartRadius = p.CurrentDistance - simParams.PulsePropagationSpeed
	if shellStartRadius < 0 {
		shellStartRadius = 0
	}
	return shellStartRadius, shellEndRadius
}

// PulseList gerencia uma coleção de pulsos ativos na rede.
type PulseList struct {
	pulses []*Pulse
}

// NewPulseList cria uma nova lista de pulsos vazia.
func NewPulseList() *PulseList {
	return &PulseList{
		pulses: make([]*Pulse, 0),
	}
}

// Add adiciona um novo pulso à lista.
func (pl *PulseList) Add(p *Pulse) {
	pl.pulses = append(pl.pulses, p)
}

// AddAll adiciona múltiplos pulsos à lista.
func (pl *PulseList) AddAll(newPulses []*Pulse) {
	pl.pulses = append(pl.pulses, newPulses...)
}

// Clear remove todos os pulsos da lista.
func (pl *PulseList) Clear() {
	pl.pulses = make([]*Pulse, 0)
}

// GetAll retorna todos os pulsos atualmente na lista.
func (pl *PulseList) GetAll() []*Pulse {
    return pl.pulses
}

// Count retorna o número de pulsos ativos.
func (pl *PulseList) Count() int {
    return len(pl.pulses)
}

// processSinglePulseOnTargetNeuron processa o efeito de um único pulso em um único neurônio alvo.
func processSinglePulseOnTargetNeuron(
	p *Pulse,
	targetNeuron *neuron.Neuron,
	weights synaptic.NetworkWeights,
	currentCycle common.CycleCount,
	simParams *config.SimulationParameters,
	shellStartRadius, shellEndRadius float64,
) (newlyGeneratedPulse *Pulse) {

	if targetNeuron.ID == p.EmittingNeuronID {
		return nil
	}

	distanceToTarget := space.EuclideanDistance(p.OriginPosition, targetNeuron.Position)

	if distanceToTarget >= shellStartRadius && distanceToTarget < shellEndRadius {
		weight := weights.GetWeight(p.EmittingNeuronID, targetNeuron.ID)
		// Cast to float64 for multiplication, then back to PulseValue
		effectivePotential := common.PulseValue(float64(p.BaseSignalValue) * float64(weight))

		if effectivePotential == 0 {
			return nil
		}

		if targetNeuron.IntegrateIncomingPotential(effectivePotential, currentCycle) {
			emittedSignal := targetNeuron.EmittedPulseSign()
			if emittedSignal != 0 {
				return New(
					targetNeuron.ID,
					targetNeuron.Position,
					emittedSignal,
					currentCycle,
					simParams.SpaceMaxDimension*2.0,
				)
			}
		}
	}
	return nil
}

// ProcessCycle propaga todos os pulsos, processa seus efeitos nos neurônios
func (pl *PulseList) ProcessCycle(
	neurons []*neuron.Neuron,
	weights synaptic.NetworkWeights,
	currentCycle common.CycleCount,
	simParams *config.SimulationParameters,
) (newlyGeneratedPulses []*Pulse) {

	remainingActivePulses := make([]*Pulse, 0, len(pl.pulses))
	newlyGeneratedPulses = make([]*Pulse, 0)

	for _, p := range pl.pulses {
		if !p.Propagate(simParams) {
			continue
		}
		remainingActivePulses = append(remainingActivePulses, p)

		shellStartRadius, shellEndRadius := p.GetEffectShellForCycle(simParams)

		for _, targetNeuron := range neurons {
			if newPulse := processSinglePulseOnTargetNeuron(p, targetNeuron, weights, currentCycle, simParams, shellStartRadius, shellEndRadius); newPulse != nil {
				newlyGeneratedPulses = append(newlyGeneratedPulses, newPulse)
			}
		}
	}
	pl.pulses = remainingActivePulses
	return newlyGeneratedPulses
}
