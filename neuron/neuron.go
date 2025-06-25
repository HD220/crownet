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

// Constants for emitted pulse signals define the base value of a pulse emitted by neurons of certain types.
const (
	ExcitatoryPulseSignal common.PulseValue = 1.0  // Signal value for excitatory effects.
	InhibitoryPulseSignal common.PulseValue = -1.0 // Signal value for inhibitory effects.
	NeutralPulseSignal    common.PulseValue = 0.0  // Signal value for neutral or non-standard effects (e.g., from Dopaminergic neurons).
)

// nearZeroThreshold is a small value used to clamp accumulated potential to zero if it's very close,
// helping to manage floating-point inaccuracies.
const nearZeroThreshold = 1e-5

// Neuron representa uma unidade computacional individual na rede neural.
// Contém o estado e propriedades de um neurônio, incluindo sua posição, tipo, estado de disparo,
// potencial elétrico, limiares e informações de atividade recente.
type Neuron struct {
	ID common.NeuronID // Identificador único global do neurônio.
	// Type define o papel funcional do neurônio na rede (e.g., Excitatory, Input).
	Type Type
	// Position representa as coordenadas do neurônio no espaço N-dimensional da simulação.
	Position common.Point
	// CurrentState indica o estado operacional atual do neurônio (e.g., Resting, Firing).
	CurrentState State
	// AccumulatedPotential é o potencial elétrico atual acumulado pelo neurônio a partir de pulsos recebidos.
	AccumulatedPotential common.PulseValue
	// BaseFiringThreshold é o limiar de potencial base que o neurônio deve alcançar para disparar.
	BaseFiringThreshold common.Threshold
	// CurrentFiringThreshold é o limiar de disparo atual, que pode ser modulado por neuroquímicos.
	CurrentFiringThreshold common.Threshold
	// LastFiredCycle registra o ciclo da simulação em que o neurônio disparou pela última vez (-1 se nunca).
	LastFiredCycle common.CycleCount
	// CyclesInCurrentState rastreia há quantos ciclos o neurônio está em seu estado atual (útil para períodos refratários).
	CyclesInCurrentState common.CycleCount
	// Velocity representa o vetor de velocidade do neurônio, usado para o mecanismo de sinaptogênese (movimento).
	Velocity common.Vector
}

