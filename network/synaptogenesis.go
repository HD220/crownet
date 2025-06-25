package network

import (
	"crownet/common"
	"crownet/config"
	"crownet/neuron"
	"crownet/space"
	"math"
)

// minSynaptogenesisModulationFactorThreshold is a small value below which the
// neurochemical modulation factor for synaptogenesis is considered negligible,
// potentially skipping the computationally intensive synaptogenesis process.
const minSynaptogenesisModulationFactorThreshold = 1e-6

// epsilonVelocityMagnitude is a small value used to avoid division by zero or
// issues with normalizing a near-zero velocity vector when clamping velocity.
const epsilonVelocityMagnitude = 1e-9

// calculateNetForceOnNeuron computes the net force exerted on neuron n1 by all other neurons n2 in the network.
// Forces are modulated by the chemical environment (modulationFactor).
// Active (firing or refractory) neurons exert an attractive force.
// Resting neurons exert a repulsive force.
// The force is only considered if n2 is within n1's SynaptogenesisInfluenceRadius.
func calculateNetForceOnNeuron(n1 *neuron.Neuron, allNeurons []*neuron.Neuron, simParams *config.SimulationParameters, modulationFactor float64) common.Vector {
	// REFACTOR-007: Add nil check for simParams
	if simParams == nil {
		// Log or handle error appropriately. Returning zero vector if simParams are missing.
		// Consider logging: log.Println("Warning: calculateNetForceOnNeuron called with nil simParams")
		return common.Vector{}
	}
	netForce := common.Vector{} // Initialize zero vector for accumulating forces

	for _, n2 := range allNeurons {
		if n1.ID == n2.ID { // A neuron does not exert force on itself
			continue
		}

		distance := space.EuclideanDistance(n1.Position, n2.Position)

		// Skip if neurons are co-located (distance is 0) or if n2 is outside the influence radius.
		// An influence radius of 0 or less means it's not used (global influence, though unlikely for this formula).
		if distance == 0 || (simParams.SynaptogenesisInfluenceRadius > 0 && distance > simParams.SynaptogenesisInfluenceRadius) {
			continue
		}

		// Calculate the unit vector for the direction from n1 to n2.
		directionUnitVector := common.Vector{}
		for i := range n1.Position { // Loop over dimensions of the point/vector
			directionUnitVector[i] = common.Coordinate(float64(n2.Position[i]-n1.Position[i]) / distance)
		}

		forceMagnitude := 0.0
		// Determine force type (attraction/repulsion) based on the state of neuron n2.
		if n2.CurrentState == neuron.Firing || n2.CurrentState == neuron.AbsoluteRefractory || n2.CurrentState == neuron.RelativeRefractory {
			// Active or recently active neurons attract.
			forceMagnitude = simParams.AttractionForceFactor * modulationFactor
		} else if n2.CurrentState == neuron.Resting {
			// Resting neurons repel.
			forceMagnitude = -simParams.RepulsionForceFactor * modulationFactor // Negative sign indicates repulsion
		}

		// Add the calculated force vector (direction * magnitude) to the net force.
		for i := range netForce { // Loop over dimensions
			netForce[i] += common.Coordinate(float64(directionUnitVector[i]) * forceMagnitude)
		}
	}
	return netForce
}

// updateNeuronMovement calculates the new position and velocity of a neuron based on the net force acting on it.
// It applies damping to the current velocity, adds the effect of the net force,
// clamps the new velocity to MaxMovementPerCycle, and then updates the position.
// The new position is also clamped to stay within the simulation space boundaries.
// Time step is implicitly 1 cycle.
func updateNeuronMovement(n *neuron.Neuron, netForce common.Vector, simParams *config.SimulationParameters) (newPosition common.Point, newVelocity common.Vector) {
	// REFACTOR-007: Add nil check for simParams
	if simParams == nil {
		// Log or handle error appropriately. Returning current position and velocity if simParams are missing.
		// Consider logging: log.Println("Warning: updateNeuronMovement called with nil simParams")
		return n.Position, n.Velocity
	}
	currentVelocity := n.Velocity
	updatedVelocity := common.Vector{}
	var velocityMagnitudeSq float64 // Use float64 for sum of squares

	// Calculate new velocity: v_new_unclamped = v_old * dampingFactor + F_net * (time_step=1)
	// Loop over dimensions
	for i := range currentVelocity {
		vComponent := float64(currentVelocity[i])*simParams.DampeningFactor + float64(netForce[i])
		updatedVelocity[i] = common.Coordinate(vComponent)
		velocityMagnitudeSq += vComponent * vComponent
	}

	velocityMagnitude := math.Sqrt(velocityMagnitudeSq)

	// Clamp velocity magnitude if it exceeds MaxMovementPerCycle.
	// The epsilon check avoids division by zero if velocityMagnitude is extremely small.
	if velocityMagnitude > simParams.MaxMovementPerCycle && velocityMagnitude > epsilonVelocityMagnitude {
		scaleFactor := simParams.MaxMovementPerCycle / velocityMagnitude
		for i := range updatedVelocity { // Loop over dimensions
			updatedVelocity[i] = common.Coordinate(float64(updatedVelocity[i]) * scaleFactor)
		}
	}
	newVelocity = updatedVelocity

	// Calculate new position: p_new = p_old + v_new_clamped * (time_step=1)
	currentPosition := n.Position
	calculatedPosition := currentPosition // Start with current, then add velocity components
	for i := range currentPosition {      // Loop over dimensions
		calculatedPosition[i] += newVelocity[i] // Both are common.Coordinate
	}

	// Clamp new position to be within the hypersphere boundary defined by SpaceMaxDimension.
	// The second return value from ClampToHyperSphere (bool indicating if clamped) is currently ignored.
	clampedPosition, _ := space.ClampToHyperSphere(calculatedPosition, simParams.SpaceMaxDimension)
	newPosition = clampedPosition
	return // Named return values newPosition, newVelocity are assigned
}

// applySynaptogenesis handles the movement of neurons within the network, a process
// influenced by inter-neuronal forces (attraction/repulsion) and modulated by the
// neurochemical environment. This simulates structural plasticity.
// The update is performed in two phases to ensure all calculations are based on the state
// at the beginning of the current cycle (simultaneous update):
// 1. Calculate new positions and velocities for all neurons.
// 2. Apply these new positions and velocities to all neurons.
func (cn *CrowNet) applySynaptogenesis() {
	// Get the current modulation factor from the chemical environment.
	modulationFactor := float64(cn.ChemicalEnv.SynaptogenesisModulationFactor)

	// If modulation factor is negligible, skip neuron movement to save computation.
	if modulationFactor < minSynaptogenesisModulationFactorThreshold {
		return
	}

	// Temporary storage for calculated new states to ensure simultaneous updates.
	tempNewPositions := make(map[common.NeuronID]common.Point)
	tempNewVelocities := make(map[common.NeuronID]common.Vector)

	// Phase 1: Calculate new positions and velocities for all neurons.
	for _, n1 := range cn.Neurons {
		netForce := calculateNetForceOnNeuron(n1, cn.Neurons, cn.SimParams, modulationFactor)
		newPos, newVel := updateNeuronMovement(n1, netForce, cn.SimParams)
		tempNewPositions[n1.ID] = newPos
		tempNewVelocities[n1.ID] = newVel
	}

	// Phase 2: Apply the calculated new positions and velocities.
	for _, n := range cn.Neurons {
		n.Position = tempNewPositions[n.ID]
		n.Velocity = tempNewVelocities[n.ID]
	}
}
