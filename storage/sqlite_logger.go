package storage

import (
	"crownet/network"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteLogger struct {
	db *sql.DB
}

func NewSQLiteLogger(dataSourceName string) (*SQLiteLogger, error) {
	_ = os.Remove(dataSourceName)

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

const pointDimension = 16

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

func (sl *SQLiteLogger) DBForTest() *sql.DB {
	return sl.db
}

func (sl *SQLiteLogger) LogNetworkState(net *network.CrowNet) error {
	if sl.db == nil {
		return fmt.Errorf("logger SQLite não inicializado")
	}

	tx, err := sl.db.Begin()
	if err != nil {
		return fmt.Errorf("falha ao iniciar transação SQLite: %w", err)
	}
	defer tx.Rollback()

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

func (sl *SQLiteLogger) Close() error {
	if sl.db != nil {
		return sl.db.Close()
	}
	return nil
}
