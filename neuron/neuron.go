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
// O ID é fornecido externamente para garantir unicidade na rede.
func New(id common.NeuronID, neuronType Type, initialPosition common.Point, simParams *config.SimulationParameters) *Neuron {
	n := &Neuron{
		ID:                     id,
		Type:                   neuronType,
		Position:               initialPosition,
		CurrentState:           Resting,
		AccumulatedPotential:   0.0,
		BaseFiringThreshold:    common.Threshold(simParams.BaseFiringThreshold),
		CurrentFiringThreshold: common.Threshold(simParams.BaseFiringThreshold), // Inicialmente igual ao base
		LastFiredCycle:         -1, // -1 indica que nunca disparou
		CyclesInCurrentState:   0,
		Velocity:               common.Vector{}, // Velocidade inicial zero
	}
	return n
}

// IntegrateIncomingPotential processa um potencial sináptico recebido.
// Retorna true se o neurônio disparou como resultado desta integração.
// O cicloAtual é passado para registrar quando ocorreu o disparo.
func (n *Neuron) IntegrateIncomingPotential(potential common.PulseValue, currentCycle common.CycleCount) (fired bool) {
	if n.CurrentState == AbsoluteRefractory {
		return false // Não pode integrar nem disparar em estado refratário absoluto
	}

	n.AccumulatedPotential += potential

	if n.AccumulatedPotential < n.CurrentFiringThreshold {
		return false // Não atingiu o limiar
	}
	// Atingiu o limiar e não está em AbsoluteRefractory
	n.CurrentState = Firing // Transição imediata para o estado de disparo
	n.CyclesInCurrentState = 0
	// n.LastFiredCycle é atualizado em AdvanceState após o estado Firing.
	return true
}

const nearZeroThreshold = 1e-5

// AdvanceState progride o estado do neurônio para o próximo ciclo.
// Deve ser chamado uma vez por ciclo para cada neurônio, após todos os inputs terem sido integrados.
// O cicloAtual é usado para registrar o LastFiredCycle corretamente.
func (n *Neuron) AdvanceState(currentCycle common.CycleCount, simParams *config.SimulationParameters) {
	n.CyclesInCurrentState++

	// Transições de estado baseadas no estado atual e no tempo decorrido nele
	if n.CurrentState == Firing {
		n.CurrentState = AbsoluteRefractory
		n.CyclesInCurrentState = 0
		n.LastFiredCycle = currentCycle
		n.AccumulatedPotential = 0.0 // Reset do potencial após o disparo
		return                       // Transição feita
	}

	if n.CurrentState == AbsoluteRefractory {
		// simParams.AbsoluteRefractoryCycles agora é common.CycleCount, sem necessidade de cast.
		if n.CyclesInCurrentState >= simParams.AbsoluteRefractoryCycles {
			n.CurrentState = RelativeRefractory
			n.CyclesInCurrentState = 0
			// O limiar em RelativeRefractory pode ser maior; por enquanto, CurrentFiringThreshold é
			// modificado externamente por neuroquímicos. Se não houver modulação,
			// poderia-se aumentar o BaseFiringThreshold temporariamente aqui.
		}
		return // Transição (ou não) feita
	}

	if n.CurrentState == RelativeRefractory {
		// simParams.RelativeRefractoryCycles agora é common.CycleCount, sem necessidade de cast.
		if n.CyclesInCurrentState >= simParams.RelativeRefractoryCycles {
			n.CurrentState = Resting
			n.CyclesInCurrentState = 0
			// Restaura o limiar para o valor base (se não estiver sendo modulado por químicos)
			// n.CurrentFiringThreshold = n.BaseFiringThreshold; // Essa lógica é mais complexa devido à modulação externa
		}
		return // Transição (ou não) feita
	}
	// Se Resting, permanece Resting a menos que um disparo ocorra via IntegrateIncomingPotential.
}

// DecayPotential aplica o decaimento ao potencial acumulado do neurônio.
// Deve ser chamado uma vez por ciclo.
func (n *Neuron) DecayPotential(simParams *config.SimulationParameters) {
	decayRate := common.Rate(simParams.AccumulatedPulseDecayRate)
	if n.AccumulatedPotential > 0 {
		n.AccumulatedPotential -= common.PulseValue(float64(n.AccumulatedPotential) * float64(decayRate))
		if n.AccumulatedPotential < nearZeroThreshold { // Usar constante
			n.AccumulatedPotential = 0
		}
		return // Decaimento aplicado
	}
	// Se o potencial for negativo (devido a inputs inibitórios), também decai em direção a zero.
	if n.AccumulatedPotential < 0 {
		n.AccumulatedPotential += common.PulseValue(float64(n.AccumulatedPotential) * float64(decayRate) * -1.0) // decayRate é positivo
		if n.AccumulatedPotential > -nearZeroThreshold { // Usar constante
			n.AccumulatedPotential = 0
		}
	}
}

// EmittedPulseSign retorna o sinal base do pulso que este neurônio emite ao disparar.
// +1.0 para excitatório, -1.0 para inibitório.
// Neurônios dopaminérgicos não emitem pulsos ponderados desta forma; seu efeito é químico.
func (n *Neuron) EmittedPulseSign() common.PulseValue {
	if n.Type == Excitatory || n.Type == Input || n.Type == Output {
		return 1.0
	}
	if n.Type == Inhibitory {
		return -1.0
	}
	return 0.0 // Dopaminergic ou tipo desconhecido não emite pulso padrão
}

// UpdatePosition atualiza a posição do neurônio com base em sua velocidade.
// dx = v * dt. Como dt = 1 ciclo, dx = v.
func (n *Neuron) UpdatePosition() {
	for i := 0; i < 16; i++ {
		n.Position[i] += common.Coordinate(n.Velocity[i])
		// Condições de contorno (clamping ao espaço) serão tratadas no pacote `network` ou `space`
		// que tem conhecimento das dimensões do espaço.
	}
}
```
