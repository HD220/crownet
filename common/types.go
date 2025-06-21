package common

// NeuronID representa o identificador único de um neurônio.
type NeuronID int

// CycleCount representa um contador de ciclos da simulação.
type CycleCount int

// Coordinate representa um valor em uma das dimensões do espaço.
type Coordinate float64

// PulseValue representa o valor base de um pulso (antes da ponderação sináptica).
type PulseValue float64

// SynapticWeight representa o peso de uma conexão sináptica.
type SynapticWeight float64

// Percentage representa um valor percentual (0.0 a 1.0).
type Percentage float64

// Rate representa uma taxa (ex: taxa de aprendizado, taxa de decaimento).
type Rate float64

// Factor representa um fator de multiplicação.
type Factor float64

// Threshold representa um valor de limiar.
type Threshold float64

// Level representa o nível de uma substância química.
type Level float64

// Point representa um ponto no espaço N-dimensional (especificamente 16D para CrowNet).
// Usar um array de tamanho fixo garante a dimensionalidade.
type Point [16]Coordinate

// Vector representa um vetor no espaço N-dimensional, usado para velocidade ou força.
type Vector [16]float64
