package datagen_test

import (
	"crownet/config"
	"crownet/datagen"
	"reflect"
	"testing"
)

func TestGetDigitPattern_ValidDigit(t *testing.T) {
	simParams := config.DefaultSimulationParameters() // Usa PatternHeight=7, PatternWidth=5, PatternSize=35

	pattern, err := datagen.GetDigitPatternFn(0, &simParams)
	if err != nil {
		t.Fatalf("GetDigitPatternFn(0) returned error: %v", err)
	}

	if len(pattern) != simParams.PatternSize {
		t.Errorf("Expected pattern length %d, got %d", simParams.PatternSize, len(pattern))
	}

	// Verificar alguns valores do padrão do dígito 0 (primeira e última linha de 1s)
	// Padrão 0:
	// {1,1,1,1,1}, (0-4)
	// {1,0,0,0,1}, (5-9)
	// ...
	// {1,1,1,1,1}, (30-34)
	expectedPrefix := []float64{1,1,1,1,1, 1,0,0,0,1}
	if !reflect.DeepEqual(pattern[:10], expectedPrefix) {
		t.Errorf("Pattern for digit 0, prefix mismatch. Expected %v, got %v", expectedPrefix, pattern[:10])
	}
	expectedSuffix := []float64{1,0,0,0,1, 1,1,1,1,1} // Últimas duas linhas do '0'
	if !reflect.DeepEqual(pattern[simParams.PatternSize-10:], expectedSuffix) {
		t.Errorf("Pattern for digit 0, suffix mismatch. Expected %v, got %v", expectedSuffix, pattern[simParams.PatternSize-10:])
	}
}

func TestGetDigitPattern_InvalidDigit(t *testing.T) {
	simParams := config.DefaultSimulationParameters()

	_, err := datagen.GetDigitPatternFn(10, &simParams)
	if err == nil {
		t.Errorf("Expected error for invalid digit 10, got nil")
	}

	_, err = datagen.GetDigitPatternFn(-1, &simParams)
	if err == nil {
		t.Errorf("Expected error for invalid digit -1, got nil")
	}
}

func TestGetDigitPattern_DimensionMismatch(t *testing.T) {
	// Caso 1: PatternHeight em simParams não corresponde aos dados
	simParams1 := config.DefaultSimulationParameters()
	simParams1.PatternHeight = 6 // Dados são 7
	_, err1 := datagen.GetDigitPatternFn(0, &simParams1)
	if err1 == nil {
		t.Errorf("Expected error when simParams.PatternHeight mismatches data, got nil")
	} else {
		// t.Logf("Got expected error for height mismatch: %v", err1) // Log opcional
	}

	// Caso 2: PatternWidth em simParams não corresponde aos dados
	simParams2 := config.DefaultSimulationParameters()
	simParams2.PatternWidth = 4 // Dados são 5
	_, err2 := datagen.GetDigitPatternFn(0, &simParams2)
	if err2 == nil {
		t.Errorf("Expected error when simParams.PatternWidth mismatches data, got nil")
	} else {
		// t.Logf("Got expected error for width mismatch: %v", err2)
	}

	// Caso 3: PatternSize em simParams não é PatternHeight * PatternWidth
	simParams3 := config.DefaultSimulationParameters() // H=7, W=5, Size=35
	simParams3.PatternSize = 30 // Inconsistente
	_, err3 := datagen.GetDigitPatternFn(0, &simParams3)
	if err3 == nil {
		t.Errorf("Expected error when simParams.PatternSize is inconsistent, got nil")
	} else {
		// t.Logf("Got expected error for size inconsistency: %v", err3)
	}

	// Caso 4: PatternHeight ou PatternWidth em simParams é zero ou negativo
	simParams4 := config.DefaultSimulationParameters()
	simParams4.PatternHeight = 0
	_, err4 := datagen.GetDigitPatternFn(0, &simParams4)
	if err4 == nil {
		t.Errorf("Expected error for zero PatternHeight, got nil")
	}
}

func TestGetAllDigitPatterns_Valid(t *testing.T) {
	simParams := config.DefaultSimulationParameters()
	allPatterns, err := datagen.GetAllDigitPatterns(&simParams)

	if err != nil {
		t.Fatalf("GetAllDigitPatterns returned error: %v", err)
	}

	if len(allPatterns) != 10 {
		t.Errorf("Expected 10 patterns (0-9), got %d", len(allPatterns))
	}

	for i := 0; i <= 9; i++ {
		if _, ok := allPatterns[i]; !ok {
			t.Errorf("Pattern for digit %d missing from GetAllDigitPatterns result", i)
		}
	}

	// Verificar um padrão específico, por exemplo, dígito 1
	pattern1, ok := allPatterns[1]
	if !ok {
		t.Fatalf("Pattern for digit 1 not found in map") // Já coberto acima, mas para segurança
	}
	// Padrão 1 (achatado):
	// {0,0,1,0,0}, (0-4)
	// {0,1,1,0,0}, (5-9)
	// {0,0,1,0,0}, (10-14)
	// {0,0,1,0,0}, (15-19)
	// {0,0,1,0,0}, (20-24)
	// {0,0,1,0,0}, (25-29)
	// {0,1,1,1,0}, (30-34)
	expectedPattern1_prefix := []float64{0,0,1,0,0, 0,1,1,0,0}
	if len(pattern1) != simParams.PatternSize {
		t.Errorf("Pattern for digit 1 has incorrect length %d, expected %d", len(pattern1), simParams.PatternSize)
	} else if !reflect.DeepEqual(pattern1[:10], expectedPattern1_prefix) {
		t.Errorf("Pattern for digit 1, prefix mismatch. Expected %v, got %v", expectedPattern1_prefix, pattern1[:10])
	}
}

func TestGetAllDigitPatterns_ErrorPropagation(t *testing.T) {
	// Forçar um erro em GetDigitPattern alterando simParams para ser inválido
	simParamsError := config.DefaultSimulationParameters()
	simParamsError.PatternHeight = -1 // Isso causará erro em GetDigitPattern

	_, err := datagen.GetAllDigitPatterns(&simParamsError)
	if err == nil {
		t.Errorf("Expected GetAllDigitPatterns to propagate error from GetDigitPattern, but got nil error")
	} else {
		// t.Logf("Got expected propagated error: %v", err) // Log opcional
	}
}
