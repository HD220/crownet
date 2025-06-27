package space

import (
	"math"
	"math/rand"
	"testing"

	"crownet/common" // For common.Point
)

func TestEuclideanDistance(t *testing.T) {
	tests := []struct {
		name     string
		p1, p2   common.Point
		expected float64
	}{
		{"same point", common.Point{1, 2}, common.Point{1, 2}, 0.0},
		{"2D distance", common.Point{0, 0}, common.Point{3, 4}, 5.0}, // 3-4-5 triangle
		{"different dimensions", common.Point{1, 2}, common.Point{1, 2, 3}, 0.0},
		{"negative coords", common.Point{-1, -1}, common.Point{1, 1},
			math.Sqrt(8)}, // sqrt( (1 - -1)^2 + (1 - -1)^2 ) = sqrt(2^2 + 2^2) = sqrt(4+4)
		{"1D distance", common.Point{5}, common.Point{2}, 3.0},
		{"empty points", common.Point{}, common.Point{}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dist := EuclideanDistance(tt.p1, tt.p2)
			if math.Abs(dist-tt.expected) > 1e-9 {
				t.Errorf("EuclideanDistance(%v, %v) = %f, want %f", tt.p1, tt.p2, dist, tt.expected)
			}
		})
	}
}

func TestMagnitude(t *testing.T) {
	tests := []struct {
		name     string
		p        common.Point
		expected float64
	}{
		{"origin", common.Point{0, 0, 0}, 0.0},
		{"unit vector x", common.Point{1, 0, 0}, 1.0},
		{"unit vector y", common.Point{0, -1, 0}, 1.0},
		{"3-4-5 vector", common.Point{3, 4, 0}, 5.0},
		{"negative components", common.Point{-1, -2, -2}, 3.0}, // sqrt(1+4+4) = sqrt(9) = 3
		{"1D vector", common.Point{-5}, 5.0},
		{"empty vector", common.Point{}, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mag := Magnitude(tt.p)
			if math.Abs(mag-tt.expected) > 1e-9 {
				t.Errorf("Magnitude(%v) = %f, want %f", tt.p, mag, tt.expected)
			}
		})
	}
}

func TestGenerateRandomPositionInHyperSphere(t *testing.T) {
	rng := rand.New(rand.NewSource(0))

	tests := []struct {
		name      string
		maxRadius float64
		dimension int
	}{
		{"zero radius", 0.0, common.PointDimension},
		{"negative radius", -5.0, common.PointDimension},
		{"positive radius", 10.0, common.PointDimension},
		{"small radius", 0.1, common.PointDimension},
	}

	numSamples := 1000

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isOriginExpected := tt.maxRadius <= 0
			sumDist := 0.0
			var firstPoint common.Point
			allPointsSame := true

			for i := 0; i < numSamples; i++ {
				p := GenerateRandomPositionInHyperSphere(tt.maxRadius, rng)

				if len(p) != tt.dimension {
					t.Errorf("Generated point with dimension %d, want %d", len(p), tt.dimension)
					return
				}

				dist := Magnitude(p)
				sumDist += dist

				isOrigin := true
				for _, val := range p {
					if val != 0 {
						isOrigin = false
						break
					}
				}

				if isOriginExpected {
					if !isOrigin {
						t.Errorf("Radius %v, expected origin, got %v (dist %v)",
							tt.maxRadius, p, dist)
					}
				} else {
					if dist > tt.maxRadius+1e-9 {
						t.Errorf("Radius %v generated point %v with distance %v, outside expected radius",
							tt.maxRadius, p, dist)
					}
				}
				if i == 0 {
					firstPoint = p
				} else if allPointsSame && !pointsEqual(p, firstPoint) {
					allPointsSame = false
				}
			}

			if !isOriginExpected && tt.maxRadius > 0 {
				if sumDist/float64(numSamples) == 0 && numSamples > 1 {
					t.Errorf("For R=%.1f generated all points at origin", tt.maxRadius)
				}
				if allPointsSame && numSamples > 10 {
					t.Errorf("For R=%.1f generated %d identical points: %v",
						tt.maxRadius, numSamples, firstPoint)
				}
			}
		})
	}
}

func pointsEqual(p1, p2 common.Point) bool {
	if len(p1) != len(p2) {
		return false
	}
	for i := range p1 {
		if p1[i] != p2[i] {
			return false
		}
	}
	return true
}

func TestClampToHyperSphere(t *testing.T) {
	tests := []struct {
		name      string
		p         common.Point
		maxRadius float64
		expectedP common.Point
	}{
		{"already inside", common.Point{1, 1}, 2.0, common.Point{1, 1}},
		{"on surface", common.Point{2, 0}, 2.0, common.Point{2, 0}},
		{"outside, needs clamping", common.Point{3, 4}, 2.5, common.Point{1.5, 2.0}},
		{"origin, positive radius", common.Point{0, 0}, 5.0, common.Point{0, 0}},
		{"origin, zero radius", common.Point{0, 0}, 0.0, common.Point{0, 0}},
		{"non-origin, zero radius", common.Point{1, 1}, 0.0, common.Point{0, 0}},
		{"negative radius, no clamp", common.Point{10, 10}, -1.0, common.Point{10, 10}},
		{"1D point inside", common.Point{3}, 5.0, common.Point{3}},
		{"1D point outside", common.Point{7}, 5.0, common.Point{5}},
		{"1D point outside negative", common.Point{-7}, 5.0, common.Point{-5}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clampedP := ClampToHyperSphere(tt.p, tt.maxRadius)
			if len(clampedP) != len(tt.expectedP) {
				t.Fatalf("Returned point of len %d, want %d", len(clampedP), len(tt.expectedP))
			}
			for i := range clampedP {
				if math.Abs(float64(clampedP[i]-tt.expectedP[i])) > 1e-9 {
					t.Errorf("ClampToHyperSphere(%v, %.1f) = %v, want %v. Mismatch at index %d.",
						tt.p, tt.maxRadius, clampedP, tt.expectedP, i)
					break
				}
			}
		})
	}
}
