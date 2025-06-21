package io

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"crownet/src/core" // Ajuste conforme a estrutura do seu módulo
)

// ConsoleManager lida com input e output via console.
type ConsoleManager struct {
	network *core.NeuralNetwork
	reader  *bufio.Reader
}

// NewConsoleManager cria um novo gerenciador de console.
func NewConsoleManager(network *core.NeuralNetwork) *ConsoleManager {
	return &ConsoleManager{
		network: network,
		reader:  bufio.NewReader(os.Stdin),
	}
}

// CheckForInput verifica se há inputs do console para a rede.
// Esta função pode ser chamada em cada ciclo de simulação ou em intervalos.
// Para o MVP, vamos fazer um input simples que pode ser lido antes de iniciar um lote de ciclos.
// Retorna true se a simulação deve continuar, false se o usuário digitou 'quit'.
func (cm *ConsoleManager) CheckForInput() bool {
	fmt.Println("\nComandos disponíveis: 'run <ciclos>', 'input <neuron_id> <strength>', 'status', 'outputs', 'quit'")
	fmt.Print("> ")
	text, _ := cm.reader.ReadString('\n')
	text = strings.TrimSpace(text)
	parts := strings.Fields(text) // Divide a string por espaços

	if len(parts) == 0 {
		return true // Nenhum comando, continuar
	}

	command := parts[0]

	switch command {
	case "run":
		if len(parts) < 2 {
			fmt.Println("Uso: run <numero_de_ciclos>")
			return true
		}
		cycles, err := strconv.Atoi(parts[1])
		if err != nil {
			fmt.Println("Número de ciclos inválido.")
			return true
		}
		if cycles <= 0 {
			fmt.Println("Número de ciclos deve ser positivo.")
			return true
		}
		fmt.Printf("Executando %d ciclos...\n", cycles)
		for i := 0; i < cycles; i++ {
			cm.network.SimulateCycle()
			// Opcional: Mostrar output a cada ciclo ou a cada N ciclos
			if (i+1)%10 == 0 || cycles < 10 { // Exibe a cada 10 ciclos ou se menos de 10 ciclos no total
				cm.DisplayOutput()
			}
		}
		fmt.Printf("%d ciclos concluídos.\n", cycles)

	case "input":
		if len(parts) < 3 {
			fmt.Println("Uso: input <neuron_id> <strength>")
			return true
		}
		neuronID, errID := strconv.Atoi(parts[1])
		strength, errStr := strconv.ParseFloat(parts[2], 64)
		if errID != nil || errStr != nil {
			fmt.Println("ID do neurônio ou força do input inválidos.")
			return true
		}
		// Verifica se o neurônio é do tipo Input (opcional, mas bom para seguir o modelo)
		neuron, exists := cm.network.Neurons[neuronID]
		if !exists {
			fmt.Printf("Neurônio com ID %d não encontrado.\n", neuronID)
			return true
		}
		if neuron.Type != core.Input {
			fmt.Printf("Aviso: Neurônio %d não é do tipo Input (é %v). Input será aplicado mesmo assim.\n", neuronID, neuron.Type)
		}

		cm.network.AddExternalInput(neuronID, strength)
		fmt.Printf("Input de %.2f aplicado ao neurônio %d.\n", strength, neuronID)

	case "status":
		cm.DisplayNetworkStatus()

	case "outputs":
		cm.DisplayOutput()

	case "quit":
		fmt.Println("Encerrando simulação.")
		return false

	default:
		fmt.Println("Comando desconhecido.")
	}
	return true
}

// DisplayOutput mostra a atividade dos neurônios de output.
// "A codificação de input e output dos neurônios é baseada na frequência de pulsos (Hz)."
// Para o MVP, a "frequência" é simplificada. Podemos mostrar o potencial atual
// ou se o neurônio disparou no último ciclo.
func (cm *ConsoleManager) DisplayOutput() {
	outputs := cm.network.GetOutputActivity() // Retorna potencial atual dos neurônios de output
	if len(outputs) == 0 {
		fmt.Println("Nenhum neurônio de Output definido ou ativo.")
		return
	}
	fmt.Println("\n--- Atividade dos Neurônios de Output ---")
	for id, activity := range outputs {
		neuron := cm.network.Neurons[id] // Acesso seguro, pois GetOutputActivity filtra
		stateStr := fmt.Sprintf("Pot: %.2f", activity)
		if neuron.LastFiringCycle == cm.network.CurrentCycle {
			stateStr += " (DISPAROU NESTE CICLO)"
		}
		fmt.Printf("Neurônio Output ID %d: %s\n", id, stateStr)
	}
	fmt.Println("------------------------------------")
}

// DisplayNetworkStatus mostra um resumo do estado da rede.
func (cm *ConsoleManager) DisplayNetworkStatus() {
	fmt.Println("\n--- Status da Rede Neural ---")
	fmt.Printf("Ciclo Atual: %d\n", cm.network.CurrentCycle)
	fmt.Printf("Total de Neurônios: %d\n", len(cm.network.Neurons))
	fmt.Printf("Nível de Cortisol: %.4f\n", cm.network.CortisolGland.CortisolLevel)

	numFiring := 0
	var sumDopamine float64
	activeDopamineNeurons := 0
	for _, neuron := range cm.network.Neurons {
		if neuron.State == core.Firing || (neuron.State == core.AbsoluteRefractory && neuron.CyclesInState == 0) { // Disparou neste ciclo
			numFiring++
		}
		if level, ok := cm.network.DopamineLevels[neuron.ID]; ok && level > 0.01 {
			sumDopamine += level
			activeDopamineNeurons++
		}
	}
	avgDopamine := 0.0
	if activeDopamineNeurons > 0 {
		avgDopamine = sumDopamine / float64(activeDopamineNeurons)
	}
	fmt.Printf("Neurônios disparando neste ciclo (aprox.): %d\n", numFiring) // Aproximado pois o estado pode mudar durante o ciclo
	fmt.Printf("Média de Dopamina (em neurônios afetados): %.4f\n", avgDopamine)
	fmt.Printf("Pulsos ativos na rede: %d\n", len(cm.network.Pulses))
	fmt.Println("-----------------------------")
}

// GetInputNeuronIDs retorna os IDs dos neurônios de input.
func (cm *ConsoleManager) GetInputNeuronIDs() []int {
	var ids []int
	for id, n := range cm.network.Neurons {
		if n.Type == core.Input {
			ids = append(ids, id)
		}
	}
	log.Printf("Neurônios de Input disponíveis: %v", ids)
	return ids
}

// GetOutputNeuronIDs retorna os IDs dos neurônios de output.
func (cm *ConsoleManager) GetOutputNeuronIDs() []int {
	var ids []int
	for id, n := range cm.network.Neurons {
		if n.Type == core.Output {
			ids = append(ids, id)
		}
	}
	log.Printf("Neurônios de Output disponíveis: %v", ids)
	return ids
}
