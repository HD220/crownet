package network

// SpaceMaxDimension defines the maximum coordinate value in any dimension.
// The space is assumed to be a hypercube centered at origin, from -SpaceMaxDimension to +SpaceMaxDimension.
// The README mentions "Distância máxima do espaço: 8 unidades".
// If this means from center to edge, then SpaceMaxDimension is 8.
// If it means from one end to the other, then SpaceMaxDimension is 4.
// Assuming it's from center to one edge for now.
const SpaceMaxDimension = 4.0 // Max distance from center (0,0,...) to edge of the space

// Neuron distribution percentages
const (
	DopaminergicPercent = 0.01
	InhibitoryPercent   = 0.30
	ExcitatoryPercent   = 0.69 // This will be adjusted if Input/Output are part of this 69% or separate
	InputPercent        = 0.05
	OutputPercent       = 0.05
)

// Radial constraints for neuron types (percentage of SpaceMaxDimension)
const (
	DopaminergicRadiusFactor = 0.60 // 60% of space
	InhibitoryRadiusFactor   = 0.10 // 10% of space
	ExcitatoryRadiusFactor   = 0.30 // 30% of space
	// Input/Output neurons might not have specific radial constraints mentioned,
	// or they might be distributed within the general excitatory area or throughout the space.
	// For now, let's assume they can be anywhere or within the largest (dopaminergic) radius.
)

// CortisolGlandPosition is at the center of the space.
// This needs to be imported from neuron package or Point type defined locally if used here.
// For now, let's assume network.go will handle the actual Point type.
// var CortisolGlandPosition = /* neuron.Point{0,0,...} */

// Initial firing thresholds (can be made configurable later)
const DefaultFiringThreshold = 1.0 // Note: This is also in neuron/config.go, consider consolidating. For now, network uses neuron's.

// Total number of neurons in the network (example, can be configurable)
const TotalNeurons = 1000

// Synaptogenesis parameters
const (
	// AttractionForceFactor determines strength of pull towards active (firing/refractory) neurons
	AttractionForceFactor = 0.01
	// RepulsionForceFactor determines strength of push from resting neurons
	RepulsionForceFactor = 0.005
	// MaxMovementPerCycle limits how much a neuron can move in one cycle
	MaxMovementPerCycle = 0.1
	// SynaptogenesisInfluenceRadius defines max distance for neurons to affect each other.
	// If 0 or very large, all neurons can influence all others (computationally expensive).
	// Let's start with a reasonably large radius, e.g., a fraction of SpaceMaxDimension.
	SynaptogenesisInfluenceRadius = SpaceMaxDimension * 0.5 // e.g., 50% of space radius
	// DampeningFactor for movement to prevent excessive oscillations and ensure stability
	DampeningFactor = 0.9
)

// Chemical Modulation Parameters - Cortisol
const (
	CortisolGlandRadius           = 0.5  // Radius around gland position to detect pulse "hits"
	CortisolProductionPerHit      = 0.1  // Amount of cortisol produced per excitatory pulse hit
	CortisolDecayRate             = 0.05 // Percentage decay per cycle (e.g., 5%)
	CortisolMinEffectThreshold    = 0.2  // Cortisol level above which effects start
	CortisolOptimalLowThreshold   = 0.8  // Cortisol level for max threshold decrease
	CortisolOptimalHighThreshold  = 1.2  // Cortisol level for max threshold decrease
	CortisolHighEffectThreshold   = 1.5  // Cortisol level above which negative effects (increased threshold, reduced synap) start
	CortisolMaxLevel              = 2.0  // A ceiling for cortisol
	MaxThresholdReductionFactor   = 0.7  // e.g., threshold becomes 70% of base at optimal cortisol
	ThresholdIncreaseFactorHigh   = 1.3  // e.g., threshold becomes 130% of base at very high cortisol
	SynaptogenesisReductionFactor = 0.5  // e.g., movement factors become 50% at high cortisol
)

// Chemical Modulation Parameters - Dopamine
const (
	DopamineProductionPerEvent           = 0.2  // Amount of dopamine produced when a dopaminergic neuron "fires" or releases
	DopamineDecayRate                    = 0.15 // Percentage decay per cycle (e.g., 15%, faster than cortisol)
	DopamineMaxLevel                     = 2.0  // A ceiling for dopamine
	DopamineThresholdIncreaseFactor      = 1.5  // Factor by which threshold increases at max dopamine (e.g., 1.0 baseline -> 1.5)
	DopamineSynaptogenesisIncreaseFactor = 1.5  // Factor by which synaptogenesis movement increases at max dopamine
)

// Input/Output Encoding Parameters
const (
	// CyclesPerSecond represents the number of simulation cycles that correspond to one second of real time.
	// The README mentions "10 ciclos por segundo (framerate)".
	CyclesPerSecond = 10.0
	// OutputFrequencyWindowCycles is the number of past cycles to consider when calculating output neuron firing frequency.
	OutputFrequencyWindowCycles = CyclesPerSecond * 2 // e.g., 2 seconds window
)

// Learning Parameters
const (
	BaseLearningRate     = 0.01 // Base learning rate for Hebbian updates
	ClassificationCycles = 3    // Number of cycles to run network for activity propagation before classification/observation

	// Neuromodulation factors for learning rate
	DopamineLearningEnhancementFactor = 1.5 // At max dopamine, LR can be (1+factor) * base, or simply factor * base. Let's try (1 + factor * normalized_dopamine)
	MaxDopamineLearningMultiplier     = 2.0 // Max multiplier for LR due to dopamine (e.g. 2*BaseLR)
	CortisolLearningSuppressionFactor = 0.2 // At max high cortisol, LR becomes this fraction of current LR (e.g. 0.2 * currentLR)
	MinLearningRateFactor             = 0.1 // Ensure learning rate doesn't become zero or negative due to cortisol.

	// Hebbian learning parameters
	HebbianWeightMin   = -1.0   // Minimum synaptic weight
	HebbianWeightMax   = 1.0    // Maximum synaptic weight
	HebbianWeightDecay = 0.0001 // Small decay factor for all weights per cycle to prevent runaway
	// CoincidenceWindow defines how many cycles apart pre and post-synaptic firing can be to be considered "coincident"
	HebbianCoincidenceWindow = 1 // 0 means same cycle, 1 means up to 1 cycle apart
)
