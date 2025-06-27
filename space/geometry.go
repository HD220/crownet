// Package space provides geometric types and functions for operating in an
// N-dimensional space, primarily focused on calculations relevant to the
// spatial arrangement and interaction of neurons in the CrowNet simulation.
// It includes utilities for distance calculation, point clamping, and random
// position generation within hyperspheres.
package space

import (
	"math"
	"math/rand"

	"crownet/common"
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

// GenerateRandomPositionInHyperSphere creates a random position uniformly distributed
// within an N-dimensional hypersphere (N-ball) of radius `maxRadius`, centered at the origin.
// This implementation uses a method suitable for high dimensions:
//  1. Generate N standard normal deviates (Gaussian distribution).
//  2. Normalize the resulting N-dimensional vector to get a point on the surface of the unit N-sphere.
//  3. Scale this point by a radius R' = R * u^(1/N), where R is maxRadius and u is a uniform random
//     number in [0,1). This scaling ensures uniform distribution by volume within the N-ball.
//
// This approach is significantly more efficient for high dimensions (like the 16D space
// used in CrowNet, as defined by `pointDimension`) than methods like rejection sampling.
//
// Parameters:
//   - maxRadius: The radius of the hypersphere. If negative or zero, the origin (a zero point) is returned.
//   - rng: A source of randomness (*rand.Rand) for generating normal and uniform random numbers.
//
// Returns:
//
//	A common.Point representing a randomly generated point within the specified N-dimensional hypersphere.
func GenerateRandomPositionInHyperSphere(maxRadius float64, rng *rand.Rand) common.Point {
	var p common.Point // common.Point is [16]float64, initialized to all zeros.

	if maxRadius <= 0 { // Handles negative or zero radius by returning the origin.
		return p // p is {0,0,...,0} by default
	}

	// Step 1: Generate N (pointDimension) random variables from a standard normal distribution.
	normDeviates := make([]float64, pointDimension)
	sumSq := 0.0
	for i := 0; i < pointDimension; i++ {
		// Changed from rng.NormFloat64() to package-level rand.NormFloat64()
		// This uses the global math/rand source, not the passed rng instance for these deviates.
		val := rand.NormFloat64()
		normDeviates[i] = val
		sumSq += val * val
	}

	// If sumSq is zero (extremely unlikely for N > 1), it means all normDeviates were zero.
	// In this case, the point is the origin, which is already within the sphere.
	// This also protects against division by zero if math.Sqrt(sumSq) is zero.
	if sumSq == 0 { // or sumSq < epsilon for floating-point robustness
		return p // Return the origin
	}

	// Step 2: Normalize the vector to obtain a point on the surface of the unit N-sphere.
	invMagnitude := 1.0 / math.Sqrt(sumSq)
	pointOnUnitSphere := common.Point{}
	for i := 0; i < pointDimension; i++ {
		pointOnUnitSphere[i] = common.Coordinate(normDeviates[i] * invMagnitude)
	}

	// Step 3: Generate a scalar radius for the point within the N-ball.
	// u^(1/N) is used to ensure uniform distribution by volume.
	u := rng.Float64() // Uniformly random number in [0.0, 1.0)
	// math.Pow(0.0, 1.0/N) is 0.0. math.Pow(u, 1.0/N) for u close to 1.0 is close to 1.0.
	scaledRadius := maxRadius * math.Pow(u, 1.0/float64(pointDimension))

	// Step 4: Scale the point on the unit sphere by the calculated radius.
	for i := 0; i < pointDimension; i++ {
		p[i] = common.Coordinate(float64(pointOnUnitSphere[i]) * scaledRadius)
	}

	return p
}
