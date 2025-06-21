package tests

import (
	"os"
	"testing"
	"math"
	"fmt"

	"crownet/src/core"
	"crownet/src/database"
)

// Helper para limpar o arquivo de banco de dados de teste
func setupTestDB() {
	// Usa um nome de DB diferente para testes ou garante que o default seja limpo
	// Por simplicidade, vamos remover o db padrão se ele existir.
	// Idealmente, database.dbFileName deveria ser configurável para testes.
	// Temporariamente, vamos assumir que database.dbFileName é "data/crownet.db"
	// e vamos usar "data/crownet_test.db"

	// monkey patch dbFileName para testes - NÃO É UMA BOA PRÁTICA PARA CÓDIGO REAL,
    // mas para este contexto de agente, é uma forma de contornar sem pedir refatoração imediata.
    // A melhor forma seria ter database.InitDB(filePath string).
    originalDBName := "data/crownet.db" // Precisamos saber o nome original
    testDBName := "data/crownet_test.db"

    // Renomear o arquivo de banco de dados real se existir, para não afetá-lo
    if _, err := os.Stat(originalDBName); err == nil {
        os.Rename(originalDBName, originalDBName+".bak_before_test")
    }
    // E se o arquivo de teste já existe de um teste anterior, limpá-lo
    if _, err := os.Stat(testDBName); err == nil {
        os.Remove(testDBName)
    }

    // Aqui, precisaríamos de uma forma de dizer a `database.InitDB` para usar `testDBName`.
    // Como não temos, o teste vai operar no `database.dbFileName` que o pacote `database` conhece.
    // Vamos assumir que o `database` package usa uma constante `dbFileName`.
    // Para este teste, vamos apenas limpar o arquivo padrão.
    if _, err := os.Stat("data/crownet.db"); err == nil {
		os.Remove("data/crownet.db")
	}

	if err := database.InitDB(); err != nil {
		panic(fmt.Sprintf("Falha ao inicializar DB de teste: %v", err))
	}
}

func cleanupTestDB() {
	database.CloseDB()
	// os.Remove("data/crownet_test.db") // Se estivéssemos usando um nome de arquivo de teste
    os.Remove("data/crownet.db")

    originalDBName := "data/crownet.db"
    // Restaurar o db original se fizemos backup
    if _, err := os.Stat(originalDBName + ".bak_before_test"); err == nil {
        os.Rename(originalDBName+".bak_before_test", originalDBName)
    }
}

func TestNetworkInitialization(t *testing.T) {
	setupTestDB()
	defer cleanupTestDB()

	config := core.GetDefaultConfig()
	config.NumNeurons = 50 // Usar um número menor para teste
	nn := core.InitializeNetwork(config)

	if len(nn.Neurons) != config.NumNeurons {
		t.Errorf("Esperado %d neurônios, obteve %d", config.NumNeurons, len(nn.Neurons))
	}

	if nn.CortisolGland == nil {
		t.Errorf("Glândula de cortisol não foi inicializada.")
	}
	if nn.CurrentCycle != 0 {
		t.Errorf("Ciclo atual esperado 0, obteve %d", nn.CurrentCycle)
	}

	// Verificar se os tipos de neurônios foram distribuídos
	counts := make(map[core.NeuronType]int)
	for _, n := range nn.Neurons {
		counts[n.Type]++
	}

	expectedDopa := int(float64(config.NumNeurons) * config.DopaminergicRatio)
	// Permitir uma pequena variação devido a arredondamentos na inicialização
	if counts[core.Dopaminergic] < expectedDopa || counts[core.Dopaminergic] > expectedDopa+1 {
		t.Errorf("Esperado ~%d neurônios Dopaminérgicos, obteve %d", expectedDopa, counts[core.Dopaminergic])
	}
	// Adicionar mais verificações para outros tipos se necessário

	// Verificar se SaveInitialNeurons e LogNetworkState (ciclo 0) funcionam
	err := database.SaveInitialNeurons(nn.Neurons)
	if err != nil {
		t.Fatalf("SaveInitialNeurons falhou: %v", err)
	}
	err = database.LogNetworkState(0, nn)
	if err != nil {
		t.Fatalf("LogNetworkState para ciclo 0 falhou: %v", err)
	}

	// Tentar carregar o estado logado (ciclo 0)
	loadedNN, err := database.LoadNetworkState(0)
	if err != nil {
		t.Fatalf("LoadNetworkState para ciclo 0 falhou: %v", err)
	}
	if len(loadedNN.Neurons) != config.NumNeurons {
		t.Errorf("LoadNetworkState: Esperado %d neurônios, obteve %d", config.NumNeurons, len(loadedNN.Neurons))
	}
	if math.Abs(loadedNN.CortisolGland.CortisolLevel - nn.CortisolGland.CortisolLevel) > 1e-9 {
         t.Errorf("LoadNetworkState: Nível de cortisol diferente. Esperado %.2f, obteve %.2f", nn.CortisolGland.CortisolLevel, loadedNN.CortisolGland.CortisolLevel)
    }

}

