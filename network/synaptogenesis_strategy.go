// Package network (specifically this file for synaptogenesis strategies)
// provides different strategies for structural plasticity, including how neurons
// calculate forces between them and how they update their positions based on these forces.
package network

import (
	"math"
	"math/rand"

	"crownet/common"
	"crownet/config"
	"crownet/neuron"
	"crownet/space"
	"crownet/synaptic"
)

// ForceCalculator defines an interface for calculating forces between neurons,
// which can be used to drive neuron movement during synaptogenesis.
type ForceCalculator interface {
	CalculateForces(
		neurons map[common.NeuronID]*neuron.Neuron,
		grid *space.SpatialGrid,
		simParams *config.SimulationParameters,
	) map[common.NeuronID]common.Point
}

// MovementUpdater defines an interface for updating neuron positions based on
// calculated forces and other simulation parameters.
type MovementUpdater interface {
	ApplyMovements(
		neurons map[common.NeuronID]*neuron.Neuron,
		forces map[common.NeuronID]common.Point,
		modulationFactor float64,
		simParams *config.SimulationParameters,
		rng *rand.Rand,
	)
}

// DefaultForceCalculator implements a basic force calculation strategy.
// It considers attractive forces between connected neurons (based on synaptic weights)
// and repulsive forces between all nearby neurons to prevent clumping.
// Neurochemical modulation can influence the strength of these forces.
type DefaultForceCalculator struct{}

// CalculateForces calculates the net force on each neuron.
// TODO: Implement actual force calculation logic. This is a placeholder.
// For now, it only considers repulsion from nearby neurons and a weak attraction to connected ones.
func (dfc *DefaultForceCalculator) CalculateForces(
	neurons map[common.NeuronID]*neuron.Neuron,
	grid *space.SpatialGrid,
	simParams *config.SimulationParameters,
) map[common.NeuronID]common.Point {
	forces := make(map[common.NeuronID]common.Point)
	if simParams == nil {
		// Log error or handle: simParams are essential
		return forces // Return empty forces if params are missing
	}

	for id, n := range neurons {
		netForce := make(common.Point, common.PointDimension) // Initialize net force for neuron n

		// 1. Repulsive forces from all nearby neurons (within SynaptogenesisInfluenceRadius)
		// Use spatial grid to find neighbors efficiently.
		// Neighbors are other neurons, not necessarily connected by synapses.
		influenceRadius := float64(simParams.Synaptogenesis.SynaptogenesisInfluenceRadius)
		neighbors := grid.GetNeighborsWithinRadius(n, influenceRadius)

		for _, otherNeuron := range neighbors {
			if n.ID == otherNeuron.ID {
				continue // Skip self
			}
			distance := space.EuclideanDistance(n.Position, otherNeuron.Position)
			// Avoid division by zero if distance is very small (though grid cell size should help)
			if distance < 1e-6 { // epsilon distance
				distance = 1e-6
			}

			// Repulsive force: F_rep = k_rep / d^2 (inverse square law)
			// Modulate by RepulsionForceFactor from SimParams.
			// The direction is away from otherNeuron.
			repulsionStrength := float64(simParams.Synaptogenesis.RepulsionForceFactor) / (distance * distance)
			for i := 0; i < common.PointDimension; i++ {
				directionComponent := (n.Position[i] - otherNeuron.Position[i]) / common.Coordinate(distance)
				netForce[i] += directionComponent * common.Coordinate(repulsionStrength)
			}
		}

		// 2. Attractive forces to synaptically connected neurons (simplified)
		// This part needs access to synaptic weights to determine connection strength.
		// For simplicity, let's assume a weak constant attraction if any connection exists.
		// A more detailed model would use synaptic.NetworkWeights.
		// For now, this is a conceptual placeholder.
		// Example:
		// for otherID, otherNeuron := range neurons {
		//    if n.ID == otherID { continue }
		//    weight, connected := synWeights.GetWeight(n.ID, otherID) // Assuming synWeights is accessible
		//    if connected && weight > 0 { // Attraction for excitatory connections
		//        distance := space.EuclideanDistance(n.Position, otherNeuron.Position)
		//        if distance < 1e-6 { distance = 1e-6 }
		//        attractionStrength := simParams.Synaptogenesis.AttractionForceFactor * float64(weight) / distance
		//        for i := 0; i < common.PointDimension; i++ {
		//            directionComponent := (otherNeuron.Position[i] - n.Position[i]) / distance
		//            netForce[i] += directionComponent * attractionStrength
		//        }
		//    }
		// }
		forces[id] = netForce
	}
	return forces
}

