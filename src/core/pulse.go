package core

// UpdatePulses processa a propagação de todos os pulsos ativos na rede para o ciclo atual.
// Retorna uma lista de pulsos que alcançaram seus alvos e novos pulsos gerados por disparos.
func (nn *NeuralNetwork) UpdatePulsesAndProcessTargets() (newlyGeneratedPulses []*Pulse) {

	// Lista para armazenar pulsos que ainda estão em trânsito
	stillActivePulses := make([]*Pulse, 0, len(nn.Pulses))
	// Mapa para agregar pulsos que chegam ao mesmo neurônio alvo neste ciclo
	pulsesArrivingAtTarget := make(map[int][]*Pulse) // neuronID -> lista de pulsos

	// 1. Propagar pulsos existentes e identificar chegadas
	for _, p := range nn.Pulses {
		if p.Processed { // Já foi processado neste ciclo ou em ciclo anterior
			continue
		}

		// Se ArrivalTime já foi calculado e é o ciclo atual
		if p.ArrivalTime > 0 && p.ArrivalTime == nn.CurrentCycle {
			// Se o pulso tem um TargetNeuronID específico (não usado no modelo atual de área)
			// ou se é um pulso de área que precisa ser aplicado aos vizinhos agora.
			// No modelo do README, o pulso não tem um único TargetNeuronID, mas afeta uma área.
			// A lógica de "chegada" é quando ele atinge neurônios dentro de seu raio de propagação.

			// A lógica do README (passos 1-9 do loop) sugere que a cada ciclo,
			// para cada pulso *existente*, verificamos quais neurônios estão na "casca"
			// de propagação para aquele ciclo.

			// Esta função é para atualizar o estado dos *pulsos* e identificar quem eles afetam *neste ciclo*.
			// A aplicação do efeito do pulso é separada.

			// Vamos simplificar: se um pulso foi emitido, ele viaja.
			// A cada ciclo, ele cobre nn.PulsePropagationSpeed.
			// A "CurrentPosition" do pulso não é usada no modelo do README,
			// que é baseado na distância do emissor e no tempo.

			// A lógica de propagação conforme o README (passos 6-8):
			// range de distancia de propagação do pulso para a iteração
			// inicio = (int8(raio/0.6)*0.6) * iteração-1
			// fim = (int8(raio/0.6)*0.6) * iteração
			// Onde "iteração" parece ser o número de ciclos desde a emissão do pulso.
			// E "raio" é o raio de efeito máximo do pulso (não da glândula/neurônio).
			// Este "raio" não está definido. Vamos assumir que é nn.MaxSpaceDistance por enquanto,
			// ou um valor menor se quisermos pulsos com alcance limitado.
			// O README menciona "Distância máxima do espaço: 8 unidades", "velocidade 0.6 unidades/ciclo".
			// Isso significa que um pulso pode levar até 8/0.6 = ~13.33 ciclos para cruzar o espaço.

			cyclesSinceEmission := nn.CurrentCycle - p.EmittedCycle

			// O "raio" no cálculo do README parece ser o raio de alcance do *pulso*,
			// não o raio da distribuição dos neurônios.
			// Vamos assumir um raio de alcance máximo para um pulso, por exemplo, 3.0 unidades.
			// Se não, ele se propagaria indefinidamente ou até nn.MaxSpaceDistance.
			// O README não é claro sobre o "raio" em "int8(raio/0.6)".
			// Vamos interpretar "raio" como a distância máxima que este pulso específico pode influenciar.
			// Para o MVP, vamos assumir que todos os pulsos podem, teoricamente, cruzar todo o espaço.
			// O "raio" na fórmula do README (passo 6) é confuso.
			// "int8(raio/0.6)" - se raio é o MaxSpaceDistance (8), 8/0.6 = 13.33. int8(13.33) = 13.
			// Então, max_iterations = 13.
			// distancia_por_iteracao_fixa = 13 * 0.6 = 7.8. Isso não parece certo.

			// Reinterpretando o passo 6 do README:
			// A cada ciclo (iteração `t` desde a emissão), o pulso afeta uma "casca esférica".
			// Distância percorrida até o início do ciclo atual: `dist_covered_start = (cyclesSinceEmission - 1) * nn.PulsePropagationSpeed`
			// Distância percorrida até o fim do ciclo atual: `dist_covered_end = cyclesSinceEmission * nn.PulsePropagationSpeed`
			// Neurônios afetados são aqueles cuja distância ao `p.SourceNeuronID` está entre `dist_covered_start` e `dist_covered_end`.

			distCoveredEnd := float64(cyclesSinceEmission) * nn.PulsePropagationSpeed
			distCoveredStart := float64(cyclesSinceEmission-1) * nn.PulsePropagationSpeed
			if cyclesSinceEmission == 0 { // Primeiro ciclo de propagação (ou ciclo de emissão)
				distCoveredStart = 0 // Afeta neurônios muito próximos imediatamente
			}


			sourceNeuron, ok := nn.Neurons[p.SourceNeuronID]
			if !ok {
				p.Processed = true // Neurônio fonte não existe mais, marcar como processado
				continue
			}

			// Verificar quais neurônios são afetados por este pulso neste ciclo
			for targetID, targetNeuron := range nn.Neurons {
				if targetID == p.SourceNeuronID {
					continue // Não afeta a si mesmo
				}

				// A busca de vizinhos otimizada (passos 1-4 do README) seria usada aqui.
				// Para o MVP, faremos o cálculo direto da distância.
				distance := EuclideanDistance(sourceNeuron.Position, targetNeuron.Position)

				if distance >= distCoveredStart && distance < distCoveredEnd {
					// Este neurônio está na "casca" de propagação do pulso para este ciclo.
					// O pulso "chega" a este neurônio.
					pulsesArrivingAtTarget[targetID] = append(pulsesArrivingAtTarget[targetID], p)
				}
			}

			// Se o pulso já viajou a distância máxima, ele se dissipa.
			if distCoveredEnd >= nn.MaxSpaceDistance { // Ou um raio de alcance máximo do pulso
				p.Processed = true // Marcar como processado para remoção
			} else {
				stillActivePulses = append(stillActivePulses, p) // Continua propagando no próximo ciclo
			}

		} else if p.ArrivalTime == 0 { // Primeira vez processando este pulso (emitido neste ciclo)
			// Define ArrivalTime para o próximo ciclo para que ele comece a propagar.
			// Ou, se quisermos que ele afete imediatamente os vizinhos muito próximos:
			p.ArrivalTime = nn.CurrentCycle // Será processado na lógica acima neste mesmo ciclo.
			stillActivePulses = append(stillActivePulses, p)
		} else { // p.ArrivalTime > nn.CurrentCycle - ainda não chegou
			stillActivePulses = append(stillActivePulses, p)
		}
	}
	nn.Pulses = stillActivePulses


	// 2. Aplicar pulsos que chegaram aos seus alvos
	for targetNeuronID, arrivingPulses := range pulsesArrivingAtTarget {
		targetNeuron, ok := nn.Neurons[targetNeuronID]
		if !ok {
			continue
		}
		for _, arrivedPulse := range arrivingPulses {
			// Se o pulso é de um neurônio dopaminérgico, seu efeito é especial
			sourceOfPulse, exists := nn.Neurons[arrivedPulse.SourceNeuronID]
			if !exists { continue }

			if sourceOfPulse.Type == Dopaminergic {
				// O README diz: "se for dopamina soma a quantidade de dopamina do neuronio"
				// Isso é interpretado como: o neurônio *alvo* atingido por um pulso de um neurônio dopaminérgico
				// tem seu nível local de dopamina aumentado.
				// A "quantidade de dopamina do neurônio" (emissor) não é uma propriedade direta.
				// Vamos assumir que `arrivedPulse.Strength` para um pulso dopaminérgico é a quantidade de dopamina liberada.
				// Esta dopamina afeta o neurônio alvo.
				if _, ok := nn.DopamineLevels[targetNeuronID]; !ok {
					nn.DopamineLevels[targetNeuronID] = 0.0
				}
				nn.DopamineLevels[targetNeuronID] += arrivedPulse.Strength
				// Dopamina também decai, isso será tratado em outra parte do ciclo.
			} else {
				// Para pulsos excitatórios/inibitórios
				targetNeuron.ApplyPulse(arrivedPulse, nn)
			}
			// Marcamos o pulso como processado para este alvo específico.
			// No entanto, um pulso de área pode afetar múltiplos alvos.
			// A flag `Processed` no pulso em si indica que ele se dissipou completamente.
		}
	}

	// 3. Atualizar estados dos neurônios e coletar novos pulsos
	// A modulação do limiar de disparo (cortisol, dopamina) deve ser considerada aqui.
	// Por enquanto, passamos um fator de modulação de limiar de 1.0.
	// Este valor será ajustado com base nos níveis de cortisol e dopamina.
	globalFiringThresholdFactor := 1.0 // TODO: Calcular com base em cortisol e dopamina globais/locais.

	newlyGeneratedPulses = make([]*Pulse, 0)
	for _, neuron := range nn.Neurons {
		fired, newPulse := neuron.UpdateNeuronState(nn.CurrentCycle, globalFiringThresholdFactor)
		if fired && newPulse != nil {
			// O pulso é gerado. Seu ArrivalTime e TargetNeuronID não são definidos aqui,
			// mas sim durante a propagação no próximo ciclo (ou neste mesmo, se processado imediatamente).
			// Para consistência, vamos adicionar à lista de pulsos da rede e ele será
			// processado no início do próximo ciclo de `UpdatePulsesAndProcessTargets`.
			newPulse.ArrivalTime = nn.CurrentCycle // Marcar para ser processado na próxima chamada (ou no mesmo ciclo se a lógica permitir)
			newlyGeneratedPulses = append(newlyGeneratedPulses, newPulse)

			// Se o neurônio que disparou for a glândula de cortisol (ou um neurônio que a estimula)
			// ou um neurônio dopaminérgico, tratar seus efeitos específicos.
			// A glândula de cortisol não é um neurônio, mas é afetada por eles.
			// Neurônios dopaminérgicos geram pulsos que são tratados acima para aumentar a dopamina.
		}
	}
	nn.Pulses = append(nn.Pulses, newlyGeneratedPulses...)


	// 4. Lógica da glândula de cortisol (simplificada)
	// A glândula de cortisol é afetada por pulsos excitatórios que a atingem.
	// Ela não é um neurônio, então não "dispara".
	// Precisamos verificar se algum pulso excitatório atingiu a área da glândula.
	// A variável cortisolGlandRadius não está sendo usada na lógica atual, pois
	// a estimulação da glândula é verificada diretamente pela sua posição
	// em propagateAndUpdateNetworkStates. Removendo-a para evitar "declared and not used".
	// cortisolGlandRadius := 0.5 // Raio arbitrário para a glândula
	for _, p := range nn.Pulses { // Iterar sobre todos os pulsos ativos
		if p.Processed { continue } // Ignorar pulsos já dissipados

		sourceNeuron, ok := nn.Neurons[p.SourceNeuronID]
		if !ok || sourceNeuron.Type != Excitatory { // Apenas pulsos excitatórios afetam
			continue
		}

		// Calcular distância do *pulso* à glândula.
		// A posição do pulso não é explicitamente rastreada no modelo do README.
		// Em vez disso, verificamos se a glândula está dentro da "casca" de propagação do pulso.
		cyclesSinceEmission := nn.CurrentCycle - p.EmittedCycle
		distCoveredEnd := float64(cyclesSinceEmission) * nn.PulsePropagationSpeed
		distCoveredStart := float64(cyclesSinceEmission-1) * nn.PulsePropagationSpeed
		if cyclesSinceEmission == 0 { distCoveredStart = 0 }

		distanceToGland := euclideanDistance(sourceNeuron.Position, nn.CortisolGland.Position)

		if distanceToGland >= distCoveredStart && distanceToGland < distCoveredEnd {
			// Pulso excitatório atingiu a glândula de cortisol
			nn.CortisolGland.CortisolLevel += 0.1 // Aumento arbitrário
			if nn.CortisolGland.CortisolLevel > 2.0 { // Limite máximo
				nn.CortisolGland.CortisolLevel = 2.0
			}
		}
	}


	return newlyGeneratedPulses // Retorna apenas os que foram gerados *neste* ciclo.
}


