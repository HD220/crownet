package neuron_test

import (
	"crownet/common"
	"crownet/config"
	"crownet/neuron"
	"testing"
)

func TestNewNeuron(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	id := common.NeuronID(1)
	pos := common.Point{1, 2, 3}
	n := neuron.New(id, neuron.Excitatory, pos, &simParams)

	if n.ID != id {
		t.Errorf("Esperado ID %d, obteve %d", id, n.ID)
	}
	if n.Type != neuron.Excitatory {
		t.Errorf("Esperado tipo Excitatory, obteve %s", n.Type)
	}
	if n.Position != pos {
		t.Errorf("Esperado Posição %v, obteve %v", pos, n.Position)
	}
	if n.CurrentState != neuron.Resting {
		t.Errorf("Esperado estado Resting, obteve %s", n.CurrentState)
	}
	if n.AccumulatedPotential != 0.0 {
		t.Errorf("Esperado Potencial Acumulado 0.0, obteve %f", n.AccumulatedPotential)
	}
	if n.BaseFiringThreshold != common.Threshold(simParams.BaseFiringThreshold) {
		t.Errorf("Esperado Limiar Base %f, obteve %f", simParams.BaseFiringThreshold, n.BaseFiringThreshold)
	}
	if n.LastFiredCycle != -1 {
		t.Errorf("Esperado Último Ciclo de Disparo -1, obteve %d", n.LastFiredCycle)
	}
}

func TestNeuronIntegrateIncomingPotential(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	n := neuron.New(1, neuron.Excitatory, common.Point{}, &simParams)

	fired := n.IntegrateIncomingPotential(common.PulseValue(simParams.BaseFiringThreshold-0.1), 0)
	if fired {
		t.Errorf("Neurônio disparou com potencial abaixo do limiar")
	}
	if n.AccumulatedPotential != common.PulseValue(simParams.BaseFiringThreshold-0.1) {
		t.Errorf("Potencial acumulado incorreto: esperado %f, obteve %f", simParams.BaseFiringThreshold-0.1, n.AccumulatedPotential)
	}
	if n.CurrentState != neuron.Resting {
		t.Errorf("Estado incorreto após potencial abaixo do limiar: esperado Resting, obteve %s", n.CurrentState)
	}

	n.AccumulatedPotential = 0
	fired = n.IntegrateIncomingPotential(common.PulseValue(simParams.BaseFiringThreshold), 1)
	if !fired {
		t.Errorf("Neurônio não disparou com potencial no limiar")
	}
	if n.CurrentState != neuron.Firing {
		t.Errorf("Estado incorreto após disparo: esperado Firing, obteve %s", n.CurrentState)
	}
}

func TestNeuronAdvanceState(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	n := neuron.New(1, neuron.Excitatory, common.Point{}, &simParams)

	n.CurrentState = neuron.Firing
	n.CyclesInCurrentState = 0
	n.AccumulatedPotential = common.PulseValue(simParams.BaseFiringThreshold + 0.1)

	n.AdvanceState(0, &simParams)

	if n.CurrentState != neuron.AbsoluteRefractory {
		t.Errorf("Esperado estado AbsoluteRefractory após Firing, obteve %s", n.CurrentState)
	}
	if n.LastFiredCycle != 0 {
		t.Errorf("Esperado LastFiredCycle 0, obteve %d", n.LastFiredCycle)
	}
	if n.AccumulatedPotential != 0.0 {
		t.Errorf("Esperado Potencial Acumulado 0.0 após Firing, obteve %f", n.AccumulatedPotential)
	}
	if n.CyclesInCurrentState != 0 {
		t.Errorf("Esperado CyclesInCurrentState 0 para novo estado AbsoluteRefractory, obteve %d", n.CyclesInCurrentState)
	}

	n.CurrentState = neuron.AbsoluteRefractory
	n.LastFiredCycle = 0
	n.CyclesInCurrentState = simParams.AbsoluteRefractoryCycles - 1

	n.AdvanceState(simParams.AbsoluteRefractoryCycles, &simParams)

	if n.CurrentState != neuron.RelativeRefractory {
		t.Errorf("Esperado estado RelativeRefractory após AbsoluteRefractory, obteve %s", n.CurrentState)
	}
	if n.CyclesInCurrentState != 0 {
		t.Errorf("Esperado CyclesInCurrentState 0 para novo estado RelativeRefractory, obteve %d", n.CyclesInCurrentState)
	}

	n.CurrentState = neuron.RelativeRefractory
	n.LastFiredCycle = 0
	n.CyclesInCurrentState = simParams.RelativeRefractoryCycles - 1

	n.AdvanceState(simParams.AbsoluteRefractoryCycles + simParams.RelativeRefractoryCycles, &simParams)

	if n.CurrentState != neuron.Resting {
		t.Errorf("Esperado estado Resting após RelativeRefractory, obteve %s", n.CurrentState)
	}
	if n.CyclesInCurrentState != 0 {
		t.Errorf("Esperado CyclesInCurrentState 0 para novo estado Resting, obteve %d", n.CyclesInCurrentState)
	}
}