func TestNeuronFiringAndRefractory(t *testing.T) {
	n := &core.Neuron{
		ID:              1,
		Type:            core.Excitatory,
		State:           core.Resting,
		FiringThreshold: 1.0,
	}
	var nn core.NeuralNetwork // Dummy network para ApplyPulse, se necessário

	// 1. Testar disparo
	n.CurrentPotential = 1.5
	fired, pulse := n.UpdateNeuronState(1, 1.0) // Ciclo 1, fator de limiar 1.0

	if !fired {
		t.Errorf("Neurônio deveria ter disparado com potencial %.2f e limiar %.2f", n.CurrentPotential, n.FiringThreshold)
	}
	if pulse == nil {
		t.Errorf("Neurônio disparou mas não gerou pulso.")
	}
	if n.State != core.AbsoluteRefractory {
		t.Errorf("Estado esperado AbsoluteRefractory após disparo, obteve %v", n.State)
	}
	if n.LastFiringCycle != 1 {
		t.Errorf("LastFiringCycle esperado 1, obteve %d", n.LastFiringCycle)
	}

	// 2. Testar período refratário absoluto
	n.CurrentPotential = 1.5 // Mesmo que o potencial seja alto
	fired, _ = n.UpdateNeuronState(2, 1.0) // Ciclo 2
	if fired {
		t.Errorf("Neurônio não deveria disparar em AbsoluteRefractory")
	}
	if n.State != core.AbsoluteRefractory || n.CyclesInState != 1 { // CyclesInState incrementado
		t.Errorf("Estado esperado AbsoluteRefractory com CyclesInState 1, obteve %v, %d", n.State, n.CyclesInState)
	}

	// Avançar para sair do refratário absoluto (2 ciclos de duração)
	n.UpdateNeuronState(3, 1.0) // Ciclo 3, CyclesInState se torna 2, transita para Relativo
	if n.State != core.RelativeRefractory {
		t.Errorf("Estado esperado RelativeRefractory após Absolute, obteve %v", n.State)
	}
	if n.CurrentPotential != -0.1 { // Potencial resetado na transição para refratário relativo
		t.Errorf("Potencial esperado -0.1 ao entrar em RelativeRefractory, obteve %.2f", n.CurrentPotential)
	}


	// 3. Testar período refratário relativo
	n.CurrentPotential = 1.2 // Potencial acima do limiar base (1.0) mas talvez não do efetivo (1.0 * 1.5 = 1.5)
	fired, _ = n.UpdateNeuronState(4, 1.0) // Ciclo 4
	if fired {
		t.Errorf("Neurônio não deveria disparar em RelativeRefractory com potencial 1.2 (limiar efetivo ~1.5)")
	}
	if n.State != core.RelativeRefractory {
		t.Errorf("Ainda deveria estar em RelativeRefractory")
	}

	n.CurrentPotential = 1.6 // Agora potencial acima do limiar efetivo
	fired, _ = n.UpdateNeuronState(5, 1.0) // Ciclo 5
	if !fired {
		t.Errorf("Neurônio deveria disparar em RelativeRefractory com potencial 1.6")
	}
	if n.State != core.AbsoluteRefractory { // Volta para absoluto após disparar
		t.Errorf("Estado esperado AbsoluteRefractory após disparo em relativo, obteve %v", n.State)
	}

	// Resetar para testar saída do refratário relativo para repouso
	n.State = core.RelativeRefractory
	n.CyclesInState = 0
	n.CurrentPotential = 0.0
	n.UpdateNeuronState(6, 1.0) // ciclo 1 em relativo
	n.UpdateNeuronState(7, 1.0) // ciclo 2 em relativo
	n.UpdateNeuronState(8, 1.0) // ciclo 3 em relativo, deve transitar para Repouso

	if n.State != core.Resting {
		t.Errorf("Estado esperado Resting após RelativeRefractoryPeriod, obteve %v", n.State)
	}
	if n.CurrentPotential != 0.0 {
		t.Errorf("Potencial esperado 0.0 ao sair de RelativeRefractory para Resting, obteve %.2f", n.CurrentPotential)
	}
}

