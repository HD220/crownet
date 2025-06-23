package storage_test

import (
	// "crownet/common" // Unused
	"crownet/config" // Used by simParams
	"crownet/network"
	"crownet/neuron" // Used by net.Neurons[0].Type
	"crownet/storage"
	"database/sql"
	"fmt"
	"math" // For floatEquals
	"os"
	"path/filepath"
	"testing"
	// "time" // Unused

	_ "github.com/mattn/go-sqlite3" // Driver SQLite
)

// Helper para verificar se uma tabela existe e tem colunas específicas (simplificado)
func tableExistsAndHasColumns(db *sql.DB, tableName string, expectedCols []string) (bool, error) {
	rows, err := db.Query(fmt.Sprintf("PRAGMA table_info(%s);", tableName))
	if err != nil {
		return false, fmt.Errorf("falha ao consultar PRAGMA table_info para %s: %w", tableName, err)
	}
	defer rows.Close()

	foundCols := make(map[string]bool)
	for rows.Next() {
		var cid int
		var name string
		var typeStr string
		var notnull int
		var dflt_value sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &typeStr, &notnull, &dflt_value, &pk); err != nil {
			return false, fmt.Errorf("falha ao escanear linha de PRAGMA table_info para %s: %w", tableName, err)
		}
		foundCols[name] = true
	}
	if err = rows.Err(); err != nil {
		return false, fmt.Errorf("erro após iteração de PRAGMA table_info para %s: %w", tableName, err)
	}

	if len(foundCols) == 0 && len(expectedCols) > 0 {
		return false, nil
	}

	for _, col := range expectedCols {
		if !foundCols[col] {
			return false, fmt.Errorf("coluna esperada '%s' não encontrada na tabela '%s'", col, tableName)
		}
	}
	return true, nil
}

func TestNewSQLiteLogger_InMemory(t *testing.T) {
	logger, err := storage.NewSQLiteLogger(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteLogger(\":memory:\") failed: %v", err)
	}
	defer logger.Close()

	if logger.DBForTest() == nil {
		t.Fatalf("Logger DB não foi inicializado")
	}

	expectedSnapshotCols := []string{"SnapshotID", "CycleCount", "Timestamp", "CortisolLevel", "DopamineLevel"}
	exists, err := tableExistsAndHasColumns(logger.DBForTest(), "NetworkSnapshots", expectedSnapshotCols)
	if err != nil {
		t.Fatalf("Erro ao verificar tabela NetworkSnapshots: %v", err)
	}
	if !exists {
		t.Errorf("Tabela NetworkSnapshots não foi criada ou não tem colunas esperadas")
	}

	expectedNeuronCols := []string{"StateID", "SnapshotID", "NeuronID", "Type", "CurrentState", "Position0", "Velocity0"}
	exists, err = tableExistsAndHasColumns(logger.DBForTest(), "NeuronStates", expectedNeuronCols)
	if err != nil {
		t.Fatalf("Erro ao verificar tabela NeuronStates: %v", err)
	}
	if !exists {
		t.Errorf("Tabela NeuronStates não foi criada ou não tem colunas esperadas")
	}
}

func TestSQLiteLogger_LogNetworkState(t *testing.T) {
	logger, err := storage.NewSQLiteLogger(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteLogger failed: %v", err)
	}
	defer logger.Close()

	simParams := config.DefaultSimulationParameters()
	net := network.NewCrowNet(2, 0.01, &simParams, 123)
	net.CycleCount = 100
	net.ChemicalEnv.CortisolLevel = 0.5
	net.ChemicalEnv.DopamineLevel = 0.25
	net.ChemicalEnv.LearningRateModulationFactor = 1.1
	net.ChemicalEnv.SynaptogenesisModulationFactor = 0.9

	if len(net.Neurons) < 2 {
		t.Fatalf("Rede de teste não tem neurônios suficientes.")
	}
	net.Neurons[0].ID = 101
	net.Neurons[0].Type = neuron.Input
	net.Neurons[0].CurrentState = neuron.Firing
	net.Neurons[0].AccumulatedPotential = 0.75
	net.Neurons[0].Position[0] = 1.1
	net.Neurons[0].Velocity[0] = -0.1

	net.Neurons[1].ID = 102
	net.Neurons[1].Type = neuron.Output

	err = logger.LogNetworkState(net)
	if err != nil {
		t.Fatalf("LogNetworkState failed: %v", err)
	}

	var cycleCount int
	var cortisol float64
	err = logger.DBForTest().QueryRow("SELECT CycleCount, CortisolLevel FROM NetworkSnapshots WHERE SnapshotID = 1").Scan(&cycleCount, &cortisol)
	if err != nil {
		t.Fatalf("Falha ao consultar NetworkSnapshots: %v", err)
	}
	if cycleCount != 100 {
		t.Errorf("Snapshot CycleCount: esperado 100, got %d", cycleCount)
	}
	if !floatEquals(cortisol, 0.5, 1e-9) {
		t.Errorf("Snapshot CortisolLevel: esperado 0.5, got %f", cortisol)
	}

	var neuronID int
	var neuronType int
	var position0 float64
	err = logger.DBForTest().QueryRow(
		"SELECT NeuronID, Type, Position0 FROM NeuronStates WHERE SnapshotID = 1 AND NeuronID = ?", net.Neurons[0].ID).Scan(&neuronID, &neuronType, &position0)
	if err != nil {
		t.Fatalf("Falha ao consultar NeuronStates para neurônio %d: %v", net.Neurons[0].ID, err)
	}
	if neuronID != int(net.Neurons[0].ID) {
		t.Errorf("NeuronState NeuronID: esperado %d, got %d", net.Neurons[0].ID, neuronID)
	}
	if neuronType != int(neuron.Input) {
		t.Errorf("NeuronState Type: esperado %d, got %d", neuron.Input, neuronType)
	}
	if !floatEquals(position0, 1.1, 1e-9) {
		t.Errorf("NeuronState Position0: esperado 1.1, got %f", position0)
	}
}

func TestSQLiteLogger_Close(t *testing.T) {
	loggerMem, err := storage.NewSQLiteLogger(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteLogger(\":memory:\") failed: %v", err)
	}
	if err := loggerMem.Close(); err != nil {
		t.Errorf("Close() para DB em memória falhou: %v", err)
	}
	if err := loggerMem.Close(); err != nil {
		t.Errorf("Close() repetido para DB em memória falhou: %v", err)
	}

	tempDir := t.TempDir()
	dbFilePath := filepath.Join(tempDir, "test_close.db")

	loggerFile, err := storage.NewSQLiteLogger(dbFilePath)
	if err != nil {
		t.Fatalf("NewSQLiteLogger (arquivo) failed: %v", err)
	}
	if _, errStat := os.Stat(dbFilePath); os.IsNotExist(errStat) {
		t.Fatalf("Arquivo de DB %s não foi criado", dbFilePath)
	}

	if err := loggerFile.Close(); err != nil {
		t.Errorf("Close() para DB em arquivo falhou: %v", err)
	}
}

// Helper para comparar floats com tolerância (duplicado, mas ok para teste)
func floatEquals(a, b, tolerance float64) bool {
	if a == b {
		return true
	}
	return math.Abs(a-b) < tolerance
}
