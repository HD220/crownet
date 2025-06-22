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

	// Garantir minW <= maxW
	if minW > maxW {
		// Trocar ou logar um aviso; por ora, vamos usar um intervalo pequeno default se invertido
		minW, maxW = 0.01, 0.05 // Pequenos pesos positivos default
	}
	if minW == maxW { // Se forem iguais, rand.Float64 * (maxW-minW) será zero.
		// Para permitir alguma variação se minW=maxW, ou simplesmente usar o valor.
		// Se o objetivo é um peso fixo, então maxW-minW = 0 está ok.
		// Se uma pequena variação é desejada mesmo se minW=maxW for configurado,
		// pode-se adicionar um pequeno epsilon a maxW, ou tratar como um caso especial.
		// Por agora, se minW=maxW, todos os pesos serão minW (ou maxW).
	}


	for _, fromID := range neuronIDs {
		if _, exists := nw[fromID]; !exists {
			nw[fromID] = make(WeightMap)
		}
		for _, toID := range neuronIDs {
			if fromID == toID {
				nw[fromID][toID] = 0.0 // Sem auto-conexões diretas
			} else {
				randomFactor := rng.Float64() // Usa o rng local [0.0, 1.0)
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
	return 0.0 // Conexão não encontrada implica peso zero.
}

// SetWeight define o peso da sinapse do neurônio `fromID` para `toID`.
// O peso é clampeado entre 0.0 e simParams.MaxSynapticWeight.
func (nw NetworkWeights) SetWeight(fromID, toID common.NeuronID, weight common.SynapticWeight, simParams *config.SimulationParameters) {
	if _, ok := nw[fromID]; !ok {
		nw[fromID] = make(WeightMap)
	}

	limitedWeight := weight
	// Assumindo que pesos são não-negativos. O sinal do efeito vem do tipo de neurônio/pulso.
	// O limite inferior poderia ser simParams.InitialSynapticWeightMin se pesos devem se manter acima de um mínimo ativo.
	// Por simplicidade e generalidade, 0.0 é um limite inferior comum para magnitude de peso.
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
// A neuromodulação da taxa de aprendizado é tratada externamente (já em effectiveLearningRate).
func (nw NetworkWeights) ApplyHebbianUpdate(
	fromID, toID common.NeuronID,
	preSynapticActivity, postSynapticActivity float64, // Esperado 0.0 ou 1.0
	effectiveLearningRate common.Rate,
	simParams *config.SimulationParameters,
) {
	if fromID == toID {
		return // Não aplicar a auto-conexões
	}

	currentWeight := nw.GetWeight(fromID, toID)

	// LTP (Long-Term Potentiation) - Fortalecimento
	// LTD (Long-Term Depression) - Enfraquecimento (não implementado diretamente aqui, mas pode ser parte da regra de deltaWeight)
	// Esta é uma regra de Hebb simples: fire together, wire together.
	// preActivity * postActivity será 1 se ambos ativos, 0 caso contrário.
	// Fatores de reforço de SimParams podem ser usados para LTP/LTD mais explícitos.
	// deltaWeight := common.SynapticWeight(float64(effectiveLearningRate) * preSynapticActivity * postSynapticActivity * simParams.HebbPositiveReinforceFactor)
	// Por ora, a regra é mais simples, o fator de reforço está implícito na taxa de aprendizado.

	deltaWeight := common.SynapticWeight(0.0)
	if preSynapticActivity > 0 && postSynapticActivity > 0 { // LTP
		deltaWeight = common.SynapticWeight(float64(effectiveLearningRate) * simParams.HebbPositiveReinforceFactor)
	} else if preSynapticActivity > 0 && postSynapticActivity == 0 { // LTD pós-sináptico (pré ativo, pós inativo)
		// deltaWeight = -common.SynapticWeight(float64(effectiveLearningRate) * simParams.HebbNegativeReinforceFactor)
		// Ou outra forma de LTD. Por ora, vamos focar no LTP e decaimento passivo.
		// Para um LTD simples, se pré ativo e pós inativo, pode reduzir o peso.
		// A regra original era apenas LTP. Vamos manter assim e adicionar decaimento.
	}
	// A regra original era: deltaWeight := common.SynapticWeight(float64(effectiveLearningRate) * preSynapticActivity * postSynapticActivity)
	// Vou manter a regra original de LTP e adicionar decaimento passivo.
	// Os fatores HebbPositive/NegativeReinforceFactor podem ser usados para modular a magnitude do delta.
	// Se preSynapticActivity e postSynapticActivity são 0 ou 1:
	change := preSynapticActivity * postSynapticActivity // 1 se ambos ativos, 0 caso contrário
	if change > 0 { // Potenciação
		deltaWeight = common.SynapticWeight(float64(effectiveLearningRate) * simParams.HebbPositiveReinforceFactor * change)
	} else { // Depressão ou nenhum dos dois ativos
		// Poderia haver LTD aqui se pre ativo e pos inativo, ou vice-versa, mas a regra original não tinha.
		// Vamos focar no decaimento passivo que ocorre de qualquer forma.
		// A regra Hebbiana original era mais simples: effectiveLearningRate * pre * post
		// Vou usar a regra original de Hebb para o delta e depois aplicar o decaimento.
		deltaWeight = common.SynapticWeight(float64(effectiveLearningRate) * preSynapticActivity * postSynapticActivity)
	}


	newWeight := currentWeight + deltaWeight

	// Aplicar decaimento de peso passivo (afeta todos os pesos, tendendo a zero)
	// Este decaimento é sutil e acontece independentemente da atividade Hebbiana.
	if simParams.SynapticWeightDecayRate > 0 && currentWeight != 0 { // Não decair pesos zero
		decayAmount := currentWeight * common.SynapticWeight(simParams.SynapticWeightDecayRate)
		if currentWeight > 0 {
			newWeight -= decayAmount
			if newWeight < 0 && currentWeight > 0 { newWeight = 0 } // Evitar que decaimento torne positivo em negativo
		} else { // currentWeight < 0 (se pesos negativos forem permitidos e significativos)
			newWeight -= decayAmount // Se decayAmount for negativo (currentWeight * positivo), subtrair aumenta (move para zero)
			if newWeight > 0 && currentWeight < 0 { newWeight = 0 } // Evitar que decaimento torne negativo em positivo
		}
		// A forma mais simples: newWeight = currentWeight * (1.0 - common.SynapticWeight(simParams.SynapticWeightDecayRate))
		// Mas isso não lida bem com currentWeight + deltaWeight antes do decaimento.
		// A lógica acima aplica o decaimento ao `newWeight` após o delta hebbiano.
		// newWeight = newWeight * (1.0 - common.SynapticWeight(simParams.SynapticWeightDecayRate))
		// Esta forma é mais comum para decaimento passivo aplicado ao peso resultante.
	}
	// Refatorando o decaimento para ser mais claro:
	// O decaimento deve ser aplicado ao peso *antes* ou *depois* da atualização hebbiana.
	// Aplicar ao `newWeight` (após Hebb) é mais comum.
	newWeight = newWeight * (1.0 - common.SynapticWeight(simParams.SynapticWeightDecayRate))


	nw.SetWeight(fromID, toID, newWeight, simParams)
}
```
