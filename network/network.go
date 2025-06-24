package network

import (
	"crownet/common"
	"crownet/config"
	"crownet/neuron"
	"crownet/neurochemical"
	"crownet/pulse"
	"crownet/space"
	"crownet/synaptic"
	"fmt"
	"math"
	"math/rand"
	"sort"
)

const minEffectiveLearningRateThreshold = 1e-9

// _isNeuronRecentlyActive checks if a neuron has fired within the given coincidence window relative to the current cycle.
// It accesses CrowNet's CycleCount via the receiver.
func (cn *CrowNet) _isNeuronRecentlyActive(n *neuron.Neuron, coincidenceWindow common.CycleCount) bool {
	if n == nil || n.LastFiredCycle == -1 {
		return false
	}
	// Neuron's LastFiredCycle is the cycle number it fired in.
	// cn.CycleCount is the current cycle number (about to end).
	// If neuron fired in cycle X, and current cycle is Y, age is Y-X.
	// If this age is <= window, it's recent.
	return (cn.CycleCount - n.LastFiredCycle) <= coincidenceWindow
}

type CrowNet struct {
	// --- Core Simulation Parameters & Random Number Generation ---
	SimParams        *config.SimulationParameters // Pointer to shared simulation parameters
	baseLearningRate common.Rate                  // Base learning rate, modulated by chemical environment
	rng              *rand.Rand

	Neurons            []*neuron.Neuron
	// --- Core Simulation Parameters & Random Number Generation ---
	SimParams        *config.SimulationParameters // Pointer to shared simulation parameters
	baseLearningRate common.Rate                  // Base learning rate, modulated by chemical environment
	rng              *rand.Rand                   // Local random number generator for this network instance

	// --- Neuron Collections and Management ---
	Neurons           []*neuron.Neuron            // Slice of all neuron instances in the network
	InputNeuronIDs    []common.NeuronID           // Cache of IDs for input neurons, sorted
	OutputNeuronIDs   []common.NeuronID           // Cache of IDs for output neurons, sorted
	OutputNeuronIDSet map[common.NeuronID]struct{} // Set of output neuron IDs for efficient lookup
	InputNeuronIDSet  map[common.NeuronID]struct{} // Set of input neuron IDs for efficient lookup
	neuronMap         map[common.NeuronID]*neuron.Neuron // Map for O(1) neuron lookup by ID
	neuronIDCounter   common.NeuronID             // Counter for generating unique neuron IDs

	// --- Dynamic Sub-systems of the Network ---
	ActivePulses    *pulse.PulseList         // Manages active pulses propagating through the network
	SynapticWeights synaptic.NetworkWeights  // Manages the matrix of synaptic weights between neurons
	ChemicalEnv     *neurochemical.Environment // Manages the neurochemical environment (e.g., cortisol, dopamine levels and effects)

	// --- Simulation State ---
	CycleCount common.CycleCount // Current simulation cycle

	// Feature-specific state
	// For 'sim' mode continuous frequency input
	inputTargetFrequencies map[common.NeuronID]float64
	timeToNextInputFire    map[common.NeuronID]common.CycleCount
	// For calculating output neuron firing frequency
	outputFiringHistory    map[common.NeuronID][]common.CycleCount

	// Dynamic process toggles
	isLearningEnabled           bool // If true, Hebbian learning rule is applied
	isSynaptogenesisEnabled     bool // If true, neuron movement (synaptogenesis) occurs
	isChemicalModulationEnabled bool // If true, neurochemicals modulate learning, synaptogenesis, and firing thresholds
}

