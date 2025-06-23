package network

import (
	"crownet/common"
	"crownet/config"
	"crownet/neuron"
	"crownet/space"
	"fmt"
	"math"
	"math/rand"
	"reflect"
	"strings"
	"testing"
)

// Helper para comparar floats com tolerância
func floatEquals(a, b, tolerance float64) bool {
	if a == b { // Shortcut for exact equality, handles infinities.
		return true
	}
	return math.Abs(a-b) < tolerance
}

func TestCalculateInternalNeuronCounts(t *testing.T) {
	testCases := []struct {
		name                       string
		remainingForDistribution   int
		dopaP                      float64
		inhibP                     float64
		expectedDopa               int
		expectedInhib              int
		expectedExcit              int
		expectWarning              bool
		expectedWarningSubstring   string
	}{
		{
			name: "Distribuição normal",
			remainingForDistribution: 100,
			dopaP:      0.1, // 10
			inhibP:     0.2, // 20
			expectedDopa:  10,
			expectedInhib: 20,
			expectedExcit: 70, // 100 - 10 - 20 = 70
			expectWarning: false,
		},
		{
			name: "Sem neurônios restantes",
			remainingForDistribution: 0,
			dopaP:      0.1,
			inhibP:     0.2,
			expectedDopa:  0,
			expectedInhib: 0,
			expectedExcit: 0,
			expectWarning: false,
		},
		{
			name: "Percentuais zerados",
			remainingForDistribution: 100,
			dopaP:      0.0,
			inhibP:     0.0,
			expectedDopa:  0,
			expectedInhib: 0,
			expectedExcit: 100,
			expectWarning: false,
		},
		{
			name: "Apenas dopaminérgicos",
			remainingForDistribution: 50,
			dopaP:      1.0,
			inhibP:     0.0,
			expectedDopa:  50,
			expectedInhib: 0,
			expectedExcit: 0,
			expectWarning: false,
		},
		{
			name: "Apenas inibitórios",
			remainingForDistribution: 50,
			dopaP:      0.0,
			inhibP:     1.0,
			expectedDopa:  0,
			expectedInhib: 50,
			expectedExcit: 0,
			expectWarning: false,
		},
		{
			name: "Soma de percentuais excede 1.0, precisa de ajuste",
			remainingForDistribution: 100,
			dopaP:      0.7, // 70
			inhibP:     0.5, // 50
			// Total 120, excede 100.
			// Ajuste proporcional: Dopa (70/120 * 100) = 58.33 -> round 58
			// Inhib (50/120 * 100) = 41.66 -> round 42.  100 - 58 = 42
			expectedDopa:  58, // math.Round(100 * (0.7 / 1.2)) = 58
			expectedInhib: 42, // 100 - 58 = 42
			expectedExcit: 0,
			expectWarning: true,
			expectedWarningSubstring: "excedem 100%",
		},
		{
			name: "Percentual dopa negativo (deve ser tratado como 0)",
			remainingForDistribution: 100,
			dopaP:      -0.1,
			inhibP:     0.2,  // 20
			expectedDopa:  0,
			expectedInhib: 20,
			expectedExcit: 80, // 100 - 0 - 20 = 80
			expectWarning: false,
		},
		{
			name: "Percentual inhib negativo (deve ser tratado como 0)",
			remainingForDistribution: 100,
			dopaP:      0.1, // 10
			inhibP:     -0.2,
			expectedDopa:  10,
			expectedInhib: 0,
			expectedExcit: 90, // 100 - 10 - 0 = 90
			expectWarning: false,
		},
		{
			name: "Ambos percentuais negativos",
			remainingForDistribution: 100,
			dopaP:      -0.1,
			inhibP:     -0.2,
			expectedDopa:  0,
			expectedInhib: 0,
			expectedExcit: 100,
			expectWarning: false,
		},
		{
			name: "Caso com arredondamento (Floor)",
			remainingForDistribution: 10,
			dopaP:      0.33, // Floor(3.3) = 3
			inhibP:     0.33, // Floor(3.3) = 3
			expectedDopa:  3,
			expectedInhib: 3,
			expectedExcit: 4, // 10 - 3 - 3 = 4
			expectWarning: false,
		},
		{
            name: "Ajuste com um percentual maior que 1 e outro zero", // Garante que não há divisão por zero se um P for 0 e o outro > 1
            remainingForDistribution: 100,
            dopaP:      1.5, // Excede
            inhibP:     0.0,
            // Ajuste: dopaP = 1.5, inhibP = 0.0. totalInternalPercentConfigured = 1.5
            // numDopaminergic = round(100 * (1.5/1.5)) = 100
            // numInhibitory = 100 - 100 = 0
            expectedDopa:  100,
            expectedInhib: 0,
            expectedExcit: 0,
            expectWarning: true,
            expectedWarningSubstring: "excedem 100%",
        },
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			d, i, e, warnings := calculateInternalNeuronCounts(tc.remainingForDistribution, tc.dopaP, tc.inhibP)

			if d != tc.expectedDopa {
				t.Errorf("Dopaminergic: expected %d, got %d", tc.expectedDopa, d)
			}
			if i != tc.expectedInhib {
				t.Errorf("Inhibitory: expected %d, got %d", tc.expectedInhib, i)
			}
			if e != tc.expectedExcit {
				t.Errorf("Excitatory: expected %d, got %d", tc.expectedExcit, e)
			}

			if tc.expectWarning {
				if len(warnings) == 0 {
					t.Errorf("Expected a warning, but got none")
				} else if tc.expectedWarningSubstring != "" {
					found := false
					for _, w := range warnings {
						if strings.Contains(w, tc.expectedWarningSubstring) {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("Expected warning to contain '%s', got: %v", tc.expectedWarningSubstring, warnings)
					}
				}
			} else {
				if len(warnings) > 0 {
					t.Errorf("Expected no warnings, but got: %v", warnings)
				}
			}
		})
	}
}

