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

// ClampToHyperSphere garante que um ponto permaneça dentro de uma hiperesfera de um raio_maximo
// a partir da origem (0,0,...,0). Se estiver fora, o ponto é projetado para a superfície da esfera.
// Retorna o novo ponto e um booleano indicando se foi necessário fazer o clamp.
func ClampToHyperSphere(p common.Point, maxRadius float64) (clampedPoint common.Point, wasClamped bool) {
	if maxRadius < 0 { // Não faz sentido ter raio negativo
		// Poderia retornar erro ou pânico. Por ora, retorna o ponto original.
		// Ou talvez um ponto na origem se maxRadius for 0 e p não for a origem.
		// Para simplicidade, se maxRadius < 0, consideramos como se fosse raio infinito (sem clamp).
		// Se maxRadius == 0, apenas a origem é válida.
		if maxRadius == 0 {
			isOrigin := true
			for i := 0; i < pointDimension; i++ {
				if p[i] != 0 {
					isOrigin = false
					break
				}
			}
			if isOrigin { return p, false}

			// Se não for a origem e maxRadius é 0, clampar para a origem.
			var originPoint common.Point
			// originPoint já é {0,0,...}
			return originPoint, true
		}
		// Se maxRadius < 0, não clampar.
		return p, false
	}

	distFromOriginSq := 0.0
	for i := 0; i < pointDimension; i++ {
		coordVal := float64(p[i])
		distFromOriginSq += coordVal * coordVal
	}

	// Usar uma pequena tolerância para comparações de float
	epsilon := 1e-9
	if distFromOriginSq <= maxRadius*maxRadius + epsilon {
		return p, false // Já está dentro ou na superfície (com tolerância)
	}

	distFromOrigin := math.Sqrt(distFromOriginSq)
	if distFromOrigin < epsilon { // Ponto está na origem, mas maxRadius*maxRadius era menor (e.g. maxRadius muito pequeno)
		// Se maxRadius é > 0 mas o ponto na origem está "fora" (devido a epsilon), algo está estranho
		// mas matematicamente, se distFromOrigin é ~0, e maxRadius > 0, deveria ter caído no if anterior.
		// Este caso é mais para se maxRadius for 0 e o ponto não for a origem.
		// Se maxRadius é 0, já foi tratado acima.
		// Se distFromOrigin é efetivamente 0, não há direção para escalar. Retorna a origem.
		var originP common.Point
		return originP, true // Clamped para a origem se maxRadius > 0 mas o ponto está na origem e distSq > maxRadiusSq
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
```
