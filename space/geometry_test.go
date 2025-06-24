package space

import (
	"crownet/common"
	"math"
	"math/rand"
	"testing"
)

func TestEuclideanDistance(t *testing.T) {
	tests := []struct {
		name string
		p1   common.Point
		p2   common.Point
		want float64
	}{
		{"zero distance", common.Point{1, 2, 3}, common.Point{1, 2, 3}, 0.0},
		{"simple case 2D (in 16D)", common.Point{3, 0}, common.Point{0, 4}, 5.0}, // 3-4-5 triangle
		{"1D case (in 16D)", common.Point{5}, common.Point{2}, 3.0},
		{"negative coords", common.Point{-1, -1}, common.Point{1, 1}, math.Sqrt(8)}, // sqrt( (1 - -1)^2 + (1 - -1)^2 ) = sqrt(2^2 + 2^2) = sqrt(4+4)
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pad points to 16D if not already
			p1Padded := tt.p1
			p2Padded := tt.p2

			if got := EuclideanDistance(p1Padded, p2Padded); math.Abs(got-tt.want) > 1e-9 {
				t.Errorf("EuclideanDistance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestIsWithinRadius(t *testing.T) {
	center := common.Point{0, 0, 0}
	tests := []struct {
		name   string
		pTest  common.Point
		radius float64
		want   bool
	}{
		{"inside", common.Point{1, 0, 0}, 2.0, true},
		{"on boundary", common.Point{2, 0, 0}, 2.0, true},
		{"outside", common.Point{3, 0, 0}, 2.0, false},
		{"zero radius, point at center", common.Point{0, 0, 0}, 0.0, true},
		{"zero radius, point not at center", common.Point{1, 0, 0}, 0.0, false},
		{"negative radius", common.Point{1, 0, 0}, -1.0, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsWithinRadius(center, tt.pTest, tt.radius); got != tt.want {
				t.Errorf("IsWithinRadius() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClampToHyperSphere(t *testing.T) {
	origin := common.Point{}
	tests := []struct {
		name            string
		p               common.Point
		maxRadius       float64
		wantClampedP    common.Point
		wantWasClamped  bool
		epsilon         float64
	}{
		{"inside, no clamp", common.Point{1,0}, 2.0, common.Point{1,0}, false, 1e-9},
		{"outside, clamp", common.Point{3,0}, 2.0, common.Point{2,0}, true, 1e-9},
		{"on boundary, no clamp", common.Point{2,0}, 2.0, common.Point{2,0}, false, 1e-9},
		{"at origin, radius > 0", common.Point{0,0}, 2.0, common.Point{0,0}, false, 1e-9},
		{"at origin, radius = 0", common.Point{0,0}, 0.0, common.Point{0,0}, false, 1e-9},
		{"not origin, radius = 0", common.Point{1,0}, 0.0, origin, true, 1e-9},
		{"negative radius, no clamp", common.Point{1,0}, -1.0, common.Point{1,0}, false, 1e-9},
		{"multi-dim outside", common.Point{3,4}, 2.5, common.Point{1.5, 2.0}, true, 1e-9}, // Dist=5, radius=2.5, scale=0.5
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotClampedP, gotWasClamped := ClampToHyperSphere(tt.p, tt.maxRadius)
			if gotWasClamped != tt.wantWasClamped {
				t.Errorf("ClampToHyperSphere() gotWasClamped = %v, want %v", gotWasClamped, tt.wantWasClamped)
			}
			distError := EuclideanDistance(gotClampedP, tt.wantClampedP)
			if distError > tt.epsilon {
				t.Errorf("ClampToHyperSphere() gotClampedP = %v, want %v (dist error %v)", gotClampedP, tt.wantClampedP, distError)
			}
		})
	}
}

func TestGenerateRandomPositionInHyperSphere(t *testing.T) {
	seed := int64(12345)
	rng := rand.New(rand.NewSource(seed))
	numSamples := 100

	tests := []struct {
		name      string
		maxRadius float64
	}{
		{"radius 0", 0.0},
		{"negative radius", -5.0},
		{"positive radius 1", 1.0},
		{"positive radius 10", 10.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedRadius := tt.maxRadius
			if expectedRadius < 0 {
				expectedRadius = 0 // Function treats negative radius as 0
			}

			for i := 0; i < numSamples; i++ {
				p := GenerateRandomPositionInHyperSphere(tt.maxRadius, rng)
				dist := EuclideanDistance(common.Point{}, p) // Distance from origin

				if expectedRadius == 0 {
					isOrigin := true
					for _, coord := range p {
						if math.Abs(float64(coord)) > 1e-9 {
							isOrigin = false
							break
						}
					}
					if !isOrigin {
						t.Errorf("GenerateRandomPositionInHyperSphere() with radius %v, expected origin, got %v (dist %v)", tt.maxRadius, p, dist)
					}
				} else {
					// Allow for a tiny bit of floating point error, hence <=
					if dist > expectedRadius+1e-9 {
						t.Errorf("GenerateRandomPositionInHyperSphere() with radius %v generated point %v with distance %v, outside expected radius", tt.maxRadius, p, dist)
					}
				}
			}
		})
	}

	// Basic check for distribution (very naive - just check not all points are same or at origin for positive radius)
	t.Run("distribution sanity check R=5", func(t *testing.T) {
		radius := 5.0
		points := make([]common.Point, numSamples)
		allSame := true
		allOrigin := true

		firstPoint := GenerateRandomPositionInHyperSphere(radius, rng)
		points[0] = firstPoint
		isFirstPointOrigin := true
		for _, coord := range firstPoint {
			if math.Abs(float64(coord)) > 1e-9 {
				isFirstPointOrigin = false
				break
			}
		}
		if !isFirstPointOrigin {
			allOrigin = false
		}

		for i := 1; i < numSamples; i++ {
			p := GenerateRandomPositionInHyperSphere(radius, rng)
			points[i] = p
			if EuclideanDistance(p, firstPoint) > 1e-9 { // Not same as first point
				allSame = false
			}
			isCurrentPointOrigin := true
			for _, coord := range p {
				if math.Abs(float64(coord)) > 1e-9 {
					isCurrentPointOrigin = false
					break
				}
			}
			if !isCurrentPointOrigin {
				allOrigin = false
			}
		}

		if radius > 0 && allSame {
			t.Errorf("GenerateRandomPositionInHyperSphere() for R=%.1f generated %d identical points: %v", radius, numSamples, firstPoint)
		}
		if radius > 0 && allOrigin {
			t.Errorf("GenerateRandomPositionInHyperSphere() for R=%.1f generated %d points at origin", radius, numSamples)
		}
	})
}
