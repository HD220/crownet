package network

import (
	"crownet/common"
	"crownet/config"
	"crownet/datagen"
	"crownet/neuron"
	"crownet/neurochemical"
	"crownet/pulse"
	"crownet/space"
	"crownet/synaptic"
	"fmt"
	"math" // Adicionado
	"math/rand"
	"sort"
)

// CrowNet é o orquestrador principal da simulação da rede neural.
type CrowNet struct {
	SimParams        *config.SimulationParameters // Parâmetros da simulação
	baseLearningRate common.Rate                  // Taxa de aprendizado base da CLI
	rng              *rand.Rand                   // Fonte de aleatoriedade local

	Neurons            []*neuron.Neuron
	InputNeuronIDs     []common.NeuronID
	OutputNeuronIDs    []common.NeuronID
	neuronIDCounter    common.NeuronID
	CortisolGlandPosition common.Point // Posição da glândula de cortisol

	ActivePulses   *pulse.PulseList
	SynapticWeights synaptic.NetworkWeights
	ChemicalEnv    *neurochemical.Environment

	CycleCount common.CycleCount

	// Campos para I/O de frequência (modo 'sim')
	inputTargetFrequencies map[common.NeuronID]float64
	timeToNextInputFire    map[common.NeuronID]common.CycleCount
	outputFiringHistory    map[common.NeuronID][]common.CycleCount

	// Controle de dinâmicas para modos específicos (ex: 'observe')
	isLearningEnabled      bool
	isSynaptogenesisEnabled bool
	isChemicalModulationEnabled bool
}

