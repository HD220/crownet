package pulse

import (
	"crownet/common"
	"crownet/config"
	"crownet/neuron"   // Adicionado
	"crownet/space"    // Adicionado
	"crownet/synaptic" // Adicionado
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
	shellEndRadius = p.CurrentDistance
	shellStartRadius = p.CurrentDistance - simParams.PulsePropagationSpeed
	if shellStartRadius < 0 {
		shellStartRadius = 0
	}
	return shellStartRadius, shellEndRadius
}

// --- PulseList ---

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
// Usado principalmente para depuração ou cenários onde acesso externo é necessário.
func (pl *PulseList) GetAll() []*Pulse {
    return pl.pulses
}

// Count retorna o número de pulsos ativos.
func (pl *PulseList) Count() int {
    return len(pl.pulses)
}

// ProcessCycle propaga todos os pulsos, processa seus efeitos nos neurônios
// e retorna uma lista de novos pulsos gerados por neurônios que dispararam.
// A lista interna de pulsos é atualizada (removendo os dissipados).
func (pl *PulseList) ProcessCycle(
	neurons []*neuron.Neuron,
	weights synaptic.NetworkWeights,
	currentCycle common.CycleCount,
	simParams *config.SimulationParameters,
	// spaceOps space.Operations // Se tivéssemos uma interface spaceOps
) (newlyGeneratedPulses []*Pulse) {

	remainingActivePulses := make([]*Pulse, 0, len(pl.pulses))
	newlyGeneratedPulses = make([]*Pulse, 0)

	for _, p := range pl.pulses {
		if !p.Propagate(simParams) {
			continue // Pulso dissipou (excedeu MaxTravelRadius)
		}
		remainingActivePulses = append(remainingActivePulses, p)

		shellStartRadius, shellEndRadius := p.GetEffectShellForCycle(simParams)

		for _, targetNeuron := range neurons {
			if targetNeuron.ID == p.EmittingNeuronID {
				continue // Um pulso não afeta o neurônio que o emitiu diretamente desta forma
			}

			distanceToTarget := space.EuclideanDistance(p.OriginPosition, targetNeuron.Position)

			// Verifica se o targetNeuron está dentro da casca de efeito do pulso
			if distanceToTarget >= shellStartRadius && distanceToTarget < shellEndRadius {
				weight := weights.GetWeight(p.EmittingNeuronID, targetNeuron.ID)
				effectivePotential := p.BaseSignalValue * weight

				if effectivePotential == 0 { // Se peso for zero ou BaseSignalValue for zero
					continue
				}

				if targetNeuron.IntegrateIncomingPotential(effectivePotential, currentCycle) {
					// Neurônio disparou!
					// O estado do targetNeuron já foi mudado para Firing.

					emittedSignal := targetNeuron.EmittedPulseSign()
					if emittedSignal != 0 {
						newP := New( // Usa o construtor New do pacote pulse
							targetNeuron.ID,
							targetNeuron.Position,
							emittedSignal,
							currentCycle,
							simParams.SpaceMaxDimension*2.0, // MaxTravelRadius pode ser o diâmetro do espaço
						)
						newlyGeneratedPulses = append(newlyGeneratedPulses, newP)
					}
				}
			}
		}
	}
	pl.pulses = remainingActivePulses // Atualiza a lista de pulsos ativos da rede
	return newlyGeneratedPulses
}
```