// calculateDistanceToNearestReferencePoints e findNeighborsInCoverageArea
// seriam as implementações dos passos 1-5 do loop do README para otimizar a busca de vizinhos.
// Para o MVP, a busca de vizinhos é feita por iteração simples e cálculo de distância (acima).

// Placeholder para a função de cálculo de distância (já existe em initialization.go, pode ser movida para um utils)
// func EuclideanDistance(p1, p2 Vector16D) float64 { ... } // Movido para initialization.go e exportado


// findNeighborsInRange identifica todos os neurônios dentro de um raio de um ponto de origem.
// Esta é uma abordagem mais simples do que a "casca esférica" para aplicar o efeito de um pulso.
// No entanto, o README descreve a propagação em cascas.
// Esta função não está sendo usada atualmente, mas pode ser útil.
func (nn *NeuralNetwork) findNeighborsInRange(origin Positionable, pulseRadius float64) []*Neuron {
	neighbors := []*Neuron{}
	originPos := origin.GetPosition()
	for _, neuron := range nn.Neurons {
		if neuron.GetPosition() == originPos { // Não considerar o próprio emissor
			continue
		}
		dist := EuclideanDistance(originPos, neuron.GetPosition()) // Corrigido para usar a função exportada
		if dist <= pulseRadius {
			neighbors = append(neighbors, neuron)
		}
	}
	return neighbors
}

