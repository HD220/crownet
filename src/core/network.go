package core

import (
	"log"
)

// SimulateCycle executa um único ciclo de simulação da rede neural.
func (nn *NeuralNetwork) SimulateCycle() {
	nn.CurrentCycle++
	log.Printf("--- Iniciando Ciclo %d ---", nn.CurrentCycle)

	// 1. Propagar pulsos existentes, aplicar seus efeitos aos neurônios alvo
	// e identificar quais neurônios são atingidos.
	// Esta função também lida com a geração de dopamina por pulsos dopaminérgicos
	// e estímulo à glândula de cortisol por pulsos excitatórios.
	// A função `UpdatePulsesAndProcessTargets` foi redesenhada para incorporar isso.
	// Ela também coleta novos pulsos gerados por disparos de neurônios.
	nn.UpdatePulsesAndProcessTargets() // Esta função agora retorna void e modifica nn.Pulses diretamente

	// 2. Atualizar o estado de cada neurônio (disparo, refratário, etc.)
	// Esta etapa foi integrada em UpdatePulsesAndProcessTargets, que chama neuron.UpdateNeuronState.
	// Os novos pulsos gerados já foram adicionados a nn.Pulses.

	// 3. Aplicar decaimento de substâncias químicas (Cortisol, Dopamina)
	nn.decayChemicals()

	// 4. Calcular o fator de modulação global do limiar de disparo
	// Este fator é usado por neuron.UpdateNeuronState.
	// Idealmente, este cálculo deveria preceder a atualização do estado do neurônio.
	// Vamos ajustar a ordem:
	//    a. Calcular modulações (cortisol, dopamina) e seus efeitos nos limiares.
	//    b. Propagar pulsos e aplicar seus efeitos (potencial).
	//    c. Atualizar estado dos neurônios (disparar se limiar atingido).
	//    d. Decair químicas.
	// A estrutura atual em `pulse.go` e `neuron.go` já segue uma lógica similar:
	// `UpdatePulsesAndProcessTargets` em `pulse.go`:
	//    - Propaga pulsos.
	//    - Aplica pulsos aos alvos (mudando potencial ou dopamina local).
	//    - Estimula glândula de cortisol.
	//    - Chama `neuron.UpdateNeuronState` para cada neurônio, que:
	//        - Usa o `globalFiringThresholdFactor` (que precisa ser calculado *antes* desta chamada).
	//        - Verifica disparo, gera novo pulso.
	//        - Atualiza estado (refratário, etc.).
	//    - Adiciona novos pulsos à lista da rede.
	// `decayChemicals` é chamado depois.

	// Para corrigir a ordem, o `globalFiringThresholdFactor` deve ser calculado ANTES
	// de `neuron.UpdateNeuronState` ser chamado dentro de `UpdatePulsesAndProcessTargets`.
	// Vamos passar o fator calculado para `UpdatePulsesAndProcessTargets`.

	// Recalculando a ordem correta das operações dentro de SimulateCycle:
	// nn.CurrentCycle++
	// log.Printf("--- Iniciando Ciclo %d ---", nn.CurrentCycle)

	// ETAPA A: Calcular o fator de modulação do limiar de disparo com base nos níveis atuais
	// de cortisol e dopamina (que são do final do ciclo anterior).
	globalFactor := nn.calculateGlobalFiringThresholdFactor() // Esta função está em pulse.go

	// ETAPA B & C: Propagar pulsos, aplicar efeitos, atualizar neurônios e gerar novos pulsos.
	// Esta função interna agora precisa do globalFactor.
	// A função `UpdatePulsesAndProcessTargets` em `pulse.go` precisa ser ajustada para aceitar `globalFactor`
	// e passá-lo para `neuron.UpdateNeuronState`.
	nn.propagateAndUpdateNetworkStates(globalFactor) // Nova função wrapper em network.go

	// ETAPA D: Aplicar decaimento de substâncias químicas para o próximo ciclo.
	nn.decayChemicals() // Esta função está em pulse.go

	// ETAPA E: Sinaptogênese (movimentação dos neurônios)
	nn.ApplySynaptogenesis()

	log.Printf("--- Fim do Ciclo %d ---", nn.CurrentCycle)
	log.Printf("Nível de Cortisol: %.4f", nn.CortisolGland.CortisolLevel)
	// Poderia logar média de dopamina ou outros status da rede.
}


