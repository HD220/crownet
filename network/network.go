// Package network provides the core simulation engine for the CrowNet neural network.
// It defines the CrowNet struct, which orchestrates neuron interactions, pulse propagation,
// learning, synaptogenesis, and neurochemical modulation.
package network

import (
	"crownet/common"
	"crownet/config"
	"crownet/neuron"
	"crownet/neurochemical"
	"crownet/pulse"
	"crownet/space"
	// "crownet/space/grid" // If grid becomes its own sub-package. For now, space.SpatialGrid
	"crownet/synaptic"
	"fmt"
	"math"
	"math/rand"
	"sort"
)

// minEffectiveLearningRateThreshold defines the minimum threshold below which the
// effective learning rate is considered negligible, potentially skipping
// computationally intensive learning processes.
const minEffectiveLearningRateThreshold = 1e-9

// isNeuronRecentlyActive checks if a neuron has fired within the given coincidence window
// relative to the current cycle.
// It accesses CrowNet's CycleCount via the receiver.
func (cn *CrowNet) isNeuronRecentlyActive(n *neuron.Neuron, coincidenceWindow common.CycleCount) bool {
	if n == nil || n.LastFiredCycle == -1 {
		return false
	}
	// Neuron's LastFiredCycle is the cycle number it fired in.
	// cn.CycleCount is the current cycle number (about to end).
	// If neuron fired in cycle X, and current cycle is Y, age is Y-X.
	// If this age is <= window, it's recent.
	return (cn.CycleCount - n.LastFiredCycle) <= coincidenceWindow
}

// CrowNet represents the entire neural network simulation environment.
// It encapsulates all neurons, their connections (synaptic weights),
// active pulses, the neurochemical environment, spatial indexing,
// simulation parameters, and the overall simulation state.
type CrowNet struct {
	// SimParams holds the shared simulation parameters for the network.
	SimParams *config.SimulationParameters
	// baseLearningRate is the initial learning rate, which can be modulated by the chemical environment.
	baseLearningRate common.Rate
	// rng is a local random number generator for this network instance, ensuring deterministic behavior if seeded.
	rng *rand.Rand

	// Neurons is a slice containing all neuron instances in the network.
	Neurons []*neuron.Neuron
	// InputNeuronIDs caches the IDs of input neurons, sorted for consistent access.
	InputNeuronIDs []common.NeuronID
	// OutputNeuronIDs caches the IDs of output neurons, sorted for consistent access.
	OutputNeuronIDs []common.NeuronID
	// OutputNeuronIDSet provides efficient O(1) lookup for output neuron IDs.
	OutputNeuronIDSet map[common.NeuronID]struct{}
	// InputNeuronIDSet provides efficient O(1) lookup for input neuron IDs.
	InputNeuronIDSet map[common.NeuronID]struct{}
	// neuronMap allows for O(1) retrieval of a neuron instance by its ID.
	neuronMap map[common.NeuronID]*neuron.Neuron
	// neuronIDCounter is used to generate unique IDs for new neurons.
	neuronIDCounter common.NeuronID

	// ActivePulses manages all pulses currently propagating through the network.
	ActivePulses *pulse.PulseList
	// SynapticWeights manages the matrix of synaptic weights between neurons.
	SynapticWeights *synaptic.NetworkWeights
	// ChemicalEnv manages the neurochemical environment (e.g., cortisol, dopamine) and their effects.
	ChemicalEnv *neurochemical.Environment
	// SpatialGrid provides a spatial index for neurons, optimizing proximity queries.
	SpatialGrid *space.SpatialGrid

	// CycleCount tracks the current simulation cycle number.
	CycleCount common.CycleCount

	// inputTargetFrequencies stores target firing frequencies for input neurons under continuous stimulus.
	inputTargetFrequencies map[common.NeuronID]float64
	// timeToNextInputFire tracks remaining cycles until the next scheduled firing for frequency-stimulated inputs.
	timeToNextInputFire map[common.NeuronID]common.CycleCount
	// outputFiringHistory records recent firing times for output neurons, used for frequency calculation.
	outputFiringHistory map[common.NeuronID][]common.CycleCount

	// isLearningEnabled, if true, allows Hebbian learning rules to be applied.
	isLearningEnabled bool
	// isSynaptogenesisEnabled, if true, allows neuron movement and structural plasticity.
	isSynaptogenesisEnabled bool
	// isChemicalModulationEnabled, if true, allows neurochemicals to modulate network behavior.
	isChemicalModulationEnabled bool

	// SynaptogenesisStrategy components
	SynaptogenesisForceCalculator ForceCalculator // REFACTOR-006
	SynaptogenesisMovementUpdater MovementUpdater // REFACTOR-006
}

