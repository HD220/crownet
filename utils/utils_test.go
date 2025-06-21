package utils

import (
	"crownet/neuron" // For neuron.Point
	"math"
	"testing"
)

func TestEuclideanDistance(t *testing.T) {
	p1 := neuron.Point{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	p2 := neuron.Point{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	expected1 := 1.0
	dist1 := EuclideanDistance(p1, p2)
	if math.Abs(dist1-expected1) > 1e-9 {
		t.Errorf("EuclideanDistance failed for single dimension. Expected %.2f, got %.2f", expected1, dist1)
	}

	p3 := neuron.Point{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	p4 := neuron.Point{3, 4, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0} // 3-4-5 triangle
	expected2 := 5.0
	dist2 := EuclideanDistance(p3, p4)
	if math.Abs(dist2-expected2) > 1e-9 {
		t.Errorf("EuclideanDistance failed for 3-4-5 triangle. Expected %.2f, got %.2f", expected2, dist2)
	}

	p5 := neuron.Point{1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1, 1}
	p6 := neuron.Point{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	expected3 := math.Sqrt(16) // sqrt(1*1 * 16)
	dist3 := EuclideanDistance(p5, p6)
	if math.Abs(dist3-expected3) > 1e-9 {
		t.Errorf("EuclideanDistance failed for 16D unit vector. Expected %.2f, got %.2f", expected3, dist3)
	}

	// Test with identical points
	distSame := EuclideanDistance(p1, p1)
	if math.Abs(distSame-0.0) > 1e-9 {
		t.Errorf("EuclideanDistance failed for identical points. Expected 0.0, got %.2f", distSame)
	}
}

// Test for GenerateRandomPosition and GenerateStructuredPosition would involve checking bounds
// or statistical properties, which is more involved. For now, focusing on deterministic functions.