// propagateAndUpdateNetworkStates é uma função interna para encapsular a lógica de atualização de pulsos e neurônios.
// Ela utiliza o globalFiringThresholdFactor calculado no início do ciclo.
func (nn *NeuralNetwork) propagateAndUpdateNetworkStates(globalFiringThresholdFactor float64) {
    // Lista para armazenar pulsos que ainda estão em trânsito após este ciclo
    stillActivePulses := make([]*Pulse, 0, len(nn.Pulses))
    // Mapa para agregar pulsos que chegam ao mesmo neurônio alvo neste ciclo
    pulsesArrivingAtTarget := make(map[int][]*Pulse) // neuronID -> lista de pulsos

    // 1. Propagar pulsos existentes e identificar chegadas
    for _, p := range nn.Pulses {
        if p.Processed { continue }

        // Lógica de propagação baseada em "casca esférica"
        cyclesSinceEmission := nn.CurrentCycle - p.EmittedCycle
        distCoveredEnd := float64(cyclesSinceEmission) * nn.PulsePropagationSpeed
        distCoveredStart := float64(cyclesSinceEmission-1) * nn.PulsePropagationSpeed
        if cyclesSinceEmission == 0 { // Primeiro ciclo de propagação (ou ciclo de emissão)
            distCoveredStart = 0
			// Se um pulso é emitido e processado no mesmo ciclo, cyclesSinceEmission pode ser 0.
			// Se o pulso foi emitido no ciclo C, e estamos no ciclo C, cyclesSinceEmission = 0.
			// Se ele propaga a partir do ciclo C+1, então no ciclo C+1, cyclesSinceEmission = 1.
			// Assumimos que um pulso emitido no ciclo C começa a propagar no ciclo C.
			// Efeitos em distCoveredStart=0 a distCoveredEnd=PulsePropagationSpeed.
        }


        sourceNeuron, ok := nn.Neurons[p.SourceNeuronID]
        if !ok {
            p.Processed = true
            continue
        }

        // Verificar quais neurônios são afetados por este pulso neste ciclo
        for targetID, targetNeuron := range nn.Neurons {
            if targetID == p.SourceNeuronID { continue }
            distance := EuclideanDistance(sourceNeuron.Position, targetNeuron.Position)
            if distance >= distCoveredStart && distance < distCoveredEnd {
                pulsesArrivingAtTarget[targetID] = append(pulsesArrivingAtTarget[targetID], p)
            }
        }

		// Estimular glândula de cortisol se estiver na área de efeito de um pulso excitatório
		// A glândula de cortisol é um "ponto", não um neurônio.
		// Verificamos se a posição da glândula está na casca de propagação do pulso excitatório.
		if sourceNeuron.Type == Excitatory {
			distanceToGland := EuclideanDistance(sourceNeuron.Position, nn.CortisolGland.Position)
			if distanceToGland >= distCoveredStart && distanceToGland < distCoveredEnd {
				nn.CortisolGland.CortisolLevel += 0.1 // Aumento arbitrário
				if nn.CortisolGland.CortisolLevel > 2.0 { nn.CortisolGland.CortisolLevel = 2.0 }
				log.Printf("Ciclo %d: Glândula de Cortisol estimulada por pulso de %d. Nível: %.2f", nn.CurrentCycle, p.SourceNeuronID, nn.CortisolGland.CortisolLevel)
			}
		}


        if distCoveredEnd >= nn.MaxSpaceDistance {
            p.Processed = true
        } else {
            stillActivePulses = append(stillActivePulses, p)
        }
    }

    currentPulses := stillActivePulses // nn.Pulses será atualizado no final com novos pulsos

    // 2. Aplicar pulsos que chegaram aos seus alvos
    for targetNeuronID, arrivingPulsesList := range pulsesArrivingAtTarget {
        targetNeuron, ok := nn.Neurons[targetNeuronID]
        if !ok { continue }
        for _, arrivedPulse := range arrivingPulsesList {
            sourceOfPulse, exists := nn.Neurons[arrivedPulse.SourceNeuronID]
            if !exists { continue }

            if sourceOfPulse.Type == Dopaminergic {
                if _, ok := nn.DopamineLevels[targetNeuronID]; !ok {
                    nn.DopamineLevels[targetNeuronID] = 0.0
                }
                nn.DopamineLevels[targetNeuronID] += arrivedPulse.Strength
				if nn.DopamineLevels[targetNeuronID] > 1.0 { // Limitar dopamina local
					nn.DopamineLevels[targetNeuronID] = 1.0
				}
                log.Printf("Ciclo %d: Neurônio %d recebeu dopamina de %d. Nível local: %.2f", nn.CurrentCycle, targetNeuronID, arrivedPulse.SourceNeuronID, nn.DopamineLevels[targetNeuronID])
            } else {
                targetNeuron.ApplyPulse(arrivedPulse, nn)
            }
        }
    }

    // 3. Atualizar estados dos neurônios e coletar novos pulsos
    newlyGeneratedPulses := make([]*Pulse, 0)
    for _, neuron := range nn.Neurons {
		// O fator de limiar específico para este neurônio pode incluir efeitos locais de dopamina.
		neuronSpecificThresholdFactor := globalFiringThresholdFactor
		if localDopamine, ok := nn.DopamineLevels[neuron.ID]; ok {
			// Exemplo: dopamina local aumenta o limiar.
			// O README diz: "Dopamina ... serve para aumentar o limiar de disparo dos neurônios"
			neuronSpecificThresholdFactor += localDopamine * 0.5 // Ajustar este multiplicador
		}


        fired, newPulse := neuron.UpdateNeuronState(nn.CurrentCycle, neuronSpecificThresholdFactor)
        if fired {
			log.Printf("Ciclo %d: Neurônio %d (Tipo: %d, Pot: %.2f, Thr: %.2f) disparou!", nn.CurrentCycle, neuron.ID, neuron.Type, neuron.CurrentPotential, neuron.FiringThreshold * neuronSpecificThresholdFactor)
			if newPulse != nil {
				newPulse.EmittedCycle = nn.CurrentCycle // Garantir que o ciclo de emissão está correto
				// ArrivalTime será tratado na próxima iteração de propagação.
				// Para simplificar, um pulso emitido no ciclo C pode começar a afetar alvos no mesmo ciclo C se estiverem próximos.
				newPulse.ArrivalTime = nn.CurrentCycle
				newlyGeneratedPulses = append(newlyGeneratedPulses, newPulse)
			}
		}
    }
    nn.Pulses = append(currentPulses, newlyGeneratedPulses...) // Adiciona novos pulsos aos que ainda estão ativos
	log.Printf("Ciclo %d: %d pulsos ativos. %d novos pulsos gerados.", nn.CurrentCycle, len(nn.Pulses), len(newlyGeneratedPulses))
}