// NewCrowNet creates and initializes a new CrowNet simulation environment.
// It sets up the network structure, including neurons, synaptic weights,
// the chemical environment, and spatial indexing, based on the provided AppConfig.
//
// Parameters:
//   appCfg: An *config.AppConfig containing all necessary configuration parameters
//           (simulation settings and CLI options).
//
// Returns:
//   A pointer to the newly created CrowNet instance, or an error if initialization
//   of any core component (e.g., spatial grid, synaptic weights, neurons) fails.
func NewCrowNet(appCfg *config.AppConfig) (*CrowNet, error) {
	if appCfg == nil {
		return nil, fmt.Errorf("NewCrowNet: appConfig cannot be nil")
	}
	simParams := &appCfg.SimParams // Use a pointer to SimParams

	localRng := rand.New(rand.NewSource(appCfg.Cli.Seed))

	// Spatial Grid Initialization
	// Define the grid origin (min corner of the simulation space)
	var gridMinBound common.Point
	for i := 0; i < common.PointDimension; i++ {
		gridMinBound[i] = common.Coordinate(-simParams.General.SpaceMaxDimension)
	}
	// Determine cell size, e.g., based on pulse propagation speed or a fraction of space size.
	// Using a factor of pulse speed for now. Ensure it's not zero.
	const defaultGridCellSizeMultiplier = 2.0
	cellSize := simParams.General.PulsePropagationSpeed * defaultGridCellSizeMultiplier
	if cellSize < 1e-6 { // Avoid zero or too small cell size
		cellSize = simParams.General.SpaceMaxDimension / 10.0 // Fallback to a fraction of space dimension
		if cellSize < 1e-6 {
			cellSize = 1.0 // Absolute fallback
		}
	}
	spatialGridInstance, err := space.NewSpatialGrid(cellSize, common.PointDimension, gridMinBound)
	if err != nil {
		return nil, fmt.Errorf("failed to create spatial grid: %w", err)
	}

	net := &CrowNet{
		// Simulation parameters and core components
		SimParams:        simParams,
		baseLearningRate: appCfg.Cli.BaseLearningRate,
		rng:              localRng,

		// Neuron collections and management
		Neurons:           make([]*neuron.Neuron, 0, appCfg.Cli.TotalNeurons),
		InputNeuronIDs:    make([]common.NeuronID, 0, simParams.Structure.MinInputNeurons),
		OutputNeuronIDs:   make([]common.NeuronID, 0, simParams.Structure.MinOutputNeurons),
		OutputNeuronIDSet: make(map[common.NeuronID]struct{}),
		InputNeuronIDSet:  make(map[common.NeuronID]struct{}),
		neuronMap:         make(map[common.NeuronID]*neuron.Neuron),
		neuronIDCounter:   0,

		// Dynamic sub-systems
		ActivePulses:    pulse.NewPulseList(),
		SynapticWeights: synaptic.NewNetworkWeights(), // Assuming constructor doesn't fail or is handled if it can
		ChemicalEnv:     neurochemical.NewEnvironment(),
		SpatialGrid:     spatialGridInstance,

		// Simulation state
		CycleCount:      0,

		// Feature-specific state
		inputTargetFrequencies: make(map[common.NeuronID]float64),
		timeToNextInputFire:    make(map[common.NeuronID]common.CycleCount),
		outputFiringHistory:    make(map[common.NeuronID][]common.CycleCount),

		// Dynamic process toggles (default to true)
		isLearningEnabled:      true,
		isSynaptogenesisEnabled: true,
		isChemicalModulationEnabled: true,

		// REFACTOR-006: Initialize synaptogenesis strategy components
		// Placeholder for actual default implementation structs/constructors
		SynaptogenesisForceCalculator:   &DefaultForceCalculator{},
		SynaptogenesisMovementUpdater: &DefaultMovementUpdater{},
	}
	// Ensure SynapticWeights is properly initialized (it was missing its params before)
	// This assumes NewNetworkWeights might also return an error or needs simParams & rng
	synapticWeightsInstance, err := synaptic.NewNetworkWeights(simParams, localRng)
	if err != nil {
		return nil, fmt.Errorf("failed to create synaptic weights: %w", err)
	}
	net.SynapticWeights = synapticWeightsInstance


	// Initialize neurons - this might return an error
	if err := net.initializeNeurons(appCfg.Cli.TotalNeurons); err != nil {
		return nil, fmt.Errorf("failed to initialize neurons: %w", err)
	}
	allNeuronIDs := make([]common.NeuronID, len(net.Neurons))
	for i, n := range net.Neurons {
		allNeuronIDs[i] = n.ID
	}

	net.SynapticWeights.InitializeAllToAllWeights(allNeuronIDs) // Now uses internal simParams and rng

	net.finalizeInitialization() // Populates neuronMap, ID sets

	// Initial build of the spatial grid after neurons are positioned
	net.SpatialGrid.Build(net.Neurons)


	return net, nil
}

