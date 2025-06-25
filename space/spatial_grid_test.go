package space

import (
	"crownet/common"
	"crownet/config"
	"crownet/neuron"
	"math/rand"
	"reflect"
	"sort"
	"testing"
)

// Helper to create a neuron for testing.
func newTestNeuron(id common.NeuronID, pos common.Point) *neuron.Neuron {
	// Using default sim params for neuron creation, not relevant for grid tests usually.
	dummySimParams := config.DefaultSimulationParameters()
	return neuron.New(id, neuron.Excitatory, pos, &dummySimParams)
}

func TestNewSpatialGrid(t *testing.T) {
	minBound := common.Point{}
	t.Run("Valid params", func(t *testing.T) {
		sg, err := NewSpatialGrid(10.0, common.PointDimension, minBound)
		if err != nil {
			t.Fatalf("NewSpatialGrid() error = %v, wantErr false", err)
		}
		if sg == nil {
			t.Fatal("NewSpatialGrid() returned nil sg")
		}
		if sg.cellSize != 10.0 {
			t.Errorf("sg.cellSize = %f, want 10.0", sg.cellSize)
		}
		if sg.numDims != common.PointDimension {
			t.Errorf("sg.numDims = %d, want %d", sg.numDims, common.PointDimension)
		}
		if sg.gridOriginOffset != minBound {
			t.Errorf("sg.gridOriginOffset = %v, want %v", sg.gridOriginOffset, minBound)
		}
	})

	_, err := NewSpatialGrid(0.0, common.PointDimension, minBound)
	if err == nil {
		t.Error("NewSpatialGrid() with cellSize=0 expected error, got nil")
	}
	_, err = NewSpatialGrid(10.0, 0, minBound)
	if err == nil {
		t.Error("NewSpatialGrid() with numDims=0 expected error, got nil")
	}
	_, err = NewSpatialGrid(10.0, common.PointDimension-1, minBound)
	if err == nil {
		t.Error("NewSpatialGrid() with numDims != common.PointDimension expected error, got nil")
	}
}

func TestGetCellID(t *testing.T) {
	minBound := common.Point{} // Grid origin at (0,0,...)
	for i := range minBound { minBound[i] = -100.0 } // Grid covers space from -100 in each dim

	sg, _ := NewSpatialGrid(10.0, common.PointDimension, minBound)

	tests := []struct {
		name  string
		point common.Point
		want  CellID
	}{
		{"origin of space -> cell 0,0,.. relative to gridOriginOffset", common.Point{-100, -100}, CellID{0, 0}}, // If gridOriginOffset is (-100,-100), point (-100,-100) is cell (0,0)
		{"point in first cell", common.Point{-95, -95}, CellID{0, 0}}, // (-95 - (-100)) / 10 = 0.5 -> floor(0.5) = 0
		{"point crossing to next cell", common.Point{-90, -90}, CellID{1, 1}}, // (-90 - (-100)) / 10 = 1.0 -> floor(1.0) = 1
		{"positive coords", common.Point{5, 5}, CellID{10, 10}}, // (5 - (-100)) / 10 = 10.5 -> floor(10.5) = 10
		{"exact boundary", common.Point{-80, -70}, CellID{2, 3}}, // (-80 - (-100))/10=2, (-70 - (-100))/10=3
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Pad want CellID to 16D if test case is simpler
			var fullWant CellID
			for i := 0; i < common.PointDimension; i++ {
				if i < len(tt.want) { // Check if tt.want has this dimension defined
					fullWant[i] = tt.want[i]
				} else {
					// If point coord is 0 and grid origin is -100, cell index is 10.
					// For unspecified higher dims, assume point coord is 0.
					// (0 - (-100)) / 10 = 10. So cell index is 10.
					pointCoordForDim := 0.0
					if i < len(tt.point) { // if point has a value for this dim
						pointCoordForDim = float64(tt.point[i])
					}
					fullWant[i] = int(math.Floor((pointCoordForDim - float64(sg.gridOriginOffset[i])) / sg.cellSize))
				}
			}

			if got := sg.GetCellID(tt.point); got != fullWant {
				t.Errorf("GetCellID(%v) = %v, want %v", tt.point, got, fullWant)
			}
		})
	}
}

