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
// Os pesos são aleatórios dentro dos limites definidos em SimParams, e auto-conexões têm peso zero.
func (nw NetworkWeights) InitializeAllToAllWeights(neuronIDs []common.NeuronID, simParams *config.SimulationParameters, rng *rand.Rand) {
	minW := simParams.InitialSynapticWeightMin
	maxW := simParams.InitialSynapticWeightMax

	if minW > maxW {
		minW, maxW = 0.01, 0.05
	}

	for _, fromID := range neuronIDs {
		if _, exists := nw[fromID]; !exists {
			nw[fromID] = make(WeightMap)
		}
		for _, toID := range neuronIDs {
			if fromID == toID {
				nw[fromID][toID] = 0.0
			} else {
				randomFactor := rng.Float64()
				weightValue := minW + randomFactor*(maxW-minW)
				nw[fromID][toID] = common.SynapticWeight(weightValue)
			}
		}
	}
}

// GetWeight retorna o peso da sinapse do neurônio `fromID` para `toID`.
// Retorna 0.0 se a conexão não existir explicitamente, assumindo peso zero para não-conexões.
func (nw NetworkWeights) GetWeight(fromID, toID common.NeuronID) common.SynapticWeight {
	if fromMap, ok := nw[fromID]; ok {
		if weight, ok2 := fromMap[toID]; ok2 {
			return weight
		}
	}
	return 0.0
}

// SetWeight define o peso da sinapse do neurônio `fromID` para `toID`.
// O peso é clampeado entre 0.0 e simParams.MaxSynapticWeight.
func (nw NetworkWeights) SetWeight(fromID, toID common.NeuronID, weight common.SynapticWeight, simParams *config.SimulationParameters) {
	if _, ok := nw[fromID]; !ok {
		nw[fromID] = make(WeightMap)
	}

	limitedWeight := weight
	minPossibleWeight := common.SynapticWeight(0.0)
	maxPossibleWeight := common.SynapticWeight(simParams.MaxSynapticWeight)

	if limitedWeight < minPossibleWeight {
		limitedWeight = minPossibleWeight
	}
	if limitedWeight > maxPossibleWeight {
		limitedWeight = maxPossibleWeight
	}
	nw[fromID][toID] = limitedWeight
}

// ApplyHebbianUpdate atualiza o peso de uma sinapse específica com base na atividade pré e pós-sináptica.
func (nw NetworkWeights) ApplyHebbianUpdate(
	fromID, toID common.NeuronID,
	preSynapticActivity, postSynapticActivity float64,
	effectiveLearningRate common.Rate,
	simParams *config.SimulationParameters,
) {
	if fromID == toID {
		return
	}

	currentWeight := nw.GetWeight(fromID, toID)
	deltaWeight := common.SynapticWeight(0.0)

	if preSynapticActivity > 0 && postSynapticActivity > 0 {
		deltaWeight = common.SynapticWeight(float64(effectiveLearningRate) * simParams.HebbPositiveReinforceFactor)
	} else {
		// Considerar LTD se pre ativo e pos inativo, ou vice-versa.
		// Por ora, apenas LTP e decaimento passivo.
		// Se HebbNegativeReinforceFactor for positivo, este delta seria negativo.
		// deltaWeight = -common.SynapticWeight(float64(effectiveLearningRate) * simParams.HebbNegativeReinforceFactor * (preSynapticActivity + postSynapticActivity)) // Exemplo de LTD
	}

	newWeight := currentWeight + deltaWeight
	newWeight = newWeight * (1.0 - common.SynapticWeight(simParams.SynapticWeightDecayRate))

	nw.SetWeight(fromID, toID, newWeight, simParams)
}
