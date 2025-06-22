package storage_test

import (
	"crownet/common"
	"crownet/config"
	"crownet/network"
	"crownet/neuron"
	"crownet/storage"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

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
		var typeStr string // Renomeado de type para typeStr
		var notnull int
		var dflt_value sql.NullString // Alterado para sql.NullString
		var pk int
		if err := rows.Scan(&cid, &name, &typeStr, &notnull, &dflt_value, &pk); err != nil {
			return false, fmt.Errorf("falha ao escanear linha de PRAGMA table_info para %s: %w", tableName, err)
		}
		foundCols[name] = true
	}
	if err = rows.Err(); err != nil {
		return false, fmt.Errorf("erro após iteração de PRAGMA table_info para %s: %w", tableName, err)
	}

	if len(foundCols) == 0 && len(expectedCols) > 0 { // Se não encontrou colunas mas esperava, tabela provavelmente não existe
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

	if logger.DBForTest() == nil { // Método DBForTest() precisaria ser adicionado ao SQLiteLogger
		t.Fatalf("Logger DB não foi inicializado")
	}

	// Verificar se as tabelas foram criadas
	expectedSnapshotCols := []string{"SnapshotID", "CycleCount", "Timestamp", "CortisolLevel", "DopamineLevel"}
	exists, err := tableExistsAndHasColumns(logger.DBForTest(), "NetworkSnapshots", expectedSnapshotCols)
	if err != nil {
		t.Fatalf("Erro ao verificar tabela NetworkSnapshots: %v", err)
	}
	if !exists {
		t.Errorf("Tabela NetworkSnapshots não foi criada ou não tem colunas esperadas")
	}

	// Para NeuronStates, verificar algumas colunas chave além das dinâmicas de posição/velocidade
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
	// Criar uma rede de teste simples
	// A semente é importante para posições/velocidades se elas forem aleatórias na inicialização
	net := network.NewCrowNet(2, 0.01, &simParams, 123)
	net.CycleCount = 100
	net.ChemicalEnv.CortisolLevel = 0.5
	net.ChemicalEnv.DopamineLevel = 0.25
	net.ChemicalEnv.LearningRateModulationFactor = 1.1
	net.ChemicalEnv.SynaptogenesisModulationFactor = 0.9

	if len(net.Neurons) < 2 {
		t.Fatalf("Rede de teste não tem neurônios suficientes.")
	}
	// Modificar alguns neurônios para ter dados distintos
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

	// Verificar dados em NetworkSnapshots
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

	// Verificar dados em NeuronStates para o primeiro neurônio
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
	// Teste com DB em memória
	loggerMem, err := storage.NewSQLiteLogger(":memory:")
	if err != nil {
		t.Fatalf("NewSQLiteLogger(\":memory:\") failed: %v", err)
	}
	if err := loggerMem.Close(); err != nil {
		t.Errorf("Close() para DB em memória falhou: %v", err)
	}
	// Tentar fechar de novo (deve ser seguro)
	if err := loggerMem.Close(); err != nil {
		t.Errorf("Close() repetido para DB em memória falhou: %v", err)
	}

	// Teste com DB em arquivo
	tempDir := t.TempDir()
	dbFilePath := filepath.Join(tempDir, "test_close.db")

	loggerFile, err := storage.NewSQLiteLogger(dbFilePath)
	if err != nil {
		t.Fatalf("NewSQLiteLogger (arquivo) failed: %v", err)
	}
	// Verificar se o arquivo foi criado
	if _, errStat := os.Stat(dbFilePath); os.IsNotExist(errStat) {
		t.Fatalf("Arquivo de DB %s não foi criado", dbFilePath)
	}

	if err := loggerFile.Close(); err != nil {
		t.Errorf("Close() para DB em arquivo falhou: %v", err)
	}
	// O arquivo ainda existe após Close, mas a conexão deve estar fechada.
	// Tentar reabrir pode ser um teste, ou simplesmente verificar se não há pânico.
}


// Nota: Para que os testes compilem, SQLiteLogger precisa de um método DBForTest() *sql.DB
// que retorne o *sql.DB interno. Isso é comum para permitir que os testes inspecionem o DB.
// Exemplo em sqlite_logger.go:
// func (sl *SQLiteLogger) DBForTest() *sql.DB { return sl.db }
// Vou assumir que este método será adicionado.
```
