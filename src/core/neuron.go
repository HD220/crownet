package core

import (
	"math"
)

const (
	potentialDecayRate      = 0.1   // Taxa de decaimento do potencial por ciclo se não houver input
	absoluteRefractoryPeriod = 2    // Ciclos
	relativeRefractoryPeriod = 3    // Ciclos (após o absoluto)
	baseFiringThreshold     = 1.0   // Limiar base, pode ser modulado
	maxPotential            = 2.0   // Potencial máximo para evitar runaway
	minPotential            = -1.0  // Potencial mínimo (inibição forte)
)

// UpdateNeuronState atualiza o estado de um neurônio com base em seu potencial atual e estado anterior.
// Esta função é chamada a cada ciclo para cada neurônio.
func (n *Neuron) UpdateNeuronState(currentCycle uint64, networkFiringThreshold float64) (fired bool, newPulse *Pulse) {
	fired = false

	// Aplicar modulação de limiar (simplificado por enquanto, será expandido com cortisol/dopamina)
	n.FiringThreshold = baseFiringThreshold * networkFiringThreshold

	// 1. Gerenciar Períodos Refratários
	if n.State == AbsoluteRefractory || n.State == RelativeRefractory {
		n.CyclesInState++
		if n.State == AbsoluteRefractory && n.CyclesInState >= absoluteRefractoryPeriod {
			n.State = RelativeRefractory
			n.CyclesInState = 0
			// Durante o refratário absoluto, o potencial pode ser resetado ou mantido baixo.
			// Vamos resetá-lo para um pouco abaixo do basal para simular hiperpolarização.
			n.CurrentPotential = -0.1
		} else if n.State == RelativeRefractory && n.CyclesInState >= relativeRefractoryPeriod {
			n.State = Resting
			n.CyclesInState = 0
			n.CurrentPotential = 0 // Reset ao sair do período refratário
		}
		// Durante o período refratário (especialmente absoluto), o neurônio não dispara.
		// Durante o relativo, o limiar é efetivamente mais alto.
		if n.State == AbsoluteRefractory {
			return false, nil
		}
	}

	// 2. Verificar Disparo
	// Durante o período refratário relativo, é mais difícil disparar.
	effectiveThreshold := n.FiringThreshold
	if n.State == RelativeRefractory {
		effectiveThreshold *= 1.5 // Exemplo: 50% mais difícil de disparar
	}

	if n.CurrentPotential >= effectiveThreshold && n.State != AbsoluteRefractory {
		n.State = Firing
		n.CyclesInState = 0
		n.LastFiringCycle = currentCycle
		fired = true

		// Criar um novo pulso
		// A força do pulso pode ser fixa ou baseada no "excesso" de potencial.
		// Por simplicidade, vamos usar uma força fixa baseada no tipo de neurônio.
		var pulseStrength float64
		switch n.Type {
		case Excitatory:
			pulseStrength = 0.3 // Conforme README (soma 0.3)
		case Inhibitory:
			pulseStrength = -0.3 // Conforme README (subtrai 0.3)
		case Dopaminergic:
			// Dopamina tem efeito modulatório, não diretamente somatório/subtrativo no potencial do alvo da mesma forma.
			// O pulso de dopamina será tratado de forma especial na propagação.
			// Por enquanto, vamos dar uma força simbólica. A lógica de aplicação será em `ApplyPulse`.
			pulseStrength = 0.1 // Representa a liberação de dopamina
		default:
			pulseStrength = 0.0 // Neurônios de Input/Output podem não gerar pulsos internos dessa forma.
		}

		if n.Type == Excitatory || n.Type == Inhibitory || n.Type == Dopaminergic {
			newPulse = &Pulse{
				SourceNeuronID: n.ID,
				Strength:       pulseStrength,
				EmittedCycle:   currentCycle,
				CurrentPosition: n.Position, // Pulso começa na posição do neurônio
				// TargetNeuronID e ArrivalTime serão definidos durante a propagação.
			}
		}

		// Após o disparo, o neurônio entra no período refratário absoluto.
		n.State = AbsoluteRefractory
		n.CyclesInState = 0
		// O potencial não necessariamente volta a zero imediatamente, mas não pode causar outro disparo.
		// Alguns modelos resetam o potencial, outros deixam decair.
		// Vamos resetar para perto de zero para significar gasto de energia.
		n.CurrentPotential = 0.0

	} else if n.State == Firing {
		// Se estava disparando no ciclo anterior, agora entra em refratário absoluto.
		n.State = AbsoluteRefractory
		n.CyclesInState = 0
		// O potencial foi "gasto" no disparo.
		n.CurrentPotential = 0.0
	}

	// 3. Decaimento do Potencial / Retorno ao Repouso
	if n.State == Resting || n.State == RelativeRefractory { // Potencial também decai no refratário relativo
		if n.CurrentPotential > 0 {
			n.CurrentPotential -= potentialDecayRate
			if n.CurrentPotential < 0 {
				n.CurrentPotential = 0
			}
		} else if n.CurrentPotential < 0 { // Potencial inibitório também decai para o basal
			n.CurrentPotential += potentialDecayRate
			if n.CurrentPotential > 0 {
				n.CurrentPotential = 0
			}
		}
		if n.State == Resting {
			n.CyclesInState++
		}
	}

	// Limitar o potencial para evitar valores extremos.
	n.CurrentPotential = math.Max(minPotential, math.Min(n.CurrentPotential, maxPotential))

	return fired, newPulse
}

// ApplyPulse aplica o efeito de um pulso a este neurônio.
// Esta função é chamada quando um pulso alcança o neurônio.
func (n *Neuron) ApplyPulse(pulse *Pulse, nn *NeuralNetwork) {
	// Neurônios em período refratário absoluto não são afetados por novos pulsos.
	if n.State == AbsoluteRefractory {
		return
	}

	// Efeito do pulso baseado no tipo do neurônio fonte (implícito na força do pulso)
	// e no tipo do neurônio alvo (receptores, etc. - não modelado explicitamente ainda).

	// Se o pulso vem de um neurônio dopaminérgico, seu efeito é na modulação, não diretamente no potencial.
	// O README diz: "se for dopamina soma a quantidade de dopamina do neuronio" - isso parece
	// referir-se ao *neurônio emissor* (dopaminérgico) e seu output de dopamina, não ao
	// potencial elétrico do neurônio *receptor*.
	// A dopamina afeta o limiar e a sinaptogênese.
	// Vamos assumir que `pulse.Strength` para pulsos de dopamina representa a quantidade de dopamina liberada.
	// Esta dopamina deve ser adicionada ao `DopamineLevels` da rede, possivelmente de forma localizada.

	if pulse.Strength != 0 { // Ignorar pulsos "vazios"
		// Para pulsos excitatórios/inibitórios diretos:
		if nn.Neurons[pulse.SourceNeuronID].Type == Excitatory || nn.Neurons[pulse.SourceNeuronID].Type == Inhibitory {
			n.CurrentPotential += pulse.Strength
			// Limitar o potencial
			n.CurrentPotential = math.Max(minPotential, math.Min(n.CurrentPotential, maxPotential))
		}
		// Se o pulso é de um neurônio dopaminérgico, o `pulse.Strength` é a quantidade de dopamina.
		// Esta dopamina será gerenciada a nível de rede (ex: `nn.DopamineLevels`) ou zonalmente.
		// A função `ApplyPulse` no neurônio receptor não lida diretamente com o aumento de dopamina no ambiente,
		// isso é feito no loop principal da rede ao processar pulsos dopaminérgicos.
	}
}
