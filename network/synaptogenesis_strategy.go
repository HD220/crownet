// Package network contains components for the neural network simulation.
// This file defines interfaces and default strategies for synaptogenesis.
package network

import (
	"math" // Added for Sqrt for DefaultMovementUpdater

	"crownet/common"
	"crownet/config"
	"crownet/neuron"
	"crownet/space" // Added for EuclideanDistance and ClampToHyperSphere
)

// ForceCalculator defines the interface for components responsible for calculating
// the net force acting on a target neuron due to its interactions with other neurons
// in the network, considering simulation parameters and chemical modulation.
type ForceCalculator interface {
	// CalculateForce computes the net force on targetNeuron.
	// - targetNeuron: The neuron for which to calculate the force.
	// - allNeurons: A slice of all neurons in the network for context.
	// - simParams: Global simulation parameters.
	// - modulationFactor: The current chemical modulation factor affecting synaptogenesis.
	// Returns the calculated net force as a common.Vector.
	CalculateForce(
		targetNeuron *neuron.Neuron,
		allNeurons []*neuron.Neuron,
		simParams *config.SimulationParameters,
		modulationFactor float64,
	) common.Vector
}

// MovementUpdater defines the interface for components responsible for updating
// a neuron's position and velocity based on a calculated net force and
// simulation parameters.
type MovementUpdater interface {
	// UpdateMovement calculates the new position and velocity for targetNeuron.
	// - targetNeuron: The neuron whose movement is to be updated.
	// - netForce: The net force acting on the targetNeuron.
	// - simParams: Global simulation parameters.
	// Returns the new position (common.Point) and new velocity (common.Vector).
	UpdateMovement(
		targetNeuron *neuron.Neuron,
		netForce common.Vector,
		simParams *config.SimulationParameters,
	) (newPosition common.Point, newVelocity common.Vector)
}

// DefaultForceCalculator provides the standard implementation of the ForceCalculator interface.
// It uses the original logic based on attraction/repulsion forces modulated by neuron state
// and chemical environment, within a specified influence radius.
type DefaultForceCalculator struct{}

// CalculateForce computes the net force exerted on targetNeuron by all other neurons,
// using the default attraction/repulsion model.
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

		if distance == 0 || (simParams.Synaptogenesis.SynaptogenesisInfluenceRadius > 0 && distance > float64(simParams.Synaptogenesis.SynaptogenesisInfluenceRadius)) {
			continue
		}

		directionUnitVector := common.Vector{}
		for i := range targetNeuron.Position {
			directionUnitVector[i] = common.Coordinate(float64(otherNeuron.Position[i]-targetNeuron.Position[i]) / distance)
		}

		forceMagnitude := 0.0
		if otherNeuron.CurrentState == neuron.Firing || otherNeuron.CurrentState == neuron.AbsoluteRefractory || otherNeuron.CurrentState == neuron.RelativeRefractory {
			forceMagnitude = float64(simParams.Synaptogenesis.AttractionForceFactor) * modulationFactor
		} else if otherNeuron.CurrentState == neuron.Resting {
			forceMagnitude = -float64(simParams.Synaptogenesis.RepulsionForceFactor) * modulationFactor
		}

		for i := range netForce {
			netForce[i] += common.Coordinate(float64(directionUnitVector[i]) * forceMagnitude)
		}
	}
	return netForce
}

// DefaultMovementUpdater provides the standard implementation of the MovementUpdater interface.
// It applies damping, force effects, clamps velocity to MaxMovementPerCycle,
// and ensures the new position is within simulation space boundaries.
type DefaultMovementUpdater struct{}

// UpdateMovement calculates the new position and velocity of a neuron based on the net force acting on it,
// using the default physics model.
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
		vComponent := float64(currentVelocity[i])*float64(simParams.Synaptogenesis.DampeningFactor) + float64(netForce[i])
		updatedVelocity[i] = common.Coordinate(vComponent)
		velocityMagnitudeSq += vComponent * vComponent
	}

	velocityMagnitude := math.Sqrt(velocityMagnitudeSq)

	// Epsilon defined in synaptogenesis.go, can be a local const or imported if made public
	const epsilonVelocityMagnitude = 1e-9 // Copied from original file for now
	if velocityMagnitude > float64(simParams.Synaptogenesis.MaxMovementPerCycle) && velocityMagnitude > epsilonVelocityMagnitude {
		scaleFactor := float64(simParams.Synaptogenesis.MaxMovementPerCycle) / velocityMagnitude
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
	// SpaceMaxDimension is in General sub-struct
	clampedPosition, _ := space.ClampToHyperSphere(calculatedPosition, simParams.General.SpaceMaxDimension)
	newPosition = clampedPosition
	return // Named return values
}
