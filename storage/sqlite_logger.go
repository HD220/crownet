package storage

import (
	"crownet/common"
	"crownet/config"
	"crownet/neuron" // Para neuron.Neuron e neuron.State/Type
	"crownet/network" // Para network.CrowNet (ou uma interface/DTO se quisermos desacoplar)
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // Driver SQLite
)

// SQLiteLogger é responsável por registrar o estado da rede em um banco de dados SQLite.
type SQLiteLogger struct {
	db *sql.DB
}

// NewSQLiteLogger inicializa uma nova conexão com o banco de dados SQLite.
// Recria o banco de dados se ele já existir, para cada sessão de logging.
func NewSQLiteLogger(dataSourceName string) (*SQLiteLogger, error) {
	// Remover banco de dados existente para começar do zero a cada execução.
	// Em uma aplicação real, pode-se querer anexar ou gerenciar dados existentes.
	_ = os.Remove(dataSourceName) // Ignorar erro se o arquivo não existir

	dbConn, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("falha ao abrir banco de dados SQLite em %s: %w", dataSourceName, err)
	}

	if err = dbConn.Ping(); err != nil {
		dbConn.Close()
		return nil, fmt.Errorf("falha ao pingar banco de dados SQLite em %s: %w", dataSourceName, err)
	}

	logger := &SQLiteLogger{db: dbConn}
	if err = logger.createTables(); err != nil {
		dbConn.Close()
		return nil, fmt.Errorf("falha ao criar tabelas no SQLite: %w", err)
	}

	return logger, nil
}

const pointDimension = 16 // Assumindo a mesma dimensionalidade de common.Point

// getDimensionSQLParts gera as strings SQL para nomes de colunas e placeholders
// para um dado prefixo e dimensão.
// Ex: prefix="Position", dimension=16 -> "Position0 REAL, Position1 REAL...", "?,?,..."
func getDimensionSQLParts(prefix string, dimension int, forTableCreation bool) (colNamesSQL string, placeholdersSQL string) {
	colNames := make([]string, dimension)
	placeholders := make([]string, dimension)
	for i := 0; i < dimension; i++ {
		if forTableCreation {
			colNames[i] = fmt.Sprintf("%s%d REAL", prefix, i)
		} else {
			colNames[i] = fmt.Sprintf("%s%d", prefix, i)
		}
		placeholders[i] = "?"
	}
	return strings.Join(colNames, ", "), strings.Join(placeholders, ", ")
}

// createTables define e executa o SQL para a criação das tabelas.
func (sl *SQLiteLogger) createTables() error {
	networkSnapshotsTableSQL := `
    CREATE TABLE IF NOT EXISTS NetworkSnapshots (
        SnapshotID INTEGER PRIMARY KEY AUTOINCREMENT,
        CycleCount INTEGER NOT NULL,
        Timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
        CortisolLevel REAL,
        DopamineLevel REAL,
        LearningRateModFactor REAL,
        SynaptogenesisModFactor REAL
    );`
	if _, err := sl.db.Exec(networkSnapshotsTableSQL); err != nil {
		return fmt.Errorf("falha ao criar tabela NetworkSnapshots: %w", err)
	}

	positionSchemaSQL, _ := getDimensionSQLParts("Position", pointDimension, true)
	velocitySchemaSQL, _ := getDimensionSQLParts("Velocity", pointDimension, true)

	neuronStatesTableSQL := fmt.Sprintf(`
    CREATE TABLE IF NOT EXISTS NeuronStates (
        StateID INTEGER PRIMARY KEY AUTOINCREMENT,
        SnapshotID INTEGER NOT NULL,
        NeuronID INTEGER NOT NULL,
        %s,
        %s,
        Type INTEGER,
        CurrentState INTEGER,
        AccumulatedPotential REAL,
        BaseFiringThreshold REAL,
        CurrentFiringThreshold REAL,
        LastFiredCycle INTEGER,
        CyclesInCurrentState INTEGER,
        FOREIGN KEY (SnapshotID) REFERENCES NetworkSnapshots (SnapshotID) ON DELETE CASCADE
    );`, positionSchemaSQL, velocitySchemaSQL)

	if _, err := sl.db.Exec(neuronStatesTableSQL); err != nil {
		return fmt.Errorf("falha ao criar tabela NeuronStates: %w", err)
	}
	return nil
}

// DBForTest retorna a instância do banco de dados para uso em testes.
// Este método só deve ser usado em contextos de teste.
func (sl *SQLiteLogger) DBForTest() *sql.DB {
	return sl.db
}

