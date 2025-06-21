package network

import (
	"crownet/neuron"
	"crownet/utils"
	"math"
)

// applySynaptogenesis updates neuron positions based on their activity states.
func (cn *CrowNet) applySynaptogenesis() {
	// Store new positions temporarily to avoid using updated positions in the same cycle's calculations
	newPositions := make(map[int]neuron.Point)
	newVelocities := make(map[int]neuron.Point)

	for _, n1 := range cn.Neurons {
		totalForceVector := neuron.Point{} // Initialize as zero vector

		for _, n2 := range cn.Neurons {
			if n1.ID == n2.ID {
				continue
			}

			dist := utils.EuclideanDistance(n1.Position, n2.Position)

			if dist == 0 || (SynaptogenesisInfluenceRadius > 0 && dist > SynaptogenesisInfluenceRadius) {
				continue
			}

			// Normalized direction vector from n1 to n2
			directionVector := neuron.Point{}
			for i := 0; i < 16; i++ {
				directionVector[i] = (n2.Position[i] - n1.Position[i]) / dist
			}

			var forceMagnitude float64
			isAttraction := false

			// Determine force type based on n2's state
			switch n2.State {
			case neuron.FiringState, neuron.AbsoluteRefractoryState, neuron.RelativeRefractoryState:
				// Attract to active neurons
				// Force can be simple or distance-dependent, e.g., F/dist or F/dist^2
				// Using a simpler model for now: AttractionForceFactor is the magnitude
				forceMagnitude = AttractionForceFactor * cn.GetSynaptogenesisModulationFactor()
				isAttraction = true
			case neuron.RestingState:
				// Repel from resting neurons
				forceMagnitude = RepulsionForceFactor * cn.GetSynaptogenesisModulationFactor()
				isAttraction = false
			default: // Should not happen
				continue
			}

			// Apply force: if attraction, add; if repulsion, subtract direction vector (effectively adding opposite)
			for i := 0; i < 16; i++ {
				if isAttraction {
					totalForceVector[i] += directionVector[i] * forceMagnitude
				} else {
					totalForceVector[i] -= directionVector[i] * forceMagnitude // Repulsion is away from n2
				}
			}
		}

		// Update velocity: v_new = v_old * damping + force_scaled_by_time (time_step=1 cycle)
		currentVelocity := n1.Velocity
		for i := 0; i < 16; i++ {
			currentVelocity[i] = (currentVelocity[i] * DampeningFactor) + totalForceVector[i]
		}

		// Limit velocity magnitude (movement speed per cycle)
		velocityMagnitude := 0.0
		for i := 0; i < 16; i++ {
			velocityMagnitude += currentVelocity[i] * currentVelocity[i]
		}
		velocityMagnitude = math.Sqrt(velocityMagnitude)

		if velocityMagnitude > MaxMovementPerCycle {
			scale := MaxMovementPerCycle / velocityMagnitude
			for i := 0; i < 16; i++ {
				currentVelocity[i] *= scale
			}
		}
		newVelocities[n1.ID] = currentVelocity

		// Update position: p_new = p_old + v_new*time_step (time_step=1 cycle)
		currentPosition := n1.Position
		for i := 0; i < 16; i++ {
			currentPosition[i] += currentVelocity[i]
		}

		// Boundary conditions: Keep neurons within the space (e.g., clamp or reflect)
		// Clamping to a hyper-sphere of SpaceRadius from origin
		distFromOriginSq := 0.0
		for i := 0; i < 16; i++ {
			distFromOriginSq += currentPosition[i] * currentPosition[i]
		}
		if distFromOriginSq > cn.SpaceRadius*cn.SpaceRadius {
			distFromOrigin := math.Sqrt(distFromOriginSq)
			scale := cn.SpaceRadius / distFromOrigin
			for i := 0; i < 16; i++ {
				currentPosition[i] *= scale
			}
			// Optional: Reflect velocity if hitting boundary
			// This is complex in 16D. Simpler to just clamp position and zero out relevant velocity component or overall velocity.
			// For now, just clamping position. Velocity will naturally adjust due to forces or dampening.
		}
		newPositions[n1.ID] = currentPosition
	}

	// Apply all calculated new positions and velocities
	for _, n := range cn.Neurons {
		if newPos, ok := newPositions[n.ID]; ok {
			n.Position = newPos
		}
		if newVel, ok := newVelocities[n.ID]; ok {
			n.Velocity = newVel
		}
	}
	// fmt.Printf("Synaptogenesis applied. Neuron 0 new pos: %.2f, %.2f ...\n", cn.Neurons[0].Position[0], cn.Neurons[0].Position[1])
}