func TestNewCrowNet_Initialization(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	simParams.MinInputNeurons = 10
	simParams.MinOutputNeurons = 5
	simParams.DopaminergicPercent = 0.1
	simParams.InhibitoryPercent = 0.2

	totalNeuronsCLI := 100
	baseLR := common.Rate(0.01)
	seed := int64(42) // Semente fixa para o teste

	net := NewCrowNet(totalNeuronsCLI, baseLR, &simParams, seed)

	if net == nil {
		t.Fatalf("NewCrowNet returned nil")
	}

	if net.rng == nil {
		t.Errorf("Expected net.rng to be initialized, got nil")
	}

	if net.SimParams == nil {
		t.Errorf("Expected SimParams to be initialized, got nil")
	} else {
		if net.SimParams.MinInputNeurons != 10 {
			t.Errorf("SimParams.MinInputNeurons: expected %d, got %d", 10, net.SimParams.MinInputNeurons)
		}
	}

	if net.baseLearningRate != baseLR {
		t.Errorf("baseLearningRate: expected %f, got %f", baseLR, net.baseLearningRate)
	}

	expectedTotalNeurons := totalNeuronsCLI
	if totalNeuronsCLI < simParams.MinInputNeurons+simParams.MinOutputNeurons {
		expectedTotalNeurons = simParams.MinInputNeurons + simParams.MinOutputNeurons
	}

	if len(net.Neurons) != expectedTotalNeurons {
		t.Errorf("Total neurons: expected %d, got %d", expectedTotalNeurons, len(net.Neurons))
	}

	if len(net.InputNeuronIDs) != simParams.MinInputNeurons {
		t.Errorf("InputNeuronIDs count: expected %d, got %d", simParams.MinInputNeurons, len(net.InputNeuronIDs))
	}
	if len(net.OutputNeuronIDs) != simParams.MinOutputNeurons {
		t.Errorf("OutputNeuronIDs count: expected %d, got %d", simParams.MinOutputNeurons, len(net.OutputNeuronIDs))
	}

	if len(net.SynapticWeights) != expectedTotalNeurons {
		t.Errorf("Expected %d entries in SynapticWeights map, got %d", expectedTotalNeurons, len(net.SynapticWeights))
	}
	if expectedTotalNeurons > 1 {
		foundNonZeroWeight := false
		for fromID, toMap := range net.SynapticWeights {
			for toID, weight := range toMap {
				if fromID != toID && weight != 0 {
					if float64(weight) < simParams.InitialSynapticWeightMin || float64(weight) > simParams.InitialSynapticWeightMax {
						if !(simParams.InitialSynapticWeightMin == simParams.InitialSynapticWeightMax && floatEquals(float64(weight), simParams.InitialSynapticWeightMin, 1e-9)) {
							t.Errorf("Initial weight %f for %d->%d out of expected range [%f, %f]",
								weight, fromID, toID, simParams.InitialSynapticWeightMin, simParams.InitialSynapticWeightMax)
						}
					}
					foundNonZeroWeight = true
					break
				}
			}
			if foundNonZeroWeight {
				break
			}
		}
		if !foundNonZeroWeight && (expectedTotalNeurons > 1 && simParams.InitialSynapticWeightMax > 0) {
			t.Errorf("SynapticWeights seem to be all zero, InitializeAllToAllWeights might not have run as expected.")
		}
	}

	if net.CycleCount != 0 {
		t.Errorf("Initial CycleCount: expected 0, got %d", net.CycleCount)
	}
	if net.ChemicalEnv == nil {
		t.Errorf("ChemicalEnv should be initialized, got nil")
	}
	if net.ActivePulses == nil {
		t.Errorf("ActivePulses should be initialized, got nil")
	}
}

