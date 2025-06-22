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
			expected: 3.0, // sqrt((4-1)^2) = 3
		},
		{
			name:     "Distância simples em 2D (Pitágoras 3-4-5)",
			p1:       common.Point{0, 0},
			p2:       common.Point{3, 4},
			expected: 5.0, // sqrt(3^2 + 4^2) = sqrt(9+16) = sqrt(25) = 5
		},
		{
			name:     "Distância com coordenadas negativas",
			p1:       common.Point{-1, -1},
			p2:       common.Point{1, 1},
			expected: math.Sqrt(8.0), // sqrt((1 - -1)^2 + (1 - -1)^2) = sqrt(2^2 + 2^2) = sqrt(4+4) = sqrt(8)
		},
		{
			name:     "Distância em mais dimensões (parcial)",
			p1:       common.Point{1, 1, 1, 1},
			p2:       common.Point{2, 2, 2, 2},
			expected: 2.0, // sqrt(1^2+1^2+1^2+1^2) = sqrt(4) = 2
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Preencher dimensões restantes para garantir que não afetem o teste se não definidas
			// fillRemainingDimensions(&tc.p1, len(tc.p1), 0) // Isso não funciona pois tc.p1 é um array, não um slice
			// A forma como os pontos são definidos já considera as 16 dimensões (inicializadas com 0 se não especificadas).

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
		{name: "Dentro do raio", point: common.Point{1, 0}, radius: 2.0, expected: true},  // dist = 1
		{name: "Na borda do raio", point: common.Point{2, 0}, radius: 2.0, expected: true}, // dist = 2
		{name: "Fora do raio", point: common.Point{3, 0}, radius: 2.0, expected: false}, // dist = 3
		{name: "Raio zero, ponto na origem", point: common.Point{0, 0}, radius: 0.0, expected: true},
		{name: "Raio zero, ponto fora da origem", point: common.Point{1, 0}, radius: 0.0, expected: false},
		{name: "Raio negativo", point: common.Point{1, 0}, radius: -1.0, expected: false},
		{name: "Ponto na origem, raio positivo", point: common.Point{0,0}, radius: 1.0, expected: true},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			fillRemainingDimensions(&tc.point, 2, 0) // Assumindo testes primariamente em 2D para simplicidade

			// A lógica em IsWithinRadius foi alterada para não usar Sqrt, vamos testar isso.
			// No entanto, o fallback atual ainda usa EuclideanDistance. Se a otimização for reintroduzida,
			// este teste precisará ser mais cuidadoso com a comparação de quadrados.
			// Por agora, o teste é contra o comportamento atual (que usa EuclideanDistance).

			within := space.IsWithinRadius(center, tc.point, tc.radius)
			if within != tc.expected {
				dist := space.EuclideanDistance(center, tc.point)
				t.Errorf("Point %v, radius %f: expected %t, got %t (distance: %f)", tc.point, tc.radius, tc.expected, within, dist)
			}
		})
	}
}

func TestClampToHyperSphere(t *testing.T) {
	// Testes usam principalmente 2D para facilitar o raciocínio, preenchendo o resto com 0
	var p_origin common.Point // {0,0,...}

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
			expectClamped: false, // A implementação considera na superfície como "não clampado" se já estiver lá
		},
		{
			name:          "Fora da esfera, precisa clampar (2D)",
			point:         common.Point{3, 4}, // Distância 5
			maxRadius:     2.5,                // Metade da distância
			expectedPoint: common.Point{1.5, 2.0}, // (3,4) * (2.5/5) = (3,4) * 0.5 = (1.5, 2.0)
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
			// Este caso testa se distFromOrigin < epsilon e o ponto é a origem.
			// Se maxRadius for muito pequeno mas positivo, ex: 1e-12
			// e o ponto é a origem (distFromOriginSq = 0).
			// distFromOriginSq (0) <= maxRadius*maxRadius (1e-24) + epsilon (1e-9) -> true, retorna p, false
			point:         p_origin,
			maxRadius:     1e-12,
			expectedPoint: p_origin,
			expectClamped: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Preencher dimensões não especificadas com 0 para consistência
			var p_test_full, expected_full common.Point
			copy(p_test_full[:], tc.point[:])
			copy(expected_full[:], tc.expectedPoint[:])

			clampedP, wasClamped := space.ClampToHyperSphere(p_test_full, tc.maxRadius)

			if wasClamped != tc.expectClamped {
				t.Errorf("Expected wasClamped to be %t, got %t", tc.expectClamped, wasClamped)
			}

			// Verificar as coordenadas do ponto clampeado
			for i := 0; i < 16; i++ { // pointDimension
				if !floatEquals(float64(clampedP[i]), float64(expected_full[i]), 1e-9) {
					t.Errorf("Coordinate %d: expected %f, got %f. Original point: %v, maxRadius: %f",
						i, expected_full[i], clampedP[i], tc.point, tc.maxRadius)
					break // Parar no primeiro erro de coordenada
				}
			}
		})
	}
}


func TestGenerateRandomPositionInHyperSphere(t *testing.T) {
	maxRadius := 10.0
	numSamples := 1000

	// Fonte de aleatoriedade determinística para o teste
	rng := rand.New(rand.NewSource(12345))
	randomSource := func() float64 { return rng.Float64() }

	for i := 0; i < numSamples; i++ {
		p := space.GenerateRandomPositionInHyperSphere(maxRadius, randomSource)

		distSq := 0.0
		for j := 0; j < 16; j++ { // pointDimension
			coord := float64(p[j])
			distSq += coord * coord
		}
		// Verificar se o ponto está dentro do raio (com uma pequena tolerância para erros de float)
		if distSq > maxRadius*maxRadius + 1e-9 {
			t.Errorf("Generated point %v is outside maxRadius %f (distSq: %f)", p, maxRadius, distSq)
		}
	}

	// Teste com raio zero
	pZero := space.GenerateRandomPositionInHyperSphere(0.0, randomSource)
	var originPoint common.Point // {0,0,...}
	if pZero != originPoint {
		t.Errorf("Expected point at origin for maxRadius 0, got %v", pZero)
	}

	// Teste com raio negativo (deve ser tratado como zero)
	pNegativeRadius := space.GenerateRandomPositionInHyperSphere(-5.0, randomSource)
	if pNegativeRadius != originPoint {
		t.Errorf("Expected point at origin for negative maxRadius, got %v", pNegativeRadius)
	}
}
```
