package neuron

import (
	"crownet/common"
	"crownet/config"
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

	if float64(n.AccumulatedPotential) < float64(n.CurrentFiringThreshold) {
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
func (n *Neuron) DecayPotential(simParams *config.SimulationParameters) {
	decayRate := common.Rate(simParams.AccumulatedPulseDecayRate)
	if n.AccumulatedPotential > 0 {
		n.AccumulatedPotential -= common.PulseValue(float64(n.AccumulatedPotential) * float64(decayRate))
		if n.AccumulatedPotential < nearZeroThreshold {
			n.AccumulatedPotential = 0
		}
		return
	}
	if n.AccumulatedPotential < 0 {
		n.AccumulatedPotential += common.PulseValue(float64(n.AccumulatedPotential) * float64(decayRate) * -1.0)
		if n.AccumulatedPotential > -nearZeroThreshold {
			n.AccumulatedPotential = 0
		}
	}
}

// EmittedPulseSign retorna o sinal base do pulso que este neurônio emite ao disparar.
func (n *Neuron) EmittedPulseSign() common.PulseValue {
	if n.Type == Excitatory || n.Type == Input || n.Type == Output {
		return 1.0
	}
	if n.Type == Inhibitory {
		return -1.0
	}
	return 0.0
}

// UpdatePosition atualiza a posição do neurônio com base em sua velocidade.
func (n *Neuron) UpdatePosition() {
	for i := 0; i < 16; i++ {
		n.Position[i] += common.Coordinate(n.Velocity[i])
	}
}
