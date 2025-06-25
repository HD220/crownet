// Package datagen provides utilities for generating or retrieving predefined
// data patterns, such as digit representations, for use in network simulations
// and training.
package datagen

import (
	"crownet/config" // To access PatternSize, PatternHeight, PatternWidth from SimParams
	"fmt"
)

// digitPatternData stores the raw 2D patterns for digits 0-9.
// Using a map of int to [][]int provides a straightforward way to define them.
// '1' represents an active pixel, '0' an inactive one.
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

// GetDigitPatternFn is a function variable that points to the actual implementation
// for retrieving a digit pattern. This allows for mocking during testing.
// By default, it points to getDigitPatternInternal.
var GetDigitPatternFn = getDigitPatternInternal

// getDigitPatternInternal is the core implementation for retrieving a digit's pattern.
// It fetches the 2D pattern from digitPatternData, validates its dimensions against
// the provided SimParams (PatternHeight, PatternWidth), and flattens it into a 1D slice
// of float64 values (1.0 for active, 0.0 for inactive).
//
// Parameters:
//   digit: The integer digit (0-9) for which the pattern is requested.
//   simParams: Simulation parameters used for validating pattern dimensions.
//
// Returns:
//   A 1D slice of float64 representing the flattened digit pattern, or an error
//   if the digit is not found or if the pattern dimensions mismatch SimParams.
func getDigitPatternInternal(digit int, simParams *config.SimulationParameters) ([]float64, error) {
	pattern2D, ok := digitPatternData[digit]
	if !ok {
		return nil, fmt.Errorf("digit pattern for %d not found", digit)
	}

	// Assumes simParams have been validated by config.Validate() regarding positive PatternHeight/Width
	// and consistency of PatternSize with Height*Width.
	// We still need to check that the actual pattern data matches these validated simParams.

	if len(pattern2D) != simParams.PatternHeight {
		return nil, fmt.Errorf("pattern for digit %d has incorrect height %d, expected %d (from simParams)", digit, len(pattern2D), simParams.PatternHeight)
	}

	flattenedPattern := make([]float64, 0, simParams.PatternSize)
	for r, row := range pattern2D {
		if len(row) != simParams.PatternWidth {
			return nil, fmt.Errorf("pattern for digit %d, row %d has incorrect width %d, expected %d (from simParams)", digit, r, len(row), simParams.PatternWidth)
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
		// This should ideally not happen if height and width checks pass and PatternSize is consistent.
		return nil, fmt.Errorf("internal error: flattened pattern size (%d) does not match expected PatternSize (%d)", len(flattenedPattern), simParams.PatternSize)
	}

	return flattenedPattern, nil
}

// GetAllDigitPatterns retrieves all predefined digit patterns (0-9).
// It returns a map where the key is the digit (int) and the value is its
// corresponding flattened 1D pattern (slice of float64).
// This function relies on GetDigitPatternFn (and thus simParams for validation)
// for each digit.
//
// Parameters:
//   simParams: Simulation parameters, passed to GetDigitPatternFn for dimension validation.
//
// Returns:
//   A map of digit to its pattern, or an error if any digit pattern cannot be retrieved.
func GetAllDigitPatterns(simParams *config.SimulationParameters) (map[int][]float64, error) {
	allPatterns := make(map[int][]float64)
	for i := 0; i <= 9; i++ {
		pattern, err := GetDigitPatternFn(i, simParams) // Uses the (potentially mocked) GetDigitPatternFn
		if err != nil {
			return nil, fmt.Errorf("error getting pattern for digit %d: %w", i, err)
		}
		allPatterns[i] = pattern
	}
	return allPatterns, nil
}