func TestAddNeuronsOfType(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	simParams.MinInputNeurons = 2
	simParams.MinOutputNeurons = 1
	simParams.SpaceMaxDimension = 10.0

	seed := int64(123)
	net := NewCrowNet(10, 0.01, &simParams, seed)
	initialNeuronCount := len(net.Neurons)
	initialInputIDs := len(net.InputNeuronIDs)
	initialOutputIDs := len(net.OutputNeuronIDs)

	numToAddInput := 3
	net.addNeuronsOfType(numToAddInput, neuron.Input, simParams.ExcitatoryRadiusFactor)
	if len(net.Neurons) != initialNeuronCount+numToAddInput {
		t.Errorf("After adding %d Input neurons, expected %d total, got %d",
			numToAddInput, initialNeuronCount+numToAddInput, len(net.Neurons))
	}
	if len(net.InputNeuronIDs) != initialInputIDs+numToAddInput {
		t.Errorf("After adding %d Input neurons, expected %d InputNeuronIDs, got %d",
			numToAddInput, initialInputIDs+numToAddInput, len(net.InputNeuronIDs))
	}
	for i := initialNeuronCount; i < len(net.Neurons); i++ {
		n := net.Neurons[i]
		if n.Type != neuron.Input {
			t.Errorf("Neuron %d: expected type Input, got %s", n.ID, n.Type)
		}
		distSq := 0.0
		for _, coord := range n.Position {
			distSq += float64(coord * coord)
		}
		if distSq > simParams.SpaceMaxDimension*simParams.SpaceMaxDimension*simParams.ExcitatoryRadiusFactor*simParams.ExcitatoryRadiusFactor + 1e-9 {
			t.Errorf("Neuron %d (Input) position %v is outside radius %f (distSq: %f)",
				n.ID, n.Position, simParams.SpaceMaxDimension*simParams.ExcitatoryRadiusFactor, distSq)
		}
	}

	numToAddOutput := 2
	currentTotalNeurons := len(net.Neurons)
	net.addNeuronsOfType(numToAddOutput, neuron.Output, simParams.ExcitatoryRadiusFactor)
	if len(net.Neurons) != currentTotalNeurons+numToAddOutput {
		t.Errorf("After adding %d Output neurons, expected %d total, got %d",
			numToAddOutput, currentTotalNeurons+numToAddOutput, len(net.Neurons))
	}
	if len(net.OutputNeuronIDs) != initialOutputIDs+numToAddOutput {
		t.Errorf("After adding %d Output neurons, expected %d OutputNeuronIDs, got %d",
			numToAddOutput, initialOutputIDs+numToAddOutput, len(net.OutputNeuronIDs))
	}
}

