package core

import (
	"log"
	"math"
	"math/rand"
	"time"
)

// InitializeNetwork cria e inicializa uma nova rede neural com base na configuração fornecida.
func InitializeNetwork(config Config) *NeuralNetwork {
	rand.Seed(time.Now().UnixNano()) // Seed para aleatoriedade

	if config.NumNeurons <= 0 {
		log.Fatalf("Número de neurônios deve ser positivo.")
	}

	nn := &NeuralNetwork{
		Neurons:               make(map[int]*Neuron),
		Pulses:                make([]*Pulse, 0),
		CortisolGland:         &Gland{Position: config.CortisolGlandPosition, CortisolLevel: 0.1}, // Nível inicial de cortisol
		DopamineLevels:        make(map[int]float64),
		CurrentCycle:          0,
		NumNeurons:            config.NumNeurons,
		PulsePropagationSpeed: config.PulsePropagationSpeed,
		MaxSpaceDistance:      8.0, // Conforme README
	}

	numDopaminergic := int(float64(config.NumNeurons) * config.DopaminergicRatio)
	numInhibitory := int(float64(config.NumNeurons) * config.InhibitoryRatio)
	numExcitatory := int(float64(config.NumNeurons) * config.ExcitatoryRatio)
	numInput := int(float64(config.NumNeurons) * config.InputRatio)
	numOutput := int(float64(config.NumNeurons) * config.OutputRatio)

	// Ajustar para garantir que a soma seja igual a NumNeurons, caso haja arredondamentos
	currentTotal := numDopaminergic + numInhibitory + numExcitatory + numInput + numOutput
	if currentTotal < config.NumNeurons {
		numExcitatory += config.NumNeurons - currentTotal // Adiciona a diferença aos excitatórios
	} else if currentTotal > config.NumNeurons {
		// Lógica para reduzir, se necessário (simplificado, pode ser mais robusto)
		log.Printf("Atenção: Soma das proporções de neurônios excede o total. Ajustando...")
		for currentTotal > config.NumNeurons && numExcitatory > 0 {
			numExcitatory--
			currentTotal--
		}
	}


	neuronIDCounter := 0

	// Gerar neurônios
	assignTypeAndPosition := func(neuronType NeuronType, count int, radiusFactor float64, centralPoint Vector16D) {
		for i := 0; i < count && neuronIDCounter < config.NumNeurons; i++ {
			neuron := &Neuron{
				ID:               neuronIDCounter,
				Type:             neuronType,
				State:            Resting,
				CurrentPotential: 0.0,
				FiringThreshold:  config.DefaultFiringThreshold, // Pode variar por tipo ou ser ajustado depois
				LastFiringCycle:  0,
				CyclesInState:    0,
			}
			neuron.Position = generatePositionInSphere(centralPoint, nn.MaxSpaceDistance*radiusFactor)
			nn.Neurons[neuronIDCounter] = neuron
			nn.DopamineLevels[neuronIDCounter] = 0.0 // Nível inicial de dopamina
			neuronIDCounter++
		}
	}

	// Conforme README:
	// 1% Dopaminérgicos (Raio Maior: 60% do espaço)
	// 30% Inibitórios (Raio Menor: 10% do espaço)
	// 69% Excitatórios (Raio Médio: 30% do espaço)
	// Inputs e Outputs não têm raios especificados, vamos distribuí-los no raio médio por enquanto.

	centerPoint := Vector16D{} // Origem do espaço 16D

	assignTypeAndPosition(Dopaminergic, numDopaminergic, 0.60, centerPoint)
	assignTypeAndPosition(Inhibitory, numInhibitory, 0.10, centerPoint) // Glândula de cortisol é o centro, inibitórios mais próximos
	assignTypeAndPosition(Excitatory, numExcitatory, 0.30, centerPoint)
	assignTypeAndPosition(Input, numInput, 0.30, centerPoint) // Distribuído como excitatórios
	assignTypeAndPosition(Output, numOutput, 0.30, centerPoint) // Distribuído como excitatórios


	// Verifica se todos os neurônios foram criados
	if neuronIDCounter != config.NumNeurons {
		log.Printf("Alerta: %d neurônios criados, esperado %d. Verifique as proporções.", neuronIDCounter, config.NumNeurons)
	}


	log.Printf("Rede Neural inicializada com %d neurônios.", len(nn.Neurons))
	log.Printf("Tipos: %d Dopaminérgicos, %d Inibitórios, %d Excitatórios, %d Input, %d Output",
		numDopaminergic, numInhibitory, numExcitatory, numInput, numOutput)
	log.Printf("Glândula de Cortisol em: %v", nn.CortisolGland.Position)

	return nn
}