// New cria e inicializa um novo Neurônio com os parâmetros fornecidos.
// O estado inicial é Resting, potencial acumulado é 0, e LastFiredCycle é -1.
// CurrentFiringThreshold é inicializado com o BaseFiringThreshold dos simParams.
// A velocidade inicial é um vetor zero.
func New(id common.NeuronID, neuronType Type, initialPosition common.Point, simParams *config.SimulationParameters) *Neuron {
	if simParams == nil {
		// Handle nil simParams, perhaps by returning an error or panicking,
		// as BaseFiringThreshold cannot be determined.
		// For now, assume simParams is always provided correctly by the caller (NewCrowNet).
		// Consider adding error return if this assumption might be violated.
	}
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

// IntegrateIncomingPotential atualiza o potencial acumulado do neurônio com um pulso recebido
// e determina se o neurônio dispara.
// Se o neurônio estiver em período refratário absoluto, ele não pode integrar potencial nem disparar.
// Se o potencial acumulado exceder o CurrentFiringThreshold, o neurônio entra no estado Firing.
// Retorna true se o neurônio disparou, false caso contrário.
func (n *Neuron) IntegrateIncomingPotential(potential common.PulseValue, currentCycle common.CycleCount) (fired bool) {
	// Neurons in absolute refractory period cannot integrate new potentials or fire.
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

// AdvanceState atualiza o estado do neurônio com base em sua lógica de máquina de estados interna
// e nos parâmetros de simulação (e.g., durações dos períodos refratários).
// As transições de estado são:
// Firing -> AbsoluteRefractory
// AbsoluteRefractory -> RelativeRefractory (após simParams.AbsoluteRefractoryCycles)
// RelativeRefractory -> Resting (após simParams.RelativeRefractoryCycles)
// O potencial acumulado é resetado ao entrar em AbsoluteRefractory.
func (n *Neuron) AdvanceState(currentCycle common.CycleCount, simParams *config.SimulationParameters) {
	n.CyclesInCurrentState++ // Increment cycles spent in the current state.

	switch n.CurrentState {
	case Firing:
		// After firing, neuron enters absolute refractory period.
		n.CurrentState = AbsoluteRefractory
		n.CyclesInCurrentState = 0 // Reset counter for the new state.
		n.LastFiredCycle = currentCycle // Record the cycle of this firing event.
		n.AccumulatedPotential = 0.0    // Reset potential after firing.
	case AbsoluteRefractory:
		// If absolute refractory period has ended, transition to relative refractory.
		if simParams != nil && n.CyclesInCurrentState >= simParams.AbsoluteRefractoryCycles {
			n.CurrentState = RelativeRefractory
			n.CyclesInCurrentState = 0
		}
	case RelativeRefractory:
		// If relative refractory period has ended, transition back to resting.
		if simParams != nil && n.CyclesInCurrentState >= simParams.RelativeRefractoryCycles {
			n.CurrentState = Resting
			n.CyclesInCurrentState = 0
		}
	case Resting:
		// No state change based on time alone when resting; stays resting until potential causes firing.
		// CyclesInCurrentState will continue to increment, which is fine.
	}
}

// DecayPotential aplica decaimento exponencial ao potencial acumulado do neurônio.
// O potencial decai em direção a zero a uma taxa definida em simParams.AccumulatedPulseDecayRate.
// Se o potencial resultante estiver muito próximo de zero, é fixado em zero para evitar imprecisões de ponto flutuante.
func (n *Neuron) DecayPotential(simParams *config.SimulationParameters) {
	if simParams == nil { // Defensive check
		return
	}
	decayRate := simParams.AccumulatedPulseDecayRate
	if decayRate <= 0 { // No decay if rate is zero or negative.
		return
	}
	if decayRate >= 1.0 { // Full decay if rate is 1.0 or more.
		n.AccumulatedPotential = 0.0
		return
	}

	// Potential decays towards zero by the decayRate factor.
	n.AccumulatedPotential *= common.PulseValue(1.0 - float64(decayRate)) // Ensure decayRate is float64 for calculation

	// Clamp to zero if very close, to avoid floating point inaccuracies.
	if math.Abs(float64(n.AccumulatedPotential)) < nearZeroThreshold {
		n.AccumulatedPotential = 0.0
	}
}

// EmittedPulseSign retorna o sinal base (+1.0, -1.0, ou 0.0) do pulso que este neurônio
// emite ao disparar, com base em seu tipo.
// Neurônios Input e Output são considerados excitatórios por padrão para esta emissão.
// Neurônios Dopaminérgicos emitem um sinal neutro, pois seu efeito é tipicamente via modulação química.
func (n *Neuron) EmittedPulseSign() common.PulseValue {
	switch n.Type {
	case Excitatory, Input, Output: // Input/Output neurons treated as excitatory for pulse emission.
		return ExcitatoryPulseSignal
	case Inhibitory:
		return InhibitoryPulseSignal
	case Dopaminergic:
		// Dopaminergic neurons primarily act via chemical modulation, not direct synaptic pulses in this model.
		return NeutralPulseSignal
	default:
		// Unknown neuron types also emit a neutral signal by default.
		return NeutralPulseSignal
	}
}

// UpdatePosition atualiza a posição do neurônio com base em sua velocidade.
// new_position = old_position + velocity * (time_step=1)
func (n *Neuron) UpdatePosition() {
	for i := range n.Position { // Iterate over dimensions using range
		n.Position[i] += n.Velocity[i] // Both are common.Coordinate, direct addition is fine
	}
}