// Positionable é uma interface para objetos que têm uma posição.
// Exportada para que possa ser usada por outros pacotes se necessário.
type Positionable interface {
	GetPosition() Vector16D
}

// GetPosition para Neuron - já existe em neuron.go ou é acessado diretamente.
// Se for necessário como método da interface, deve estar em neuron.go.
// Como Neuron.Position é público, a interface pode não ser estritamente necessária aqui.

// GetPosition para Pulse (se rastrearmos sua posição exata)
// func (p *Pulse) GetPosition() Vector16D {
// 	return p.CurrentPosition
// }


// updateNetworkFiringThreshold calcula o fator de modulação global do limiar de disparo
// com base nos níveis de cortisol e dopamina.
// Esta é uma implementação de exemplo e pode ser muito mais complexa.
func (nn *NeuralNetwork) calculateGlobalFiringThresholdFactor() float64 {
	factor := 1.0

	// Efeito do Cortisol (exemplo, baseado na descrição do README)
	// Cortisol diminui o limiar inicialmente, mas em excesso aumenta.
	// Pico de cortisol não está definido, vamos usar um valor arbitrário, ex: 1.0
	cortisolEffect := 0.0
	if nn.CortisolGland.CortisolLevel < 0.5 { // Abaixo do "normal" ou inicial
		cortisolEffect = -0.2 * nn.CortisolGland.CortisolLevel // Diminui limiar (torna fator < 1)
	} else if nn.CortisolGland.CortisolLevel < 1.5 { // Faixa "ótima"
		cortisolEffect = -0.3 + (nn.CortisolGland.CortisolLevel-0.5)*0.1 // Diminui mais, depois menos
	} else { // Níveis altos/de pico
		cortisolEffect = 0.1 + (nn.CortisolGland.CortisolLevel-1.5)*0.4 // Aumenta limiar
	}
	factor += cortisolEffect

	// Efeito da Dopamina (exemplo)
	// Dopamina aumenta o limiar de disparo.
	// Precisamos de um nível médio de dopamina na rede ou considerar efeitos locais.
	// Para um fator global, podemos usar uma média.
		// No entanto, o efeito da dopamina no limiar é agora tratado localmente em network.go/propagateAndUpdateNetworkStates
		// com base no nn.DopamineLevels[neuron.ID].
		// Portanto, não precisamos adicionar um efeito global de dopamina aqui novamente,
		// a menos que queiramos um efeito de "background" de dopamina além do local.
		// Por ora, vamos remover o efeito global duplicado da dopamina nesta função.
		// avgDopamine := 0.0
		// if len(nn.DopamineLevels) > 0 {
		// 	sumDopamine := 0.0
		// 	numWithDopamine := 0
		// 	for _, level := range nn.DopamineLevels {
		// 		if level > 0 { // Considerar apenas onde há dopamina
		// 			sumDopamine += level
		// 			numWithDopamine++
		// 		}
		// 	}
		// 	if numWithDopamine > 0 {
		// 		avgDopamine = sumDopamine / float64(numWithDopamine)
		// 	}
		// }
		// dopamineEffect := avgDopamine * 0.2
		// factor += dopamineEffect

	if factor < 0.1 { // Limiar mínimo para o fator
		factor = 0.1
	}
	if factor > 3.0 { // Limiar máximo
		factor = 3.0
	}
	return factor
}

// decayChemicals lida com o decaimento natural do cortisol e da dopamina.
func (nn *NeuralNetwork) decayChemicals() {
	// Decaimento do Cortisol
	// "Caso não haja pulsos na glândula de cortisol, a quantidade de cortisol diminui com o tempo."
	// Vamos assumir uma taxa de decaimento mesmo com estímulos, mas menor.
	// Se não estimulado, decai mais rápido. (Esta parte da lógica não está aqui, mas no aumento do cortisol)
	nn.CortisolGland.CortisolLevel *= 0.98 // Decai 2% por ciclo
	if nn.CortisolGland.CortisolLevel < 0.01 {
		nn.CortisolGland.CortisolLevel = 0.01 // Nível basal mínimo
	}

	// Decaimento da Dopamina
	// "A dopamina tem uma taxa de decaimento mais acentuada ao longo do tempo"
	for id, level := range nn.DopamineLevels {
		nn.DopamineLevels[id] = level * 0.90 // Decai 10% por ciclo
		if nn.DopamineLevels[id] < 0.01 {
			nn.DopamineLevels[id] = 0.0 // Pode decair a zero
		}
	}
}
