package core

import (
	"log"
	"math"
	"math/rand"
)

// Constantes exportadas para serem acessíveis por testes, se necessário.
const (
	BaseMovementRate       = 0.01  // Taxa base de movimento por ciclo para sinaptogênese
	AttractionFactor       = 1.5   // Fator pelo qual a atração é mais forte que a repulsão base
	RepulsionFactor        = 1.0   // Fator para repulsão
	InfluenceRadiusSynapto = 2.0   // Raio de influência para sinaptogênese (menor que o de pulso talvez)
	MaxMovementStep        = 0.1   // Movimento máximo em uma única dimensão por ciclo para um neurônio
	MinDistanceThreshold   = 0.05  // Distância mínima para evitar sobreposição ou divisão por zero
)

// ApplySynaptogenesis atualiza as posições dos neurônios com base na atividade da rede.
func (nn *NeuralNetwork) ApplySynaptogenesis() {
	if nn.CurrentCycle == 0 { // Não aplicar no primeiro ciclo
		return
	}

	log.Printf("Ciclo %d: Aplicando Sinaptogênese...", nn.CurrentCycle)

	cortisolLevel := nn.CortisolGland.CortisolLevel
	cortisolModulation := 1.0
	if cortisolLevel > 1.5 {
		cortisolModulation = math.Max(0.1, 1.0-(cortisolLevel-1.5)*0.5)
	}

	avgDopamine := 0.0
	if len(nn.DopamineLevels) > 0 {
		sumDopamine := 0.0
		numWithDopamine := 0
		for _, level := range nn.DopamineLevels {
			if level > 0 {
				sumDopamine += level
				numWithDopamine++
			}
		}
		if numWithDopamine > 0 {
			avgDopamine = sumDopamine / float64(numWithDopamine)
		}
	}
	dopamineModulation := 1.0 + avgDopamine*0.5

	overallMovementRate := BaseMovementRate * cortisolModulation * dopamineModulation
	if overallMovementRate <= 0 {
		log.Printf("Ciclo %d: Sinaptogênese inibida por moduladores (taxa: %.4f)", nn.CurrentCycle, overallMovementRate)
		return
	}
	log.Printf("Ciclo %d: Taxa de movimento para sinaptogênese: %.4f (CortisolMod: %.2f, DopaminaMod: %.2f)", nn.CurrentCycle, overallMovementRate, cortisolModulation, dopamineModulation)

	deltaPositions := make(map[int]*Vector16D) // Usar ponteiros para modificar o vetor diretamente

	for id, neuron := range nn.Neurons {
		// Inicializar deltaPosition para este neurônio se ainda não existir
		if _, ok := deltaPositions[id]; !ok {
			deltaPositions[id] = &Vector16D{}
		}
		currentDelta := deltaPositions[id]


		for otherID, otherNeuron := range nn.Neurons {
			if id == otherID {
				continue
			}

			distance := EuclideanDistance(neuron.Position, otherNeuron.Position) // Corrigido
			if distance > InfluenceRadiusSynapto || distance < MinDistanceThreshold {
				continue
			}

			directionVectorToOther := Vector16D{}
			for i := 0; i < 16; i++ {
				directionVectorToOther[i] = (otherNeuron.Position[i] - neuron.Position[i]) / distance
			}

			movementMagnitude := 0.0
			isOtherActive := otherNeuron.State == Firing ||
				otherNeuron.State == AbsoluteRefractory ||
				otherNeuron.State == RelativeRefractory ||
				(otherNeuron.LastFiringCycle > 0 && (nn.CurrentCycle-otherNeuron.LastFiringCycle) < 5)

			if isOtherActive {
				movementMagnitude = overallMovementRate * AttractionFactor * (1.0 / (distance + 0.1))
				for i := 0; i < 16; i++ {
					(*currentDelta)[i] += directionVectorToOther[i] * movementMagnitude // Modificar através do ponteiro
				}
			} else {
				movementMagnitude = overallMovementRate * RepulsionFactor * (1.0 / (distance*distance + 0.1))
				for i := 0; i < 16; i++ {
					(*currentDelta)[i] -= directionVectorToOther[i] * movementMagnitude // Modificar através do ponteiro
				}
			}
		}
	}

	numMoved := 0
	for id, deltaVec := range deltaPositions {
		ntm := nn.Neurons[id] // Obter o ponteiro para o neurônio (RENOMEADO PARA TESTE)
		moved := false
		for i := 0; i < 16; i++ {
			dimMovement := (*deltaVec)[i]
			if math.Abs(dimMovement) > MaxMovementStep {
				dimMovement = math.Copysign(MaxMovementStep, dimMovement)
			}
			if math.Abs(dimMovement) > 1e-5 {
				ntm.Position[i] += dimMovement // Modificar a posição do neurônio diretamente
				moved = true
			}
		}

		if moved {
			numMoved++
		}
	}

	// Aplicar os deltas de posição, com um limite máximo de movimento
	numMoved := 0
	for id, delta := range deltaPositions {
		neuron := nn.Neurons[id]
		newPosition := neuron.Position
		moved := false
		for i := 0; i < 16; i++ {
			// Limitar o movimento em cada dimensão
			dimMovement := delta[i]
			if math.Abs(dimMovement) > maxMovementStep {
				dimMovement = math.Copysign(maxMovementStep, dimMovement)
			}
			if math.Abs(dimMovement) > 1e-5 { // Apenas mover se o movimento for significativo
				newPosition[i] += dimMovement
				moved = true
			}
		}

		// Garantir que os neurônios permaneçam dentro de um limite espacial, se houver.
		// O modelo não especifica limites rígidos, mas podemos adicionar se necessário,
		// por exemplo, refletindo nas bordas de um hipercubo de MaxSpaceDistance/2.
		// Por enquanto, eles podem se mover livremente.

		if moved {
			// log.Printf("Ciclo %d: Neurônio %d moveu de %v para %v", nn.CurrentCycle, id, neuron.Position, newPosition)
			neuron.Position = newPosition
			numMoved++
		}
	}
	if numMoved > 0 {
		log.Printf("Ciclo %d: Sinaptogênese aplicada. %d neurônios moveram.", nn.CurrentCycle, numMoved)
	}
}

// randomNoiseVector (não usado atualmente, mas mantido para possível uso futuro)
func randomNoiseVector(magnitude float64) Vector16D {
	var noise Vector16D
	norm := 0.0
	for i := 0; i < 16; i++ {
		noise[i] = rand.NormFloat64()
		norm += noise[i] * noise[i]
	}
	norm = math.Sqrt(norm)
	if norm == 0 {
		return Vector16D{}
	}
	for i := 0; i < 16; i++ {
		noise[i] = (noise[i] / norm) * magnitude
	}
	return noise
}
