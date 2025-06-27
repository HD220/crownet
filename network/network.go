// Package network provides the core simulation engine for the CrowNet neural network.
// It defines the CrowNet struct, which orchestrates neuron interactions, pulse propagation,
// learning, synaptogenesis, and neurochemical modulation.
package network

import (
	"crownet/common"
	"crownet/config"
	"crownet/neurochemical"
	"crownet/neuron"
	"crownet/pulse"
	"crownet/space"

	// "crownet/space/grid" // If grid becomes its own sub-package. For now, space.SpatialGrid
	"fmt"
	"math"
	"math/rand"
	"sort"

	"crownet/synaptic"
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
	SynaptogenesisForceCalculator ForceCalculator
	SynaptogenesisMovementUpdater MovementUpdater
	inputTargetFrequencies        map[common.NeuronID]float64
	SpatialGrid                   *space.SpatialGrid
	rng                           *rand.Rand
	outputFiringHistory           map[common.NeuronID][]common.CycleCount
	OutputNeuronIDSet             map[common.NeuronID]struct{}
	InputNeuronIDSet              map[common.NeuronID]struct{}
	neuronMap                     map[common.NeuronID]*neuron.Neuron
	timeToNextInputFire           map[common.NeuronID]common.CycleCount
	ActivePulses                  *pulse.PulseList
	SynapticWeights               *synaptic.NetworkWeights
	ChemicalEnv                   *neurochemical.Environment
	SimParams                     *config.AppConfig // Changed to hold full AppConfig
	Neurons                       []*neuron.Neuron  // TODO: Deprecate in favor of neuronMap access
	OutputNeuronIDs               []common.NeuronID // TODO: Deprecate if OutputNeuronIDSet is sufficient
	InputNeuronIDs                []common.NeuronID // TODO: Deprecate if InputNeuronIDSet is sufficient
	CycleCount                    common.CycleCount
	neuronIDCounter               common.NeuronID
	logger                        Logger // Interface for logging
	isLearningEnabled             bool
	isSynaptogenesisEnabled       bool
	isChemicalModulationEnabled   bool
}

// NewCrowNet creates and initializes a new CrowNet simulation environment.
// It sets up the network structure, including neurons, synaptic weights,
// the chemical environment, and spatial indexing, based on the provided AppConfig.
//
// Parameters:
//
//	appCfg: An *config.AppConfig containing all necessary configuration parameters
//	        (simulation settings and CLI options).
//
// Returns:
//
//	A pointer to the newly created CrowNet instance, or an error if initialization
//	of any core component (e.g., spatial grid, synaptic weights, neurons) fails.
func NewCrowNet(appCfg *config.AppConfig) (*CrowNet, error) {
	if appCfg == nil {
		return nil, fmt.Errorf("NewCrowNet: appConfig cannot be nil")
	}
	// Store the whole AppConfig, not just SimParams, if CLI parameters are needed by CrowNet methods.
	// simParams := &appCfg.SimParams // Use a pointer to SimParams

	localRng := rand.New(rand.NewSource(appCfg.Cli.Seed))

	// Spatial Grid Initialization
	var gridMinBound common.Point
	for i := 0; i < common.PointDimension; i++ {
		gridMinBound[i] = common.Coordinate(-appCfg.SimParams.General.SpaceMaxDimension)
	}
	const defaultGridCellSizeMultiplier = 2.0
	cellSize := float64(appCfg.SimParams.General.PulsePropagationSpeed) * defaultGridCellSizeMultiplier
	if cellSize < 1e-6 {
		cellSize = appCfg.SimParams.General.SpaceMaxDimension / 10.0
		if cellSize < 1e-6 {
			cellSize = 1.0
		}
	}
	// Use common.PointDimension
	spatialGridInstance, err := space.NewSpatialGrid(cellSize, common.PointDimension, gridMinBound)
	if err != nil {
		return nil, fmt.Errorf("failed to create spatial grid: %w", err)
	}

	net := &CrowNet{
		SimParams:         appCfg, // Store entire AppConfig
		rng:               localRng,
		Neurons:           make([]*neuron.Neuron, 0, appCfg.Cli.TotalNeurons),
		InputNeuronIDs:    make([]common.NeuronID, 0, appCfg.SimParams.Structure.MinInputNeurons),
		OutputNeuronIDs:   make([]common.NeuronID, 0, appCfg.SimParams.Structure.MinOutputNeurons),
		OutputNeuronIDSet: make(map[common.NeuronID]struct{}),
		InputNeuronIDSet:  make(map[common.NeuronID]struct{}),
		neuronMap:         make(map[common.NeuronID]*neuron.Neuron),
		neuronIDCounter:   0,
		ActivePulses:      pulse.NewPulseList(),
		ChemicalEnv:       neurochemical.NewEnvironment(),
		SpatialGrid:       spatialGridInstance,
		CycleCount:        0,
		inputTargetFrequencies: make(map[common.NeuronID]float64),
		timeToNextInputFire:    make(map[common.NeuronID]common.CycleCount),
		outputFiringHistory:    make(map[common.NeuronID][]common.CycleCount),
		isLearningEnabled:           true,
		isSynaptogenesisEnabled:     true,
		isChemicalModulationEnabled: true,
		SynaptogenesisForceCalculator: &DefaultForceCalculator{},
		SynaptogenesisMovementUpdater: &DefaultMovementUpdater{},
	}
	synapticWeightsInstance, err := synaptic.NewNetworkWeights(&appCfg.SimParams, localRng)
	if err != nil {
		return nil, fmt.Errorf("failed to create synaptic weights: %w", err)
	}
	net.SynapticWeights = synapticWeightsInstance

	if err := net.initializeNeurons(appCfg.Cli.TotalNeurons); err != nil {
		return nil, fmt.Errorf("failed to initialize neurons: %w", err)
	}
	allNeuronIDs := make([]common.NeuronID, len(net.Neurons))
	for i, n := range net.Neurons {
		allNeuronIDs[i] = n.ID
	}
	net.SynapticWeights.InitializeAllToAllWeights(allNeuronIDs)
	net.finalizeInitialization()
	net.SpatialGrid.Build(net.Neurons) // Initial build after neurons are positioned

	return net, nil
}

