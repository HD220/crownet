package utils

import (
	"math"
	"math/rand"
	"time"

	"crownet/neuron"
)

// Seed random number generator
func init() {
	rand.Seed(time.Now().UnixNano())
}

// EuclideanDistance calculates the Euclidean distance between two points in 16D space.
func EuclideanDistance(p1, p2 neuron.Point) float64 {
	var sum float64
	for i := 0; i < 16; i++ {
		diff := p1[i] - p2[i]
		sum += diff * diff
	}
	return math.Sqrt(sum)
}

// GenerateRandomPosition creates a random position in 16D space within a given radius/bound.
// For now, it generates within a hypercube of [-maxCoord, maxCoord]^16
func GenerateRandomPosition(maxCoord float64) neuron.Point {
	var p neuron.Point
	for i := 0; i < 16; i++ {
		p[i] = (rand.Float64() * 2 * maxCoord) - maxCoord
	}
	return p
}

// Placeholder for OpenNoise or custom noise generation for structured neuron placement.
// For now, it will just return a random point, similar to GenerateRandomPosition.
// In a real implementation, this would use a noise algorithm (e.g., Perlin, Simplex)
// to create more structured, less random, initial positions.
func GenerateStructuredPosition(seed float64, scale float64, offset neuron.Point, maxCoord float64) neuron.Point {
	// This is a placeholder. A proper noise function would be more complex.
	// For example, using each dimension as an input to a 16D noise function,
	// or generating each coordinate based on a noise value derived from the seed and index.
	var p neuron.Point
	// Simple approach: generate random but could be biased by a "seed" or global offset.
	// This doesn't use OpenNoise yet.
	for i := 0; i < 16; i++ {
		// A very basic "noise" - could be improved with actual noise libraries
		p[i] = (rand.Float64()*2*maxCoord - maxCoord) + offset[i]
		// Ensure it stays within bounds if necessary, though noise might naturally handle this
		// or the distribution logic would place it.
		if p[i] > maxCoord {
			p[i] = maxCoord
		}
		if p[i] < -maxCoord {
			p[i] = -maxCoord
		}
	}
	return p
}
