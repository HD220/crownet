package space

import (
	"crownet/common"
	"math"
)

// EuclideanDistance calcula a distância Euclidiana entre dois pontos no espaço 16D.
func EuclideanDistance(p1, p2 common.Point) float64 {
	var sumOfSquares float64
	for i := 0; i < 16; i++ {
		diff := float64(p1[i] - p2[i])
		sumOfSquares += diff * diff
	}
	return math.Sqrt(sumOfSquares)
}

// IsWithinRadius verifica se um ponto pTest está dentro de um raio de um ponto central pCenter.
func IsWithinRadius(pCenter, pTest common.Point, radius float64) bool {
	return EuclideanDistance(pCenter, pTest) <= radius
}

// ClampToHyperSphere garante que um ponto permaneça dentro de uma hiperesfera de um raio_maximo
// a partir da origem (0,0,...,0). Se estiver fora, o ponto é projetado para a superfície da esfera.
// Retorna o novo ponto e um booleano indicando se foi necessário fazer o clamp.
func ClampToHyperSphere(p common.Point, maxRadius float64) (clampedPoint common.Point, wasClamped bool) {
	distFromOriginSq := 0.0
	for i := 0; i < 16; i++ {
		distFromOriginSq += float64(p[i] * p[i])
	}

	if distFromOriginSq <= maxRadius*maxRadius {
		return p, false // Já está dentro ou na superfície
	}

	distFromOrigin := math.Sqrt(distFromOriginSq)
	scaleFactor := maxRadius / distFromOrigin
	clampedP := common.Point{}
	for i := 0; i < 16; i++ {
		clampedP[i] = common.Coordinate(float64(p[i]) * scaleFactor)
	}
	return clampedP, true
}

// GenerateRandomPositionInHyperSphere cria uma posição aleatória dentro de uma hiperesfera
// de raio `maxRadius` centrada na origem.
// Utiliza amostragem por rejeição dentro de um hipercubo envolvente.
func GenerateRandomPositionInHyperSphere(maxRadius float64, randomSource func() float64) common.Point {
	for {
		var p common.Point
		distSq := 0.0
		for i := 0; i < 16; i++ {
			// Gera coordenada entre -maxRadius e +maxRadius
			coord := (randomSource()*2*maxRadius - maxRadius)
			p[i] = common.Coordinate(coord)
			distSq += coord * coord
		}
		// Se o ponto gerado estiver dentro da hiperesfera, retorna
		if distSq <= maxRadius*maxRadius {
			return p
		}
		// Caso contrário, tenta novamente (rejeição)
	}
}
```