// getNextNeuronID returns the next available unique ID for a new neuron and increments the internal counter.
func (cn *CrowNet) getNextNeuronID() common.NeuronID {
	id := cn.neuronIDCounter
	cn.neuronIDCounter++
	return id
}

// addNeuronsOfType creates and adds a specified number of neurons of a given type to the network.
// Neurons are positioned randomly within a sphere whose radius is determined by radiusFactor * SimParams.SpaceMaxDimension.
// It also updates the InputNeuronIDs or OutputNeuronIDs slices if applicable.
func (cn *CrowNet) addNeuronsOfType(count int, neuronType neuron.Type, radiusFactor float64) {
	if count <= 0 {
		return
	}
	// Ensure SimParams is not nil to prevent panic, though it should always be set in a valid CrowNet.
	if cn.SimParams == nil {
		// This case should ideally not be reached if CrowNet is constructed properly.
		// Consider logging a critical error or panicking if SimParams is unexpectedly nil.
		return
	}
	for i := 0; i < count; i++ {
		id := cn.getNextNeuronID()
		effectiveRadius := radiusFactor * cn.SimParams.General.SpaceMaxDimension
		pos := space.GenerateRandomPositionInHyperSphere(effectiveRadius, cn.rng) // cn.rng is *rand.Rand

		n := neuron.New(id, neuronType, pos, cn.SimParams)
		cn.Neurons = append(cn.Neurons, n)

		if neuronType == neuron.Input {
			cn.InputNeuronIDs = append(cn.InputNeuronIDs, id)
		} else if neuronType == neuron.Output {
			cn.OutputNeuronIDs = append(cn.OutputNeuronIDs, id)
		}
	}
}

// calculateInternalNeuronCounts determines the number of dopaminergic, inhibitory, and
// excitatory neurons to create based on the total remaining neurons to be distributed and
// the configured percentages for dopaminergic and inhibitory types. Excitatory neurons
// make up the remainder.
func calculateInternalNeuronCounts(remainingForDistribution int, dopaP, inhibP float64) (numDopaminergic, numInhibitory, numExcitatory int) {
	if remainingForDistribution <= 0 {
		return 0, 0, 0
	}
	numDopaminergic = int(math.Floor(float64(remainingForDistribution) * dopaP))
	numInhibitory = int(math.Floor(float64(remainingForDistribution) * inhibP))
	currentAllocated := numDopaminergic + numInhibitory
	numExcitatory = remainingForDistribution - currentAllocated
	return
}

// initializeNeurons populates the network with neurons of different types based on configuration.
// It ensures the minimum number of input and output neurons are created and distributes
// the remaining neurons among internal types (excitatory, inhibitory, dopaminergic).
// Assumes totalNeuronsInput has been validated by config.Validate() to be >= MinInputNeurons and MinOutputNeurons.
// It returns an error if the final neuron count does not match the expected count, indicating an internal logic issue.
func (cn *CrowNet) initializeNeurons(totalNeuronsInput int) error {
	simParams := cn.SimParams // This is *config.SimulationParameters
	actualTotalNeurons := totalNeuronsInput // totalNeuronsInput is pre-validated by config.Validate()
	numInput := simParams.Structure.MinInputNeurons
	numOutput := simParams.Structure.MinOutputNeurons

	// Note: The check for actualTotalNeurons < numInput+numOutput and associated adjustment/warning
	// has been moved to config.Validate() to handle configuration errors upfront.
	// initializeNeurons now assumes totalNeuronsInput is sufficient.

	cn.addNeuronsOfType(numInput, neuron.Input, simParams.Distribution.ExcitatoryRadiusFactor)
	cn.addNeuronsOfType(numOutput, neuron.Output, simParams.Distribution.ExcitatoryRadiusFactor)

	remainingForInternalDistribution := actualTotalNeurons - numInput - numOutput
	numDopaminergic, numInhibitory, numExcitatory := calculateInternalNeuronCounts(
		remainingForInternalDistribution,
		simParams.Distribution.DopaminergicPercent,
		simParams.Distribution.InhibitoryPercent,
	)

	cn.addNeuronsOfType(numDopaminergic, neuron.Dopaminergic, simParams.Distribution.DopaminergicRadiusFactor)
	cn.addNeuronsOfType(numInhibitory, neuron.Inhibitory, simParams.Distribution.InhibitoryRadiusFactor)
	cn.addNeuronsOfType(numExcitatory, neuron.Excitatory, simParams.Distribution.ExcitatoryRadiusFactor)

	if len(cn.Neurons) != actualTotalNeurons {
		// This is a critical failure in setup.
		return fmt.Errorf("critical alert: final neuron count (%d) does not match expected (%d) in initializeNeurons", len(cn.Neurons), actualTotalNeurons)
	}
	return nil
}

