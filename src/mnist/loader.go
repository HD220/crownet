package mnist

import (
	"compress/gzip"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// MNIST dataset file names
const (
	trainImagesFile = "train-images-idx3-ubyte.gz"
	trainLabelsFile = "train-labels-idx1-ubyte.gz"
	testImagesFile  = "t10k-images-idx3-ubyte.gz"
	testLabelsFile  = "t10k-labels-idx1-ubyte.gz"
	baseURL         = "http://yann.lecun.com/exdb/mnist/"
)

// Image represents a single MNIST image.
// Pixels are flattened (28x28 = 784 values).
type Image struct {
	Pixels []byte // Grayscale pixel values (0-255)
	Label  byte   // The digit (0-9)
}

// Dataset represents the MNIST dataset.
type Dataset struct {
	TrainImages []Image
	TestImages  []Image
	Rows        int
	Cols        int
}

// downloadFile downloads a file from a URL to a local path.
func downloadFile(url, path string) error {
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("File %s already exists. Skipping download.\n", filepath.Base(path))
		return nil // File already exists
	}
	fmt.Printf("Downloading %s...\n", filepath.Base(path))

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("failed to download %s: %w", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("bad status for %s: %s", url, resp.Status)
	}

	out, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", path, err)
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to save %s: %w", url, err)
	}
	return nil
}

// loadImages reads MNIST image data from a .gz file.
func loadImages(filePath string) (images [][]byte, rows, cols int, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to open image file %s: %w", filePath, err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("failed to create gzip reader for %s: %w", filePath, err)
	}
	defer gzReader.Close()

	var magic, numImages, numRows, numCols int32
	if err := binary.Read(gzReader, binary.BigEndian, &magic); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read magic number: %w", err)
	}
	if magic != 2051 { // Magic number for image files
		return nil, 0, 0, fmt.Errorf("invalid magic number for image file: %d", magic)
	}

	if err := binary.Read(gzReader, binary.BigEndian, &numImages); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read number of images: %w", err)
	}
	if err := binary.Read(gzReader, binary.BigEndian, &numRows); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read number of rows: %w", err)
	}
	if err := binary.Read(gzReader, binary.BigEndian, &numCols); err != nil {
		return nil, 0, 0, fmt.Errorf("failed to read number of columns: %w", err)
	}

	rows, cols = int(numRows), int(numCols)
	pixelDataSize := rows * cols
	images = make([][]byte, numImages)

	for i := 0; i < int(numImages); i++ {
		images[i] = make([]byte, pixelDataSize)
		n, err := io.ReadFull(gzReader, images[i])
		if err != nil {
			return nil, 0, 0, fmt.Errorf("failed to read image data for image %d: %w", i, err)
		}
		if n != pixelDataSize {
			return nil, 0, 0, fmt.Errorf("short read for image data for image %d: read %d, expected %d", i, n, pixelDataSize)
		}
	}
	return images, rows, cols, nil
}

// loadLabels reads MNIST label data from a .gz file.
func loadLabels(filePath string) (labels []byte, err error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open label file %s: %w", filePath, err)
	}
	defer file.Close()

	gzReader, err := gzip.NewReader(file)
	if err != nil {
		return nil, fmt.Errorf("failed to create gzip reader for %s: %w", filePath, err)
	}
	defer gzReader.Close()

	var magic, numLabels int32
	if err := binary.Read(gzReader, binary.BigEndian, &magic); err != nil {
		return nil, fmt.Errorf("failed to read magic number: %w", err)
	}
	if magic != 2049 { // Magic number for label files
		return nil, fmt.Errorf("invalid magic number for label file: %d", magic)
	}

	if err := binary.Read(gzReader, binary.BigEndian, &numLabels); err != nil {
		return nil, fmt.Errorf("failed to read number of labels: %w", err)
	}

	labels = make([]byte, numLabels)
	n, err := io.ReadFull(gzReader, labels)
	if err != nil {
		return nil, fmt.Errorf("failed to read label data: %w", err)
	}
	if n != int(numLabels) {
		return nil, fmt.Errorf("short read for label data: read %d, expected %d", n, numLabels)
	}
	return labels, nil
}

// Load downloads (if necessary) and loads the MNIST dataset from the specified directory.
func Load(dataDir string) (*Dataset, error) {
	if err := os.MkdirAll(dataDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create data directory %s: %w", dataDir, err)
	}

	filesToDownload := []string{trainImagesFile, trainLabelsFile, testImagesFile, testLabelsFile}
	for _, file := range filesToDownload {
		if err := downloadFile(baseURL+file, filepath.Join(dataDir, file)); err != nil {
			return nil, err
		}
	}

	trainPixelData, rows, cols, err := loadImages(filepath.Join(dataDir, trainImagesFile))
	if err != nil {
		return nil, fmt.Errorf("failed to load train images: %w", err)
	}
	trainLabelData, err := loadLabels(filepath.Join(dataDir, trainLabelsFile))
	if err != nil {
		return nil, fmt.Errorf("failed to load train labels: %w", err)
	}

	testPixelData, _, _, err := loadImages(filepath.Join(dataDir, testImagesFile)) // rows/cols should be same
	if err != nil {
		return nil, fmt.Errorf("failed to load test images: %w", err)
	}
	testLabelData, err := loadLabels(filepath.Join(dataDir, testLabelsFile))
	if err != nil {
		return nil, fmt.Errorf("failed to load test labels: %w", err)
	}

	if len(trainPixelData) != len(trainLabelData) {
		return nil, fmt.Errorf("mismatch between number of train images (%d) and labels (%d)", len(trainPixelData), len(trainLabelData))
	}
	if len(testPixelData) != len(testLabelData) {
		return nil, fmt.Errorf("mismatch between number of test images (%d) and labels (%d)", len(testPixelData), len(testLabelData))
	}

	dataset := &Dataset{
		TrainImages: make([]Image, len(trainPixelData)),
		TestImages:  make([]Image, len(testPixelData)),
		Rows:        rows,
		Cols:        cols,
	}

	for i := 0; i < len(trainPixelData); i++ {
		dataset.TrainImages[i] = Image{Pixels: trainPixelData[i], Label: trainLabelData[i]}
	}
	for i := 0; i < len(testPixelData); i++ {
		dataset.TestImages[i] = Image{Pixels: testPixelData[i], Label: testLabelData[i]}
	}

	fmt.Printf("MNIST dataset loaded: %d train images, %d test images.\n", len(dataset.TrainImages), len(dataset.TestImages))
	return dataset, nil
}

// NormalizePixels scales pixel values from [0, 255] to [0.0, 1.0].
func NormalizePixels(pixels []byte) []float64 {
	normalized := make([]float64, len(pixels))
	for i, p := range pixels {
		normalized[i] = float64(p) / 255.0
	}
	return normalized
}