// LogNetworkState salva o estado atual da rede no banco de dados.
// Aceita um CrowNet diretamente, mas idealmente poderia aceitar uma interface ou DTO
// para melhor desacoplamento, se o pacote `network` não devesse depender de `storage`.
// No entanto, para este projeto, a dependência circular gerenciada por interfaces ou
// o `storage` conhecendo `network` é aceitável.
func (sl *SQLiteLogger) LogNetworkState(net *network.CrowNet) error {
	if sl.db == nil {
		return fmt.Errorf("logger SQLite não inicializado")
	}

	tx, err := sl.db.Begin()
	if err != nil {
		return fmt.Errorf("falha ao iniciar transação SQLite: %w", err)
	}
	// Deferir Rollback é importante para garantir que a transação seja desfeita em caso de erro.
	// Se Commit for bem-sucedido, o Rollback não terá efeito.
	defer tx.Rollback()

	// Inserir na tabela NetworkSnapshots
	snapshotRes, err := tx.Exec(`INSERT INTO NetworkSnapshots
                                     (CycleCount, Timestamp, CortisolLevel, DopamineLevel, LearningRateModFactor, SynaptogenesisModFactor)
                                 VALUES (?, ?, ?, ?, ?, ?)`,
		net.CycleCount,
		time.Now(),
		net.ChemicalEnv.CortisolLevel,
		net.ChemicalEnv.DopamineLevel,
		net.ChemicalEnv.LearningRateModulationFactor,
		net.ChemicalEnv.SynaptogenesisModulationFactor,
	)
	if err != nil {
		return fmt.Errorf("falha ao inserir em NetworkSnapshots: %w", err)
	}
	snapshotID, err := snapshotRes.LastInsertId()
	if err != nil {
		return fmt.Errorf("falha ao obter LastInsertId para snapshot: %w", err)
	}

	// Preparar statement para NeuronStates
	posColNames := make([]string, 16)
	velColNames := make([]string, 16)
	posPlaceholders := make([]string, 16)
	velPlaceholders := make([]string, 16)
	for i := 0; i < 16; i++ {
		posColNames[i] = fmt.Sprintf("Position%d", i)
		velColNames[i] = fmt.Sprintf("Velocity%d", i)
		posPlaceholders[i] = "?"
		velPlaceholders[i] = "?"
	}
	sqlQuery := fmt.Sprintf(`INSERT INTO NeuronStates (
                                SnapshotID, NeuronID, %s, %s,
                                Type, CurrentState, AccumulatedPotential, BaseFiringThreshold,
                                CurrentFiringThreshold, LastFiredCycle, CyclesInCurrentState
                             ) VALUES (?, ?, %s, %s, ?, ?, ?, ?, ?, ?, ?)`,
		strings.Join(posColNames, ", "), strings.Join(velColNames, ", "),
		strings.Join(posPlaceholders, ", "), strings.Join(velPlaceholders, ", "),
	)

	stmt, err := tx.Prepare(sqlQuery)
	if err != nil {
		return fmt.Errorf("falha ao preparar statement para NeuronStates: %w", err)
	}
	defer stmt.Close()

	for _, n := range net.Neurons {
		// Args: SnapshotID, NeuronID, P0..P15, V0..V15, Type, State, AccPot, BaseThr, CurrThr, LastFired, CyclesInState
		args := make([]interface{}, 0, 2+pointDimension*2+7)
		args = append(args, snapshotID, n.ID)
		for i := 0; i < pointDimension; i++ {
			args = append(args, float64(n.Position[i]))
		}
		for i := 0; i < pointDimension; i++ {
			args = append(args, float64(n.Velocity[i]))
		}
		args = append(args, int(n.Type), int(n.CurrentState), float64(n.AccumulatedPotential),
			float64(n.BaseFiringThreshold), float64(n.CurrentFiringThreshold),
			int(n.LastFiredCycle), int(n.CyclesInCurrentState))

		if _, err = stmt.Exec(args...); err != nil {
			return fmt.Errorf("falha ao inserir estado para neurônio %d: %w", n.ID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("falha ao commitar transação SQLite: %w", err)
	}
	return nil
}

// Close encerra a conexão com o banco de dados.
func (sl *SQLiteLogger) Close() error {
	if sl.db != nil {
		// fmt.Println("Fechando conexão com o banco de dados SQLite.")
		return sl.db.Close()
	}
	return nil
}
```
