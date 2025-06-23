package neuron

import (
	"crownet/common"
	"crownet/config"
)

// Constants for emitted pulse signals
const (
	ExcitatoryPulseSignal common.PulseValue = 1.0
	InhibitoryPulseSignal common.PulseValue = -1.0
	NeutralPulseSignal    common.PulseValue = 0.0
)

// Neuron representa uma unidade computacional individual na rede neural.
type Neuron struct {
	ID                     common.NeuronID
	Type                   Type // Excitatory, Inhibitory, Dopaminergic, Input, Output
	Position               common.Point
	CurrentState           State // Resting, Firing, AbsoluteRefractory, RelativeRefractory
	AccumulatedPotential   common.PulseValue
	BaseFiringThreshold    common.Threshold
	CurrentFiringThreshold common.Threshold // Pode ser modulado por neuroquímicos
	LastFiredCycle         common.CycleCount
	CyclesInCurrentState   common.CycleCount
	Velocity               common.Vector // Para sinaptogênese (movimento)
}

// New cria e inicializa um novo Neurônio.
func New(id common.NeuronID, neuronType Type, initialPosition common.Point, simParams *config.SimulationParameters) *Neuron {
	n := &Neuron{
		ID:                     id,
		Type:                   neuronType,
		Position:               initialPosition,
		CurrentState:           Resting,
		AccumulatedPotential:   0.0,
		BaseFiringThreshold:    common.Threshold(simParams.BaseFiringThreshold),
		CurrentFiringThreshold: common.Threshold(simParams.BaseFiringThreshold),
		LastFiredCycle:         -1,
		CyclesInCurrentState:   0,
		Velocity:               common.Vector{},
	}
	return n
}

// IntegrateIncomingPotential processa um potencial sináptico recebido.
func (n *Neuron) IntegrateIncomingPotential(potential common.PulseValue, currentCycle common.CycleCount) (fired bool) {
	if n.CurrentState == AbsoluteRefractory {
		return false
	}

	n.AccumulatedPotential += potential

	// Check if the neuron fires
	if n.AccumulatedPotential < n.CurrentFiringThreshold { // Direct comparison as both are float64 underlying types
		return false
	}
	n.CurrentState = Firing
	n.CyclesInCurrentState = 0
	return true
}

const nearZeroThreshold = 1e-5

// AdvanceState progride o estado do neurônio para o próximo ciclo.
func (n *Neuron) AdvanceState(currentCycle common.CycleCount, simParams *config.SimulationParameters) {
	n.CyclesInCurrentState++

	if n.CurrentState == Firing {
		n.CurrentState = AbsoluteRefractory
		n.CyclesInCurrentState = 0
		n.LastFiredCycle = currentCycle
		n.AccumulatedPotential = 0.0
		return
	}

	if n.CurrentState == AbsoluteRefractory {
		if n.CyclesInCurrentState >= simParams.AbsoluteRefractoryCycles {
			n.CurrentState = RelativeRefractory
			n.CyclesInCurrentState = 0
		}
		return
	}

	if n.CurrentState == RelativeRefractory {
		if n.CyclesInCurrentState >= simParams.RelativeRefractoryCycles {
			n.CurrentState = Resting
			n.CyclesInCurrentState = 0
		}
		return
	}
}

// DecayPotential aplica o decaimento ao potencial acumulado do neurônio.
// O potencial decai em direção a zero.
func (n *Neuron) DecayPotential(simParams *config.SimulationParameters) {
	decayRate := simParams.AccumulatedPulseDecayRate // This is already a float64 in simParams
	if decayRate <= 0 { // No decay if rate is zero or negative
		return
	}
	if decayRate >= 1.0 { // Full decay if rate is 1.0 or more
		n.AccumulatedPotential = 0.0
		return
	}

	// Potential decays towards zero by the decayRate factor
	n.AccumulatedPotential *= common.PulseValue(1.0 - decayRate)

	// Clamp to zero if very close, to avoid floating point inaccuracies.
	if n.AccumulatedPotential > -nearZeroThreshold && n.AccumulatedPotential < nearZeroThreshold {
		n.AccumulatedPotential = 0.0
	}
}

// EmittedPulseSign retorna o sinal base do pulso que este neurônio emite ao disparar.
func (n *Neuron) EmittedPulseSign() common.PulseValue {
	switch n.Type {
	case Excitatory, Input, Output:
		return ExcitatoryPulseSignal
	case Inhibitory:
		return InhibitoryPulseSignal
	case Dopaminergic: // Dopaminergic neurons might not emit direct excitatory/inhibitory signals this way
		return NeutralPulseSignal
	default:
		return NeutralPulseSignal // Default to neutral for any unknown types
	}
}

// UpdatePosition atualiza a posição do neurônio com base em sua velocidade.
func (n *Neuron) UpdatePosition() {
	for i := 0; i < 16; i++ {
		n.Position[i] += common.Coordinate(n.Velocity[i])
	}
}
