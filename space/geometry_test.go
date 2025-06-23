package space_test

import (
	"crownet/common"
	"crownet/space"
	"math"
	"math/rand"
	"testing"
)

// Helper para comparar floats com tolerância
func floatEquals(a, b, tolerance float64) bool {
	if a == b { // Lida com Infinitos
		return true
	}
	return math.Abs(a-b) < tolerance
}

// Helper para preencher o restante das dimensões de um ponto/vetor
func fillRemainingDimensions(p *common.Point, startIdx int, val common.Coordinate) {
	for i := startIdx; i < 16; i++ { // Assumindo 16 dimensões
		p[i] = val
	}
}


func TestEuclideanDistance(t *testing.T) {
	testCases := []struct {
		name     string
		p1       common.Point
		p2       common.Point
		expected float64
	}{
		{
			name:     "Distância zero (mesmo ponto)",
			p1:       common.Point{1, 2, 3},
			p2:       common.Point{1, 2, 3},
			expected: 0.0,
		},
		{
			name:     "Distância simples em 1D",
			p1:       common.Point{1},
			p2:       common.Point{4},
			expected: 3.0,
		},
		{
			name:     "Distância simples em 2D (Pitágoras 3-4-5)",
			p1:       common.Point{0, 0},
			p2:       common.Point{3, 4},
			expected: 5.0,
		},
		{
			name:     "Distância com coordenadas negativas",
			p1:       common.Point{-1, -1},
			p2:       common.Point{1, 1},
			expected: math.Sqrt(8.0),
		},
		{
			name:     "Distância em mais dimensões (parcial)",
			p1:       common.Point{1, 1, 1, 1},
			p2:       common.Point{2, 2, 2, 2},
			expected: 2.0,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			dist := space.EuclideanDistance(tc.p1, tc.p2)
			if !floatEquals(dist, tc.expected, 1e-9) {
				t.Errorf("Expected distance %f, got %f", tc.expected, dist)
			}
		})
	}
}

func TestIsWithinRadius(t *testing.T) {
	center := common.Point{0, 0}
	fillRemainingDimensions(&center, 2, 0)

	testCases := []struct {
		name     string
		point    common.Point
		radius   float64
		expected bool
	}{
		{name: "Dentro do raio", point: common.Point{1, 0}, radius: 2.0, expected: true},
		{name: "Na borda do raio", point: common.Point{2, 0}, radius: 2.0, expected: true},
		{name: "Fora do raio", point: common.Point{3, 0}, radius: 2.0, expected: false},
		{name: "Raio zero, ponto na origem", point: common.Point{0, 0}, radius: 0.0, expected: true},
		{name: "Raio zero, ponto fora da origem", point: common.Point{1, 0}, radius: 0.0, expected: false},
		{name: "Raio negativo", point: common.Point{1, 0}, radius: -1.0, expected: false},
		{name: "Ponto na origem, raio positivo", point: common.Point{0,0}, radius: 1.0, expected: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fillRemainingDimensions(&tc.point, 2, 0)

			within := space.IsWithinRadius(center, tc.point, tc.radius)
			if within != tc.expected {
				dist := space.EuclideanDistance(center, tc.point)
				t.Errorf("Point %v, radius %f: expected %t, got %t (distance: %f)", tc.point, tc.radius, tc.expected, within, dist)
			}
		})
	}
}

func TestClampToHyperSphere(t *testing.T) {
	var p_origin common.Point

	testCases := []struct {
		name          string
		point         common.Point
		maxRadius     float64
		expectedPoint common.Point
		expectClamped bool
	}{
		{
			name:          "Já dentro da esfera",
			point:         common.Point{1, 0},
			maxRadius:     2.0,
			expectedPoint: common.Point{1, 0},
			expectClamped: false,
		},
		{
			name:          "Na superfície da esfera",
			point:         common.Point{2, 0},
			maxRadius:     2.0,
			expectedPoint: common.Point{2, 0},
			expectClamped: false,
		},
		{
			name:          "Fora da esfera, precisa clampar (2D)",
			point:         common.Point{3, 4},
			maxRadius:     2.5,
			expectedPoint: common.Point{1.5, 2.0},
			expectClamped: true,
		},
		{
			name:          "Ponto na origem, raio positivo",
			point:         p_origin,
			maxRadius:     5.0,
			expectedPoint: p_origin,
			expectClamped: false,
		},
		{
			name:          "Ponto na origem, raio zero",
			point:         p_origin,
			maxRadius:     0.0,
			expectedPoint: p_origin,
			expectClamped: false,
		},
		{
			name:          "Ponto fora, raio zero (deve clampar para origem)",
			point:         common.Point{1,1},
			maxRadius:     0.0,
			expectedPoint: p_origin,
			expectClamped: true,
		},
		{
			name:          "Raio negativo (não deve clampar)",
			point:         common.Point{10,10},
			maxRadius:     -1.0,
			expectedPoint: common.Point{10,10},
			expectClamped: false,
		},
		{
			name:          "Ponto {0,0,...0} fora de uma esfera de raio > 0 (impossível, mas testar edge case de cálculo de dist)",
			point:         p_origin,
			maxRadius:     1e-12,
			expectedPoint: p_origin,
			expectClamped: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			var p_test_full, expected_full common.Point
			copy(p_test_full[:], tc.point[:])
			copy(expected_full[:], tc.expectedPoint[:])

			clampedP, wasClamped := space.ClampToHyperSphere(p_test_full, tc.maxRadius)

			if wasClamped != tc.expectClamped {
				t.Errorf("Expected wasClamped to be %t, got %t", tc.expectClamped, wasClamped)
			}

			for i := 0; i < 16; i++ {
				if !floatEquals(float64(clampedP[i]), float64(expected_full[i]), 1e-9) {
					t.Errorf("Coordinate %d: expected %f, got %f. Original point: %v, maxRadius: %f",
						i, expected_full[i], clampedP[i], tc.point, tc.maxRadius)
					break
				}
			}
		})
	}
}

func TestGenerateRandomPositionInHyperSphere(t *testing.T) {
	maxRadius := 10.0
	numSamples := 1000

	rng := rand.New(rand.NewSource(12345))
	randomSource := func() float64 { return rng.Float64() }

	for i := 0; i < numSamples; i++ {
		p := space.GenerateRandomPositionInHyperSphere(maxRadius, randomSource)

		distSq := 0.0
		for j := 0; j < 16; j++ {
			coord := float64(p[j])
			distSq += coord * coord
		}
		if distSq > maxRadius*maxRadius + 1e-9 {
			t.Errorf("Generated point %v is outside maxRadius %f (distSq: %f)", p, maxRadius, distSq)
		}
	}

	pZero := space.GenerateRandomPositionInHyperSphere(0.0, randomSource)
	var originPoint common.Point
	if pZero != originPoint {
		t.Errorf("Expected point at origin for maxRadius 0, got %v", pZero)
	}

	pNegativeRadius := space.GenerateRandomPositionInHyperSphere(-5.0, randomSource)
	if pNegativeRadius != originPoint {
		t.Errorf("Expected point at origin for negative maxRadius, got %v", pNegativeRadius)
	}
}
