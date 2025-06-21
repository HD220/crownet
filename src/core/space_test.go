package core

import (
	"math"
	"testing"
)

const float64EqualityThreshold = 1e-9

func AreEqualFloat64(a, b float64) bool {
	return math.Abs(a-b) <= float64EqualityThreshold
}

func AreEqualVector(v1, v2 [SpaceDimensions]float64) bool {
	for i := 0; i < SpaceDimensions; i++ {
		if !AreEqualFloat64(v1[i], v2[i]) {
			return false
		}
	}
	return true
}

func TestDistanceEuclidean(t *testing.T) {
	p1 := [SpaceDimensions]float64{1, 2, 3}
	p2 := [SpaceDimensions]float64{4, 5, 6}

	expected := math.Sqrt(27)
	actual := DistanceEuclidean(p1, p2)

	if !AreEqualFloat64(expected, actual) {
		t.Errorf("DistanceEuclidean: expected %f, got %f", expected, actual)
	}

	p3 := [SpaceDimensions]float64{}
	p4 := [SpaceDimensions]float64{}
	expectedZero := 0.0
	actualZero := DistanceEuclidean(p3, p4)
	if !AreEqualFloat64(expectedZero, actualZero) {
		t.Errorf("DistanceEuclidean (zero): expected %f, got %f", expectedZero, actualZero)
	}
}

func TestAddVectors(t *testing.T) {
	v1 := [SpaceDimensions]float64{1, 2, 3}
	v2 := [SpaceDimensions]float64{4, 5, 6}
	expected := [SpaceDimensions]float64{5, 7, 9}
	actual := AddVectors(v1, v2)
	if !AreEqualVector(expected, actual) {
		t.Errorf("AddVectors: expected %v, got %v", expected, actual)
	}
}

func TestSubtractVectors(t *testing.T) {
	v1 := [SpaceDimensions]float64{5, 7, 9}
	v2 := [SpaceDimensions]float64{1, 2, 3}
	expected := [SpaceDimensions]float64{4, 5, 6}
	actual := SubtractVectors(v1, v2)
	if !AreEqualVector(expected, actual) {
		t.Errorf("SubtractVectors: expected %v, got %v", expected, actual)
	}
}

func TestScaleVector(t *testing.T) {
	v := [SpaceDimensions]float64{1, 2, 3}
	scalar := 2.5
	expected := [SpaceDimensions]float64{2.5, 5.0, 7.5}
	actual := ScaleVector(v, scalar)
	if !AreEqualVector(expected, actual) {
		t.Errorf("ScaleVector: expected %v, got %v", expected, actual)
	}
}

func TestNormalizeVector(t *testing.T) {
	v1 := [SpaceDimensions]float64{3, 4}
	expected1 := [SpaceDimensions]float64{3.0 / 5.0, 4.0 / 5.0}
	actual1 := NormalizeVector(v1)
	if !AreEqualVector(expected1, actual1) {
		t.Errorf("NormalizeVector (v1): expected %v, got %v", expected1, actual1)
	}

	vZero := [SpaceDimensions]float64{}
	expectedZero := [SpaceDimensions]float64{}
	actualZero := NormalizeVector(vZero)
	if !AreEqualVector(expectedZero, actualZero) {
		t.Errorf("NormalizeVector (zero): expected %v, got %v", expectedZero, actualZero)
	}
}

func TestClampPosition(t *testing.T) {
	spaceSize := 100.0

	pos1 := [SpaceDimensions]float64{10, 20, 30}
	clamped1 := ClampPosition(pos1, spaceSize)
	if !AreEqualVector(pos1, clamped1) {
		t.Errorf("ClampPosition (within bounds): expected %v, got %v", pos1, clamped1)
	}

	pos2 := [SpaceDimensions]float64{-10, 50, 120}
	expected2 := [SpaceDimensions]float64{0, 50, 100}
	clamped2 := ClampPosition(pos2, spaceSize)
	if !AreEqualVector(expected2, clamped2) {
		t.Errorf("ClampPosition (mixed out of bounds): expected %v, got %v", expected2, clamped2)
	}

	pos3 := [SpaceDimensions]float64{-5, -10}
	for i := 2; i < SpaceDimensions; i++ {
		pos3[i] = spaceSize + float64(i)
	}
	expected3 := [SpaceDimensions]float64{0,0}
	for i := 2; i < SpaceDimensions; i++ {
		expected3[i] = spaceSize
	}
	clamped3 := ClampPosition(pos3, spaceSize)
	for i := 0; i < SpaceDimensions; i++ {
		if !AreEqualFloat64(expected3[i], clamped3[i]) {
			t.Errorf("ClampPosition (all out): Mismatch at index %d. Expected %f, got %f. Full: exp %v, got %v", i, expected3[i], clamped3[i], expected3, clamped3)
			break
		}
	}
}