// NewCrowNet creates and initializes a new CrowNet instance.
// It now takes an AppConfig to centralize configuration sourcing and returns an error if initialization fails.
func NewCrowNet(appCfg *config.AppConfig) (*CrowNet, error) {
	if appCfg == nil {
		return nil, fmt.Errorf("NewCrowNet: appConfig cannot be nil")
	}
	simParams := &appCfg.SimParams // Use a pointer to SimParams

	localRng := rand.New(rand.NewSource(appCfg.Cli.Seed))
	net := &CrowNet{
		// Simulation parameters and core components
		SimParams:        simParams,
		baseLearningRate: appCfg.Cli.BaseLearningRate,
		rng:              localRng,

		// Neuron collections and management
		Neurons:           make([]*neuron.Neuron, 0, appCfg.Cli.TotalNeurons),
		InputNeuronIDs:    make([]common.NeuronID, 0, simParams.MinInputNeurons),
		OutputNeuronIDs:   make([]common.NeuronID, 0, simParams.MinOutputNeurons),
		OutputNeuronIDSet: make(map[common.NeuronID]struct{}),
		InputNeuronIDSet:  make(map[common.NeuronID]struct{}),
		neuronMap:         make(map[common.NeuronID]*neuron.Neuron),
		neuronIDCounter:   0,

		// Dynamic sub-systems
		ActivePulses:    pulse.NewPulseList(),
		SynapticWeights: synaptic.NewNetworkWeights(),
		ChemicalEnv:     neurochemical.NewEnvironment(), // Note: CortisolGlandPosition is now in SimParams.
		                                                // ChemicalEnv.UpdateLevels will need to access it via SimParams.
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
	}

	// Initialize neurons - this might return an error
	if err := net.initializeNeurons(appCfg.Cli.TotalNeurons); err != nil {
		return nil, fmt.Errorf("failed to initialize neurons: %w", err)
	}
	allNeuronIDs := make([]common.NeuronID, len(net.Neurons))
	for i, n := range net.Neurons {
		allNeuronIDs[i] = n.ID
	}
	// InitializeAllToAllWeights uses rng and SimParams.
	// Consider if InitializeAllToAllWeights could also return an error. For now, assuming it doesn't.
	net.SynapticWeights.InitializeAllToAllWeights(allNeuronIDs, net.SimParams, net.rng)

	net.finalizeInitialization()

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
		effectiveRadius := radiusFactor * cn.SimParams.SpaceMaxDimension
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
// It returns an error if the final neuron count does not match the expected count.
// initializeNeurons populates the network with neurons of different types based on configuration.
// It ensures the minimum number of input and output neurons are created and distributes
// the remaining neurons among internal types (excitatory, inhibitory, dopaminergic).
// Assumes totalNeuronsInput has been validated by config.Validate() to be >= MinInputNeurons and MinOutputNeurons.
// It returns an error if the final neuron count does not match the expected count, indicating an internal logic issue.
func (cn *CrowNet) initializeNeurons(totalNeuronsInput int) error {
	simParams := cn.SimParams
	actualTotalNeurons := totalNeuronsInput // totalNeuronsInput is pre-validated by config.Validate()
	numInput := simParams.MinInputNeurons
	numOutput := simParams.MinOutputNeurons

	// Note: The check for actualTotalNeurons < numInput+numOutput and associated adjustment/warning
	// has been moved to config.Validate() to handle configuration errors upfront.
	// initializeNeurons now assumes totalNeuronsInput is sufficient.

	cn.addNeuronsOfType(numInput, neuron.Input, simParams.ExcitatoryRadiusFactor)
	cn.addNeuronsOfType(numOutput, neuron.Output, simParams.ExcitatoryRadiusFactor)

	remainingForInternalDistribution := actualTotalNeurons - numInput - numOutput
	numDopaminergic, numInhibitory, numExcitatory := calculateInternalNeuronCounts(
		remainingForInternalDistribution,
		simParams.DopaminergicPercent,
		simParams.InhibitoryPercent,
	)

	cn.addNeuronsOfType(numDopaminergic, neuron.Dopaminergic, simParams.DopaminergicRadiusFactor)
	cn.addNeuronsOfType(numInhibitory, neuron.Inhibitory, simParams.InhibitoryRadiusFactor)
	cn.addNeuronsOfType(numExcitatory, neuron.Excitatory, simParams.ExcitatoryRadiusFactor)

	if len(cn.Neurons) != actualTotalNeurons {
		// This is a critical failure in setup.
		return fmt.Errorf("critical alert: final neuron count (%d) does not match expected (%d) in initializeNeurons", len(cn.Neurons), actualTotalNeurons)
	}
	return nil
}

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

func (cn *CrowNet) SetDynamicState(learning, synaptogenesis, chemicalModulation bool) {
	cn.isLearningEnabled = learning
	cn.isSynaptogenesisEnabled = synaptogenesis
	cn.isChemicalModulationEnabled = chemicalModulation
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
		// Note: CortisolGlandPosition is accessed from SimParams.
		cn.ChemicalEnv.UpdateLevels(cn.Neurons, cn.ActivePulses.GetAll(), cn.CortisolGlandPosition, cn.SimParams)
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

func (cn *CrowNet) RunCycle() {
	cn.processFrequencyInputs()         // Step 1: Process any continuous/frequency-based inputs
	cn._updateAllNeuronStates()         // Step 2: Update base states of all neurons

	cn.processActivePulses()            // Step 3: Process pulse propagation and effects

	cn._applyChemicalModulationEffects() // Step 4: Apply or reset neurochemical effects

	if cn.isLearningEnabled {           // Step 5: Apply Hebbian learning if enabled
		cn.applyHebbianLearning()
	}
	if cn.isSynaptogenesisEnabled {     // Step 6: Apply Synaptogenesis (neuron movement) if enabled
		cn.applySynaptogenesis()
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
	newlyGeneratedPulses := cn.ActivePulses.ProcessCycle(
		cn.Neurons,
		cn.SynapticWeights,
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

	coincidenceWindow := common.CycleCount(cn.SimParams.HebbianCoincidenceWindow)

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

func (cn *CrowNet) processFrequencyInputs() {
	for neuronID, timeLeft := range cn.timeToNextInputFire {
		newTimeLeft := timeLeft - 1
		cn.timeToNextInputFire[neuronID] = newTimeLeft
		if newTimeLeft <= 0 {
			var inputNeuron *neuron.Neuron
			for _, n := range cn.Neurons {
				if n.ID == neuronID && n.Type == neuron.Input {
					inputNeuron = n
					break
				}
			}
			if inputNeuron != nil {
				inputNeuron.CurrentState = neuron.Firing
				emittedSignal := inputNeuron.EmittedPulseSign()
				if emittedSignal != 0 {
					newP := pulse.New(
						inputNeuron.ID,
						inputNeuron.Position,
						emittedSignal,
						cn.CycleCount,
						cn.SimParams.SpaceMaxDimension*2.0,
					)
					cn.ActivePulses.Add(newP)
				}
			}
			targetHz := cn.inputTargetFrequencies[neuronID]
			if targetHz > 0 {
				cyclesPerFiring := cn.SimParams.CyclesPerSecond / targetHz
				cn.timeToNextInputFire[neuronID] = common.CycleCount(math.Max(1.0, math.Round(cn.rng.Float64()*cyclesPerFiring)+1))
			} else {
				delete(cn.timeToNextInputFire, neuronID)
			}
		}
	}
}

func (cn *CrowNet) recordOutputFiring(neuronID common.NeuronID) {
	isOutput := false
	for _, id := range cn.OutputNeuronIDs {
		if id == neuronID {
			isOutput = true
			break
		}
	}
	if !isOutput {
		return
	}
	history, exists := cn.outputFiringHistory[neuronID]
	if !exists {
		history = make([]common.CycleCount, 0)
	}
	history = append(history, cn.CycleCount)
	cutoffCycle := cn.CycleCount - common.CycleCount(cn.SimParams.OutputFrequencyWindowCycles)
	prunedHistory := make([]common.CycleCount, 0, len(history))
	for _, fireCycle := range history {
		if fireCycle >= cutoffCycle {
			prunedHistory = append(prunedHistory, fireCycle)
		}
	}
	cn.outputFiringHistory[neuronID] = prunedHistory
}
