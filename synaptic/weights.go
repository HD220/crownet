package synaptic

import (
	"crownet/common"
	"crownet/config"
	"fmt"
	"math/rand"
)

// WeightMap representa um mapa de NeuronID de destino para o peso da sinapse.
// A chave é o NeuronID do neurônio pós-sináptico (destino).
type WeightMap map[common.NeuronID]common.SynapticWeight

// NetworkWeights armazena e gerencia todos os pesos sinápticos na rede.
// Ele encapsula o mapa de pesos e os parâmetros de simulação relevantes.
type NetworkWeights struct {
	weights   map[common.NeuronID]WeightMap
	simParams *config.SimulationParameters
	rng       *rand.Rand
}

// NewNetworkWeights cria e retorna uma nova instância de NetworkWeights.
// Requer os parâmetros de simulação e uma fonte de aleatoriedade (rng).
func NewNetworkWeights(simParams *config.SimulationParameters, rng *rand.Rand) (*NetworkWeights, error) {
	if simParams == nil {
		return nil, fmt.Errorf("NewNetworkWeights: simParams não pode ser nulo")
	}
	if rng == nil {
		return nil, fmt.Errorf("NewNetworkWeights: rng não pode ser nulo")
	}
	return &NetworkWeights{
		weights:   make(map[common.NeuronID]WeightMap),
		simParams: simParams,
		rng:       rng,
	}, nil
}

// InitializeAllToAllWeights configura pesos iniciais entre todos os neurônios fornecidos.
// Os pesos são aleatórios dentro dos limites definidos em simParams (InitialSynapticWeightMin/Max).
// Auto-conexões (de um neurônio para si mesmo) são inicializadas com peso zero.
func (nw *NetworkWeights) InitializeAllToAllWeights(neuronIDs []common.NeuronID) {
	minW := nw.simParams.InitialSynapticWeightMin
	maxW := nw.simParams.InitialSynapticWeightMax

	// Validação básica dos limites de peso inicial (idealmente, isso também é feito no config.Validate)
	if minW >= maxW {
		// Fallback ou log de aviso se os parâmetros estiverem inconsistentes
		// Por ora, usamos um fallback seguro, mas isso deve ser um erro de configuração.
		minW = 0.01
		maxW = 0.05
		// Considerar logar um aviso aqui: log.Printf("Aviso: InitialSynapticWeightMin (%f) >= InitialSynapticWeightMax (%f). Usando fallback %f-%f.", nw.simParams.InitialSynapticWeightMin, nw.simParams.InitialSynapticWeightMax, minW, maxW)
	}

	for _, fromID := range neuronIDs {
		if _, exists := nw.weights[fromID]; !exists {
			nw.weights[fromID] = make(WeightMap)
		}
		for _, toID := range neuronIDs {
			if fromID == toID {
				nw.weights[fromID][toID] = 0.0
			} else {
				randomFactor := nw.rng.Float64() // Usa o rng da struct
				weightValue := minW + randomFactor*(maxW-minW)
				nw.weights[fromID][toID] = common.SynapticWeight(weightValue)
			}
		}
	}
}

// GetWeight retorna o peso da sinapse do neurônio `fromID` para `toID`.
// Retorna 0.0 se a conexão não existir explicitamente, assumindo peso zero para não-conexões.
func (nw *NetworkWeights) GetWeight(fromID, toID common.NeuronID) common.SynapticWeight {
	if fromMap, ok := nw.weights[fromID]; ok {
		if weight, ok2 := fromMap[toID]; ok2 {
			return weight
		}
	}
	// Retorna 0.0 para conexões não existentes, o que é um comportamento comum
	// em redes esparsas ou para simplificar a lógica de cálculo de potencial.
	return 0.0
}

// SetWeight define o peso da sinapse do neurônio `fromID` para `toID`.
// O peso é clampeado entre 0.0 e simParams.MaxSynapticWeight.
// Nota: Para regras de aprendizado específicas (como Hebbian), os limites HebbianWeightMin/Max
// devem ser aplicados *antes* de chamar este SetWeight geral, se necessário.
// Este método aplica um clamp global definido por MaxSynapticWeight e um mínimo de 0.
func (nw *NetworkWeights) SetWeight(fromID, toID common.NeuronID, weight common.SynapticWeight) {
	if _, ok := nw.weights[fromID]; !ok {
		nw.weights[fromID] = make(WeightMap)
	}

	limitedWeight := weight
	// Usa os limites globais definidos nos parâmetros de simulação.
	// HebbianWeightMin pode ser negativo, mas MaxSynapticWeight é o teto absoluto.
	// Um peso sináptico geralmente não é menor que 0 a menos que explicitamente permitido
	// por regras como HebbianWeightMin. Este SetWeight é um setter mais genérico.
	// Para este setter genérico, vamos assumir que o peso não pode ser negativo a menos que
	// o HebbianWeightMin seja explicitamente usado e seja negativo.
	// A lógica original usava 0.0 como mínimo. Vamos manter isso para o setter genérico,
	// e a lógica Hebbiana pode definir pesos negativos se HebbianMin for < 0.
	// O clamp mais importante aqui é o MaxSynapticWeight.
	minApplicableWeight := common.SynapticWeight(0.0) // Defaulting to non-negative for general set
	if nw.simParams.HebbianWeightMin < 0 {             // Allow negative if Hebbian rules permit
		minApplicableWeight = common.SynapticWeight(nw.simParams.HebbianWeightMin)
	}

	maxApplicableWeight := common.SynapticWeight(nw.simParams.MaxSynapticWeight)

	if limitedWeight < minApplicableWeight && fromID != toID { // Auto-conexões são sempre 0
		limitedWeight = minApplicableWeight
	}
	if limitedWeight > maxApplicableWeight {
		limitedWeight = maxApplicableWeight
	}

	// Auto-conexões devem sempre ser zero
	if fromID == toID {
		nw.weights[fromID][toID] = 0.0
	} else {
		nw.weights[fromID][toID] = limitedWeight
	}
}

