// Package common defines shared data types used throughout the CrowNet
// application. These types provide a consistent representation for fundamental
// concepts like identifiers, spatial coordinates, and simulation metrics.
// Package common defines shared data types used throughout the CrowNet
// application. These types provide a consistent representation for fundamental
// concepts like identifiers, spatial coordinates, and simulation metrics.
package common

// NeuronID is a unique identifier for a neuron.
type NeuronID int

// CycleCount represents a simulation cycle counter.
type CycleCount int

// Coordinate represents a value in one of the N spatial dimensions.
type Coordinate float64

// PulseValue represents the base value of a pulse before synaptic weighting.
type PulseValue float64

// SynapticWeight represents the strength of a synaptic connection.
type SynapticWeight float64

// Percentage represents a percentage value, typically ranging from 0.0 to 1.0.
type Percentage float64

// Rate represents a rate, such as a learning rate or decay rate.
type Rate float64

// Factor represents a multiplication factor, often used for modulation and typically positive.
type Factor float64

// Threshold represents a threshold value, e.g., for neuron firing.
type Threshold float64

// Level represents the concentration or level of a neurochemical substance, typically non-negative.
type Level float64

// Point represents a point in an N-dimensional space. For CrowNet, this is
// specifically a 16-dimensional space, enforced by the fixed-size array.
type Point [16]Coordinate

// Vector represents a vector in an N-dimensional space, used for quantities
// like velocity or force. It uses Coordinate for consistency with Point and is
// fixed to 16 dimensions for CrowNet.
type Vector [16]Coordinate
