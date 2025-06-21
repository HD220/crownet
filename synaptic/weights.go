package synaptic

import (
	"crownet/common"
	"crownet/config"
	"math/rand"
)

// WeightMap é um mapa de NeuronID de destino para o peso da sinapse.
type WeightMap map[common.NeuronID]common.SynapticWeight

// NetworkWeights armazena todos os pesos sinápticos na rede.
// A primeira chave é o NeuronID do neurônio pré-sináptico (origem).
// A segunda chave é o NeuronID do neurônio pós-sináptico (destino).
type NetworkWeights map[common.NeuronID]WeightMap

// NewNetworkWeights cria uma nova estrutura NetworkWeights vazia.
func NewNetworkWeights() NetworkWeights {
	return make(NetworkWeights)
}

// InitializeAllToAllWeights configura pesos iniciais entre todos os neurônios fornecidos.
// Os pesos são pequenos e aleatórios, e auto-conexões têm peso zero.
func (nw NetworkWeights) InitializeAllToAllWeights(neuronIDs []common.NeuronID, simParams *config.SimulationParameters) {
	minInitialWeight := simParams.HebbianWeightMin / 10.0 // Exemplo: pesos iniciais menores que os limites
	maxInitialWeight := simParams.HebbianWeightMax / 10.0 // Exemplo: pesos iniciais menores que os limites
	// Garantir que min < max, mesmo que os limites sejam invertidos ou iguais
	if minInitialWeight >= maxInitialWeight {
		minInitialWeight = -0.01
		maxInitialWeight = 0.01
	}


	for _, fromID := range neuronIDs {
		if _, exists := nw[fromID]; !exists {
			nw[fromID] = make(WeightMap)
		}
		for _, toID := range neuronIDs {
			if fromID == toID {
				nw[fromID][toID] = 0.0 // Sem auto-conexões diretas via esta matriz
			} else {
				// Inicializa com peso aleatório pequeno
				// rand.Float64() retorna [0.0, 1.0)
				randomFactor := rand.Float64()
				weightValue := minInitialWeight + randomFactor*(maxInitialWeight-minInitialWeight)
				nw[fromID][toID] = common.SynapticWeight(weightValue)
			}
		}
	}
}

// GetWeight retorna o peso da sinapse do neurônio `fromID` para `toID`.
// Retorna 0.0 se a conexão não existir explicitamente.
func (nw NetworkWeights) GetWeight(fromID, toID common.NeuronID) common.SynapticWeight {
	if fromMap, ok := nw[fromID]; ok {
		if weight, ok2 := fromMap[toID]; ok2 {
			return weight
		}
	}
	return 0.0 // Conexão não encontrada ou não definida, assume peso zero.
}

// SetWeight define o peso da sinapse do neurônio `fromID` para `toID`.
func (nw NetworkWeights) SetWeight(fromID, toID common.NeuronID, weight common.SynapticWeight, simParams *config.SimulationParameters) {
	if _, ok := nw[fromID]; !ok {
		nw[fromID] = make(WeightMap)
	}
	// Aplica limites de peso
	limitedWeight := weight
	if limitedWeight < common.SynapticWeight(simParams.HebbianWeightMin) {
		limitedWeight = common.SynapticWeight(simParams.HebbianWeightMin)
	}
	if limitedWeight > common.SynapticWeight(simParams.HebbianWeightMax) {
		limitedWeight = common.SynapticWeight(simParams.HebbianWeightMax)
	}
	nw[fromID][toID] = limitedWeight
}

// ApplyHebbianUpdate atualiza o peso de uma sinapse específica com base na atividade pré e pós-sináptica.
// Esta função encapsula a regra de Hebb básica, incluindo decaimento.
// A neuromodulação da taxa de aprendizado é tratada externamente.
func (nw NetworkWeights) ApplyHebbianUpdate(
	fromID, toID common.NeuronID,
	preSynapticActivity, postSynapticActivity float64, // Geralmente 0.0 ou 1.0
	effectiveLearningRate common.Rate,
	simParams *config.SimulationParameters,
) {
	if fromID == toID {
		return // Não aplicar a auto-conexões
	}

	currentWeight := nw.GetWeight(fromID, toID)
	deltaWeight := common.SynapticWeight(float64(effectiveLearningRate) * preSynapticActivity * postSynapticActivity)
	newWeight := currentWeight + deltaWeight

	// Aplicar decaimento de peso
	decay := common.SynapticWeight(simParams.HebbianWeightDecay)
	newWeight -= newWeight * decay // newWeight * (1 - decay) seria mais preciso se decay fosse grande

	nw.SetWeight(fromID, toID, newWeight, simParams) // SetWeight já aplica os limites Min/Max
}
```
