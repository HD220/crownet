package network

import (
	"crownet/datagen" // For PatternSize
	"crownet/neuron"
	"crownet/utils" // Will be needed for EuclideanDistance
	"log"           // For Fatalf
	"math"
	"math/rand"
	// "fmt" // No longer needed after removing debug prints
)

// CortisolGlandPosition is at the center of the space.
var CortisolGlandPosition = neuron.Point{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

// CrowNet represents the entire neural network.
type CrowNet struct {
	Neurons                []*neuron.Neuron
	ActivePulses           []*Pulse // List of currently traveling pulses
	InputNeuronIDs         []int
	OutputNeuronIDs        []int
	CortisolLevel          float64
	DopamineLevel          float64
	CycleCount             int
	neuronIDCounter        int
	SpaceRadius            float64 // Max radius of the entire space from center (e.g. 4.0)
	PulseMaxTravelDistance float64 // Max distance a pulse can travel (e.g. 8.0, diameter of space)
	DopaminergicMaxRadius  float64
	InhibitoryMaxRadius    float64
	ExcitatoryMaxRadius    float64
	// Chemical modulation state
	currentSynaptogenesisModulationFactor float64
	// Input/Output state
	inputTargetFrequencies map[int]float64 // neuronID (must be an InputNeuron) -> target frequency in Hz
	timeToNextInputFire    map[int]int     // neuronID -> cycles remaining until it should fire
	outputFiringHistory    map[int][]int   // neuronID (must be an OutputNeuron) -> list of cycle counts when it fired
	// Synaptic weights: map[from_neuron_id]map[to_neuron_id]weight
	SynapticWeights map[int]map[int]float64
	// Flags to control complex dynamics for different modes
	EnableSynaptogenesis     bool
	EnableChemicalModulation bool
	// Learning parameter
	BaseLearningRate float64
	// ReferencePoints for advanced pulse propagation (TODO)
}

// NewCrowNet initializes a new CrowNet instance with neurons distributed.
func NewCrowNet(totalNeurons int) *CrowNet {
	net := &CrowNet{
		Neurons:                make([]*neuron.Neuron, 0, totalNeurons),
		ActivePulses:           make([]*Pulse, 0),
		InputNeuronIDs:         make([]int, 0), // Will be populated by initializeNeurons
		OutputNeuronIDs:        make([]int, 0), // Will be populated by initializeNeurons
		CortisolLevel:          0.0,
		DopamineLevel:          0.0,
		CycleCount:             0,
		neuronIDCounter:        0,
		SpaceRadius:            SpaceMaxDimension,
		PulseMaxTravelDistance: SpaceMaxDimension * 2.0, // Diameter
		DopaminergicMaxRadius:  SpaceMaxDimension * DopaminergicRadiusFactor,
		InhibitoryMaxRadius:    SpaceMaxDimension * InhibitoryRadiusFactor,
		ExcitatoryMaxRadius:    SpaceMaxDimension * ExcitatoryRadiusFactor,
		// I/O
		inputTargetFrequencies:   make(map[int]float64),
		timeToNextInputFire:      make(map[int]int),
		outputFiringHistory:      make(map[int][]int),
		SynapticWeights:          make(map[int]map[int]float64),
		EnableSynaptogenesis:     true,             // Default to on
		EnableChemicalModulation: true,             // Default to on
		BaseLearningRate:         BaseLearningRate, // From config.go
	}

	net.initializeNeurons(totalNeurons)         // Populates InputNeuronIDs and OutputNeuronIDs
	net.initializeSynapticWeights(totalNeurons) // Initialize weights after neurons are created
	net.initializeChemicalModulation()          // Initialize factors
	net.finalizeInitialization()                // Initialize IO related structures and sort IDs
	return net
}

func (cn *CrowNet) getNextID() int {
	id := cn.neuronIDCounter
	cn.neuronIDCounter++
	return id
}

// initializeNeurons creates and distributes neurons in the network.
func (cn *CrowNet) initializeNeurons(totalNeurons int) {
	// For digit recognition MVP, ensure at least datagen.PatternSize input neurons and 10 output neurons.
	numRequiredInput := datagen.PatternSize // 35
	numRequiredOutput := 10

	numInput := numRequiredInput
	if totalNeurons < numInput+numRequiredOutput { // Basic sanity for very small totalNeurons
		log.Fatalf("Total neurons (%d) too small for required input (%d) and output (%d) neurons.", totalNeurons, numInput, numRequiredOutput)
	}

	// Ensure enough input neurons are designated if user requests fewer via percentages
	// The percentage calculation will be for the *remaining* neurons after I/O are set.
	// For MVP, we prioritize fixed numbers for I/O for classification.

	numOutput := numRequiredOutput

	remainingNeurons := totalNeurons - numInput - numOutput
	if remainingNeurons < 0 {
		remainingNeurons = 0
	}

	numDopaminergic := int(math.Round(float64(remainingNeurons) * DopaminergicPercent))
	numInhibitory := int(math.Round(float64(remainingNeurons) * InhibitoryPercent))
	numExcitatory := remainingNeurons - numDopaminergic - numInhibitory

	if numExcitatory < 0 {
		numExcitatory = 0
	}

	currentTotal := numInput + numOutput + numDopaminergic + numInhibitory + numExcitatory
	if currentTotal != totalNeurons {
		// fmt.Printf("Warning: Neuron count mismatch. Target: %d, Calculated: %d. Adjusting Excitatory.\n", totalNeurons, currentTotal)
		numExcitatory += (totalNeurons - currentTotal)
		if numExcitatory < 0 {
			numExcitatory = 0
		}
	}

	// Create Input neurons
	for i := 0; i < numInput; i++ {
		pos := cn.generatePositionInRadius(cn.ExcitatoryMaxRadius) // Placed like excitatory
		id := cn.getNextID()
		n := neuron.NewNeuron(id, pos, neuron.InputNeuron, neuron.DefaultFiringThreshold)
		cn.Neurons = append(cn.Neurons, n)
		cn.InputNeuronIDs = append(cn.InputNeuronIDs, id)
	}
	// ... (rest of neuron type creation as before) ...
	// Create Output neurons
	for i := 0; i < numOutput; i++ {
		pos := cn.generatePositionInRadius(cn.ExcitatoryMaxRadius) // Placed like excitatory
		id := cn.getNextID()
		n := neuron.NewNeuron(id, pos, neuron.OutputNeuron, neuron.DefaultFiringThreshold)
		cn.Neurons = append(cn.Neurons, n)
		cn.OutputNeuronIDs = append(cn.OutputNeuronIDs, id)
	}

	// Create Dopaminergic neurons
	for i := 0; i < numDopaminergic; i++ {
		pos := cn.generatePositionInRadius(cn.DopaminergicMaxRadius)
		n := neuron.NewNeuron(cn.getNextID(), pos, neuron.DopaminergicNeuron, neuron.DefaultFiringThreshold)
		cn.Neurons = append(cn.Neurons, n)
	}

	// Create Inhibitory neurons
	for i := 0; i < numInhibitory; i++ {
		pos := cn.generatePositionInRadius(cn.InhibitoryMaxRadius)
		n := neuron.NewNeuron(cn.getNextID(), pos, neuron.InhibitoryNeuron, neuron.DefaultFiringThreshold)
		cn.Neurons = append(cn.Neurons, n)
	}

	// Create Excitatory neurons
	for i := 0; i < numExcitatory; i++ {
		pos := cn.generatePositionInRadius(cn.ExcitatoryMaxRadius)
		n := neuron.NewNeuron(cn.getNextID(), pos, neuron.ExcitatoryNeuron, neuron.DefaultFiringThreshold)
		cn.Neurons = append(cn.Neurons, n)
	}
	// fmt.Printf("Successfully initialized %d neurons.\n", len(cn.Neurons)) // Moved to main
}

func (cn *CrowNet) generatePositionInRadius(maxRadius float64) neuron.Point {
	for {
		var p neuron.Point
		for i := 0; i < 16; i++ {
			p[i] = (rand.Float64()*2*maxRadius - maxRadius)
		}
		var distSq float64
		for i := 0; i < 16; i++ {
			distSq += p[i] * p[i]
		}
		if distSq <= maxRadius*maxRadius {
			return p
		}
	}
}

// RunCycle simulates one cycle of the network.
func (cn *CrowNet) RunCycle() {
	// fmt.Printf("Cycle %d starting. Cortisol: %.2f, Dopamine: %.2f. Active pulses: %d\n", cn.CycleCount, cn.CortisolLevel, cn.DopamineLevel, len(cn.ActivePulses))
	newlyFiredPulses := make([]*Pulse, 0)

	// 0. Process Inputs (fire input neurons based on frequency)
	cn.processInputs()

	// 1. Update neuron states and decay accumulated pulses
	for _, n := range cn.Neurons {
		n.DecayPulseAccumulation()   // Decay first
		n.UpdateState(cn.CycleCount) // Then update refractory state, etc.
	}

	// 2. Process active pulses
	remainingActivePulses := make([]*Pulse, 0, len(cn.ActivePulses))
	for _, p := range cn.ActivePulses {
		if !p.Propagate() { // Advance pulse, check if it's still active
			//  fmt.Printf("Pulse from %d (orig cycle %d) dissipated at dist %.2f.\n", p.EmittingNeuronID, p.CreationCycle, p.CurrentDistance)
			continue // Pulse dissipated or out of range
		}
		remainingActivePulses = append(remainingActivePulses, p)

		pulseEffectStartDist, pulseEffectEndDist := p.GetEffectRangeForCycle()
		// fmt.Printf("Pulse from %d (val %.2f) effect range: [%.2f, %.2f)\n", p.EmittingNeuronID, p.Value, pulseEffectStartDist, pulseEffectEndDist)

		// Find neurons in the pulse's effect shell for this cycle
		for _, targetNeuron := range cn.Neurons {
			if targetNeuron.ID == p.EmittingNeuronID {
				continue // Neuron cannot stimulate itself with its own pulse this way
			}
			if targetNeuron.State == neuron.AbsoluteRefractoryState {
				continue // Cannot be affected by pulses in absolute refractory state
			}

			distanceToTarget := utils.EuclideanDistance(p.OriginPosition, targetNeuron.Position)

			if distanceToTarget >= pulseEffectStartDist && distanceToTarget < pulseEffectEndDist {
				// This neuron is hit by the pulse in this cycle
				// fmt.Printf("Neuron %d (state %d, acc %.2f) hit by pulse from %d (value: %.2f). Dist: %.2f. Range: [%.2f, %.2f)\n", targetNeuron.ID, targetNeuron.State, targetNeuron.AccumulatedPulse, p.EmittingNeuronID, p.Value, distanceToTarget, pulseEffectStartDist, pulseEffectEndDist)

				// Special handling for dopamine if this were a dopamine "pulse"
				// For now, dopaminergic neurons' "firing" will be handled in chemical modulation step.
				// Here we only process standard excitatory/inhibitory pulses.
				// p.Value is the base signal sign (+1.0 or -1.0) from the emitting neuron
				if p.Value == 0 {
					continue
				}

				weight := cn.GetWeight(p.EmittingNeuronID, targetNeuron.ID)
				effectiveValue := p.Value * weight

				if math.Abs(effectiveValue) < 1e-6 { // If weight is effectively zero, no effect
					continue
				}

				if targetNeuron.ReceivePulse(effectiveValue, cn.CycleCount) {
					// Target neuron fired!
					// fmt.Printf("Neuron %d Fired! (type %d, threshold: %.2f, accumulated: %.2f)\n", targetNeuron.ID, targetNeuron.Type, targetNeuron.CurrentFiringThreshold, targetNeuron.AccumulatedPulse)

					if targetNeuron.Type == neuron.OutputNeuron {
						cn.recordOutputFiring(targetNeuron.ID, cn.CycleCount)
					}

					baseSignalForNewPulse := targetNeuron.GetBasePulseSign()
					if baseSignalForNewPulse != 0 { // Only create pulses that have a non-zero base signal type
						newP := NewPulse(targetNeuron.ID, targetNeuron.Position, baseSignalForNewPulse, cn.CycleCount, cn.PulseMaxTravelDistance)
						newlyFiredPulses = append(newlyFiredPulses, newP)
						// fmt.Printf("New pulse created from neuron %d (base signal: %.2f) at cycle %d\n", newP.EmittingNeuronID, newP.Value, cn.CycleCount)
					}
				}
			}
		}
	}
	cn.ActivePulses = remainingActivePulses
	cn.ActivePulses = append(cn.ActivePulses, newlyFiredPulses...) // Append all newly created pulses

	// TODO: Implement the README's 10-step pulse propagation if the current model is too simple.

	if cn.EnableChemicalModulation {
		// 3. Update Chemical Levels (Cortisol, Dopamine)
		cn.updateCortisolLevel()
		cn.updateDopamineLevel()
		cn.applyCortisolEffects()
		cn.applyDopamineEffects()
	} else {
		// Ensure synaptogenesis factor is reset if chemicals are off
		cn.currentSynaptogenesisModulationFactor = 1.0
		// Ensure neuron thresholds are at base if chemicals are off
		for _, n := range cn.Neurons {
			n.CurrentFiringThreshold = n.BaseFiringThreshold
		}
	}

	// 6. Apply Hebbian Plasticity (neuromodulated)
	// This should happen after neuron states (like LastFiredCycle) and chemical levels for the current cycle are determined.
	cn.ApplyHebbianPlasticity() // Assumes this method exists and uses current net.CortisolLevel, net.DopamineLevel

	if cn.EnableSynaptogenesis {
		// 7. Apply Synaptogenesis (now modulated by chemicals if they are enabled)
		cn.applySynaptogenesis()
	}

	// 8. Handle Input/Output (already handled: input by processInputs, output by recordOutputFiring)
	//    The GetOutputFrequency method is for external querying.

	// Removed temporary firing of input neuron 0, now handled by processInputs via SetInputFrequency.

	cn.CycleCount++
}

// GetSynaptogenesisModulationFactor returns the combined modulation factor for synaptogenesis
// based on current chemical levels (cortisol, dopamine, etc.).
func (cn *CrowNet) GetSynaptogenesisModulationFactor() float64 {
	return cn.currentSynaptogenesisModulationFactor
}