// generatePositionInSphere gera uma posição aleatória dentro de uma n-esfera.
// Para simplificar, estamos gerando dentro de um hipercubo e normalizando para a esfera,
// o que não é perfeitamente uniforme na esfera, mas é um começo.
// Uma melhor abordagem seria usar uma distribuição Gaussiana para cada coordenada e normalizar o vetor.
func generatePositionInSphere(center Vector16D, radius float64) Vector16D {
	var p Vector16D
	// Gerador de ruído como OpenNoise seria usado aqui para posições "organizadas".
	// Para o MVP, usamos posições aleatórias dentro de um raio.
	// Gerar um ponto em uma N-esfera uniformemente é não trivial.
	// Abordagem simples (não perfeitamente uniforme, mas suficiente para começar):
	// Gerar cada coordenada entre -radius e +radius, depois verificar se está dentro da esfera.
	// Se não, tentar novamente. Isso pode ser ineficiente para altas dimensões.

	// Alternativa mais comum: gerar a partir de uma distribuição normal e normalizar.
	norm := 0.0
	for i := 0; i < 16; i++ {
		p[i] = rand.NormFloat64() // Amostra de uma distribuição normal padrão
		norm += p[i] * p[i]
	}
	norm = math.Sqrt(norm)

	// Escalar para o raio desejado e adicionar o deslocamento do centro
	// u é um fator aleatório para distribuir pontos *dentro* da esfera, não apenas na superfície.
	// rand.Float64() ^ (1/16.0) para distribuição mais uniforme em volumes N-dimensionais.
	u := math.Pow(rand.Float64(), 1.0/16.0) * radius

	for i := 0; i < 16; i++ {
		if norm == 0 { // Evitar divisão por zero, embora improvável com 16D
			p[i] = center[i]
		} else {
			p[i] = center[i] + (p[i]/norm)*u
		}
	}
	return p
}

// EuclideanDistance calcula a distância Euclidiana entre dois vetores 16D.
// Exportado para que possa ser usado por testes ou outros pacotes, se necessário.
func EuclideanDistance(p1, p2 Vector16D) float64 {
	sumSq := 0.0
	for i := 0; i < 16; i++ {
		diff := p1[i] - p2[i]
		sumSq += diff * diff
	}
	return math.Sqrt(sumSq)
}

// GetDefaultConfig retorna uma configuração padrão para a rede.
func GetDefaultConfig() Config {
	// Posição da glândula de cortisol no centro do espaço.
	// O espaço vetorial é conceitual; não tem um "tamanho" fixo até definirmos
	// como as posições são geradas. Se as posições são, por exemplo, entre -4 e 4,
	// o centro é (0,0,...,0).
	cortisolPos := Vector16D{} // {0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}

	return Config{
		NumNeurons:            100, // Número pequeno para testes iniciais
		DopaminergicRatio:     0.01,
		InhibitoryRatio:       0.30,
		ExcitatoryRatio:       0.59, // Ajustado para que input+output somem 10%
		InputRatio:            0.05,
		OutputRatio:           0.05,
		PulsePropagationSpeed: 0.6, // unidades por ciclo
		DefaultFiringThreshold: 1.0, // Valor arbitrário inicial
		CortisolGlandPosition: cortisolPos,
	}
}