func TestProcessFrequencyInputs(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	simParams.CyclesPerSecond = 100
	net := NewCrowNet(5, 0.01, &simParams, 123)

	if len(net.Neurons) < 2 {
		t.Fatalf("Necessário pelo menos 2 neurônios para o teste, tem %d", len(net.Neurons))
	}
	inputID1 := net.Neurons[0].ID
	inputID2 := net.Neurons[1].ID
	net.Neurons[0].Type = neuron.Input
	net.Neurons[1].Type = neuron.Input
	net.InputNeuronIDs = []common.NeuronID{inputID1, inputID2}

	net.inputTargetFrequencies[inputID1] = 50.0
	net.timeToNextInputFire[inputID1] = 2
	net.inputTargetFrequencies[inputID2] = 25.0
	net.timeToNextInputFire[inputID2] = 4

	net.CycleCount = 0
	net.processFrequencyInputs()
	if net.Neurons[0].CurrentState == neuron.Firing || net.Neurons[1].CurrentState == neuron.Firing {
		t.Errorf("Ciclo 1: Nenhum neurônio deveria ter disparado por frequência")
	}
	if net.ActivePulses.Count() != 0 {
		t.Errorf("Ciclo 1: Nenhum pulso deveria ter sido gerado, got %d", net.ActivePulses.Count())
	}
	if net.timeToNextInputFire[inputID1] != 1 || net.timeToNextInputFire[inputID2] != 3 {
		t.Errorf("Ciclo 1: timeToNextInputFire incorreto. ID1: %d (exp 1), ID2: %d (exp 3)",
			net.timeToNextInputFire[inputID1], net.timeToNextInputFire[inputID2])
	}

	net.CycleCount = 1
	net.processFrequencyInputs()
	if net.Neurons[0].CurrentState != neuron.Firing {
		t.Errorf("Ciclo 2: Neurônio ID1 (index 0) deveria ter disparado. Estado: %s", net.Neurons[0].CurrentState)
	}
	if net.Neurons[1].CurrentState == neuron.Firing {
		t.Errorf("Ciclo 2: Neurônio ID2 (index 1) NÃO deveria ter disparado. Estado: %s", net.Neurons[1].CurrentState)
	}
	if net.ActivePulses.Count() != 1 {
		t.Errorf("Ciclo 2: Esperado 1 pulso gerado, got %d", net.ActivePulses.Count())
	} else if net.ActivePulses.GetAll()[0].EmittingNeuronID != inputID1 {
		t.Errorf("Ciclo 2: Pulso gerado com ID emissor incorreto")
	}
	if net.timeToNextInputFire[inputID1] != 2 || net.timeToNextInputFire[inputID2] != 2 {
		t.Errorf("Ciclo 2: timeToNextInputFire incorreto. ID1: %d (exp 2), ID2: %d (exp 2)",
			net.timeToNextInputFire[inputID1], net.timeToNextInputFire[inputID2])
	}
	net.Neurons[0].CurrentState = neuron.Resting
	net.ActivePulses.Clear()

	net.CycleCount = 2
	net.processFrequencyInputs()
	if net.Neurons[0].CurrentState == neuron.Firing || net.Neurons[1].CurrentState == neuron.Firing {
		t.Errorf("Ciclo 3: Nenhum neurônio deveria ter disparado por frequência")
	}
	if net.ActivePulses.Count() != 0 {
		t.Errorf("Ciclo 3: Nenhum pulso deveria ter sido gerado, got %d", net.ActivePulses.Count())
	}
	if net.timeToNextInputFire[inputID1] != 1 || net.timeToNextInputFire[inputID2] != 1 {
		t.Errorf("Ciclo 3: timeToNextInputFire incorreto. ID1: %d (exp 1), ID2: %d (exp 1)",
			net.timeToNextInputFire[inputID1], net.timeToNextInputFire[inputID2])
	}

	net.CycleCount = 3
	net.processFrequencyInputs()
	if net.Neurons[0].CurrentState != neuron.Firing {
		t.Errorf("Ciclo 4: Neurônio ID1 (index 0) deveria ter disparado. Estado: %s", net.Neurons[0].CurrentState)
	}
	if net.Neurons[1].CurrentState != neuron.Firing {
		t.Errorf("Ciclo 4: Neurônio ID2 (index 1) deveria ter disparado. Estado: %s", net.Neurons[1].CurrentState)
	}
	if net.ActivePulses.Count() != 2 {
		t.Errorf("Ciclo 4: Esperado 2 pulsos gerados, got %d", net.ActivePulses.Count())
	}
	if net.timeToNextInputFire[inputID1] != 2 || net.timeToNextInputFire[inputID2] != 4 {
		t.Errorf("Ciclo 4: timeToNextInputFire incorreto. ID1: %d (exp 2), ID2: %d (exp 4)",
			net.timeToNextInputFire[inputID1], net.timeToNextInputFire[inputID2])
	}
}