// AddExternalInput simula um input externo para um neurônio específico.
// Isso pode ser feito aumentando diretamente o potencial do neurônio de input.
func (nn *NeuralNetwork) AddExternalInput(neuronID int, strength float64) {
	neuron, ok := nn.Neurons[neuronID]
	if !ok {
		log.Printf("Tentativa de adicionar input ao neurônio inexistente ID %d", neuronID)
		return
	}

	// Garantir que é um neurônio de Input, embora qualquer um possa ser estimulado.
	// if neuron.Type != Input {
	// 	log.Printf("Aviso: Adicionando input externo a um neurônio que não é do tipo Input (ID %d, Tipo %d)", neuronID, neuron.Type)
	// }

	neuron.CurrentPotential += strength
	// Limitar o potencial, se necessário (já feito em ApplyPulse e UpdateNeuronState)
	log.Printf("Input externo de %.2f adicionado ao neurônio %d. Novo potencial: %.2f", strength, neuronID, neuron.CurrentPotential)
}

// GetOutputActivity retorna a atividade (frequência de disparo) dos neurônios de output.
// Para o MVP, podemos simplesmente retornar se eles dispararam no último ciclo ou seus potenciais.
// Uma medida de frequência real exigiria observar ao longo de vários ciclos.
func (nn *NeuralNetwork) GetOutputActivity() map[int]float64 {
	activity := make(map[int]float64)
	for id, neuron := range nn.Neurons {
		if neuron.Type == Output {
			// Para o MVP, vamos retornar o potencial atual como uma proxy da atividade.
			// Ou, se disparou no ciclo atual.
			// if neuron.LastFiringCycle == nn.CurrentCycle {
			// 	activity[id] = 1.0 // Disparou
			// } else {
			// 	activity[id] = 0.0 // Não disparou
			// }
			activity[id] = neuron.CurrentPotential
		}
	}
	return activity
}