// DefaultMovementUpdater implements a basic movement update strategy.
// It applies the calculated forces to neurons, considering a dampening factor
// and a maximum movement per cycle to maintain simulation stability.
type DefaultMovementUpdater struct{}

// ApplyMovements updates neuron positions based on forces.
// modulationFactor can be used to scale forces (e.g., from neurochemicals).
func (dmu *DefaultMovementUpdater) ApplyMovements(
	neurons map[common.NeuronID]*neuron.Neuron,
	forces map[common.NeuronID]common.Point,
	modulationFactor float64, // Can be from neurochemical environment
	simParams *config.SimulationParameters,
	rng *rand.Rand, // For any stochastic elements in movement
) {
	if simParams == nil {
		return // Essential parameters missing
	}

	for id, n := range neurons {
		force, ok := forces[id]
		if !ok || force == nil {
			continue // No force calculated for this neuron
		}

		// Apply modulation factor to the force
		modulatedForce := make(common.Point, common.PointDimension)
		for i := 0; i < common.PointDimension; i++ {
			modulatedForce[i] = force[i] * common.Coordinate(modulationFactor)
		}

		// Calculate displacement: dPos = F * (1 - dampening)
		// SimParams.Synaptogenesis.DampeningFactor
		displacementScale := 1.0 - float64(simParams.Synaptogenesis.DampeningFactor)
		if displacementScale < 0 {
			displacementScale = 0 // Avoid negative scaling
		}

		displacement := make(common.Point, common.PointDimension)
		for i := 0; i < common.PointDimension; i++ {
			displacement[i] = modulatedForce[i] * common.Coordinate(displacementScale)
		}

		// Limit maximum movement per cycle
		// SimParams.Synaptogenesis.MaxMovementPerCycle
		displacementMagnitude := space.Magnitude(displacement)
		maxMove := float64(simParams.Synaptogenesis.MaxMovementPerCycle)

		if displacementMagnitude > maxMove && maxMove > 0 { // maxMove > 0 to avoid division by zero if it's not set
			scale := common.Coordinate(maxMove / displacementMagnitude)
			for i := 0; i < common.PointDimension; i++ {
				displacement[i] *= scale
			}
		}

		// Update neuron position
		newPosition := make(common.Point, common.PointDimension)
		for i := 0; i < common.PointDimension; i++ {
			newPosition[i] = n.Position[i] + displacement[i]
		}

		// Clamp position to simulation boundaries (e.g., a sphere or cube)
		// Assuming SimParams.General.SpaceMaxDimension defines a cubic boundary centered at origin
		boundary := simParams.General.SpaceMaxDimension
		for i := 0; i < common.PointDimension; i++ {
			if newPosition[i] > common.Coordinate(boundary) {
				newPosition[i] = common.Coordinate(boundary)
			} else if newPosition[i] < common.Coordinate(-boundary) {
				newPosition[i] = common.Coordinate(-boundary)
			}
		}
		n.Position = newPosition
	}
}

// PruningAndFormationStrategy defines how new synapses might be formed or existing ones pruned.
// This is a conceptual placeholder for a more detailed structural plasticity mechanism.
type PruningAndFormationStrategy interface {
	ApplyPruningAndFormation(
		neurons map[common.NeuronID]*neuron.Neuron,
		weights *synaptic.NetworkWeights,
		grid *space.SpatialGrid,
		simParams *config.SimulationParameters,
		rng *rand.Rand,
		modulationFactor float64, // From neurochemicals, affects probability/rate
	)
}

// DefaultPruningAndFormation implements a basic strategy.
// Example: Prune weak synapses, form new ones randomly between nearby active neurons.
type DefaultPruningAndFormation struct{}