func TestPresentPattern(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	simParams.PatternSize = 3
	simParams.MinInputNeurons = 3
	net := NewCrowNet(5, 0.01, &simParams, 456)

	for i := 0; i < simParams.MinInputNeurons; i++ {
		if i < len(net.Neurons) {
			net.Neurons[i].Type = neuron.Input
		} else {
			t.Fatalf("Rede não tem neurônios suficientes para configurar como Input para o teste.")
		}
	}
	net.InputNeuronIDs = make([]common.NeuronID, 0, simParams.MinInputNeurons)
	for i:=0; i < simParams.MinInputNeurons; i++ {
		net.InputNeuronIDs = append(net.InputNeuronIDs, net.Neurons[i].ID)
	}

	pattern1 := []float64{1.0, 0.0, 1.0}
	net.CycleCount = 0
	err := net.PresentPattern(pattern1)
	if err != nil {
		t.Fatalf("PresentPattern com padrão válido retornou erro: %v", err)
	}
	if net.Neurons[0].CurrentState != neuron.Firing {
		t.Errorf("Neurônio de Input 0 deveria estar Firing")
	}
	if net.Neurons[1].CurrentState == neuron.Firing {
		t.Errorf("Neurônio de Input 1 NÃO deveria estar Firing")
	}
	if net.Neurons[2].CurrentState != neuron.Firing {
		t.Errorf("Neurônio de Input 2 deveria estar Firing")
	}
	if net.ActivePulses.Count() != 2 {
		t.Errorf("Esperado 2 pulsos ativos, got %d", net.ActivePulses.Count())
	}
	emitters := make(map[common.NeuronID]bool)
	for _, p := range net.ActivePulses.GetAll() {
		emitters[p.EmittingNeuronID] = true
	}
	if !emitters[net.InputNeuronIDs[0]] || !emitters[net.InputNeuronIDs[2]] {
		t.Errorf("Pulsos gerados com IDs emissores incorretos: %v", emitters)
	}

	for _, n := range net.Neurons { n.CurrentState = neuron.Resting }
	net.ActivePulses.Clear()

	pattern2 := []float64{1.0, 0.0}
	err = net.PresentPattern(pattern2)
	if err == nil {
		t.Errorf("PresentPattern deveria retornar erro para tamanho de padrão incorreto, mas não retornou")
	}

	originalInputIDs := net.InputNeuronIDs
	net.InputNeuronIDs = []common.NeuronID{net.Neurons[0].ID}
	simParams.PatternSize = 2
	pattern3 := []float64{1.0, 1.0}
	err = net.PresentPattern(pattern3)
	if err == nil {
		t.Errorf("PresentPattern deveria retornar erro para InputNeuronIDs insuficientes, mas não retornou")
	}
	net.InputNeuronIDs = originalInputIDs
	simParams.PatternSize = 3
}

