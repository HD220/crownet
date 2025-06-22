package datagen

import (
	"crownet/config" // Para ter acesso a PatternSize, PatternHeight, PatternWidth
	"fmt"
)

// digitPatternData armazena os padrões brutos.
// Usar um mapa de int para [][]int é simples para definição.
var digitPatternData = map[int][][]int{
	0: {
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{1, 1, 1, 1, 1},
	},
	1: {
		{0, 0, 1, 0, 0},
		{0, 1, 1, 0, 0},
		{0, 0, 1, 0, 0},
		{0, 0, 1, 0, 0},
		{0, 0, 1, 0, 0},
		{0, 0, 1, 0, 0},
		{0, 1, 1, 1, 0},
	},
	2: {
		{1, 1, 1, 1, 0},
		{0, 0, 0, 0, 1},
		{0, 0, 0, 0, 1},
		{0, 1, 1, 1, 0},
		{1, 0, 0, 0, 0},
		{1, 0, 0, 0, 0},
		{1, 1, 1, 1, 1},
	},
	3: {
		{1, 1, 1, 1, 0},
		{0, 0, 0, 0, 1},
		{0, 0, 1, 1, 0},
		{0, 0, 0, 0, 1},
		{0, 0, 0, 0, 1},
		{0, 0, 0, 0, 1},
		{1, 1, 1, 1, 0},
	},
	4: {
		{1, 0, 0, 1, 0},
		{1, 0, 0, 1, 0},
		{1, 0, 0, 1, 0},
		{1, 1, 1, 1, 1},
		{0, 0, 0, 1, 0},
		{0, 0, 0, 1, 0},
		{0, 0, 0, 1, 0},
	},
	5: {
		{1, 1, 1, 1, 1},
		{1, 0, 0, 0, 0},
		{1, 1, 1, 1, 0},
		{0, 0, 0, 0, 1},
		{0, 0, 0, 0, 1},
		{0, 0, 0, 0, 1},
		{1, 1, 1, 1, 0},
	},
	6: {
		{0, 1, 1, 1, 0},
		{1, 0, 0, 0, 0},
		{1, 0, 0, 0, 0},
		{1, 1, 1, 1, 0},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{0, 1, 1, 1, 0},
	},
	7: {
		{1, 1, 1, 1, 1},
		{0, 0, 0, 0, 1},
		{0, 0, 0, 1, 0},
		{0, 0, 1, 0, 0},
		{0, 1, 0, 0, 0},
		{0, 1, 0, 0, 0},
		{0, 1, 0, 0, 0},
	},
	8: {
		{0, 1, 1, 1, 0},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{0, 1, 1, 1, 0},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{0, 1, 1, 1, 0},
	},
	9: {
		{0, 1, 1, 1, 0},
		{1, 0, 0, 0, 1},
		{1, 0, 0, 0, 1},
		{0, 1, 1, 1, 1},
		{0, 0, 0, 0, 1},
		{0, 0, 0, 0, 1},
		{0, 1, 1, 1, 0},
	},
}

// GetDigitPatternFn é uma variável de função para permitir o mock em testes.
// A implementação real é getDigitPatternInternal.
var GetDigitPatternFn = getDigitPatternInternal

// getDigitPatternInternal é a implementação real de GetDigitPattern.
// Os valores são 1.0 para "ligado" e 0.0 para "desligado".
// Utiliza os parâmetros de SimParams para validação de dimensões.
func getDigitPatternInternal(digit int, simParams *config.SimulationParameters) ([]float64, error) {
	pattern2D, ok := digitPatternData[digit]
	if !ok {
		return nil, fmt.Errorf("padrão de dígito para %d não encontrado", digit)
	}

	// Validar dimensões com base em simParams
	if simParams.PatternHeight <= 0 || simParams.PatternWidth <= 0 {
		return nil, fmt.Errorf("PatternHeight (%d) e PatternWidth (%d) em simParams devem ser positivos", simParams.PatternHeight, simParams.PatternWidth)
	}

	expectedSize := simParams.PatternHeight * simParams.PatternWidth
	if simParams.PatternSize != expectedSize {
		// Esta é uma verificação de consistência interna da configuração.
		return nil, fmt.Errorf(
			"PatternSize (%d) em simParams não corresponde a PatternHeight (%d) * PatternWidth (%d) = %d",
			simParams.PatternSize, simParams.PatternHeight, simParams.PatternWidth, expectedSize,
		)
	}

	if len(pattern2D) != simParams.PatternHeight {
		return nil, fmt.Errorf("padrão para dígito %d tem altura incorreta %d, esperado %d (de simParams)", digit, len(pattern2D), simParams.PatternHeight)
	}

	flattenedPattern := make([]float64, 0, simParams.PatternSize)
	for r, row := range pattern2D {
		if len(row) != simParams.PatternWidth {
			return nil, fmt.Errorf("padrão para dígito %d, linha %d tem largura incorreta %d, esperado %d (de simParams)", digit, r, len(row), simParams.PatternWidth)
		}
		for _, val := range row {
			if val == 1 {
				flattenedPattern = append(flattenedPattern, 1.0)
			} else {
				flattenedPattern = append(flattenedPattern, 0.0)
			}
		}
	}

	if len(flattenedPattern) != simParams.PatternSize {
		// Isso não deveria acontecer se as validações de altura/largura e os dados estiverem corretos.
		return nil, fmt.Errorf("erro interno: tamanho do padrão achatado (%d) não corresponde ao PatternSize esperado (%d)", len(flattenedPattern), simParams.PatternSize)
	}

	return flattenedPattern, nil
}

// GetAllDigitPatterns retorna um mapa de todos os padrões de dígitos, onde a chave é o dígito (0-9)
// e o valor é o padrão achatado.
func GetAllDigitPatterns(simParams *config.SimulationParameters) (map[int][]float64, error) {
	allPatterns := make(map[int][]float64)
	for i := 0; i <= 9; i++ {
		pattern, err := GetDigitPattern(i, simParams)
		if err != nil {
			// Isso indica um problema com a definição dos padrões internos, o que não deveria acontecer.
			return nil, fmt.Errorf("erro ao obter padrão para o dígito %d: %w", i, err)
		}
		allPatterns[i] = pattern
	}
	return allPatterns, nil
}
```