func TestSpatialGrid_BuildAndQuery(t *testing.T) {
	minBound := common.Point{}
	sg, _ := NewSpatialGrid(10.0, common.PointDimension, minBound) // Cell size 10, origin 0,0...

	neurons := []*neuron.Neuron{
		newTestNeuron(0, common.Point{5, 5}),    // Cell (0,0)
		newTestNeuron(1, common.Point{15, 5}),   // Cell (1,0)
		newTestNeuron(2, common.Point{5, 15}),   // Cell (0,1)
		newTestNeuron(3, common.Point{25, 25}),  // Cell (2,2)
		newTestNeuron(4, common.Point{-5, -5}),  // Cell (-1,-1)
		newTestNeuron(5, common.Point{50, 50}), // Cell (5,5) - further away
	}
	sg.Build(neurons)

	t.Run("Query sphere hitting one cell", func(t *testing.T) {
		// Query sphere centered at (6,6) with radius 3, should only hit cell (0,0) and find neuron 0
		candidates := sg.QuerySphereForCandidates(common.Point{6, 6}, 3.0)
		ids := getNeuronIDs(candidates)
		sortNeuronIDs(ids)
		expectedIDs := []common.NeuronID{0}
		if !reflect.DeepEqual(ids, expectedIDs) {
			t.Errorf("Query (6,6) R=3: got %v, want %v", ids, expectedIDs)
		}
	})

	t.Run("Query sphere hitting multiple cells", func(t *testing.T) {
		// Query sphere centered at (10,10) with radius 8
		// Min/Max cell indices for query:
		// Dim 0: center 10, R 8. Extent [2, 18]. Cell size 10. Origin 0.
		//   minCellIdx = floor(2/10) = 0. maxCellIdx = floor(18/10) = 1. Cells: 0, 1
		// Dim 1: center 10, R 8. Extent [2, 18].
		//   minCellIdx = floor(2/10) = 0. maxCellIdx = floor(18/10) = 1. Cells: 0, 1
		// Cells to check: (0,0), (0,1), (1,0), (1,1)
		// Neurons: 0 (5,5 in 0,0), 1 (15,5 in 1,0), 2 (5,15 in 0,1)
		// Neuron 3 (25,25 in 2,2) should NOT be in candidates from cell search.
		candidates := sg.QuerySphereForCandidates(common.Point{10, 10}, 8.0)
		ids := getNeuronIDs(candidates)
		sortNeuronIDs(ids)
		// Neurons in cells (0,0), (0,1), (1,0), (1,1) are 0, 1, 2.
		// Actual distance check:
		// N0 (5,5) to (10,10): dist=sqrt(5^2+5^2)=sqrt(50)=7.07 < 8. YES
		// N1 (15,5) to (10,10): dist=sqrt((-5)^2+5^2)=sqrt(50)=7.07 < 8. YES
		// N2 (5,15) to (10,10): dist=sqrt(5^2+(-5)^2)=sqrt(50)=7.07 < 8. YES
		expectedIDs := []common.NeuronID{0, 1, 2}
		if !reflect.DeepEqual(ids, expectedIDs) {
			t.Errorf("Query (10,10) R=8: got %v, want %v", ids, expectedIDs)
		}
	})

	t.Run("Query sphere hitting no neurons in cells", func(t *testing.T) {
		candidates := sg.QuerySphereForCandidates(common.Point{100, 100}, 1.0)
		if len(candidates) != 0 {
			t.Errorf("Query (100,100) R=1: got %d candidates, want 0", len(candidates))
		}
	})

	t.Run("Query with negative radius", func(t *testing.T) {
		candidates := sg.QuerySphereForCandidates(common.Point{5,5}, -1.0)
		if len(candidates) != 0 {
			t.Errorf("Query with negative radius: got %d candidates, want 0", len(candidates))
		}
	})

	t.Run("Query hitting cell with multiple neurons", func(t *testing.T) {
		sgLocal, _ := NewSpatialGrid(20.0, common.PointDimension, minBound)
		neuronsLocal := []*neuron.Neuron{
			newTestNeuron(10, common.Point{5,5}),   // Cell (0,0)
			newTestNeuron(11, common.Point{6,6}),   // Cell (0,0)
			newTestNeuron(12, common.Point{25,5}), // Cell (1,0)
		}
		sgLocal.Build(neuronsLocal)
		candidates := sgLocal.QuerySphereForCandidates(common.Point{10,10}, 10.0)
		// Query sphere center 10,10, R 10. Extent [0,20] x [0,20]. Cell size 20.
		// Min/Max cells: (0,0) to (0,0). So only cell (0,0) is checked.
		// Neurons in cell (0,0): 10, 11.
		ids := getNeuronIDs(candidates)
		sortNeuronIDs(ids)
		expectedIDs := []common.NeuronID{10, 11}
		if !reflect.DeepEqual(ids, expectedIDs) {
			t.Errorf("Query hitting cell (0,0) with R=10: got %v, want %v", ids, expectedIDs)
		}
	})
}

func getNeuronIDs(neurons []*neuron.Neuron) []common.NeuronID {
	ids := make([]common.NeuronID, len(neurons))
	for i, n := range neurons {
		ids[i] = n.ID
	}
	return ids
}

func sortNeuronIDs(ids []common.NeuronID) {
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
}

// TestGetCellID_WithOffset tests GetCellID when gridOriginOffset is not zero.
func TestGetCellID_WithOffset(t *testing.T) {
    var offset common.Point
    for i := range offset { offset[i] = -50.0 } // Grid effectively starts at (-50, -50, ...)

    sg, _ := NewSpatialGrid(10.0, common.PointDimension, offset)

    tests := []struct {
        name  string
        point common.Point
        want  CellID // Only first few dims relevant for test case clarity
    }{
        {"point at grid origin", common.Point{-50, -50}, CellID{0,0}},
        {"point in first cell from grid origin", common.Point{-45, -45}, CellID{0,0}}, // (-45 - (-50))/10 = 0.5 -> 0
        {"point crossing to next cell from grid origin", common.Point{-40, -40}, CellID{1,1}}, // (-40 - (-50))/10 = 1.0 -> 1
		{"point at true origin", common.Point{0,0}, CellID{5,5}}, // (0 - (-50))/10 = 5
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            var fullWant CellID
            // Populate fullWant based on tt.want for defined dims, and calculate for others
            for i := 0; i < common.PointDimension; i++ {
                var pCoord float64
                if i < len(tt.point) { pCoord = float64(tt.point[i]) } // Use test point coord if defined for this dim

                if i < len(tt.want) { // Use test want coord if defined for this dim
                    fullWant[i] = tt.want[i]
                } else { // Calculate expected for higher dims assuming point coord is 0 for those
                    fullWant[i] = int(math.Floor((pCoord - float64(offset[i])) / 10.0))
                }
            }

            if got := sg.GetCellID(tt.point); !reflect.DeepEqual(got, fullWant) {
                t.Errorf("GetCellID() with offset: point %v, got %v, want %v", tt.point, got, fullWant)
            }
        })
    }
}