func TestConfigureFrequencyInput(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	simParams.CyclesPerSecond = 100.0
	net := NewCrowNet(3, 0.01, &simParams, 789)

	inputID := net.Neurons[0].ID
	net.Neurons[0].Type = neuron.Input
	net.InputNeuronIDs = []common.NeuronID{inputID}

	err := net.ConfigureFrequencyInput(inputID, 10.0)
	if err != nil {
		t.Fatalf("ConfigureFrequencyInput retornou erro inesperado: %v", err)
	}
	if targetHz, ok := net.inputTargetFrequencies[inputID]; !ok || !floatEquals(targetHz, 10.0, 1e-9) {
		t.Errorf("inputTargetFrequencies incorreto: esperado 10.0, got %f (ok: %t)", targetHz, ok)
	}
	if timeLeft, ok := net.timeToNextInputFire[inputID]; !ok || timeLeft <= 0 || timeLeft > 10 {
		t.Errorf("timeToNextInputFire incorreto: esperado >0 e <=10, got %d (ok: %t)", timeLeft, ok)
	}

	err = net.ConfigureFrequencyInput(inputID, 0.0)
	if err != nil {
		t.Fatalf("ConfigureFrequencyInput (hz=0) retornou erro inesperado: %v", err)
	}
	if _, ok := net.inputTargetFrequencies[inputID]; ok {
		t.Errorf("inputTargetFrequencies deveria ter sido removido para hz=0")
	}
	if _, ok := net.timeToNextInputFire[inputID]; ok {
		t.Errorf("timeToNextInputFire deveria ter sido removido para hz=0")
	}

	invalidID := common.NeuronID(999)
	err = net.ConfigureFrequencyInput(invalidID, 10.0)
	if err == nil {
		t.Errorf("ConfigureFrequencyInput deveria retornar erro para ID inválido, mas não retornou")
	}
}

func TestGetOutputFrequency(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	simParams.CyclesPerSecond = 100.0
	simParams.OutputFrequencyWindowCycles = 50.0
	net := NewCrowNet(3, 0.01, &simParams, 101112)

	outputID := net.Neurons[0].ID
	net.Neurons[0].Type = neuron.Output
	net.OutputNeuronIDs = []common.NeuronID{outputID}

	freq, err := net.GetOutputFrequency(outputID)
	if err != nil {
		t.Fatalf("GetOutputFrequency (sem histórico) retornou erro: %v", err)
	}
	if !floatEquals(freq, 0.0, 1e-9) {
		t.Errorf("Frequência esperada 0.0 para sem histórico, got %f", freq)
	}

	net.CycleCount = 100
	net.outputFiringHistory[outputID] = []common.CycleCount{
		60, 70, 80, 90, 100,
	}
	freq, err = net.GetOutputFrequency(outputID)
	if err != nil {
		t.Fatalf("GetOutputFrequency (com histórico) retornou erro: %v", err)
	}
	if !floatEquals(freq, 10.0, 1e-9) {
		t.Errorf("Frequência esperada 10.0 Hz, got %f", freq)
	}

	net.outputFiringHistory[outputID] = []common.CycleCount{}
	freq, err = net.GetOutputFrequency(outputID)
	if err != nil {
		t.Fatalf("GetOutputFrequency (disparos fora da janela) retornou erro: %v", err)
	}
	if !floatEquals(freq, 0.0, 1e-9) {
		t.Errorf("Frequência esperada 0.0 Hz para disparos fora da janela, got %f", freq)
	}

	invalidID := common.NeuronID(999)
	_, err = net.GetOutputFrequency(invalidID)
	if err == nil {
		t.Errorf("GetOutputFrequency deveria retornar erro para ID inválido, mas não retornou")
	}

	simParamsZeroHz := simParams
	simParamsZeroHz.CyclesPerSecond = 0.0
	netZeroHz := NewCrowNet(3, 0.01, &simParamsZeroHz, 101113)
	netZeroHz.Neurons[0].Type = neuron.Output
	netZeroHz.OutputNeuronIDs = []common.NeuronID{netZeroHz.Neurons[0].ID}
	netZeroHz.outputFiringHistory[netZeroHz.Neurons[0].ID] = []common.CycleCount{1}

	_, err = netZeroHz.GetOutputFrequency(netZeroHz.Neurons[0].ID)
	if err == nil {
		t.Errorf("GetOutputFrequency deveria retornar erro se CyclesPerSecond é zero, mas não retornou")
	}
}

