package space

import (
	"crownet/common"
	"math"
)

// pointDimension defines the dimensionality of points and vectors processed by this package.
// For CrowNet, this is typically a 16-dimensional space.
const pointDimension = 16

// EuclideanDistance calculates the standard Euclidean distance between two N-dimensional points.
// It sums the squares of the differences of corresponding coordinates and returns the square root of that sum.
func EuclideanDistance(p1, p2 common.Point) float64 {
	var sumOfSquares float64
	for i := range p1 { // Iterate over dimensions of the point
		diff := float64(p1[i] - p2[i])
		sumOfSquares += diff * diff
	}
	return math.Sqrt(sumOfSquares)
}

// IsWithinRadius checks if a point pTest is within a specified Euclidean distance (radius)
// from a central point pCenter in N-dimensional space.
// It handles negative radius by returning false.
// For robustness and clarity with edge cases (like radius=0), it uses EuclideanDistance.
func IsWithinRadius(pCenter, pTest common.Point, radius float64) bool {
	if radius < 0 { // A negative radius is not meaningful for this check.
		return false
	}
	// Direct Euclidean distance comparison is preferred over squared distances
	// to avoid potential overflow with radius*radius and to simplify handling of radius = 0.
	return EuclideanDistance(pCenter, pTest) <= radius
}

// ClampToHyperSphere ensures that a given point p stays within (or on the surface of)
// an N-dimensional hypersphere centered at the origin with a specified maxRadius.
// If the point is outside this hypersphere, it is projected onto the hypersphere's surface
// by scaling its vector from the origin.
//
// Parameters:
//   - p: The point to be clamped.
//   - maxRadius: The maximum allowed distance from the origin. If negative, the original point is returned as no clamping is applied.
//
// Returns:
//   - clampedPoint: The (potentially) clamped point.
//   - wasClamped: A boolean indicating true if the point was outside and clamped, false otherwise.
func ClampToHyperSphere(p common.Point, maxRadius float64) (clampedPoint common.Point, wasClamped bool) {
	const epsilon = 1e-9 // Small value for floating-point comparisons, e.g., to check if a point is at the origin.

	if maxRadius < 0 { // A negative maxRadius implies no boundary or an invalid scenario.
		// Negative radius implies no boundary.
		return p, false
	}

	// Calculate squared distance from origin.
	distFromOriginSq := 0.0
	isOrigin := true
	for i := range p { // Iterate over dimensions of the point
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
	for i := range p { // Iterate over dimensions of the point
		clampedP[i] = common.Coordinate(float64(p[i]) * scaleFactor)
	}
	return clampedP, true
}

// GenerateRandomPositionInHyperSphere cria uma posição aleatória dentro de uma hiperesfera
// de raio `maxRadius` centrada na origem.
// Utiliza amostragem por rejeição dentro de um hipercubo envolvente.
//
// Parameters:
//   - maxRadius: The radius of the hypersphere. If negative, it's treated as 0.
//   - randomSource: A function that returns a float64 in [0.0, 1.0), used as the source of randomness.
//                  This allows for deterministic testing by providing a seeded RNG.
//
// Returns:
//   - common.Point: A randomly generated point within the specified N-dimensional hypersphere.
//
// Note on Efficiency:
// Rejection sampling becomes inefficient in high dimensions (like 16D used in CrowNet)
// due to the "curse of dimensionality," where the volume of the hypersphere becomes
// a very small fraction of the volume of its enclosing hypercube. This means many
// generated points are rejected, leading to potentially many iterations.
// Alternatives like Marsaglia's method (for 2D/3D) or Muller's method (normalizing points
// generated from a Gaussian distribution) are more efficient for higher dimensions but are more complex to implement.
// For the current scope, rejection sampling is used for its simplicity.
//
// The function below is the NEW, OPTIMIZED version.
// GenerateRandomPositionInHyperSphere cria uma posição aleatória uniformemente distribuída
// dentro de uma hiperesfera N-dimensional (N-ball) de raio `maxRadius` centrada na origem.
// Utiliza o método de gerar N variáveis aleatórias de uma distribuição normal padrão,
// normalizando o vetor resultante para obter um ponto na superfície da N-esfera unitária,
// e então escalando este ponto por um raio R * u^(1/N) para garantir distribuição uniforme por volume.
// Esta abordagem é significativamente mais eficiente para altas dimensões do que a amostragem por rejeição.
//
// Parameters:
//   - maxRadius: O raio da hiperesfera. Se negativo ou zero, a origem é retornada.
//   - rng: Uma fonte de aleatoriedade (`*rand.Rand`) para gerar os números normais e uniformes.
//
// Returns:
//   - common.Point: Um ponto gerado aleatoriamente dentro da hiperesfera N-dimensional especificada.
func GenerateRandomPositionInHyperSphere(maxRadius float64, rng *rand.Rand) common.Point {
	var p common.Point // common.Point is [16]float64

	if maxRadius <= 0 { // Handles negative or zero radius by returning the origin.
		return p // p is {0,0,...,0} by default
	}

	// 1. Gerar N (pointDimension) variáveis aleatórias de uma distribuição normal padrão.
	normDeviates := make([]float64, pointDimension)
	sumSq := 0.0
	for i := 0; i < pointDimension; i++ {
		val := rng.NormFloat64() // Gera um desvio normal padrão (média 0, stddev 1)
		normDeviates[i] = val
		sumSq += val * val
	}

	// Se sumSq for zero (extremamente improvável para N > 1), significa que todos os normDeviates foram zero.
	// Neste caso, o ponto é a origem, que já está dentro da esfera.
	// Também protege contra divisão por zero se sqrt(sumSq) for zero.
	if sumSq == 0 { // ou sumSq < epsilon para robustez de ponto flutuante
		return p // Retorna a origem
	}

	// 2. Normalizar o vetor para obter um ponto na superfície da N-esfera unitária.
	invMagnitude := 1.0 / math.Sqrt(sumSq)
	pointOnUnitSphere := common.Point{}
	for i := 0; i < pointDimension; i++ {
		pointOnUnitSphere[i] = common.Coordinate(normDeviates[i] * invMagnitude)
	}

	// 3. Gerar um raio escalar para o ponto dentro da N-ball.
	// u^(1/N) é usado para garantir distribuição uniforme por volume.
	u := rng.Float64() // Uniforme em [0.0, 1.0)
	// math.Pow(0.0, 1.0/N) é 0.0. math.Pow(u, 1.0/N) para u próximo de 1.0 é próximo de 1.0.
	scaledRadius := maxRadius * math.Pow(u, 1.0/float64(pointDimension))

	// 4. Escalar o ponto na esfera unitária pelo raio calculado.
	for i := 0; i < pointDimension; i++ {
		p[i] = common.Coordinate(float64(pointOnUnitSphere[i]) * scaledRadius)
	}

	return p
}