// NewCrowNet cria e inicializa uma nova instância de CrowNet.
func NewCrowNet(totalNeurons int, baseLR common.Rate, simParams *config.SimulationParameters, seed int64) *CrowNet {
	localRng := rand.New(rand.NewSource(seed))
	net := &CrowNet{
		SimParams:              simParams,
		baseLearningRate:       baseLR,
		rng:                    localRng,
		Neurons:                make([]*neuron.Neuron, 0, totalNeurons),
		InputNeuronIDs:         make([]common.NeuronID, 0, simParams.MinInputNeurons),
		OutputNeuronIDs:        make([]common.NeuronID, 0, simParams.MinOutputNeurons),
		neuronIDCounter:        0,
		CortisolGlandPosition:  common.Point{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // Centro
		ActivePulses:           pulse.NewPulseList(),
		SynapticWeights:        synaptic.NewNetworkWeights(),
		ChemicalEnv:            neurochemical.NewEnvironment(),
		CycleCount:             0,
		inputTargetFrequencies: make(map[common.NeuronID]float64),
		timeToNextInputFire:    make(map[common.NeuronID]common.CycleCount),
		outputFiringHistory:    make(map[common.NeuronID][]common.CycleCount),
		isLearningEnabled:      true, // Habilitado por padrão
		isSynaptogenesisEnabled: true, // Habilitado por padrão
		isChemicalModulationEnabled: true, // Habilitado por padrão
	}

	net.initializeNeurons(totalNeurons) // Usa cn.rng internamente agora
	allNeuronIDs := make([]common.NeuronID, len(net.Neurons))
	for i, n := range net.Neurons {
		allNeuronIDs[i] = n.ID
	}
	net.SynapticWeights.InitializeAllToAllWeights(allNeuronIDs, net.SimParams, net.rng) // Passa o rng
	net.finalizeInitialization()

	return net
}

func (cn *CrowNet) getNextNeuronID() common.NeuronID {
	id := cn.neuronIDCounter
	cn.neuronIDCounter++
	return id
}

// addNeuronsOfType cria 'count' neurônios do 'neuronType' especificado,
// os posiciona usando o 'radiusFactor' e os adiciona à rede.
// Também popula InputNeuronIDs e OutputNeuronIDs conforme necessário.
// Modificado para usar cn.SimParams
func (cn *CrowNet) addNeuronsOfType(count int, neuronType neuron.Type, radiusFactor float64) {
	if count <= 0 {
		return
	}
	for i := 0; i < count; i++ {
		id := cn.getNextNeuronID()
		effectiveRadius := radiusFactor * cn.SimParams.SpaceMaxDimension
		pos := space.GenerateRandomPositionInHyperSphere(effectiveRadius, cn.rng.Float64) // Usa cn.rng

		n := neuron.New(id, neuronType, pos, cn.SimParams)
		cn.Neurons = append(cn.Neurons, n)

		if neuronType == neuron.Input {
			cn.InputNeuronIDs = append(cn.InputNeuronIDs, id)
		} else if neuronType == neuron.Output {
			cn.OutputNeuronIDs = append(cn.OutputNeuronIDs, id)
		}
	}
}

// calculateInternalNeuronCounts calcula a distribuição de neurônios internos.
// Retorna as contagens e uma lista de strings de aviso.
func calculateInternalNeuronCounts(remainingForDistribution int, dopaP, inhibP float64) (numDopaminergic, numInhibitory, numExcitatory int, warnings []string) {
	if remainingForDistribution <= 0 {
		return 0, 0, 0, nil
	}

	dopaP = math.Max(0.0, dopaP) // Garantir que os percentuais não sejam negativos.
	inhibP = math.Max(0.0, inhibP)

	numDopaminergic = int(math.Floor(float64(remainingForDistribution) * dopaP))
	numInhibitory = int(math.Floor(float64(remainingForDistribution) * inhibP))

	currentAllocated := numDopaminergic + numInhibitory
	numExcitatory = remainingForInternalDistribution - currentAllocated

	if numExcitatory < 0 {
		// O warning sobre percentuais excedendo 100% é agora tratado pela validação em config.AppConfig.Validate()
		// No entanto, a lógica de ajuste ainda é necessária aqui caso a validação seja contornada ou
		// se os percentuais forem válidos individualmente mas sua soma para o remainingForDistribution cause numExcitatory < 0.
		// A mensagem de warning pode ser removida daqui se a validação de config for considerada suficiente.
		// Por ora, manteremos a lógica de ajuste, mas o warning pode ser opcional.
		// warnings = append(warnings, fmt.Sprintf("Ajuste interno: Percentuais de Dopa (%.2f) e Inhib (%.2f) recalculados para caber no espaço restante.", dopaP, inhibP))
		numExcitatory = 0
		if dopaP+inhibP > 0 { // Evitar divisão por zero
			totalInternalPercentConfigured := dopaP + inhibP
			numDopaminergic = int(math.Round(float64(remainingForInternalDistribution) * (dopaP / totalInternalPercentConfigured)))
			numInhibitory = remainingForInternalDistribution - numDopaminergic // Inhib absorve arredondamento
		} else {
			// Se ambos os percentuais configurados eram 0, mas numExcitatory deu < 0 (o que não deveria acontecer aqui),
			// zera ambos por segurança.
			numDopaminergic = 0
			numInhibitory = 0
		}
	}
	return
}

// initializeNeurons cria e distribui os neurônios na rede.
func (cn *CrowNet) initializeNeurons(totalNeuronsInput int) {
	simParams := cn.SimParams
	actualTotalNeurons := totalNeuronsInput

	numInput := simParams.MinInputNeurons
	numOutput := simParams.MinOutputNeurons

	if actualTotalNeurons < numInput+numOutput {
		warningMsg := fmt.Sprintf("Aviso: Total de neurônios configurado (%d) é menor que o mínimo para Input (%d) e Output (%d). Ajustando para o mínimo necessário (%d).",
			actualTotalNeurons, numInput, numOutput, numInput+numOutput)
		fmt.Println(warningMsg) // Ou logar de forma mais estruturada
		actualTotalNeurons = numInput + numOutput
	}

	cn.addNeuronsOfType(numInput, neuron.Input, simParams.ExcitatoryRadiusFactor)
	cn.addNeuronsOfType(numOutput, neuron.Output, simParams.ExcitatoryRadiusFactor)

	remainingForInternalDistribution := actualTotalNeurons - numInput - numOutput

	numDopaminergic, numInhibitory, numExcitatory, calcWarnings := calculateInternalNeuronCounts(
		remainingForInternalDistribution,
		simParams.DopaminergicPercent,
		simParams.InhibitoryPercent,
	)

	for _, w := range calcWarnings {
		fmt.Println("Aviso:", w) // Ou logar de forma mais estruturada
	}

	cn.addNeuronsOfType(numDopaminergic, neuron.Dopaminergic, simParams.DopaminergicRadiusFactor)
	cn.addNeuronsOfType(numInhibitory, neuron.Inhibitory, simParams.InhibitoryRadiusFactor)
	cn.addNeuronsOfType(numExcitatory, neuron.Excitatory, simParams.ExcitatoryRadiusFactor)

	if len(cn.Neurons) != actualTotalNeurons {
		// Este log é crucial e deve ser tratado com seriedade se aparecer.
		fmt.Printf("ALERTA CRÍTICO: Contagem final de neurônios (%d) não corresponde ao esperado (%d) em initializeNeurons.\n", len(cn.Neurons), actualTotalNeurons)
	}
}

// finalizeInitialization realiza tarefas de configuração final após a criação dos neurônios.
func (cn *CrowNet) finalizeInitialization() {
	sort.Slice(cn.InputNeuronIDs, func(i, j int) bool { return cn.InputNeuronIDs[i] < cn.InputNeuronIDs[j] })
	sort.Slice(cn.OutputNeuronIDs, func(i, j int) bool { return cn.OutputNeuronIDs[i] < cn.OutputNeuronIDs[j] })

	for _, outID := range cn.OutputNeuronIDs {
		cn.outputFiringHistory[outID] = make([]common.CycleCount, 0)
	}
}

// SetDynamicState permite ligar/desligar as principais dinâmicas da rede.
// Útil para o modo 'observe' ou para depuração.
func (cn *CrowNet) SetDynamicState(learning, synaptogenesis, chemicalModulation bool) {
	cn.isLearningEnabled = learning
	cn.isSynaptogenesisEnabled = synaptogenesis
	cn.isChemicalModulationEnabled = chemicalModulation
}

// RunCycle executa um único ciclo de simulação da rede.
// Modificado para usar cn.SimParams
func (cn *CrowNet) RunCycle() {
	cn.processFrequencyInputs()

	for _, n := range cn.Neurons {
		n.DecayPotential(cn.SimParams)
		n.AdvanceState(cn.CycleCount, cn.SimParams)
	}

	cn.processActivePulses()

	if cn.isChemicalModulationEnabled {
		// Passar cn.ActivePulses.GetAll() pois UpdateLevels espera []*pulse.Pulse
		cn.ChemicalEnv.UpdateLevels(cn.Neurons, cn.ActivePulses.GetAll(), cn.CortisolGlandPosition, cn.SimParams)
		cn.ChemicalEnv.ApplyEffectsToNeurons(cn.Neurons, cn.SimParams)
	} else {
		cn.ChemicalEnv.LearningRateModulationFactor = 1.0
		cn.ChemicalEnv.SynaptogenesisModulationFactor = 1.0
		for _, n := range cn.Neurons {
			n.CurrentFiringThreshold = n.BaseFiringThreshold
		}
	}

	if cn.isLearningEnabled {
		cn.applyHebbianLearning()
	}

	if cn.isSynaptogenesisEnabled {
		cn.applySynaptogenesis()
	}

	cn.CycleCount++
}

// processActivePulses gerencia o processamento de pulsos pela PulseList e lida com os novos pulsos gerados.
// Modificado para usar cn.SimParams
func (cn *CrowNet) processActivePulses() {
	newlyGeneratedPulses := cn.ActivePulses.ProcessCycle(
		cn.Neurons,
		cn.SynapticWeights,
		cn.CycleCount,
		cn.SimParams,
	)

	if len(newlyGeneratedPulses) > 0 {
		for _, newP := range newlyGeneratedPulses {
			for _, outID := range cn.OutputNeuronIDs {
				if newP.EmittingNeuronID == outID {
					cn.recordOutputFiring(outID)
					break
				}
			}
		}
		cn.ActivePulses.AddAll(newlyGeneratedPulses)
	}
}

// applyHebbianLearning itera sobre as conexões e aplica a regra de aprendizado.
// Modificado para usar cn.baseLearningRate e cn.SimParams
func (cn *CrowNet) applyHebbianLearning() {
	effectiveLR := common.Rate(float64(cn.baseLearningRate) * float64(cn.ChemicalEnv.LearningRateModulationFactor))

	if effectiveLR < 1e-9 {
		return
	}

	coincidenceWindow := common.CycleCount(cn.SimParams.HebbianCoincidenceWindow)

	for _, preSynapticNeuron := range cn.Neurons {
		preActivity := 0.0
		if preSynapticNeuron.LastFiredCycle != -1 && (cn.CycleCount-preSynapticNeuron.LastFiredCycle <= coincidenceWindow) {
			preActivity = 1.0
		}

		if preActivity == 0.0 {
			continue
		}

		for _, postSynapticNeuron := range cn.Neurons {
			if preSynapticNeuron.ID == postSynapticNeuron.ID {
				continue
			}

			postActivity := 0.0
			if postSynapticNeuron.LastFiredCycle != -1 && (cn.CycleCount-postSynapticNeuron.LastFiredCycle <= coincidenceWindow) {
				postActivity = 1.0
			}

			if postActivity > 0 {
				cn.SynapticWeights.ApplyHebbianUpdate(
					preSynapticNeuron.ID,
					postSynapticNeuron.ID,
					preActivity,
					postActivity,
					effectiveLR,
					cn.SimParams,
				)
			}
		}
	}
}

// calculateNetForceOnNeuron calcula a força líquida exercida sobre n1 pelos outros neurônios.
func calculateNetForceOnNeuron(n1 *neuron.Neuron, allNeurons []*neuron.Neuron, simParams *config.SimulationParameters, modulationFactor float64) common.Vector {
	netForce := common.Vector{}
	for _, n2 := range allNeurons {
		if n1.ID == n2.ID {
			continue
		}

		distance := space.EuclideanDistance(n1.Position, n2.Position)
		// Ignorar se muito longe ou sobreposto (distância zero implica mesmo ponto, mas IDs diferentes)
		if distance == 0 || (simParams.SynaptogenesisInfluenceRadius > 0 && distance > simParams.SynaptogenesisInfluenceRadius) {
			continue
		}

		directionUnitVector := common.Vector{}
		for i := 0; i < 16; i++ { // Assuming 16D space
			directionUnitVector[i] = float64(n2.Position[i]-n1.Position[i]) / distance
		}

		forceMagnitude := 0.0
		// Neurônios são atraídos por neurônios ativos (Firing, Refractory)
		// e repelidos por neurônios em repouso (Resting).
		if n2.CurrentState == neuron.Firing || n2.CurrentState == neuron.AbsoluteRefractory || n2.CurrentState == neuron.RelativeRefractory {
			forceMagnitude = simParams.AttractionForceFactor * modulationFactor // Atração
		} else if n2.CurrentState == neuron.Resting {
			forceMagnitude = -simParams.RepulsionForceFactor * modulationFactor // Repulsão
		}

		for i := 0; i < 16; i++ { // Assuming 16D space
			netForce[i] += directionUnitVector[i] * forceMagnitude
		}
	}
	return netForce
}

// updateNeuronMovement calcula a nova posição e velocidade de um neurônio com base na força líquida.
func updateNeuronMovement(n *neuron.Neuron, netForce common.Vector, simParams *config.SimulationParameters) (newPosition common.Point, newVelocity common.Vector) {
	// Atualizar velocidade: v_new = v_old * damping + netForce (assumindo dt = 1)
	currentVelocity := n.Velocity
	updatedVelocity := common.Vector{}
	velocityMagnitudeSq := 0.0
	for i := 0; i < 16; i++ { // Assuming 16D space
		updatedVelocity[i] = currentVelocity[i]*simParams.DampeningFactor + netForce[i]
		velocityMagnitudeSq += updatedVelocity[i] * updatedVelocity[i]
	}

	// Limitar velocidade máxima
	velocityMagnitude := math.Sqrt(velocityMagnitudeSq)
	if velocityMagnitude > simParams.MaxMovementPerCycle {
		scaleFactor := simParams.MaxMovementPerCycle / velocityMagnitude
		for i := 0; i < 16; i++ { // Assuming 16D space
			updatedVelocity[i] *= scaleFactor
		}
	}
	newVelocity = updatedVelocity

	// Atualizar posição: p_new = p_old + v_new (assumindo dt = 1)
	currentPosition := n.Position
	calculatedPosition := currentPosition
	for i := 0; i < 16; i++ { // Assuming 16D space
		calculatedPosition[i] += common.Coordinate(newVelocity[i])
	}

	// Manter dentro dos limites do espaço (hiperesfera)
	clampedPosition, _ := space.ClampToHyperSphere(calculatedPosition, simParams.SpaceMaxDimension)
	newPosition = clampedPosition
	return
}

// applySynaptogenesis move os neurônios com base na atividade da rede e modulação química.
func (cn *CrowNet) applySynaptogenesis() {
	modulationFactor := float64(cn.ChemicalEnv.SynaptogenesisModulationFactor)
	if modulationFactor < 1e-6 { // Movimento efetivamente desligado
		return
	}

	tempNewPositions := make(map[common.NeuronID]common.Point)
	tempNewVelocities := make(map[common.NeuronID]common.Vector)

	for _, n1 := range cn.Neurons {
		netForce := calculateNetForceOnNeuron(n1, cn.Neurons, cn.SimParams, modulationFactor)
		newPos, newVel := updateNeuronMovement(n1, netForce, cn.SimParams)
		tempNewPositions[n1.ID] = newPos
		tempNewVelocities[n1.ID] = newVel
	}

	// Aplicar todas as novas posições e velocidades calculadas
	for _, n := range cn.Neurons {
		n.Position = tempNewPositions[n.ID]
		n.Velocity = tempNewVelocities[n.ID]
	}
}

// --- Métodos de I/O e Controle de Modo ---

// processFrequencyInputs lida com disparos de neurônios de input baseados em frequência.
// Modificado para usar cn.SimParams
func (cn *CrowNet) processFrequencyInputs() {
	for neuronID, timeLeft := range cn.timeToNextInputFire {
		newTimeLeft := timeLeft - 1
		cn.timeToNextInputFire[neuronID] = newTimeLeft

		if newTimeLeft <= 0 {
			var inputNeuron *neuron.Neuron
			for _, n := range cn.Neurons {
				if n.ID == neuronID && n.Type == neuron.Input {
					inputNeuron = n
					break
				}
			}

			if inputNeuron != nil {
				inputNeuron.CurrentState = neuron.Firing
				emittedSignal := inputNeuron.EmittedPulseSign()
				if emittedSignal != 0 {
					newP := pulse.New(
						inputNeuron.ID,
						inputNeuron.Position,
						emittedSignal,
						cn.CycleCount,
						cn.SimParams.SpaceMaxDimension*2.0,
					)
					cn.ActivePulses.Add(newP)
				}
			}
			targetHz := cn.inputTargetFrequencies[neuronID]
			if targetHz > 0 {
				cyclesPerFiring := cn.SimParams.CyclesPerSecond / targetHz
				cn.timeToNextInputFire[neuronID] = common.CycleCount(math.Max(1.0, math.Round(cyclesPerFiring)))
			} else {
				delete(cn.timeToNextInputFire, neuronID)
			}
		}
	}
}

// recordOutputFiring registra o disparo de um neurônio de output.
// Modificado para usar cn.SimParams
func (cn *CrowNet) recordOutputFiring(neuronID common.NeuronID) {
	isOutput := false
	for _, id := range cn.OutputNeuronIDs {
		if id == neuronID {
			isOutput = true
			break
		}
	}
	if !isOutput {
		return
	}

	history, exists := cn.outputFiringHistory[neuronID]
	if !exists {
		history = make([]common.CycleCount, 0)
	}
	history = append(history, cn.CycleCount)

	cutoffCycle := cn.CycleCount - common.CycleCount(cn.SimParams.OutputFrequencyWindowCycles)
	prunedHistory := make([]common.CycleCount, 0, len(history))
	for _, fireCycle := range history {
		if fireCycle >= cutoffCycle {
			prunedHistory = append(prunedHistory, fireCycle)
		}
	}
	cn.outputFiringHistory[neuronID] = prunedHistory
}

// --- Funções para os modos de operação (`expose`, `observe`) ---

// PresentPattern ativa os neurônios de input com base no padrão fornecido.
// Modificado para usar cn.SimParams
func (cn *CrowNet) PresentPattern(patternData []float64) error {
	if len(patternData) != cn.SimParams.PatternSize {
		return fmt.Errorf("tamanho do padrão (%d) diferente do esperado (%d)", len(patternData), cn.SimParams.PatternSize)
	}
	if len(cn.InputNeuronIDs) < cn.SimParams.PatternSize {
		return fmt.Errorf("não há neurônios de input suficientes (%d) para o tamanho do padrão (%d)", len(cn.InputNeuronIDs), cn.SimParams.PatternSize)
	}

	for i := 0; i < cn.SimParams.PatternSize; i++ {
		if patternData[i] > 0.5 {
			inputNeuronID := cn.InputNeuronIDs[i]
			var targetNeuron *neuron.Neuron
			for _, n := range cn.Neurons {
				if n.ID == inputNeuronID {
					targetNeuron = n
					break
				}
			}
			if targetNeuron != nil && targetNeuron.Type == neuron.Input {
				targetNeuron.CurrentState = neuron.Firing
				emittedSignal := targetNeuron.EmittedPulseSign()
				if emittedSignal != 0 {
					newP := pulse.New(
						targetNeuron.ID,
						targetNeuron.Position,
						emittedSignal,
						cn.CycleCount,
						cn.SimParams.SpaceMaxDimension*2.0,
					)
					cn.ActivePulses.Add(newP)
				}
			} else {
				return fmt.Errorf("neurônio de input ID %d não encontrado ou não é do tipo Input", inputNeuronID)
			}
		}
	}
	return nil
}

// ResetNetworkState limpa potenciais acumulados, pulsos ativos e reseta estados de neurônios.
// Usado antes de apresentar um novo padrão no modo 'expose'.
func (cn *CrowNet) ResetNetworkStateForNewPattern() {
	for _, n := range cn.Neurons {
		n.AccumulatedPotential = 0.0
	}
	cn.ActivePulses.Clear()
}

// GetOutputActivation retorna a ativação (potencial acumulado) dos neurônios de output.
// Modificado para usar cn.SimParams
func (cn *CrowNet) GetOutputActivation() ([]float64, error) {
	if len(cn.OutputNeuronIDs) < cn.SimParams.MinOutputNeurons {
		return nil, fmt.Errorf("número de neurônios de output (%d) é menor que o esperado (%d)", len(cn.OutputNeuronIDs), cn.SimParams.MinOutputNeurons)
	}

	outputActivations := make([]float64, cn.SimParams.MinOutputNeurons)
	for i := 0; i < cn.SimParams.MinOutputNeurons; i++ {
		outputNeuronID := cn.OutputNeuronIDs[i]
		var targetNeuron *neuron.Neuron
		for _, n := range cn.Neurons {
			if n.ID == outputNeuronID {
				targetNeuron = n
				break
			}
		}
		if targetNeuron != nil && targetNeuron.Type == neuron.Output {
			outputActivations[i] = float64(targetNeuron.AccumulatedPotential)
		} else {
			return nil, fmt.Errorf("neurônio de output ID %d (esperado na posição %d da lista de outputs) não encontrado ou não é do tipo Output", outputNeuronID, i)
		}
	}
	return outputActivations, nil
}

// ConfigureFrequencyInput define a frequência de disparo para um neurônio de input específico.
// Modificado para usar cn.SimParams
func (cn *CrowNet) ConfigureFrequencyInput(neuronID common.NeuronID, hz float64) error {
	isInput := false
	for _, id := range cn.InputNeuronIDs {
		if id == neuronID {
			isInput = true
			break
		}
	}
	if !isInput {
		return fmt.Errorf("neurônio ID %d não é um neurônio de input válido", neuronID)
	}

	if hz <= 0 {
		delete(cn.inputTargetFrequencies, neuronID)
		delete(cn.timeToNextInputFire, neuronID)
	} else {
		cn.inputTargetFrequencies[neuronID] = hz
		cyclesPerFiring := cn.SimParams.CyclesPerSecond / hz
		cn.timeToNextInputFire[neuronID] = common.CycleCount(math.Max(1.0, math.Round(cn.rng.Float64()*cyclesPerFiring)+1)) // Usa cn.rng
	}
	return nil
}

// GetOutputFrequency calcula a frequência de disparo de um neurônio de output nos últimos N ciclos.
// Modificado para usar cn.SimParams
func (cn *CrowNet) GetOutputFrequency(neuronID common.NeuronID) (float64, error) {
	isOutput := false
	for _, id := range cn.OutputNeuronIDs {
		if id == neuronID {
			isOutput = true
			break
		}
	}
	if !isOutput {
		return 0, fmt.Errorf("neurônio ID %d não é um neurônio de output válido", neuronID)
	}

	history, exists := cn.outputFiringHistory[neuronID]
	if !exists || len(history) == 0 {
		return 0.0, nil
	}

	firingsInWindow := len(history)
	windowDurationSeconds := cn.SimParams.OutputFrequencyWindowCycles / cn.SimParams.CyclesPerSecond

	if windowDurationSeconds == 0 {
		return 0, fmt.Errorf("OutputFrequencyWindowCycles ou CyclesPerSecond é zero, não é possível calcular frequência")
	}

	frequencyHz := float64(firingsInWindow) / windowDurationSeconds
	return frequencyHz, nil
}