func (dpf *DefaultPruningAndFormation) ApplyPruningAndFormation(
	_ map[common.NeuronID]*neuron.Neuron, // neurons map (unused for now)
	_ *synaptic.NetworkWeights, // weights (unused for now)
	_ *space.SpatialGrid, // grid (unused for now)
	_ *config.SimulationParameters, // simParams (unused for now)
	_ *rand.Rand, // rng (unused for now)
	_ float64, // modulationFactor (unused for now)
) {
	// TODO: Implement logic for pruning weak synapses and forming new ones.
	// This could involve:
	// 1. Iterating through all existing synapses and removing those below a threshold weight.
	//    (modulated by neurochemical state or activity levels).
	// 2. Identifying pairs of neurons that are close in space (using the grid)
	//    and have correlated activity (e.g., both recently fired).
	// 3. Randomly forming new synapses between such pairs with a certain probability,
	//    again potentially modulated by neurochemicals or overall network state.
	//
	// Example placeholder for pruning:
	// for fromID, connections := range weights.GetAllWeights() { // Assuming GetAllWeights returns map[ID]map[ID]Weight
	//    for toID, weight := range connections {
	//        if math.Abs(float64(weight)) < simParams.Learning.MinSynapticWeightForPruning * modulationFactor {
	//            weights.SetWeight(fromID, toID, 0) // Or a method to remove synapse
	//        }
	//    }
	// }
	//
	// Example placeholder for formation:
	// for id1, n1 := range neurons {
	//    if !n1.IsRecentlyActive(simParams.Learning.HebbianCoincidenceWindow) { continue }
	//    neighbors := grid.GetNeighborsWithinRadius(n1, simParams.Synaptogenesis.SynaptogenesisInfluenceRadius)
	//    for _, n2 := range neighbors {
	//        if n1.ID == n2.ID { continue }
	//        if !n2.IsRecentlyActive(simParams.Learning.HebbianCoincidenceWindow) { continue }
	//        if !weights.AreConnected(n1.ID, n2.ID) && rng.Float64() < simParams.Synaptogenesis.NewSynapseFormationProbability * modulationFactor {
	//            initialWeight := (simParams.Learning.InitialSynapticWeightMin + simParams.Learning.InitialSynapticWeightMax) / 2
	//            weights.SetWeight(n1.ID, n2.ID, initialWeight)
	//        }
	//    }
	// }
}

// TODO: Remove this placeholder if the above DefaultForceCalculator is sufficient
// SimpleSynaptogenesisStrategy provides a basic implementation for neuron movement.
// It calculates forces based on proximity and connection strength (simplified)
// and updates positions accordingly.
type SimpleSynaptogenesisStrategy struct {
	ForceCalc   ForceCalculator
	MovementUpd MovementUpdater
	PruneForm   PruningAndFormationStrategy // Added component for synapse dynamics
}

// NewSimpleSynaptogenesisStrategy creates a new strategy with default components.
func NewSimpleSynaptogenesisStrategy() *SimpleSynaptogenesisStrategy {
	return &SimpleSynaptogenesisStrategy{
		ForceCalc:   &DefaultForceCalculator{},
		MovementUpd: &DefaultMovementUpdater{},
		PruneForm:   &DefaultPruningAndFormation{}, // Initialize with default
	}
}

// ApplyStructuralChanges orchestrates the synaptogenesis process for one cycle.
func (s *SimpleSynaptogenesisStrategy) ApplyStructuralChanges(
	neurons map[common.NeuronID]*neuron.Neuron,
	grid *space.SpatialGrid,
	weights *synaptic.NetworkWeights,
	simParams *config.SimulationParameters,
	rng *rand.Rand,
	modulationFactor float64, // Overall modulation from neurochemicals
) {
	// 1. Calculate forces on neurons
	forces := s.ForceCalc.CalculateForces(neurons, grid, simParams)

	// 2. Update neuron positions based on these forces
	// The modulationFactor here could be specific to movement sensitivity.
	s.MovementUpd.ApplyMovements(neurons, forces, modulationFactor, simParams, rng)

	// 3. Apply pruning of old/weak synapses and formation of new ones
	// This step might use a different or related modulationFactor.
	if s.PruneForm != nil { // Check if a pruning/formation strategy is set
		s.PruneForm.ApplyPruningAndFormation(neurons, weights, grid, simParams, rng, modulationFactor)
	}
}