func TestCalculateNetForceOnNeuron(t *testing.T) {
	simP := config.DefaultSimulationParameters()
	simP.SynaptogenesisInfluenceRadius = 2.0
	simP.AttractionForceFactor = 1.0
	simP.RepulsionForceFactor = 0.5

	n1 := neuron.New(0, neuron.Excitatory, common.Point{0, 0}, &simP)
	n2 := neuron.New(1, neuron.Excitatory, common.Point{1, 0}, &simP)
	n3 := neuron.New(2, neuron.Inhibitory, common.Point{0, 1}, &simP)

	allNeurons := []*neuron.Neuron{n1, n2, n3}
	modulationFactor := 1.0

	n2.CurrentState = neuron.Firing
	n3.CurrentState = neuron.Resting
	netForce := calculateNetForceOnNeuron(n1, allNeurons, &simP, modulationFactor)
	if !floatEquals(netForce[0], 1.0, 1e-9) || !floatEquals(netForce[1], -0.5, 1e-9) {
		t.Errorf("Test Case 1: Expected net force {1.0, -0.5}, got %v", netForce)
	}

	n2.CurrentState = neuron.Resting
	n3.CurrentState = neuron.Firing
	netForce = calculateNetForceOnNeuron(n1, allNeurons, &simP, modulationFactor)
	if !floatEquals(netForce[0], -0.5, 1e-9) || !floatEquals(netForce[1], 1.0, 1e-9) {
		t.Errorf("Test Case 2: Expected net force {-0.5, 1.0}, got %v", netForce)
	}

	n4 := neuron.New(3, neuron.Excitatory, common.Point{10, 10}, &config.SimulationParameters{})
	n4.CurrentState = neuron.Firing
	allNeuronsFar := []*neuron.Neuron{n1, n4}
	simParamsNear := simP
	simParamsNear.SynaptogenesisInfluenceRadius = 1.0

	netForceFar := calculateNetForceOnNeuron(n1, allNeuronsFar, &simParamsNear, modulationFactor)
	if !floatEquals(netForceFar[0], 0.0, 1e-9) || !floatEquals(netForceFar[1], 0.0, 1e-9) {
		 t.Errorf("Test Case 3: Expected zero force due to distance, got %v", netForceFar)
	}
}

func TestUpdateNeuronMovement(t *testing.T) {
	simParams := &config.SimulationParameters{
		DampeningFactor:     0.9,
		MaxMovementPerCycle: 1.0,
		SpaceMaxDimension:   100.0,
	}
	n := neuron.New(0, neuron.Excitatory, common.Point{0, 0}, simParams)
	n.Velocity = common.Vector{0.1, -0.1}

	netForce := common.Vector{0.5, 0.5}
	newPos, newVel := updateNeuronMovement(n, netForce, simParams)
	if !floatEquals(newVel[0], 0.59, 1e-9) || !floatEquals(newVel[1], 0.41, 1e-9) {
		t.Errorf("Test Case 1 Velocity: Expected {0.59, 0.41}, got %v", newVel)
	}
	if !floatEquals(float64(newPos[0]), 0.59, 1e-9) || !floatEquals(float64(newPos[1]), 0.41, 1e-9) {
		t.Errorf("Test Case 1 Position: Expected {0.59, 0.41}, got %v", newPos)
	}

	n.Position = common.Point{0,0}
	n.Velocity = common.Vector{0,0}
	netForceLarge := common.Vector{2.0, 0}
	newPosCapped, newVelCapped := updateNeuronMovement(n, netForceLarge, simParams)
	velMagnitude := math.Sqrt(newVelCapped[0]*newVelCapped[0] + newVelCapped[1]*newVelCapped[1])
	if !floatEquals(velMagnitude, simParams.MaxMovementPerCycle, 1e-9) {
        if !(netForceLarge[0] == 0 && netForceLarge[1] == 0) {
		    t.Errorf("Test Case 2 Velocity Magnitude: Expected to be capped at %.2f, got %.2f (vel: %v)", simParams.MaxMovementPerCycle, velMagnitude, newVelCapped)
        }
	}
    if velMagnitude > 1e-9 && simParams.MaxMovementPerCycle > 1e-9 {
        if !floatEquals(newVelCapped[0], 1.0, 1e-9) || !floatEquals(newVelCapped[1], 0.0, 1e-9) {
             t.Errorf("Test Case 2 Velocity Components: Expected {1.0, 0.0} after cap, got %v", newVelCapped)
        }
    }
	if !floatEquals(float64(newPosCapped[0]), 1.0, 1e-9) || !floatEquals(float64(newPosCapped[1]), 0.0, 1e-9) {
		t.Errorf("Test Case 2 Position: Expected {1.0, 0.0}, got %v", newPosCapped)
	}

	simParamsClamped := &config.SimulationParameters{
		DampeningFactor:     1.0,
		MaxMovementPerCycle: 10.0,
		SpaceMaxDimension:   0.5,
	}
	nClamp := neuron.New(1, neuron.Excitatory, common.Point{0.4, 0.0}, simParamsClamped)
	nClamp.Velocity = common.Vector{0.0, 0.0}
	forceToClamp := common.Vector{0.2, 0.0}
	newPosClamped, _ := updateNeuronMovement(nClamp, forceToClamp, simParamsClamped)
	if !floatEquals(float64(newPosClamped[0]), 0.5, 1e-9) || !floatEquals(float64(newPosClamped[1]), 0.0, 1e-9) {
		t.Errorf("Test Case 3 Position Clamping: Expected {0.5, 0.0}, got {%f, %f}", float64(newPosClamped[0]), float64(newPosClamped[1]))
	}
}