func TestNeuronDecayPotential(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	n := neuron.New(1, neuron.Excitatory, common.Point{}, &simParams)

	n.AccumulatedPotential = 1.0
	expectedPotential := 1.0 * (1.0 - simParams.AccumulatedPulseDecayRate)
	n.DecayPotential(&simParams)
	if n.AccumulatedPotential != common.PulseValue(expectedPotential) {
		t.Errorf("Decaimento positivo incorreto: esperado %f, obteve %f", expectedPotential, n.AccumulatedPotential)
	}

	n.AccumulatedPotential = -1.0
	expectedPotentialNegative := -1.0 * (1.0 - simParams.AccumulatedPulseDecayRate)
	n.DecayPotential(&simParams)
	if n.AccumulatedPotential != common.PulseValue(expectedPotentialNegative) {
		t.Errorf("Decaimento negativo incorreto: esperado %f, obteve %f", expectedPotentialNegative, n.AccumulatedPotential)
	}

	n.AccumulatedPotential = 0.000001
	n.DecayPotential(&simParams)
	if n.AccumulatedPotential != 0.0 {
		t.Errorf("Decaimento para zero (positivo) falhou: esperado 0.0, obteve %f", n.AccumulatedPotential)
	}

	n.AccumulatedPotential = -0.000001
	n.DecayPotential(&simParams)
	if n.AccumulatedPotential != 0.0 {
		t.Errorf("Decaimento para zero (negativo) falhou: esperado 0.0, obteve %f", n.AccumulatedPotential)
	}
}

func TestEmittedPulseSign(t *testing.T) {
	simParams := config.DefaultSimulationParameters()

	testCases := []struct {
		name         string
		neuronType   neuron.Type
		expectedSign common.PulseValue
	}{
		{"Excitatory", neuron.Excitatory, 1.0},
		{"Inhibitory", neuron.Inhibitory, -1.0},
		{"Input", neuron.Input, 1.0},
		{"Output", neuron.Output, 1.0},
		{"Dopaminergic", neuron.Dopaminergic, 0.0},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			n := neuron.New(1, tc.neuronType, common.Point{}, &simParams)
			sign := n.EmittedPulseSign()
			if sign != tc.expectedSign {
				t.Errorf("Para tipo %s, esperado sinal %f, obteve %f", tc.neuronType, tc.expectedSign, sign)
			}
		})
	}
}

func TestNeuronNoFireInAbsoluteRefractory(t *testing.T) {
    simParams := config.DefaultSimulationParameters()
    n := neuron.New(1, neuron.Excitatory, common.Point{}, &simParams)

    n.CurrentState = neuron.Firing
    n.AdvanceState(0, &simParams)

    if n.CurrentState != neuron.AbsoluteRefractory {
        t.Fatalf("Falha ao colocar neurônio em AbsoluteRefractory para o teste.")
    }

    n.AccumulatedPotential = 0
    fired := n.IntegrateIncomingPotential(common.PulseValue(simParams.BaseFiringThreshold*2), 1)

    if fired {
        t.Errorf("Neurônio disparou enquanto em AbsoluteRefractory state.")
    }
    if n.CurrentState != neuron.AbsoluteRefractory {
        t.Errorf("Estado do neurônio mudou de AbsoluteRefractory indevidamente.")
    }
    if n.LastFiredCycle != 0 {
	    t.Errorf("LastFiredCycle mudou, indicando novo disparo em AbsoluteRefractory. Esperado 0, obteve %d", n.LastFiredCycle)
    }
}

func TestNeuronUpdatePosition(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	n := neuron.New(1, neuron.Excitatory, common.Point{1.0, 2.0, 3.0}, &simParams)
	n.Velocity = common.Vector{0.5, -0.5, 0.1}

	for i := 3; i < 16; i++ {
		n.Position[i] = common.Coordinate(float64(i + 1))
		n.Velocity[i] = 0.0
	}

	expectedPosition := n.Position
	for i := 0; i < 16; i++ {
		expectedPosition[i] += common.Coordinate(n.Velocity[i])
	}

	n.UpdatePosition()

	if n.Position != expectedPosition {
		t.Errorf("UpdatePosition falhou: esperado %v, obteve %v", expectedPosition, n.Position)
	}

	n.Position = common.Point{1.0, 1.0}
	n.Velocity = common.Vector{0.0, 0.0}
	expectedPositionZeroVel := n.Position

	n.UpdatePosition()
	if n.Position != expectedPositionZeroVel {
		t.Errorf("UpdatePosition com velocidade zero falhou: esperado %v, obteve %v", expectedPositionZeroVel, n.Position)
	}
}
