package network

import (
	"crownet/common"
	// "crownet/config" // Unused
	// "crownet/neuron" // Unused
	// "crownet/space"  // Unused
	// "math"           // Unused
)

// minSynaptogenesisModulationFactorThreshold is a small value below which the
// neurochemical modulation factor for synaptogenesis is considered negligible,
// potentially skipping the computationally intensive synaptogenesis process.
const minSynaptogenesisModulationFactorThreshold = 1e-6

// epsilonVelocityMagnitude is a small value used to avoid division by zero or
// issues with normalizing a near-zero velocity vector when clamping velocity.
// const epsilonVelocityMagnitude = 1e-9 // Moved to DefaultMovementUpdater

// REFACTOR-006: calculateNetForceOnNeuron and updateNeuronMovement have been moved
// to be methods of DefaultForceCalculator and DefaultMovementUpdater respectively,
// in synaptogenesis_strategy.go.

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
		// REFACTOR-006: Use interface methods
		netForce := cn.SynaptogenesisForceCalculator.CalculateForce(n1, cn.Neurons, cn.SimParams, modulationFactor)
		newPos, newVel := cn.SynaptogenesisMovementUpdater.UpdateMovement(n1, netForce, cn.SimParams)
		tempNewPositions[n1.ID] = newPos
		tempNewVelocities[n1.ID] = newVel
	}

	// Phase 2: Apply the calculated new positions and velocities.
	for _, n := range cn.Neurons {
		n.Position = tempNewPositions[n.ID]
		n.Velocity = tempNewVelocities[n.ID]
	}
}