func TestRecordOutputFiring(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	simParams.OutputFrequencyWindowCycles = 10
	net := NewCrowNet(5, 0.01, &simParams, 131415)

	if len(net.Neurons) == 0 {
		t.Fatal("NewCrowNet não criou neurônios")
	}
	outputID := net.Neurons[0].ID
	net.Neurons[0].Type = neuron.Output
	net.OutputNeuronIDs = []common.NeuronID{outputID}
	if _, ok := net.outputFiringHistory[outputID]; !ok {
		net.outputFiringHistory[outputID] = make([]common.CycleCount, 0)
	}

	net.CycleCount = 5
	net.recordOutputFiring(outputID)
	if history, ok := net.outputFiringHistory[outputID]; !ok || len(history) != 1 || history[0] != 5 {
		t.Errorf("Histórico C1 incorreto: esperado [5], got %v (ok: %t)", history, ok)
	}

	net.CycleCount = 7
	net.recordOutputFiring(outputID)
	net.CycleCount = 9
	net.recordOutputFiring(outputID)
	expectedHistory2 := []common.CycleCount{5, 7, 9}
	if history, ok := net.outputFiringHistory[outputID]; !ok || !reflect.DeepEqual(history, expectedHistory2) {
		t.Errorf("Histórico C2 incorreto: esperado %v, got %v", expectedHistory2, history)
	}

	net.CycleCount = 16
	net.recordOutputFiring(outputID)
	expectedHistory3 := []common.CycleCount{7, 9, 16}
	if history, ok := net.outputFiringHistory[outputID]; !ok || !reflect.DeepEqual(history, expectedHistory3) {
		t.Errorf("Histórico C3 (poda): esperado %v, got %v (cutoff: %d)", expectedHistory3, history, net.CycleCount - common.CycleCount(simParams.OutputFrequencyWindowCycles))
	}

	nonOutputID := common.NeuronID(99)
	originalHistoryLen := len(net.outputFiringHistory[outputID])
	net.recordOutputFiring(nonOutputID)
	if _, ok := net.outputFiringHistory[nonOutputID]; ok {
		t.Errorf("Histórico não deveria ter sido criado para nonOutputID")
	}
	if len(net.outputFiringHistory[outputID]) != originalHistoryLen {
		t.Errorf("Histórico do outputID mudou indevidamente (%d vs %d) após registrar para nonOutputID", len(net.outputFiringHistory[outputID]), originalHistoryLen)
	}
}

func TestMain(m *testing.M) {
	rand.Seed(1)
	_ = m.Run()
}