// getNextNeuronID returns the next available unique ID for a new neuron and increments the internal counter.
func (cn *CrowNet) getNextNeuronID() common.NeuronID {
	id := cn.neuronIDCounter
	cn.neuronIDCounter++
	return id
}

// addNeuronsOfType creates and adds a specified number of neurons of a given type to the network.
// Neurons are positioned randomly within a sphere
// whose radius is determined by radiusFactor * SimParams.General.SpaceMaxDimension.
// It also updates the InputNeuronIDs or OutputNeuronIDs slices if applicable.
func (cn *CrowNet) addNeuronsOfType(count int, neuronType neuron.Type, radiusFactor float64) {
	if count <= 0 {
		return
	}
	if cn.SimParams == nil {
		return // Should not happen
	}
	for i := 0; i < count; i++ {
		id := cn.getNextNeuronID()
		effectiveRadius := radiusFactor * cn.SimParams.SimParams.General.SpaceMaxDimension
		pos := space.GenerateRandomPositionInHyperSphere(effectiveRadius, cn.rng)

		// Pass address of SimParams.SimParams (SimulationParameters) to neuron.New
		n := neuron.New(id, neuronType, pos, &cn.SimParams.SimParams)
		cn.Neurons = append(cn.Neurons, n)

		if neuronType == neuron.Input {
			cn.InputNeuronIDs = append(cn.InputNeuronIDs, id)
		} else if neuronType == neuron.Output {
			cn.OutputNeuronIDs = append(cn.OutputNeuronIDs, id)
		}
	}
}

