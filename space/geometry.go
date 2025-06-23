package space

import (
	"crownet/common"
	"math"
)

// pointDimension define a dimensionalidade dos pontos e vetores neste pacote.
const pointDimension = 16

// EuclideanDistance calcula a distância Euclidiana entre dois pontos.
func EuclideanDistance(p1, p2 common.Point) float64 {
	var sumOfSquares float64
	for i := 0; i < pointDimension; i++ {
		diff := float64(p1[i] - p2[i])
		sumOfSquares += diff * diff
	}
	return math.Sqrt(sumOfSquares)
}

// IsWithinRadius verifica se um ponto pTest está dentro de um raio de um ponto central pCenter.
func IsWithinRadius(pCenter, pTest common.Point, radius float64) bool {
	// Evitar Sqrt se possível, comparando quadrados
	if radius < 0 { return false } // Raio negativo não faz sentido
	var sumOfSquares float64
	for i := 0; i < pointDimension; i++ {
		diff := float64(pCenter[i] - pTest[i])
		sumOfSquares += diff * diff
		// Otimização: se sumOfSquares já excede radius*radius, podemos parar mais cedo.
		// No entanto, radius*radius pode dar overflow se radius for grande.
		// A comparação direta com EuclideanDistance é mais simples de manter correta.
	}
	// return sumOfSquares <= radius*radius // Alternativa sem Sqrt, mas requer cuidado com radius=0
	return EuclideanDistance(pCenter, pTest) <= radius
}

// ClampToHyperSphere garante que um ponto permaneça dentro de uma hiperesfera de raio `maxRadius`
// a partir da origem (0,0,...,0). Se estiver fora, o ponto é projetado para a superfície da esfera.
// Retorna o novo ponto e um booleano indicando se foi necessário fazer o clamp.
func ClampToHyperSphere(p common.Point, maxRadius float64) (clampedPoint common.Point, wasClamped bool) {
	const epsilon = 1e-9 // For floating point comparisons

	if maxRadius < 0 {
		// Negative radius implies no boundary.
		return p, false
	}

	// Calculate squared distance from origin.
	distFromOriginSq := 0.0
	isOrigin := true
	for i := 0; i < pointDimension; i++ {
		coordVal := float64(p[i])
		distFromOriginSq += coordVal * coordVal
		if math.Abs(coordVal) > epsilon {
			isOrigin = false
		}
	}

	if maxRadius < epsilon { // Case: maxRadius is effectively zero.
		if isOrigin {
			return p, false // Point is at origin, radius is zero. No clamp needed.
		}
		// Point is not at origin, but radius is zero. Clamp to origin.
		var originP common.Point
		return originP, true
	}

	// Case: maxRadius is positive.
	// Check if already within or on the surface (with tolerance for floating point).
	if distFromOriginSq <= maxRadius*maxRadius+epsilon {
		return p, false // Point is inside or on the surface.
	}

	// Point is outside and maxRadius is positive.
	// If the point was the origin, it would have been inside (distFromOriginSq = 0).
	// Therefore, distFromOrigin will be > 0 here.
	distFromOrigin := math.Sqrt(distFromOriginSq)

	// This check should ideally not be needed if logic is correct,
	// as an origin point (distFromOrigin < epsilon) with maxRadius > epsilon
	// should be classified as 'inside'. This is a safeguard.
	if distFromOrigin < epsilon { // Should effectively mean point p is the origin.
	    // If p is origin and maxRadius > 0, it's inside. This path indicates an unexpected state
	    // or extremely small maxRadius that wasn't caught by maxRadius < epsilon.
	    // Safest is to return p as it is effectively the origin and should be inside if maxRadius > 0.
		return p, false
	}


	scaleFactor := maxRadius / distFromOrigin
	clampedP := common.Point{}
	for i := 0; i < pointDimension; i++ {
		clampedP[i] = common.Coordinate(float64(p[i]) * scaleFactor)
	}
	return clampedP, true
}

// GenerateRandomPositionInHyperSphere cria uma posição aleatória dentro de uma hiperesfera
// de raio `maxRadius` centrada na origem.
// Utiliza amostragem por rejeição dentro de um hipercubo envolvente.
// Nota: A amostragem por rejeição pode se tornar ineficiente em altas dimensões (como 16D)
// devido ao "curse of dimensionality", onde o volume da hiperesfera se torna
// uma fração muito pequena do volume do hipercubo envolvente.
// Alternativas como o método de Muller podem ser mais eficientes, mas são mais complexas.
func GenerateRandomPositionInHyperSphere(maxRadius float64, randomSource func() float64) common.Point {
	if maxRadius < 0 { maxRadius = 0 } // Raio não pode ser negativo

	for {
		var p common.Point
		distSq := 0.0
		for i := 0; i < pointDimension; i++ {
			// Gera coordenada entre -maxRadius e +maxRadius
			coord := (randomSource()*2*maxRadius - maxRadius)
			p[i] = common.Coordinate(coord)
			distSq += coord * coord
		}

		// Se maxRadius é 0, o único ponto válido é a origem.
		if maxRadius == 0 {
			// p já será {0,0,...} devido a coord = (randomSource()*0 - 0) = 0
			return p
		}

		// Se o ponto gerado estiver dentro da hiperesfera, retorna
		if distSq <= maxRadius*maxRadius {
			return p
		}
		// Caso contrário, tenta novamente (rejeição)
	}
}
