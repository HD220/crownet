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
	"math/rand"
	"sort"
)

// CrowNet é o orquestrador principal da simulação da rede neural.
type CrowNet struct {
	Config *config.AppConfig // Configurações da simulação e CLI

	Neurons            []*neuron.Neuron
	InputNeuronIDs     []common.NeuronID
	OutputNeuronIDs    []common.NeuronID
	neuronIDCounter    common.NeuronID
	CortisolGlandPosition common.Point // Posição da glândula de cortisol

	ActivePulses   *pulse.PulseList // MODIFICADO: Agora é um ponteiro para PulseList
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
func NewCrowNet(appCfg *config.AppConfig) *CrowNet {
	net := &CrowNet{
		Config:                 appCfg,
		Neurons:                make([]*neuron.Neuron, 0, appCfg.Cli.TotalNeurons),
		InputNeuronIDs:         make([]common.NeuronID, 0, appCfg.SimParams.MinInputNeurons),
		OutputNeuronIDs:        make([]common.NeuronID, 0, appCfg.SimParams.MinOutputNeurons),
		neuronIDCounter:        0,
		CortisolGlandPosition: common.Point{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}, // Centro
		ActivePulses:           pulse.NewPulseList(), // MODIFICADO: Inicializa PulseList
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

	net.initializeNeurons()
	allNeuronIDs := make([]common.NeuronID, len(net.Neurons))
	for i, n := range net.Neurons {
		allNeuronIDs[i] = n.ID
	}
	net.SynapticWeights.InitializeAllToAllWeights(allNeuronIDs, &net.Config.SimParams)
	net.finalizeInitialization()

	return net
}

func (cn *CrowNet) getNextNeuronID() common.NeuronID {
	id := cn.neuronIDCounter
	cn.neuronIDCounter++
	return id
}

// initializeNeurons cria e distribui os neurônios na rede.
func (cn *CrowNet) initializeNeurons() {
	totalCLINeurons := cn.Config.Cli.TotalNeurons
	simParams := &cn.Config.SimParams

	numInput := simParams.MinInputNeurons
	numOutput := simParams.MinOutputNeurons

	if totalCLINeurons < numInput+numOutput {
		// Em caso de poucos neurônios, priorizar Input/Output e ajustar o restante.
		// Ou lançar erro. Por enquanto, vamos ajustar.
		fmt.Printf("Aviso: Total de neurônios (%d) é menor que o mínimo para Input (%d) e Output (%d). Ajustando para o mínimo necessário.\n",
			totalCLINeurons, numInput, numOutput)
		totalCLINeurons = numInput + numOutput
	}

	// Neurônios de Input
	for i := 0; i < numInput; i++ {
		id := cn.getNextNeuronID()
		// Posição: ExcitatoryMaxRadius é um bom proxy para uma área geral.
		pos := space.GenerateRandomPositionInHyperSphere(simParams.ExcitatoryRadiusFactor*simParams.SpaceMaxDimension, rand.Float64)
		n := neuron.New(id, neuron.Input, pos, simParams)
		cn.Neurons = append(cn.Neurons, n)
		cn.InputNeuronIDs = append(cn.InputNeuronIDs, id)
	}

	// Neurônios de Output
	for i := 0; i < numOutput; i++ {
		id := cn.getNextNeuronID()
		pos := space.GenerateRandomPositionInHyperSphere(simParams.ExcitatoryRadiusFactor*simParams.SpaceMaxDimension, rand.Float64)
		n := neuron.New(id, neuron.Output, pos, simParams)
		cn.Neurons = append(cn.Neurons, n)
		cn.OutputNeuronIDs = append(cn.OutputNeuronIDs, id)
	}

	remainingNeuronsForDistribution := totalCLINeurons - numInput - numOutput

	numDopaminergic := int(float64(remainingNeuronsForDistribution) * simParams.DopaminergicPercent)
	numInhibitory := int(float64(remainingNeuronsForDistribution) * simParams.InhibitoryPercent)
	// Excitatory pega o que sobrar para garantir o total de neurônios.
	numExcitatory := remainingNeuronsForDistribution - numDopaminergic - numInhibitory

	if numExcitatory < 0 { // Pode acontecer se as porcentagens somarem > 1 ou com arredondamentos.
	    fmt.Printf("Aviso: contagem de neurônios excitatórios negativa (%d) após distribuição. Ajustando para 0 e redistribuindo ligeiramente.\n", numExcitatory)
		// Simplificação: apenas zera e aceita um total menor se as porcentagens forem problemáticas.
		// Uma lógica mais robusta normalizaria as porcentagens ou ajustaria outras contagens.
		numExcitatory = 0
		// Recalcular dopa e inhib para preencher o que falta, se possível, ou aceitar menos neurônios.
		// Esta parte pode ser mais sofisticada. Por ora, a simplicidade prevalece.
	}


	for i := 0; i < numDopaminergic; i++ {
		id := cn.getNextNeuronID()
		pos := space.GenerateRandomPositionInHyperSphere(simParams.DopaminergicRadiusFactor*simParams.SpaceMaxDimension, rand.Float64)
		n := neuron.New(id, neuron.Dopaminergic, pos, simParams)
		cn.Neurons = append(cn.Neurons, n)
	}

	for i := 0; i < numInhibitory; i++ {
		id := cn.getNextNeuronID()
		pos := space.GenerateRandomPositionInHyperSphere(simParams.InhibitoryRadiusFactor*simParams.SpaceMaxDimension, rand.Float64)
		n := neuron.New(id, neuron.Inhibitory, pos, simParams)
		cn.Neurons = append(cn.Neurons, n)
	}

	for i := 0; i < numExcitatory; i++ {
		id := cn.getNextNeuronID()
		pos := space.GenerateRandomPositionInHyperSphere(simParams.ExcitatoryRadiusFactor*simParams.SpaceMaxDimension, rand.Float64)
		n := neuron.New(id, neuron.Excitatory, pos, simParams)
		cn.Neurons = append(cn.Neurons, n)
	}

	// Se o número total de neurônios criados não bate com totalCLINeurons devido a arredondamentos/ajustes:
	currentCreatedCount := len(cn.Neurons)
	if currentCreatedCount < totalCLINeurons {
	    // Adiciona neurônios excitatórios faltantes
	    fmt.Printf("Aviso: Criados %d neurônios, esperado %d. Adicionando %d neurônios excitatórios extras.\n", currentCreatedCount, totalCLINeurons, totalCLINeurons - currentCreatedCount)
	    for i := 0; i < (totalCLINeurons - currentCreatedCount); i++ {
	        id := cn.getNextNeuronID()
		    pos := space.GenerateRandomPositionInHyperSphere(simParams.ExcitatoryRadiusFactor*simParams.SpaceMaxDimension, rand.Float64)
		    n := neuron.New(id, neuron.Excitatory, pos, simParams)
		    cn.Neurons = append(cn.Neurons, n)
	    }
	} else if currentCreatedCount > totalCLINeurons {
	    // Remove neurônios excitatórios extras (ou do último tipo adicionado)
	     fmt.Printf("Aviso: Criados %d neurônios, esperado %d. Removendo %d neurônios extras.\n", currentCreatedCount, totalCLINeurons, currentCreatedCount - totalCLINeurons)
	    cn.Neurons = cn.Neurons[:totalCLINeurons]
	    // Nota: Isso pode bagunçar InputNeuronIDs/OutputNeuronIDs se eles forem os últimos a serem removidos.
	    // A lógica de criação de I/O primeiro mitiga isso.
	}

}

// finalizeInitialization realiza tarefas de configuração final após a criação dos neurônios.
func (cn *CrowNet) finalizeInitialization() {
	sort.Slice(cn.InputNeuronIDs, func(i, j int) bool { return cn.InputNeuronIDs[i] < cn.InputNeuronIDs[j] })
	sort.Slice(cn.OutputNeuronIDs, func(i, j int) bool { return cn.OutputNeuronIDs[i] < cn.OutputNeuronIDs[j] })

	// Inicializa o histórico de disparos para neurônios de output
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
func (cn *CrowNet) RunCycle() {
	// 0. Processar inputs baseados em frequência (para modo 'sim')
	cn.processFrequencyInputs()

	// 1. Atualizar estados dos neurônios e decair potencial acumulado
	for _, n := range cn.Neurons {
		n.DecayPotential(&cn.Config.SimParams)
		n.AdvanceState(cn.CycleCount, &cn.Config.SimParams) // Avança estados refratários, etc.
	}

	// 2. Propagar pulsos ativos e processar seus efeitos
	cn.processActivePulses() // MODIFICADO: processActivePulses agora não retorna e PulseList gerencia os pulsos

	// 3. Atualizar ambiente neuroquímico e aplicar seus efeitos
	if cn.isChemicalModulationEnabled {
		cn.ChemicalEnv.UpdateLevels(cn.Neurons, cn.ActivePulses, cn.CortisolGlandPosition, &cn.Config.SimParams)
		cn.ChemicalEnv.ApplyEffectsToNeurons(cn.Neurons, &cn.Config.SimParams)
	} else {
		// Se químicos estão desligados, garantir que fatores de modulação sejam neutros
		// e limiares dos neurônios estejam no valor base.
		cn.ChemicalEnv.LearningRateModulationFactor = 1.0
		cn.ChemicalEnv.SynaptogenesisModulationFactor = 1.0
		for _, n := range cn.Neurons {
			n.CurrentFiringThreshold = n.BaseFiringThreshold
		}
	}

	// 4. Aplicar aprendizado Hebbiano (se habilitado)
	if cn.isLearningEnabled {
		cn.applyHebbianLearning()
	}

	// 5. Aplicar sinaptogênese (movimento de neurônios, se habilitado)
	if cn.isSynaptogenesisEnabled {
		cn.applySynaptogenesis()
	}

	cn.CycleCount++
}

// processActivePulses gerencia o processamento de pulsos pela PulseList e lida com os novos pulsos gerados.
func (cn *CrowNet) processActivePulses() {
	// Delega o processamento do ciclo de pulsos para PulseList
	// PulseList.ProcessCycle irá iterar, propagar, aplicar efeitos e retornar os novos pulsos.
	newlyGeneratedPulses := cn.ActivePulses.ProcessCycle(
		cn.Neurons,
		cn.SynapticWeights,
		cn.CycleCount,
		&cn.Config.SimParams,
		// spaceOps // Se tivéssemos uma interface spaceOps, seria passada aqui
	)

	// Lidar com os pulsos recém-gerados
	if len(newlyGeneratedPulses) > 0 {
		for _, newP := range newlyGeneratedPulses {
			// Verificar se o neurônio emissor é um neurônio de output para registrar o disparo
			// A struct Pulse (newP) tem EmittingNeuronID
			for _, outID := range cn.OutputNeuronIDs {
				if newP.EmittingNeuronID == outID {
					cn.recordOutputFiring(outID) // Passa o ID do neurônio de output
					break
				}
			}
		}
		// Adicionar todos os novos pulsos à lista de ativos gerenciada por PulseList
		cn.ActivePulses.AddAll(newlyGeneratedPulses)
	}
}

// applyHebbianLearning itera sobre as conexões e aplica a regra de aprendizado.
func (cn *CrowNet) applyHebbianLearning() {
	// A taxa de aprendizado efetiva já considera a modulação química (via ChemicalEnv)
	effectiveLR := common.Rate(float64(cn.Config.Cli.BaseLearningRate) * float64(cn.ChemicalEnv.LearningRateModulationFactor))

	if effectiveLR < 1e-9 { // Praticamente zero, não fazer nada
		return
	}

	coincidenceWindow := common.CycleCount(cn.Config.SimParams.HebbianCoincidenceWindow)

	// Iterar sobre todos os neurônios como possíveis pré-sinápticos
	for _, preSynapticNeuron := range cn.Neurons {
		// Verificar atividade pré-sináptica
		// Disparou no ciclo atual ou na janela de coincidência?
		// LastFiredCycle é o ciclo em que o neurônio *completou* seu disparo e entrou em refratário.
		// Se CycleCount é o ciclo atual, e LastFiredCycle == CycleCount, significa que disparou no ciclo anterior
		// e está sendo processado neste ciclo.
		// Se LastFiredCycle == CycleCount-1, disparou há 1 ciclo.
		// Se LastFiredCycle == CycleCount, e estamos no meio do ProcessPulses, ele acabou de disparar.
		// A lógica precisa ser consistente: cn.CycleCount é o ciclo *prestes a ser concluído*.
		// Neuron.LastFiredCycle é o ciclo em que o estado Firing foi processado em AbsoluteRefractory.

		preActivity := 0.0
		// Se o neurônio disparou neste ciclo (seu estado é Firing e ainda não foi para AdvanceState),
		// ou se disparou recentemente dentro da janela.
		// Simples: se LastFiredCycle está dentro de [CycleCount - window, CycleCount]
		// (Considerando que LastFiredCycle é atualizado quando o neurônio *termina* de disparar)
		// Se CycleCount é o ciclo atual N, e o neurônio disparou no ciclo N-1, LastFiredCycle = N-1.
		// (cn.CycleCount - preSynapticNeuron.LastFiredCycle) <= coincidenceWindow

		// Para Hebbian, a atividade é geralmente se o neurônio disparou *neste* ciclo de processamento de inputs
		// ou no ciclo imediatamente anterior.
		// Se preSynapticNeuron.CurrentState == neuron.Firing, ele disparou neste ciclo.
		// Se preSynapticNeuron.LastFiredCycle == cn.CycleCount (após AdvanceState), ele disparou no ciclo que acabou de passar.
		// A forma mais robusta é olhar o LastFiredCycle.
		if preSynapticNeuron.LastFiredCycle != -1 && (cn.CycleCount - preSynapticNeuron.LastFiredCycle <= coincidenceWindow) {
		    preActivity = 1.0
		}


		if preActivity == 0.0 {
			continue // Neurônio pré-sináptico não esteve ativo recentemente
		}

		// Iterar sobre todos os neurônios como possíveis pós-sinápticos
		for _, postSynapticNeuron := range cn.Neurons {
			if preSynapticNeuron.ID == postSynapticNeuron.ID {
				continue // Sem auto-aprendizado direto na mesma sinapse
			}

			postActivity := 0.0
			if postSynapticNeuron.LastFiredCycle != -1 && (cn.CycleCount - postSynapticNeuron.LastFiredCycle <= coincidenceWindow) {
			    postActivity = 1.0
			}

			if postActivity > 0 { // Apenas atualizar se ambos estiverem ativos
				cn.SynapticWeights.ApplyHebbianUpdate(
					preSynapticNeuron.ID,
					postSynapticNeuron.ID,
					preActivity,
					postActivity,
					effectiveLR,
					&cn.Config.SimParams,
				)
			}
		}
	}
}

// applySynaptogenesis move os neurônios com base na atividade da rede e modulação química.
func (cn *CrowNet) applySynaptogenesis() {
	modulationFactor := float64(cn.ChemicalEnv.SynaptogenesisModulationFactor)
	if modulationFactor < 1e-6 { // Movimento efetivamente desligado
		return
	}

	tempNewPositions := make(map[common.NeuronID]common.Point)
	tempNewVelocities := make(map[common.NeuronID]common.Vector)

	simParams := &cn.Config.SimParams

	for _, n1 := range cn.Neurons {
		netForce := common.Vector{} // Acumulador para a força líquida em n1

		for _, n2 := range cn.Neurons {
			if n1.ID == n2.ID {
				continue
			}

			distance := space.EuclideanDistance(n1.Position, n2.Position)
			if distance == 0 || (simParams.SynaptogenesisInfluenceRadius > 0 && distance > simParams.SynaptogenesisInfluenceRadius) {
				continue // Muito longe ou sobreposto (não deveria acontecer com IDs diferentes)
			}

			// Vetor unitário de n1 para n2
			directionUnitVector := common.Vector{}
			for i := 0; i < 16; i++ {
				directionUnitVector[i] = float64(n2.Position[i]-n1.Position[i]) / distance
			}

			forceMagnitude := 0.0
			// Neurônios são atraídos por neurônios ativos (Firing, Refractory)
			// e repelidos por neurônios em repouso (Resting).
			if n2.CurrentState == neuron.Firing || n2.CurrentState == neuron.AbsoluteRefractory || n2.CurrentState == neuron.RelativeRefractory {
				forceMagnitude = simParams.AttractionForceFactor * modulationFactor // Atração
			} else if n2.CurrentState == neuron.Resting {
				forceMagnitude = -simParams.RepulsionForceFactor * modulationFactor // Repulsão (força negativa na direção de n2)
			}

			for i := 0; i < 16; i++ {
				netForce[i] += directionUnitVector[i] * forceMagnitude
			}
		}

		// Atualizar velocidade: v_new = v_old * damping + netForce (dt = 1)
		newVelocity := common.Vector{}
		currentVelocityMagnitudeSq := 0.0
		for i := 0; i < 16; i++ {
			newVelocity[i] = n1.Velocity[i]*simParams.DampeningFactor + netForce[i]
			currentVelocityMagnitudeSq += newVelocity[i] * newVelocity[i]
		}

		// Limitar velocidade máxima
		currentVelocityMagnitude := math.Sqrt(currentVelocityMagnitudeSq)
		if currentVelocityMagnitude > simParams.MaxMovementPerCycle {
		    scale := simParams.MaxMovementPerCycle / currentVelocityMagnitude
		    for i := 0; i < 16; i++ {
		        newVelocity[i] *= scale
		    }
		}
		tempNewVelocities[n1.ID] = newVelocity

		// Atualizar posição: p_new = p_old + v_new (dt = 1)
		newPosition := n1.Position
		for i := 0; i < 16; i++ {
			newPosition[i] += common.Coordinate(newVelocity[i])
		}

		// Manter dentro dos limites do espaço (hiperesfera)
		clampedPosition, _ := space.ClampToHyperSphere(newPosition, simParams.SpaceMaxDimension)
		tempNewPositions[n1.ID] = clampedPosition
	}

	// Aplicar todas as novas posições e velocidades calculadas
	for _, n := range cn.Neurons {
		n.Position = tempNewPositions[n.ID]
		n.Velocity = tempNewVelocities[n.ID]
	}
}

// --- Métodos de I/O e Controle de Modo ---

// processFrequencyInputs lida com disparos de neurônios de input baseados em frequência.
func (cn *CrowNet) processFrequencyInputs() {
    for neuronID, timeLeft := range cn.timeToNextInputFire {
        newTimeLeft := timeLeft - 1
        cn.timeToNextInputFire[neuronID] = newTimeLeft

        if newTimeLeft <= 0 {
            var inputNeuron *neuron.Neuron
            // Encontrar o neurônio de input correspondente
            for _, n := range cn.Neurons {
                if n.ID == neuronID && n.Type == neuron.Input {
                    inputNeuron = n
                    break
                }
            }

            if inputNeuron != nil {
                // Forçar o disparo do neurônio de input
                inputNeuron.CurrentState = neuron.Firing
                // O LastFiredCycle será atualizado em AdvanceState
                // (Não precisa chamar IntegrateIncomingPotential pois é um disparo forçado)

                emittedSignal := inputNeuron.EmittedPulseSign()
                if emittedSignal != 0 {
                    newP := pulse.New(
                        inputNeuron.ID,
                        inputNeuron.Position,
                        emittedSignal,
                        cn.CycleCount,
                        cn.Config.SimParams.SpaceMaxDimension*2.0,
                    )
					cn.ActivePulses.Add(newP) // MODIFICADO: Usa Add de PulseList
                }
            }
            // Resetar o timer para o próximo disparo
            targetHz := cn.inputTargetFrequencies[neuronID]
            if targetHz > 0 {
                cyclesPerFiring := cn.Config.SimParams.CyclesPerSecond / targetHz
                cn.timeToNextInputFire[neuronID] = common.CycleCount(math.Max(1.0, math.Round(cyclesPerFiring)))
            } else {
                delete(cn.timeToNextInputFire, neuronID) // Frequência zero ou negativa, remover da programação
            }
        }
    }
}

// recordOutputFiring registra o disparo de um neurônio de output.
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

    // Manter o histórico dentro da janela de frequência
    cutoffCycle := cn.CycleCount - common.CycleCount(cn.Config.SimParams.OutputFrequencyWindowCycles)
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
func (cn *CrowNet) PresentPattern(patternData []float64) error {
	if len(patternData) != cn.Config.SimParams.PatternSize {
		return fmt.Errorf("tamanho do padrão (%d) diferente do esperado (%d)", len(patternData), cn.Config.SimParams.PatternSize)
	}
	if len(cn.InputNeuronIDs) < cn.Config.SimParams.PatternSize {
		return fmt.Errorf("não há neurônios de input suficientes (%d) para o tamanho do padrão (%d)", len(cn.InputNeuronIDs), cn.Config.SimParams.PatternSize)
	}

	for i := 0; i < cn.Config.SimParams.PatternSize; i++ {
		if patternData[i] > 0.5 { // Considerar "ativo"
			inputNeuronID := cn.InputNeuronIDs[i] // Assume que os IDs estão ordenados e os primeiros são usados
			var targetNeuron *neuron.Neuron
			for _, n := range cn.Neurons {
				if n.ID == inputNeuronID {
					targetNeuron = n
					break
				}
			}
			if targetNeuron != nil && targetNeuron.Type == neuron.Input {
				targetNeuron.CurrentState = neuron.Firing // Força o disparo
				// targetNeuron.LastFiredCycle = cn.CycleCount; // Será atualizado em AdvanceState

				emittedSignal := targetNeuron.EmittedPulseSign()
                if emittedSignal != 0 {
                    newP := pulse.New(
                        targetNeuron.ID,
                        targetNeuron.Position,
                        emittedSignal,
                        cn.CycleCount, // Pulso criado no ciclo atual
                        cn.Config.SimParams.SpaceMaxDimension*2.0,
                    )
					cn.ActivePulses.Add(newP) // MODIFICADO: Usa Add de PulseList
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
        // Não resetar LastFiredCycle aqui, pois é histórico para aprendizado.
        // Resetar o estado para Resting se não for Input (Input pode ser forçado a Firing).
        // A lógica de `AdvanceState` já lida com transições de Firing/Refractory.
        // A preocupação é se um neurônio ficou "preso" em um estado de um padrão anterior.
        // Para `expose`, queremos um "fresh start" para cada padrão, mas não apagar o histórico de disparo recente
        // que pode ser usado pelo Hebbian learning (que olha para trás alguns ciclos).
        // A forma mais simples é deixar AdvanceState e DecayPotential limparem naturalmente,
        // e PresentPattern força os inputs a disparar.
        // Se um neurônio não input estiver em Firing/Refractory de um ciclo anterior de *outro* padrão,
        // isso pode ser um problema.
        // Para o MVP, vamos apenas resetar o potencial. O estado evoluirá.
    }
	cn.ActivePulses.Clear() // MODIFICADO: Usa o método Clear de PulseList
}


// GetOutputActivation retorna a ativação (potencial acumulado) dos neurônios de output.
func (cn *CrowNet) GetOutputActivation() ([]float64, error) {
	if len(cn.OutputNeuronIDs) < cn.Config.SimParams.MinOutputNeurons {
		return nil, fmt.Errorf("número de neurônios de output (%d) é menor que o esperado (%d)", len(cn.OutputNeuronIDs), cn.Config.SimParams.MinOutputNeurons)
	}

	outputActivations := make([]float64, cn.Config.SimParams.MinOutputNeurons)
	for i := 0; i < cn.Config.SimParams.MinOutputNeurons; i++ {
		outputNeuronID := cn.OutputNeuronIDs[i] // Assume que os IDs estão ordenados
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
			// Isso não deveria acontecer se OutputNeuronIDs estiver correto
			return nil, fmt.Errorf("neurônio de output ID %d (esperado na posição %d da lista de outputs) não encontrado ou não é do tipo Output", outputNeuronID, i)
		}
	}
	return outputActivations, nil
}

// ConfigureFrequencyInput define a frequência de disparo para um neurônio de input específico.
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
		cyclesPerFiring := cn.Config.SimParams.CyclesPerSecond / hz
		// Iniciar disparo no próximo ciclo ou distribuir aleatoriamente o primeiro disparo
		cn.timeToNextInputFire[neuronID] = common.CycleCount(math.Max(1.0, math.Round(rand.Float64()*cyclesPerFiring)+1))
	}
	return nil
}

// GetOutputFrequency calcula a frequência de disparo de um neurônio de output nos últimos N ciclos.
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
		return 0.0, nil // Sem histórico de disparos
	}

	// A janela de cálculo é definida por OutputFrequencyWindowCycles
	// O histórico já é podado em recordOutputFiring para manter apenas essa janela.
	// Então, o número de disparos em `history` é o número de disparos na janela.
	firingsInWindow := len(history)
	windowDurationSeconds := cn.Config.SimParams.OutputFrequencyWindowCycles / cn.Config.SimParams.CyclesPerSecond

	if windowDurationSeconds == 0 {
		return 0, fmt.Errorf("OutputFrequencyWindowCycles ou CyclesPerSecond é zero, não é possível calcular frequência")
	}

	frequencyHz := float64(firingsInWindow) / windowDurationSeconds
	return frequencyHz, nil
}
```