// calculateInternalNeuronCounts determines the number of dopaminergic, inhibitory,
// and excitatory neurons to create based on the total remaining neurons to be distributed and
// the configured percentages for dopaminergic and inhibitory types. Excitatory neurons
// make up the remainder.
func calculateInternalNeuronCounts(
	remainingForDistribution int,
	dopaP, inhibP float64,
) (numDopaminergic, numInhibitory, numExcitatory int) {
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
// Assumes totalNeuronsInput has been validated by config.Validate()
// to be >= MinInputNeurons and MinOutputNeurons.
// It returns an error if the final neuron count does not match the expected count,
// indicating an internal logic issue.
func (cn *CrowNet) initializeNeurons(totalNeuronsInput int) error {
	simParams := &cn.SimParams.SimParams       // Dereference to *config.SimulationParameters
	actualTotalNeurons := totalNeuronsInput
	numInput := simParams.Structure.MinInputNeurons
	numOutput := simParams.Structure.MinOutputNeurons

	cn.addNeuronsOfType(numInput, neuron.Input, float64(simParams.Distribution.ExcitatoryRadiusFactor))
	cn.addNeuronsOfType(numOutput, neuron.Output, float64(simParams.Distribution.ExcitatoryRadiusFactor))

	remainingForInternalDistribution := actualTotalNeurons - numInput - numOutput
	numDopaminergic, numInhibitory, numExcitatory := calculateInternalNeuronCounts(
		remainingForInternalDistribution,
		float64(simParams.Distribution.DopaminergicPercent),
		float64(simParams.Distribution.InhibitoryPercent),
	)

	cn.addNeuronsOfType(numDopaminergic, neuron.Dopaminergic,
		float64(simParams.Distribution.DopaminergicRadiusFactor))
	cn.addNeuronsOfType(numInhibitory, neuron.Inhibitory,
		float64(simParams.Distribution.InhibitoryRadiusFactor))
	cn.addNeuronsOfType(numExcitatory, neuron.Excitatory,
		float64(simParams.Distribution.ExcitatoryRadiusFactor))

	if len(cn.Neurons) != actualTotalNeurons {
		return fmt.Errorf("critical alert: final neuron count (%d) does not match expected (%d) in initializeNeurons",
			len(cn.Neurons), actualTotalNeurons)
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

	cn.OutputNeuronIDSet = make(map[common.NeuronID]struct{}, len(cn.OutputNeuronIDs))
	for _, outID := range cn.OutputNeuronIDs {
		cn.OutputNeuronIDSet[outID] = struct{}{}
		cn.outputFiringHistory[outID] = make([]common.CycleCount, 0)
	}

	cn.InputNeuronIDSet = make(map[common.NeuronID]struct{}, len(cn.InputNeuronIDs))
	for _, inID := range cn.InputNeuronIDs {
		cn.InputNeuronIDSet[inID] = struct{}{}
	}

	cn.neuronMap = make(map[common.NeuronID]*neuron.Neuron, len(cn.Neurons))
	for _, n := range cn.Neurons {
		cn.neuronMap[n.ID] = n
	}
}

// _updateAllNeuronStates handles the decay of accumulated potential and state advancement for all neurons.
func (cn *CrowNet) _updateAllNeuronStates() {
	for _, n := range cn.Neurons { // Or iterate cn.neuronMap
		n.DecayPotential(&cn.SimParams.SimParams)
		n.AdvanceState(cn.CycleCount, &cn.SimParams.SimParams)
	}
}

// _applyChemicalModulationEffects updates chemical levels and applies their effects to neurons
// if chemical modulation is enabled. Otherwise, it resets modulation factors and neuron thresholds.
func (cn *CrowNet) _applyChemicalModulationEffects() {
	simParamsPtr := &cn.SimParams.SimParams
	if cn.isChemicalModulationEnabled {
		cn.ChemicalEnv.UpdateLevels(cn.neuronMap, cn.ActivePulses.GetAll(),
			simParamsPtr.Neurochemical.CortisolGlandPosition, simParamsPtr)
		cn.ChemicalEnv.ApplyEffectsToNeurons(cn.neuronMap, simParamsPtr)
	} else {
		cn.ChemicalEnv.LearningRateModulationFactor = 1.0
		cn.ChemicalEnv.SynaptogenesisModulationFactor = 1.0
		for _, n := range cn.neuronMap { // Iterate cn.neuronMap
			n.CurrentFiringThreshold = n.BaseFiringThreshold
		}
	}
}

// RunCycle executes a single simulation cycle of the CrowNet.
// It orchestrates the various phases of the simulation including:
// 1. Processing continuous inputs (if any).
// 2. Updating base neuron states (potential decay, refractory periods).
// 3. Processing active pulse propagation and their effects on neurons
//    (using the spatial grid for optimization).
// 4. Updating and applying neurochemical effects.
// 5. Applying Hebbian learning.
// 6. Applying synaptogenesis (neuron movement), which also triggers a rebuild
//    of the spatial grid if enabled.
// 7. Incrementing the simulation cycle count.
func (cn *CrowNet) RunCycle() {
	cn.processFrequencyInputs()
	cn._updateAllNeuronStates()
	cn.processActivePulses()
	cn._applyChemicalModulationEffects()

	if cn.isLearningEnabled {
		cn.applyHebbianLearning()
	}
	if cn.isSynaptogenesisEnabled {
		cn.applySynaptogenesis()
		// Rebuild spatial grid if neurons moved.
		// This should use cn.neuronMap values or cn.Neurons if still maintained.
		// Assuming cn.Neurons is up-to-date or cn.neuronMap is used.
		var neuronsForGrid []*neuron.Neuron
		for _, n := range cn.neuronMap {
			neuronsForGrid = append(neuronsForGrid, n)
		}
		cn.SpatialGrid.Build(neuronsForGrid)
	}
	cn.CycleCount++
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
	simParamsPtr := &cn.SimParams.SimParams
	newlyGeneratedPulses := cn.ActivePulses.ProcessCycle(
		cn.SpatialGrid,
		cn.SynapticWeights,
		cn.CycleCount,
		simParamsPtr,
		cn.neuronMap, // Pass map directly
	)

	if len(newlyGeneratedPulses) > 0 {
		for _, newP := range newlyGeneratedPulses {
			if _, isOutputNeuron := cn.OutputNeuronIDSet[newP.EmittingNeuronID]; isOutputNeuron {
				cn.recordOutputFiring(newP.EmittingNeuronID)
			}
		}
		cn.ActivePulses.AddAll(newlyGeneratedPulses)
	}
}

// applyHebbianLearning applies Hebbian learning rules to update synaptic weights.
// It strengthens connections between neurons that fire together
// within a defined time window (coincidenceWindow).
// The learning rate is modulated by the neurochemical environment.
func (cn *CrowNet) applyHebbianLearning() {
	simParamsPtr := &cn.SimParams.SimParams
	effectiveLR := common.Rate(float64(cn.SimParams.Cli.BaseLearningRate) *
		float64(cn.ChemicalEnv.LearningRateModulationFactor))

	if effectiveLR < minEffectiveLearningRateThreshold {
		return
	}

	coincidenceWindow := common.CycleCount(simParamsPtr.Learning.HebbianCoincidenceWindow)

	for _, preSynapticNeuron := range cn.neuronMap { // Iterate map
		isPreActive := cn.isNeuronRecentlyActive(preSynapticNeuron, coincidenceWindow)
		if !isPreActive {
			continue
		}
		preActivityValue := 1.0

		for _, postSynapticNeuron := range cn.neuronMap { // Iterate map
			if preSynapticNeuron.ID == postSynapticNeuron.ID {
				continue
			}
			isPostActive := cn.isNeuronRecentlyActive(postSynapticNeuron, coincidenceWindow)
			if isPostActive {
				postActivityValue := 1.0
				cn.SynapticWeights.ApplyHebbianUpdate(
					preSynapticNeuron.ID,
					postSynapticNeuron.ID,
					preActivityValue,
					postActivityValue,
					effectiveLR,
				)
			}
		}
	}
}

// applySynaptogenesis handles structural plasticity, including neuron movement
// and potentially synapse formation/pruning, modulated by neurochemicals.
func (cn *CrowNet) applySynaptogenesis() {
	simParamsPtr := &cn.SimParams.SimParams
	synaptoModFactor := cn.ChemicalEnv.SynaptogenesisModulationFactor

	// Delegate to the strategy objects for force calculation and movement updates.
	// This is a simplified representation; actual implementation would be more detailed.
	forces := cn.SynaptogenesisForceCalculator.CalculateForces(cn.neuronMap, cn.SpatialGrid, simParamsPtr)
	cn.SynaptogenesisMovementUpdater.ApplyMovements(cn.neuronMap, forces, synaptoModFactor, simParamsPtr, cn.rng)

	// After neurons move, their positions in the SpatialGrid need to be updated.
	// This is typically done by clearing and rebuilding the grid or updating individual neurons.
	// Rebuilding is simpler if many neurons move, but less efficient if few move.
	// For now, assume rebuild is handled by RunCycle or a dedicated grid update method.
}

// SetLogger assigns a logger to the CrowNet instance.
func (cn *CrowNet) SetLogger(logger Logger) {
	cn.logger = logger
}

// LogSnapshot logs the current state of the network.
// This method is intended to be called periodically by the simulation runner.
// (This is a basic LogSnapshot, actual implementation might be in a storage specific logger)
func (cn *CrowNet) LogSnapshot() error {
	if cn.logger == nil {
		return fmt.Errorf("logger not set for CrowNet")
	}
	// Delegate to the specific logger implementation
	return cn.logger.LogNetworkSnapshot(cn.CycleCount, cn)
}

// Interface for logging - can be implemented by SQLite logger or others.
type Logger interface {
	LogNetworkSnapshot(cycle common.CycleCount, net *CrowNet) error
	// Add other logging methods as needed, e.g., LogSynapticWeightUpdate, LogNeuronFiring, etc.
}
