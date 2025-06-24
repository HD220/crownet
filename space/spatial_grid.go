package space

import (
	"crownet/common"
	"crownet/neuron"
	"fmt"
	"math"
)

// CellID represents the integer coordinates of a cell in the N-dimensional grid.
type CellID [common.PointDimension]int // common.PointDimension is 16, so this is [16]int

// SpatialGrid provides a uniform grid spatial index for neurons.
type SpatialGrid struct {
	cells            map[CellID][]*neuron.Neuron
	cellSize         float64
	gridOriginOffset common.Point // The world coordinate that maps to cell index (0,0,...,0).
	numDims          int
}

// NewSpatialGrid creates a new spatial grid.
// cellSize: The size of one side of a hypercubic grid cell.
// numDims: The dimensionality of the space.
// spaceMinBound: The minimum coordinate corner of the simulation space (e.g., [-maxDim, -maxDim,...]).
//                This point will correspond to the origin of the grid's cell indexing system.
func NewSpatialGrid(cellSize float64, numDims int, spaceMinBound common.Point) (*SpatialGrid, error) {
	if cellSize <= 1e-9 { // Epsilon for zero check
		return nil, fmt.Errorf("NewSpatialGrid: cellSize must be positive, got %f", cellSize)
	}
	if numDims <= 0 {
		return nil, fmt.Errorf("NewSpatialGrid: numDims must be positive, got %d", numDims)
	}
	if numDims != common.PointDimension {
		return nil, fmt.Errorf("NewSpatialGrid: numDims (%d) must match common.PointDimension (%d)", numDims, common.PointDimension)
	}

	sg := &SpatialGrid{
		cells:            make(map[CellID][]*neuron.Neuron),
		cellSize:         cellSize,
		gridOriginOffset: spaceMinBound,
		numDims:          numDims,
	}
	return sg, nil
}

// GetCellID calculates the cell ID for a given point in world coordinates.
func (sg *SpatialGrid) GetCellID(point common.Point) CellID {
	var id CellID
	for i := 0; i < sg.numDims; i++ {
		// Subtracting gridOriginOffset effectively translates the point relative to the grid's origin.
		id[i] = int(math.Floor((float64(point[i]) - float64(sg.gridOriginOffset[i])) / sg.cellSize))
	}
	return id
}

// AddNeuron adds a neuron to the grid. Not thread-safe.
func (sg *SpatialGrid) AddNeuron(n *neuron.Neuron) {
	if n == nil {
		return
	}
	cellID := sg.GetCellID(n.Position)
	sg.cells[cellID] = append(sg.cells[cellID], n)
}

// Build clears and rebuilds the grid with the given neurons. Not thread-safe.
func (sg *SpatialGrid) Build(neurons []*neuron.Neuron) {
	sg.cells = make(map[CellID][]*neuron.Neuron) // Clear existing cells
	for _, n := range neurons {
		if n != nil {
			sg.AddNeuron(n)
		}
	}
}

// QuerySphereForCandidates collects neurons from cells that could potentially
// intersect with the query sphere defined by a center and radius.
// This method identifies a hyper-rectangular region of cells that bounds the query sphere
// and returns all neurons within those cells.
// The caller is responsible for performing precise distance checks on these candidates.
//
// center: The center of the query sphere in world coordinates.
// radius: The radius of the query sphere.
//
// Returns a slice of candidate neurons.
func (sg *SpatialGrid) QuerySphereForCandidates(center common.Point, radius float64) []*neuron.Neuron {
	candidateNeurons := make([]*neuron.Neuron, 0)
	if radius < 0 {
		return candidateNeurons
	}

	minCellIndices := [common.PointDimension]int{}
	maxCellIndices := [common.PointDimension]int{}

	for i := 0; i < sg.numDims; i++ {
		sphereMinDimCoord := float64(center[i]) - radius
		sphereMaxDimCoord := float64(center[i]) + radius
		minCellIndices[i] = int(math.Floor((sphereMinDimCoord - float64(sg.gridOriginOffset[i])) / sg.cellSize))
		maxCellIndices[i] = int(math.Floor((sphereMaxDimCoord - float64(sg.gridOriginOffset[i])) / sg.cellSize))
	}

	var currentCellVisit [common.PointDimension]int
	sg.queryCellsRecursive(minCellIndices, maxCellIndices, &currentCellVisit, 0, &candidateNeurons)

	return candidateNeurons
}

// queryCellsRecursive is a helper to iterate N-dimensionally through a range of cells.
func (sg *SpatialGrid) queryCellsRecursive(
	minCellIndices, maxCellIndices [common.PointDimension]int,
	currentCellIndices *[common.PointDimension]int,
	dim int,
	candidateNeurons *[]*neuron.Neuron,
) {
	if dim == sg.numDims {
		var cellToQuery CellID
		copy(cellToQuery[:], currentCellIndices[:])

		if neuronsInCell, found := sg.cells[cellToQuery]; found {
			*candidateNeurons = append(*candidateNeurons, neuronsInCell...)
		}
		return
	}

	for i := minCellIndices[dim]; i <= maxCellIndices[dim]; i++ {
		(*currentCellIndices)[dim] = i
		sg.queryCellsRecursive(minCellIndices, maxCellIndices, currentCellIndices, dim+1, candidateNeurons)
	}
}

// Ensure common.PointDimension is defined, e.g. in common/types.go
// const PointDimension = 16
// type Point [PointDimension]float64
// type NeuronID int
// ... other common types
