package core

import (
	"math"
)

// ApplySynaptogenesis ajusta as posições dos neurônios na rede.
// Esta é uma função complexa que simula a plasticidade estrutural.
// Baseado no README:
// - Neurônios se aproximam daqueles que dispararam ou estavam em período refratário.
// - Neurônios se afastam daqueles que estavam em repouso.
// - A taxa de movimentação é afetada por cortisol e dopamina.
func ApplySynaptogenesis(net *Network) {
	if net == nil || len(net.Neurons) == 0 {
		return
	}

	// Fatores de modulação da sinaptogênese por neuroquímicos
	// Dopamina aumenta sinapto, Cortisol (alto) diminui sinapto
	// Estes são multiplicadores para as taxas base de movimento.
	dopamineFactor := 1.0 + (net.DopamineLevel * net.Config.DopamineEffectOnSynapto)
	cortisolFactor := 1.0
	if net.CortisolLevel > 1.0 { // Alto cortisol reduz sinaptogênese
		cortisolFactor = math.Max(0.1, 1.0-(net.CortisolLevel-1.0)*net.Config.CortisolEffectOnSynapto)
	} else if net.CortisolLevel > 0.5 { // Cortisol moderado pode ter um leve efeito positivo ou neutro
		// cortisolFactor = 1.0 + (net.CortisolLevel * net.Config.CortisolEffectOnSynapto * 0.1) // Pequeno aumento
	}


	movementScaleAttract := net.Config.SynaptoMovementRateAttract * dopamineFactor * cortisolFactor
	movementScaleRepel := net.Config.SynaptoMovementRateRepel * dopamineFactor * cortisolFactor

	// Para evitar recalcular o estado de cada neurônio N vezes para cada um dos N neurônios,
	// podemos primeiro classificar os neurônios ou usar seus estados atuais.
	// A sinaptogênese ocorre APÓS a propagação de pulsos no ciclo.
	// Portanto, os estados (Firing, Refractory, Resting) estão atualizados.

	newPositions := make([][SpaceDimensions]float64, len(net.Neurons))
	for i := range newPositions {
		newPositions[i] = net.Neurons[i].Position // Começa com a posição atual
	}

	for i, neuron := range net.Neurons {
		if neuron == nil {
			continue
		}

		var totalMovementVector [SpaceDimensions]float64 // Vetor de movimento acumulado para o neurônio i

		for j, otherNeuron := range net.Neurons {
			if i == j || otherNeuron == nil {
				continue
			}

			directionVector := SubtractVectors(otherNeuron.Position, neuron.Position)
			distance := DistanceEuclidean(neuron.Position, otherNeuron.Position)

			if distance == 0 { // Evitar divisão por zero; não deveria acontecer se i != j
				continue
			}

			// Normalizar o vetor de direção
			normalizedDirection := ScaleVector(directionVector, 1.0/distance)

			// Interação baseada no estado do otherNeuron
			// "Neurônios se aproximam daqueles que dispararam ou estavam em período refratário."
			// "Neurônios se afastam daqueles que estavam em repouso."
			// O "dispararam" refere-se ao ciclo atual ou a um histórico recente?
			// Assumindo que se refere ao estado atual após a fase de disparo do ciclo.
			// LastFiringCycle pode ser usado para "recentemente disparou".
			// Para MVP, usar o estado atual (FiringState, Refractory*) ou LastFiringCycle == net.CurrentCycle.

			isOtherActive := otherNeuron.State == FiringState ||
							 otherNeuron.State == RefractoryAbsoluteState ||
							 otherNeuron.State == RefractoryRelativeState ||
							 (otherNeuron.LastFiringCycle >= net.CurrentCycle - 1 && otherNeuron.LastFiringCycle <= net.CurrentCycle) // Disparou neste ciclo ou no anterior


			var movementEffect float64
			if isOtherActive {
				// Atração: mover `neuron` em direção a `otherNeuron`
				// A força da atração pode diminuir com a distância (ex: 1/distance ou 1/distance^2)
				// ou ser uma taxa constante até um certo raio.
				// Usar uma taxa constante para simplificar.
				movementEffect = movementScaleAttract
			} else { // otherNeuron está em RestingState e não disparou recentemente
				// Repulsão: mover `neuron` para longe de `otherNeuron`
				movementEffect = -movementScaleRepel // Negativo para inverter a direção
			}

			// O movimento deve ser mais fraco para neurônios distantes?
			// Ou há um "raio de influência" para sinaptogênese?
			// Para MVP, vamos assumir que todos os neurônios exercem alguma influência,
			// mas pode ser ponderado pela distância.
			// Ex: movementEffect /= (1 + distance*0.1) // Atenuar com a distância
			// Se não atenuar, a rede pode se aglomerar/dispersar demais.
			// Vamos aplicar uma atenuação simples.
			attenuation := 1.0 / (1.0 + distance*0.05) // Atenuação suave

			scaledMovement := ScaleVector(normalizedDirection, movementEffect * attenuation)
			totalMovementVector = AddVectors(totalMovementVector, scaledMovement)
		}

		// Aplicar o vetor de movimento total à posição original do neurônio
		// para calcular sua nova posição candidata.
		// É importante usar as posições originais para calcular todos os movimentos
		// antes de atualizar qualquer uma, para evitar que o movimento de um neurônio afete
		// o cálculo do movimento de outro no mesmo ciclo de sinaptogênese.
		// (Já estamos fazendo isso ao popular `newPositions` e lendo de `net.Neurons[i].Position`)

		newPositions[i] = AddVectors(neuron.Position, totalMovementVector)

		// Garantir que a nova posição esteja dentro dos limites do espaço
		newPositions[i] = ClampPosition(newPositions[i], net.Config.SpaceSize)
	}

	// Atualizar todas as posições dos neurônios de uma vez
	for i, neuron := range net.Neurons {
		if neuron != nil {
			neuron.SetPosition(newPositions[i])
		}
	}
}
