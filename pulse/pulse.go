package pulse

import (
	"crownet/common"
	"crownet/config"
)

// Pulse representa um sinal elétrico ou químico propagando-se pela rede.
type Pulse struct {
	EmittingNeuronID common.NeuronID
	OriginPosition   common.Point      // Posição de origem do neurônio emissor no momento do disparo
	BaseSignalValue  common.PulseValue // Sinal base (+1.0 excitatório, -1.0 inibitório)
	CreationCycle    common.CycleCount
	CurrentDistance  float64 // Distância que o pulso já percorreu desde a origem
	MaxTravelRadius  float64 // Distância máxima que este pulso pode percorrer antes de dissipar
}

// New cria um novo Pulso.
func New(emitterID common.NeuronID, origin common.Point, signal common.PulseValue, creationCycle common.CycleCount, maxRadius float64) *Pulse {
	return &Pulse{
		EmittingNeuronID: emitterID,
		OriginPosition:   origin,
		BaseSignalValue:  signal,
		CreationCycle:    creationCycle,
		CurrentDistance:  0.0, // Começa na origem
		MaxTravelRadius:  maxRadius,
	}
}

// Propagate avança a distância do pulso e verifica se ele ainda está ativo.
// Retorna false se o pulso se dissipou (excedeu MaxTravelRadius).
func (p *Pulse) Propagate(simParams *config.SimulationParameters) (isActive bool) {
	p.CurrentDistance += simParams.PulsePropagationSpeed
	return p.CurrentDistance < p.MaxTravelRadius
}

// GetEffectShellForCycle calcula o raio interno e externo da "casca" esférica
// onde este pulso exerce sua influência durante o ciclo atual.
// Um neurônio é afetado se sua distância da OriginPosition do pulso cair dentro desta casca.
func (p *Pulse) GetEffectShellForCycle(simParams *config.SimulationParameters) (shellStartRadius, shellEndRadius float64) {
	// O pulso afeta a região que ele varre neste ciclo.
	// Raio final da casca neste ciclo.
	shellEndRadius = p.CurrentDistance
	// Raio inicial da casca neste ciclo (onde estava no final do ciclo anterior).
	shellStartRadius = p.CurrentDistance - simParams.PulsePropagationSpeed

	// Garante que o raio inicial não seja negativo se for o primeiro ciclo de propagação do pulso.
	if shellStartRadius < 0 {
		shellStartRadius = 0
	}
	return shellStartRadius, shellEndRadius
}
```