// ApplyHebbianUpdate atualiza o peso de uma sinapse específica com base na atividade pré e pós-sináptica.
// A atualização considera a taxa de aprendizado efetiva e os fatores de reforço definidos
// nos parâmetros de simulação. O peso resultante é então clampeado.
func (nw *NetworkWeights) ApplyHebbianUpdate(
	fromID, toID common.NeuronID,
	preSynapticActivity, postSynapticActivity float64, // Valores > 0 indicam atividade
	effectiveLearningRate common.Rate,
) {
	if fromID == toID { // Sem auto-plasticidade
		return
	}

	currentWeight := nw.GetWeight(fromID, toID)
	deltaWeight := common.SynapticWeight(0.0)

	// LTP - Long-Term Potentiation
	if preSynapticActivity > 0 && postSynapticActivity > 0 {
		// Aumenta o peso se ambos os neurônios estiverem ativos
		deltaWeight = common.SynapticWeight(float64(effectiveLearningRate) * nw.simParams.HebbPositiveReinforceFactor)
	} else {
		// LTD - Long-Term Depression (opcional, baseado em parâmetros)
		// Exemplo: se HebbNegativeReinforceFactor > 0 e um dos neurônios (mas não ambos) está ativo.
		// Esta parte da lógica pode ser expandida conforme os requisitos.
		// if nw.simParams.HebbNegativeReinforceFactor > 0 && (preSynapticActivity > 0 || postSynapticActivity > 0) {
		//    deltaWeight = -common.SynapticWeight(float64(effectiveLearningRate) * nw.simParams.HebbNegativeReinforceFactor)
		// }
	}

	newWeight := currentWeight + deltaWeight

	// Aplica decaimento passivo do peso
	if nw.simParams.SynapticWeightDecayRate > 0 {
		newWeight *= (1.0 - common.SynapticWeight(nw.simParams.SynapticWeightDecayRate))
	}

	// Clampeia o novo peso usando os limites específicos para aprendizado Hebbiano
	clampedHebbianWeight := newWeight
	if clampedHebbianWeight < common.SynapticWeight(nw.simParams.HebbianWeightMin) {
		clampedHebbianWeight = common.SynapticWeight(nw.simParams.HebbianWeightMin)
	}
	if clampedHebbianWeight > common.SynapticWeight(nw.simParams.HebbianWeightMax) {
		clampedHebbianWeight = common.SynapticWeight(nw.simParams.HebbianWeightMax)
	}

	// Usa o SetWeight geral que aplica o clamp global MaxSynapticWeight.
	// Isso garante que mesmo o peso Hebbiano clampeado não exceda o máximo absoluto da rede.
	nw.SetWeight(fromID, toID, clampedHebbianWeight)
}

// GetAllWeights retorna uma cópia profunda de todos os pesos na rede.
// Útil para salvar o estado ou para análise externa.
func (nw *NetworkWeights) GetAllWeights() map[common.NeuronID]WeightMap {
	copiedWeights := make(map[common.NeuronID]WeightMap)
	for fromID, toMap := range nw.weights {
		copiedDestMap := make(WeightMap)
		for toID, weight := range toMap {
			copiedDestMap[toID] = weight
		}
		copiedWeights[fromID] = copiedDestMap
	}
	return copiedWeights
}

// LoadWeights carrega um mapa de pesos na estrutura NetworkWeights,
// substituindo quaisquer pesos existentes.
func (nw *NetworkWeights) LoadWeights(weightsToLoad map[common.NeuronID]WeightMap) {
	nw.weights = make(map[common.NeuronID]WeightMap) // Limpa pesos existentes
	for fromID, toMap := range weightsToLoad {
		nw.weights[fromID] = make(WeightMap)
		for toID, weight := range toMap {
			// Ao carregar, os pesos devem ser definidos diretamente sem clamp adicional,
			// assumindo que os pesos salvos já são válidos.
			// Ou, aplicar SetWeight para garantir que os pesos carregados estejam dentro dos limites atuais.
			// Optando por SetWeight para consistência com os limites da simulação atual.
			nw.SetWeight(fromID, toID, weight)
		}
	}
}
