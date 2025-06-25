// Package network contains components for the neural network simulation.
// This file defines interfaces and default strategies for synaptogenesis.
package network

import (
	"crownet/common"
	"crownet/config"
	"crownet/neuron"
	"crownet/space" // Added for EuclideanDistance and ClampToHyperSphere
	"math"          // Added for Sqrt for DefaultMovementUpdater
)

// ForceCalculator defines the interface for calculating the net force
// acting on a neuron due to interactions with other neurons.
type ForceCalculator interface {
	CalculateForce(
		targetNeuron *neuron.Neuron,
		allNeurons []*neuron.Neuron,
		simParams *config.SimulationParameters,
		modulationFactor float64, // Chemical modulation factor for synaptogenesis
	) common.Vector
}

// MovementUpdater defines the interface for updating a neuron's position and
// velocity based on the net force acting upon it and simulation parameters.
type MovementUpdater interface {
	UpdateMovement(
		targetNeuron *neuron.Neuron,
		netForce common.Vector,
		simParams *config.SimulationParameters,
	) (newPosition common.Point, newVelocity common.Vector)
}

// DefaultForceCalculator implements the ForceCalculator interface using the
// original logic from calculateNetForceOnNeuron.
type DefaultForceCalculator struct{}

// CalculateForce computes the net force exerted on targetNeuron by all other neurons.
// Forces are modulated by the chemical environment (modulationFactor).
// Active (firing or refractory) neurons exert an attractive force.
// Resting neurons exert a repulsive force.
// The force is only considered if other neurons are within targetNeuron's SynaptogenesisInfluenceRadius.
func (dfc *DefaultForceCalculator) CalculateForce(
	targetNeuron *neuron.Neuron,
	allNeurons []*neuron.Neuron,
	simParams *config.SimulationParameters,
	modulationFactor float64,
) common.Vector {
	// This logic is moved from the original network.calculateNetForceOnNeuron
	if simParams == nil {
		return common.Vector{}
	}
	netForce := common.Vector{}

	for _, otherNeuron := range allNeurons {
		if targetNeuron.ID == otherNeuron.ID {
			continue
		}

		distance := space.EuclideanDistance(targetNeuron.Position, otherNeuron.Position)

		if distance == 0 || (simParams.SynaptogenesisInfluenceRadius > 0 && distance > simParams.SynaptogenesisInfluenceRadius) {
			continue
		}

		directionUnitVector := common.Vector{}
		for i := range targetNeuron.Position {
			directionUnitVector[i] = common.Coordinate(float64(otherNeuron.Position[i]-targetNeuron.Position[i]) / distance)
		}

		forceMagnitude := 0.0
		if otherNeuron.CurrentState == neuron.Firing || otherNeuron.CurrentState == neuron.AbsoluteRefractory || otherNeuron.CurrentState == neuron.RelativeRefractory {
			forceMagnitude = simParams.AttractionForceFactor * modulationFactor
		} else if otherNeuron.CurrentState == neuron.Resting {
			forceMagnitude = -simParams.RepulsionForceFactor * modulationFactor
		}

		for i := range netForce {
			netForce[i] += common.Coordinate(float64(directionUnitVector[i]) * forceMagnitude)
		}
	}
	return netForce
}

// DefaultMovementUpdater implements the MovementUpdater interface using the
// original logic from updateNeuronMovement.
type DefaultMovementUpdater struct{}

// UpdateMovement calculates the new position and velocity of a neuron based on the net force acting on it.
// It applies damping to the current velocity, adds the effect of the net force,
// clamps the new velocity to MaxMovementPerCycle, and then updates the position.
// The new position is also clamped to stay within the simulation space boundaries.
// Time step is implicitly 1 cycle.
func (dmu *DefaultMovementUpdater) UpdateMovement(
	targetNeuron *neuron.Neuron,
	netForce common.Vector,
	simParams *config.SimulationParameters,
) (newPosition common.Point, newVelocity common.Vector) {
	// This logic is moved from the original network.updateNeuronMovement
	if simParams == nil {
		return targetNeuron.Position, targetNeuron.Velocity
	}
	currentVelocity := targetNeuron.Velocity
	updatedVelocity := common.Vector{}
	var velocityMagnitudeSq float64

	for i := range currentVelocity {
		vComponent := float64(currentVelocity[i])*simParams.DampeningFactor + float64(netForce[i])
		updatedVelocity[i] = common.Coordinate(vComponent)
		velocityMagnitudeSq += vComponent * vComponent
	}

	velocityMagnitude := math.Sqrt(velocityMagnitudeSq)

	// Epsilon defined in synaptogenesis.go, can be a local const or imported if made public
	const epsilonVelocityMagnitude = 1e-9 // Copied from original file for now
	if velocityMagnitude > simParams.MaxMovementPerCycle && velocityMagnitude > epsilonVelocityMagnitude {
		scaleFactor := simParams.MaxMovementPerCycle / velocityMagnitude
		for i := range updatedVelocity {
			updatedVelocity[i] = common.Coordinate(float64(updatedVelocity[i]) * scaleFactor)
		}
	}
	newVelocity = updatedVelocity

	currentPosition := targetNeuron.Position
	calculatedPosition := currentPosition
	for i := range currentPosition {
		calculatedPosition[i] += newVelocity[i]
	}

	clampedPosition, _ := space.ClampToHyperSphere(calculatedPosition, simParams.SpaceMaxDimension)
	newPosition = clampedPosition
	return // Named return values
}