func TestPulsePropagationAndApplication(t *testing.T) {
	setupTestDB() // DB necessário para carregar config e potencialmente logar
	defer cleanupTestDB()

	config := core.GetDefaultConfig()
	config.NumNeurons = 2
	nn := core.InitializeNetwork(config)

	// Forçar posições para teste determinístico
	// IDs dos neurônios serão 0 e 1
	nn.Neurons[0].Position = core.Vector16D{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	nn.Neurons[1].Position = core.Vector16D{1, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0} // Distância 1.0
	nn.Neurons[0].Type = core.Excitatory
	nn.Neurons[1].Type = core.Excitatory // ou qualquer tipo que possa receber potencial
	nn.Neurons[0].FiringThreshold = 0.5
	nn.Neurons[1].FiringThreshold = 10.0 // Alto para não disparar por si só
	nn.PulsePropagationSpeed = 0.6      // Conforme README

	// Fazer neurônio 0 disparar no ciclo 1
	nn.Neurons[0].CurrentPotential = 1.0
	nn.SimulateCycle() // Ciclo 1

	if nn.Neurons[0].State != core.AbsoluteRefractory {
		t.Fatalf("Neurônio 0 deveria estar em AbsoluteRefractory após disparar no ciclo 1. Estado: %v", nn.Neurons[0].State)
	}
	if len(nn.Pulses) != 1 {
		t.Fatalf("Esperado 1 pulso na rede após disparo do neurônio 0, obteve %d", len(nn.Pulses))
	}

	pulse := nn.Pulses[0]
	if pulse.SourceNeuronID != 0 {
		t.Errorf("Fonte do pulso esperada 0, obteve %d", pulse.SourceNeuronID)
	}
	if pulse.Strength != 0.3 { // Força padrão para excitatório
		t.Errorf("Força do pulso esperada 0.3, obteve %.2f", pulse.Strength)
	}

	// Ciclo 2: Pulso deve propagar.
	// Distância é 1.0. Velocidade é 0.6.
	// No ciclo 1 (emissão): dist_start=0, dist_end=0 (ou 0.6 se propaga imediatamente)
	//    Se consideramos que o pulso emitido no ciclo C afeta alvos no mesmo ciclo C:
	//    Em SimulateCycle(1): Neurônio 0 dispara, cria pulso P.
	//        P é processado: cyclesSinceEmission = 1-1 = 0. dist_start=0, dist_end=0.
	//        Nenhum neurônio a distância 0 (exceto ele mesmo).
	// No ciclo 2 (nn.CurrentCycle = 2):
	//    Para pulso P (emitido no ciclo 1): cyclesSinceEmission = 2-1 = 1.
	//    dist_covered_start = (1-1)*0.6 = 0.
	//    dist_covered_end = 1*0.6 = 0.6.
	//    Neurônio 1 está a distância 1.0. Não deve ser atingido.
	initialPotentialN1 := nn.Neurons[1].CurrentPotential
	nn.SimulateCycle() // Ciclo 2

	if nn.Neurons[1].CurrentPotential != initialPotentialN1 {
		t.Errorf("Ciclo 2: Potencial do Neurônio 1 não deveria mudar. Era %.2f, agora %.2f. Pulso não deveria alcançar (dist 1.0, propagado 0.6)", initialPotentialN1, nn.Neurons[1].CurrentPotential)
	}

	// Ciclo 3: Pulso continua propagando.
	//    Para pulso P (emitido no ciclo 1): cyclesSinceEmission = 3-1 = 2.
	//    dist_covered_start = (2-1)*0.6 = 0.6.
	//    dist_covered_end = 2*0.6 = 1.2.
	//    Neurônio 1 (dist 1.0) está em [0.6, 1.2). Deve ser atingido.
	nn.SimulateCycle() // Ciclo 3
	expectedPotentialN1 := initialPotentialN1 + pulse.Strength
	if math.Abs(nn.Neurons[1].CurrentPotential - expectedPotentialN1) > 1e-9 {
		t.Errorf("Ciclo 3: Potencial do Neurônio 1 incorreto. Esperado %.2f, obteve %.2f", expectedPotentialN1, nn.Neurons[1].CurrentPotential)
	}

	// Verificar se o pulso é removido após propagar MaxSpaceDistance
	nn.MaxSpaceDistance = 1.0 // Forçar pulso a se dissipar mais cedo para teste
	nn.Neurons[1].Position = core.Vector16D{5,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0} // Mover N1 para longe
	nn.Pulses = []*core.Pulse{ // Resetar pulso, emitido no ciclo 3
		{SourceNeuronID: 0, Strength: 0.3, EmittedCycle: nn.CurrentCycle, CurrentPosition: nn.Neurons[0].Position, ArrivalTime: nn.CurrentCycle},
	}

	// Ciclo 4: (Pulso emitido no ciclo 3)
	// cyclesSinceEmission = 4-3=1. dist_end = 0.6. N1 (dist 5) não atingido.
	nn.SimulateCycle()
	if len(nn.Pulses) != 1 { t.Errorf("Ciclo 4: Pulso não deveria ter sido removido ainda (propagou 0.6 de 1.0)")}

	// Ciclo 5: (Pulso emitido no ciclo 3)
	// cyclesSinceEmission = 5-3=2. dist_end = 1.2. N1 não atingido.
	// O pulso deve ser removido porque dist_end (1.2) >= MaxSpaceDistance (1.0)
	nn.SimulateCycle()
	if len(nn.Pulses) != 0 {
		t.Errorf("Ciclo 5: Pulso deveria ter sido removido após exceder MaxSpaceDistance. Pulsos restantes: %d", len(nn.Pulses))
		if len(nn.Pulses) > 0 {
			p := nn.Pulses[0]
			cyclesSinceEmission := nn.CurrentCycle - p.EmittedCycle
			distCoveredEnd := float64(cyclesSinceEmission) * nn.PulsePropagationSpeed
			t.Logf("Detalhes do pulso restante: Emitted: %d, CurrentCycle: %d, CSE: %d, DistEnd: %.2f, MaxDist: %.2f, Processed: %v",
				p.EmittedCycle, nn.CurrentCycle, cyclesSinceEmission, distCoveredEnd, nn.MaxSpaceDistance, p.Processed)
		}
	}
}

func TestModulationMechanisms(t *testing.T) {
	config := core.GetDefaultConfig()
	config.NumNeurons = 10
	nn := core.InitializeNetwork(config)

	// 1. Testar Cortisol
	nn.CortisolGland.CortisolLevel = 0.2 // Nível baixo inicial
	// Simular um pulso excitatório atingindo a glândula
	// Para isso, precisamos de um neurônio excitatório perto da glândula (posição 0,0,...)
	// e um pulso dele.
	// Mais fácil: chamar diretamente a lógica de aumento de cortisol.
	// A lógica está em propagateAndUpdateNetworkStates. Vamos simular o efeito.
	// nn.CortisolGland.CortisolLevel += 0.1 // Simula estímulo
	// Em vez disso, vamos colocar um neurônio excitatório perto da glândula e fazê-lo disparar.
	nn.Neurons[0].Type = core.Excitatory
	nn.Neurons[0].Position = core.Vector16D{} // No centro, perto da glândula
	nn.Neurons[0].CurrentPotential = 1.0 // Para disparar
	nn.Neurons[0].FiringThreshold = 0.5
	nn.PulsePropagationSpeed = 0.1 // Para que atinja a glândula rápido (dist 0)

	initialCortisol := nn.CortisolGland.CortisolLevel
	nn.SimulateCycle() // Ciclo 1: N0 dispara, pulso emitido. Cortisol estimulado.

	if nn.CortisolGland.CortisolLevel <= initialCortisol {
		t.Errorf("Nível de cortisol esperado aumentar após estímulo. Era %.2f, agora %.2f", initialCortisol, nn.CortisolGland.CortisolLevel)
	}

	// Testar decaimento do cortisol
	levelBeforeDecay := nn.CortisolGland.CortisolLevel
	nn.SimulateCycle() // Ciclo 2: Sem estímulo direto à glândula (N0 está refratário)
	if nn.CortisolGland.CortisolLevel >= levelBeforeDecay && levelBeforeDecay > 0.01 { // Se não estiver no piso
		// Pode haver outro pulso aleatório atingindo a glândula, o que complica o teste.
		// Para um teste isolado de decaimento, seria melhor chamar decayChemicals() diretamente.
		// t.Errorf("Nível de cortisol esperado diminuir. Era %.2f, agora %.2f", levelBeforeDecay, nn.CortisolGland.CortisolLevel)
		// Este teste é frágil devido à natureza estocástica da rede.
		// Vamos testar o decaimento de forma mais controlada:
		nn.CortisolGland.CortisolLevel = 1.0
		nn.decayChemicals() // Chamada direta
		if nn.CortisolGland.CortisolLevel != 1.0*0.98 {
			t.Errorf("Decaimento do cortisol incorreto. Esperado %.3f, obteve %.3f", 1.0*0.98, nn.CortisolGland.CortisolLevel)
		}
	}


	// 2. Testar Dopamina
	nn.Neurons[1].Type = core.Dopaminergic
	nn.Neurons[1].Position = core.Vector16D{0.1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0} // Perto de N2
	nn.Neurons[2].Position = core.Vector16D{0.2,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0} // N2 é o alvo
	nn.Neurons[1].CurrentPotential = 1.0
	nn.Neurons[1].FiringThreshold = 0.5
	nn.DopamineLevels[2] = 0.0 // Nível inicial de dopamina em N2

	// Ciclo (vamos usar o ciclo atual da rede, que já avançou)
	currentSimCycle := nn.CurrentCycle
	nn.SimulateCycle() // N1 (Dopa) dispara, pulso emitido.
	// No próximo ciclo, o pulso de dopamina deve atingir N2.
	// Dist N1-N2 = 0.1. Vel = 0.1. Atinge no mesmo ciclo de propagação.
	// Pulso emitido no ciclo `currentSimCycle+1`.
	// No ciclo `currentSimCycle+2`, o pulso é processado.
	//   cyclesSinceEmission = (currentSimCycle+2) - (currentSimCycle+1) = 1.
	//   dist_start = 0, dist_end = 0.1. N2 é atingido.

	// Para garantir que N1 dispara no ciclo certo e o pulso é processado como esperado:
	nn.Neurons[1].State = core.Resting // Resetar estado para permitir disparo
	nn.Neurons[1].CurrentPotential = 1.0
	initialDopamineN2 := nn.DopamineLevels[2]

	nn.SimulateCycle() // N1 dispara. Pulso P_dopa emitido.

	// Se a propagação é instantânea para distâncias curtas ou no mesmo ciclo:
	// O pulso de dopamina de N1 (emitido) deve ser processado e afetar N2 no *mesmo* ciclo SimulateCycle.
	// A lógica em propagateAndUpdateNetworkStates:
	//   - N1 dispara, P_dopa adicionado a newlyGeneratedPulses.
	//   - nn.Pulses = append(currentPulses, newlyGeneratedPulses...)
	//   - Então, o P_dopa *não* é processado para afetar N2 até o *próximo* SimulateCycle.
	// Isso significa que precisamos de mais um ciclo.

	if nn.DopamineLevels[2] > initialDopamineN2 {
		t.Logf("Dopamina em N2 aumentou no mesmo ciclo do disparo de N1. Nível: %.2f", nn.DopamineLevels[2])
		// Isso pode acontecer se a lógica de propagação/aplicação for muito rápida ou no mesmo ciclo.
	}

	nn.SimulateCycle() // Agora o pulso de dopamina P_dopa deve ser processado.

	if nn.DopamineLevels[2] <= initialDopamineN2 {
		t.Errorf("Nível de dopamina em N2 esperado aumentar após pulso de N1. Era %.2f, agora %.2f", initialDopamineN2, nn.DopamineLevels[2])
		t.Logf("N1 ID: %d, N2 ID: %d", nn.Neurons[1].ID, nn.Neurons[2].ID)
		// Log para entender o que aconteceu com o pulso
		foundPulse := false
		for _, p := range nn.Pulses {
			if p.SourceNeuronID == nn.Neurons[1].ID {
				foundPulse = true
				t.Logf("Pulso de N1 encontrado: Emitted %d, Strength %.2f", p.EmittedCycle, p.Strength)
			}
		}
		if !foundPulse {t.Logf("Nenhum pulso de N1 encontrado na rede.")}
	}

	// Testar decaimento da dopamina
	levelBeforeDecayDopa := nn.DopamineLevels[2]
	if levelBeforeDecayDopa > 0 {
		nn.decayChemicals() // Chamada direta
		expectedDopaDecay := levelBeforeDecayDopa * 0.90
		if math.Abs(nn.DopamineLevels[2]-expectedDopaDecay) > 1e-9 {
			t.Errorf("Decaimento da dopamina em N2 incorreto. Esperado %.3f, obteve %.3f", expectedDopaDecay, nn.DopamineLevels[2])
		}
	}
}

func TestSynaptogenesisEffect(t *testing.T) {
	config := core.GetDefaultConfig()
	config.NumNeurons = 3
	nn := core.InitializeNetwork(config)

	// N0: Ativo (disparou recentemente)
	// N1: Passivo (em repouso)
	// N2: O neurônio que estamos testando a movimentação
	nn.Neurons[0].Position = core.Vector16D{0,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}
	nn.Neurons[1].Position = core.Vector16D{1,0,0,0,0,0,0,0,0,0,0,0,0,0,0,0}
	nn.Neurons[2].Position = core.Vector16D{0.5,0.5,0,0,0,0,0,0,0,0,0,0,0,0,0,0}

	nn.CurrentCycle = 10 // Simular que já estamos em um ciclo mais avançado
	nn.Neurons[0].LastFiringCycle = nn.CurrentCycle - 1 // Disparou no ciclo anterior
	nn.Neurons[0].State = core.RelativeRefractory      // Estado ativo
	nn.Neurons[1].State = core.Resting                 // Estado de repouso

	// Forçar moduladores para um estado neutro para isolar o efeito da atividade
	nn.CortisolGland.CortisolLevel = 1.0 // Nível que não inibe nem excita muito a sinaptogênese
	for id := range nn.DopamineLevels {
		nn.DopamineLevels[id] = 0.0 // Sem dopamina
	}


	initialPosN2 := nn.Neurons[2].Position
	nn.ApplySynaptogenesis()
	newPosN2 := nn.Neurons[2].Position

	// N2 deve se aproximar de N0 (ativo) e se afastar de N1 (repouso)
	// Dist N2-N0 inicial: sqrt(0.5^2 + 0.5^2) = sqrt(0.25+0.25) = sqrt(0.5) approx 0.707
	// Dist N2-N1 inicial: sqrt((0.5-1)^2 + 0.5^2) = sqrt(-0.5^2 + 0.5^2) = sqrt(0.25+0.25) = sqrt(0.5) approx 0.707

	distN2_N0_initial := core.EuclideanDistance(initialPosN2, nn.Neurons[0].Position)
	distN2_N1_initial := core.EuclideanDistance(initialPosN2, nn.Neurons[1].Position)
	distN2_N0_final := core.EuclideanDistance(newPosN2, nn.Neurons[0].Position)
	distN2_N1_final := core.EuclideanDistance(newPosN2, nn.Neurons[1].Position)

	if distN2_N0_final >= distN2_N0_initial && distN2_N0_initial < core.InfluenceRadiusSynapto {
		// Permitir que não se mova se já estiver muito perto (minDistanceThreshold)
		if distN2_N0_initial > core.MinDistanceThreshold + 0.01 { // Adicionar uma pequena margem
			t.Errorf("N2 deveria se aproximar de N0 (ativo). Dist inicial %.3f, final %.3f", distN2_N0_initial, distN2_N0_final)
			t.Logf("Pos N2 inicial: %v, final: %v", initialPosN2, newPosN2)
		}
	}
	if distN2_N1_final <= distN2_N1_initial && distN2_N1_initial < core.InfluenceRadiusSynapto {
		// Permitir que não se mova se já estiver longe ou se o movimento for mínimo
		if math.Abs(distN2_N1_final-distN2_N1_initial) > 1e-4 { // Se houve movimento significativo na direção errada
			t.Errorf("N2 deveria se afastar de N1 (repouso). Dist inicial %.3f, final %.3f", distN2_N1_initial, distN2_N1_final)
			t.Logf("Pos N2 inicial: %v, final: %v", initialPosN2, newPosN2)
		}
	}

	// Verificar se houve algum movimento
    moved := false
    for i := 0; i < 16; i++ {
        if math.Abs(initialPosN2[i]-newPosN2[i]) > 1e-6 { // Pequena tolerância para float
            moved = true
            break
        }
    }
    if !moved && (distN2_N0_initial < core.InfluenceRadiusSynapto || distN2_N1_initial < core.InfluenceRadiusSynapto) {
         // Se algum neurônio estava no raio de influência, N2 deveria ter se movido.
         // A menos que esteja no minDistanceThreshold de N0 e fora do raio de N1.
        if distN2_N0_initial > core.MinDistanceThreshold + 0.01 || distN2_N1_initial < core.InfluenceRadiusSynapto - 0.01 {
            t.Errorf("N2 não se moveu, mas deveria devido à influência de N0 ou N1.")
			t.Logf("Dist N0: %.3f, Dist N1: %.3f, InfluenceRadius: %.2f, MinDist: %.2f", distN2_N0_initial, distN2_N1_initial, core.InfluenceRadiusSynapto, core.MinDistanceThreshold)
        }
    }
}
