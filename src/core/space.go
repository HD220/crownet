package core

import "math"

// DistanceEuclidean calcula a distância Euclidiana entre dois pontos no espaço N-dimensional.
func DistanceEuclidean(p1, p2 [SpaceDimensions]float64) float64 {
	var sumSqDiff float64
	for i := 0; i < SpaceDimensions; i++ {
		diff := p1[i] - p2[i]
		sumSqDiff += diff * diff
	}
	return math.Sqrt(sumSqDiff)
}

// AddVectors adiciona dois vetores.
func AddVectors(v1, v2 [SpaceDimensions]float64) [SpaceDimensions]float64 {
	var result [SpaceDimensions]float64
	for i := 0; i < SpaceDimensions; i++ {
		result[i] = v1[i] + v2[i]
	}
	return result
}

// SubtractVectors subtrai o vetor v2 de v1.
func SubtractVectors(v1, v2 [SpaceDimensions]float64) [SpaceDimensions]float64 {
	var result [SpaceDimensions]float64
	for i := 0; i < SpaceDimensions; i++ {
		result[i] = v1[i] - v2[i]
	}
	return result
}

// ScaleVector multiplica um vetor por um escalar.
func ScaleVector(v [SpaceDimensions]float64, scalar float64) [SpaceDimensions]float64 {
	var result [SpaceDimensions]float64
	for i := 0; i < SpaceDimensions; i++ {
		result[i] = v[i] * scalar
	}
	return result
}

// NormalizeVector normaliza um vetor para ter magnitude 1.
// Retorna um vetor zero se a magnitude for zero.
func NormalizeVector(v [SpaceDimensions]float64) [SpaceDimensions]float64 {
	var magnitude float64
	for _, val := range v {
		magnitude += val * val
	}
	magnitude = math.Sqrt(magnitude)

	if magnitude == 0 {
		return [SpaceDimensions]float64{} // Retorna vetor zero
	}

	var result [SpaceDimensions]float64
	for i := 0; i < SpaceDimensions; i++ {
		result[i] = v[i] / magnitude
	}
	return result
}

// ClampPosition garante que a posição de um neurônio permaneça dentro dos limites do espaço.
// Assumindo que o espaço é um cubo de 0 a spaceSize em cada dimensão.
func ClampPosition(pos [SpaceDimensions]float64, spaceSize float64) [SpaceDimensions]float64 {
	var clampedPos [SpaceDimensions]float64
	for i := 0; i < SpaceDimensions; i++ {
		if pos[i] < 0 {
			clampedPos[i] = 0
		} else if pos[i] > spaceSize {
			clampedPos[i] = spaceSize
		} else {
			clampedPos[i] = pos[i]
		}
	}
	return clampedPos
}