// finalizeInitialization completes the setup of the CrowNet instance after neurons
// have been created. It sorts InputNeuronIDs and OutputNeuronIDs for consistent ordering,
// populates lookup maps (neuronMap, InputNeuronIDSet, OutputNeuronIDSet), and
// initializes data structures like outputFiringHistory.
func (cn *CrowNet) finalizeInitialization() {
	sort.Slice(cn.InputNeuronIDs, func(i, j int) bool { return cn.InputNeuronIDs[i] < cn.InputNeuronIDs[j] })
	sort.Slice(cn.OutputNeuronIDs, func(i, j int) bool { return cn.OutputNeuronIDs[i] < cn.OutputNeuronIDs[j] })

	// Populate OutputNeuronIDSet and initialize outputFiringHistory
	cn.OutputNeuronIDSet = make(map[common.NeuronID]struct{}, len(cn.OutputNeuronIDs))
	for _, outID := range cn.OutputNeuronIDs {
		cn.OutputNeuronIDSet[outID] = struct{}{}
		cn.outputFiringHistory[outID] = make([]common.CycleCount, 0)
	}

	// Populate InputNeuronIDSet
	cn.InputNeuronIDSet = make(map[common.NeuronID]struct{}, len(cn.InputNeuronIDs))
	for _, inID := range cn.InputNeuronIDs {
		cn.InputNeuronIDSet[inID] = struct{}{}
	}

	// Populate neuronMap
	cn.neuronMap = make(map[common.NeuronID]*neuron.Neuron, len(cn.Neurons))
	for _, n := range cn.Neurons {
		cn.neuronMap[n.ID] = n
	}
}

// _updateAllNeuronStates handles the decay of accumulated potential and state advancement for all neurons.
func (cn *CrowNet) _updateAllNeuronStates() {
	for _, n := range cn.Neurons {
		n.DecayPotential(cn.SimParams)
		n.AdvanceState(cn.CycleCount, cn.SimParams)
	}
}

// _applyChemicalModulationEffects updates chemical levels and applies their effects to neurons
// if chemical modulation is enabled. Otherwise, it resets modulation factors and neuron thresholds.
func (cn *CrowNet) _applyChemicalModulationEffects() {
	if cn.isChemicalModulationEnabled {
		// Pass the specific CortisolGlandPosition from the nested SimParams struct
		cn.ChemicalEnv.UpdateLevels(cn.Neurons, cn.ActivePulses.GetAll(), cn.SimParams.Neurochemical.CortisolGlandPosition, cn.SimParams)
		cn.ChemicalEnv.ApplyEffectsToNeurons(cn.Neurons, cn.SimParams)
	} else {
		// Reset modulation factors and thresholds if chemical modulation is off
		cn.ChemicalEnv.LearningRateModulationFactor = 1.0
		cn.ChemicalEnv.SynaptogenesisModulationFactor = 1.0
		for _, n := range cn.Neurons {
			n.CurrentFiringThreshold = n.BaseFiringThreshold
		}
	}
}

// RunCycle executes a single simulation cycle of the CrowNet.
// It orchestrates the various phases of the simulation including:
// 1. Processing continuous inputs (if any).
// 2. Updating base neuron states (potential decay, refractory periods).
// 3. Processing active pulse propagation and their effects on neurons (using the spatial grid for optimization).
// 4. Updating and applying neurochemical effects.
// 5. Applying Hebbian learning.
// 6. Applying synaptogenesis (neuron movement), which also triggers a rebuild of the spatial grid if enabled.
// 7. Incrementing the simulation cycle count.
func (cn *CrowNet) RunCycle() {
	cn.processFrequencyInputs()         // Step 1: Process any continuous/frequency-based inputs
	cn._updateAllNeuronStates()         // Step 2: Update base states of all neurons

	cn.processActivePulses()            // Step 3: Process pulse propagation and effects (uses SpatialGrid)

	cn._applyChemicalModulationEffects() // Step 4: Apply or reset neurochemical effects

	if cn.isLearningEnabled {           // Step 5: Apply Hebbian learning if enabled
		cn.applyHebbianLearning()
	}
	if cn.isSynaptogenesisEnabled {     // Step 6: Apply Synaptogenesis (neuron movement) if enabled
		cn.applySynaptogenesis()
		// Rebuild spatial grid if neurons moved
		cn.SpatialGrid.Build(cn.Neurons)
	}
	cn.CycleCount++                     // Step 7: Increment cycle count
}

