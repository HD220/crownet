package core

import (
	"math"
)

// ApplySynaptogenesis ajusta as posições dos neurônios na rede.
func ApplySynaptogenesis(net *Network) {
	if net == nil || len(net.Neurons) == 0 {
		return
	}

	dopamineFactor := 1.0 + (net.DopamineLevel * net.Config.DopamineEffectOnSynapto)
	cortisolFactor := 1.0
	if net.CortisolLevel > 1.0 {
		cortisolFactor = math.Max(0.1, 1.0-(net.CortisolLevel-1.0)*net.Config.CortisolEffectOnSynapto)
	}

	movementScaleAttract := net.Config.SynaptoMovementRateAttract * dopamineFactor * cortisolFactor
	movementScaleRepel := net.Config.SynaptoMovementRateRepel * dopamineFactor * cortisolFactor

	newPositions := make([][SpaceDimensions]float64, len(net.Neurons))
	for i := range newPositions {
		newPositions[i] = net.Neurons[i].Position
	}

	for i, neuron := range net.Neurons {
		if neuron == nil {
			continue
		}

		var totalMovementVector [SpaceDimensions]float64

		for j, otherNeuron := range net.Neurons {
			if i == j || otherNeuron == nil {
				continue
			}

			directionVector := SubtractVectors(otherNeuron.Position, neuron.Position)
			distance := DistanceEuclidean(neuron.Position, otherNeuron.Position)

			if distance == 0 {
				continue
			}

			normalizedDirection := ScaleVector(directionVector, 1.0/distance)

			isOtherActive := otherNeuron.State == FiringState ||
							 otherNeuron.State == RefractoryAbsoluteState ||
							 otherNeuron.State == RefractoryRelativeState ||
							 (otherNeuron.LastFiringCycle >= net.CurrentCycle - 1 && otherNeuron.LastFiringCycle <= net.CurrentCycle)


			var movementEffect float64
			if isOtherActive {
				movementEffect = movementScaleAttract
			} else {
				movementEffect = -movementScaleRepel
			}

			attenuation := 1.0 / (1.0 + distance*0.05)

			scaledMovement := ScaleVector(normalizedDirection, movementEffect * attenuation)
			totalMovementVector = AddVectors(totalMovementVector, scaledMovement)
		}

		newPositions[i] = AddVectors(neuron.Position, totalMovementVector)
		newPositions[i] = ClampPosition(newPositions[i], net.Config.SpaceSize)
	}

	for i, neuron := range net.Neurons {
		if neuron != nil {
			neuron.SetPosition(newPositions[i])
		}
	}
}