// processActivePulses orchestrates the processing of existing pulses and handles newly generated ones.
// 1. It delegates to PulseList to process the current cycle for existing pulses, which includes:
//    - Moving pulses.
//    - Applying pulse effects to neurons they hit (updating neuron potentials).
//    - Generating new pulses from neurons that fire as a result.
// 2. It then handles these newly generated pulses by:
//    - Checking if any originated from output neurons and recording those firings.
//    - Adding all new pulses to the active list for processing in subsequent cycles.
func (cn *CrowNet) processActivePulses() {
	// Step 1: Process existing pulses and get newly generated ones from firing neurons.
	// Pass the spatialGrid and the pointer to SynapticWeights.
	newlyGeneratedPulses := cn.ActivePulses.ProcessCycle(
		cn.SpatialGrid,
		cn.SynapticWeights, // cn.SynapticWeights is already *synaptic.NetworkWeights
		cn.CycleCount,
		cn.SimParams,
	)

	// Step 2: Handle any pulses that were newly generated in this cycle.
	if len(newlyGeneratedPulses) > 0 {
		// Step 2a: Record firing if an output neuron emitted one of these new pulses.
		for _, newP := range newlyGeneratedPulses {
			if _, isOutputNeuron := cn.OutputNeuronIDSet[newP.EmittingNeuronID]; isOutputNeuron {
				cn.recordOutputFiring(newP.EmittingNeuronID)
				// Note: A single new pulse is assumed to be from one neuron, so no need to continue inner loop once found.
				// If an output neuron fires, its pulse is recorded. We don't 'break' here because other new pulses might also be from other output neurons.
			}
		}
		// Step 2b: Add all newly generated pulses to the main active list for the next cycle.
		cn.ActivePulses.AddAll(newlyGeneratedPulses)
	}
}

// applyHebbianLearning applies Hebbian learning rules to update synaptic weights.
// It strengthens connections between neurons that fire together within a defined time window (coincidenceWindow).
// The learning rate is modulated by the neurochemical environment.
func (cn *CrowNet) applyHebbianLearning() {
	// Calculate effective learning rate, modulated by chemical environment.
	effectiveLR := common.Rate(float64(cn.baseLearningRate) * float64(cn.ChemicalEnv.LearningRateModulationFactor))

	// If learning rate is negligible, skip the computationally intensive process.
	if effectiveLR < minEffectiveLearningRateThreshold {
		return
	}

	coincidenceWindow := common.CycleCount(cn.SimParams.Learning.HebbianCoincidenceWindow)

	// Iterate over all possible presynaptic neurons.
	for _, preSynapticNeuron := range cn.Neurons {
		// Determine if the presynaptic neuron was recently active.
		isPreActive := cn._isNeuronRecentlyActive(preSynapticNeuron, coincidenceWindow)

		if !isPreActive {
			continue // If presynaptic neuron wasn't active, no update based on it.
		}
		preActivityValue := 1.0 // Using a binary activity signal for this Hebbian rule.

		// Iterate over all possible postsynaptic neurons.
		for _, postSynapticNeuron := range cn.Neurons {
			if preSynapticNeuron.ID == postSynapticNeuron.ID { // No self-connections for Hebbian learning.
				continue
			}

			// Determine if the postsynaptic neuron was recently active.
			isPostActive := cn._isNeuronRecentlyActive(postSynapticNeuron, coincidenceWindow)

			if isPostActive { // If postsynaptic neuron was also active (co-activation).
				postActivityValue := 1.0 // Using a binary activity signal.

				// Delegate the actual weight update to SynapticWeights.
				cn.SynapticWeights.ApplyHebbianUpdate(
					preSynapticNeuron.ID,
					postSynapticNeuron.ID,
					preActivityValue,  // Pass 1.0 for active
					postActivityValue, // Pass 1.0 for active
					effectiveLR,
					cn.SimParams,
				)
			}
		}
	}
}
