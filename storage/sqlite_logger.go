package storage

import (
	"database/sql"
	"encoding/json" // Added for LogNetworkState
	"fmt"

	"crownet/network"

	// "os"      // Unused
	// "strings" // Unused
	"time"
	// "github.com/mattn/go-sqlite3" // Unused as sqlite3 is imported via blank import in log_exporter.go or by driver registration
	// "crownet/common" // Unused, neuron types are handled via int casting
)

// SQLiteLogger provides functionality to log network snapshots and neuron states
// to an SQLite database.
type SQLiteLogger struct {
	db *sql.DB // db holds the active database connection.
}

// NewSQLiteLogger creates or opens an SQLite database file specified by dataSourceName
// and prepares it for logging network snapshots.
// It ensures the necessary tables ('NetworkSnapshots', 'NeuronStates') are created if they don't exist.
// Unlike previous versions, this function will NOT delete an existing database file.
// It will open an existing one or create a new one if it's not found.
func NewSQLiteLogger(dataSourceName string) (*SQLiteLogger, error) {
	// The database file will be created by sql.Open if it doesn't exist.
	// os.Remove has been removed to allow persistence of logs across runs.
	dbConn, err := sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite database at %s: %w", dataSourceName, err)
	}

	if err = dbConn.Ping(); err != nil {
		dbConn.Close()
		return nil, fmt.Errorf("failed to ping SQLite database at %s: %w", dataSourceName, err)
	}

	logger := &SQLiteLogger{db: dbConn}
	if err = logger.createTables(); err != nil {
		dbConn.Close()
		return nil, fmt.Errorf("failed to create tables in SQLite: %w", err)
	}

	return logger, nil
}

// createTables ensures that the necessary tables (NetworkSnapshots, NeuronStates) exist in the database.
// If they don't exist, they are created.
// Position and Velocity are now stored as TEXT columns containing JSON arrays.
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
		return fmt.Errorf("failed to create NetworkSnapshots table: %w", err)
	}

	// Simplified schema for NeuronStates: Position and Velocity are stored as JSON strings.
	neuronStatesTableSQL := `
    CREATE TABLE IF NOT EXISTS NeuronStates (
        StateID INTEGER PRIMARY KEY AUTOINCREMENT,
        SnapshotID INTEGER NOT NULL,
        NeuronID INTEGER NOT NULL,
        Position TEXT,
        Velocity TEXT,
        Type INTEGER,
        CurrentState INTEGER,
        AccumulatedPotential REAL,
        BaseFiringThreshold REAL,
        CurrentFiringThreshold REAL,
        LastFiredCycle INTEGER,
        CyclesInCurrentState INTEGER,
        FOREIGN KEY (SnapshotID) REFERENCES NetworkSnapshots (SnapshotID) ON DELETE CASCADE
    );`

	if _, err := sl.db.Exec(neuronStatesTableSQL); err != nil {
		return fmt.Errorf("failed to create NeuronStates table: %w", err)
	}
	return nil
}

// DBForTest returns the underlying *sql.DB object.
// This method is intended ONLY for use in test suites, for purposes such as:
//   - Inspecting the database state after operations.
//   - Performing setup or cleanup tasks on the test database.
//
// Caution: Directly manipulating the database via this accessor during normal operation
// can interfere with the logger's consistency and is strongly discouraged.
func (sl *SQLiteLogger) DBForTest() *sql.DB {
	return sl.db
}

// LogNetworkState logs a complete snapshot of the current state of the provided 'net' (CrowNet instance)
// into the SQLite database.
// This involves:
//  1. Inserting a summary record into the 'NetworkSnapshots' table (cycle count, timestamp, chemical levels, modulation factors).
//  2. For each neuron in the network, inserting its detailed state into the 'NeuronStates' table,
//     linking it to the snapshot ID. Position and Velocity are stored as JSON strings.
//
// All database operations are performed within a single transaction. If any part fails,
// the transaction is rolled back.
//
// Parameters:
//   - net: A pointer to the CrowNet instance whose state is to be logged.
//
// Returns:
//   - error: An error if any database operation fails or if the logger is not initialized, nil otherwise.
func (sl *SQLiteLogger) LogNetworkState(net *network.CrowNet) error {
	if sl.db == nil {
		return fmt.Errorf("SQLiteLogger not initialized (db is nil)")
	}
	if net == nil {
		return fmt.Errorf("cannot log network state: CrowNet instance is nil")
	}
	if net.ChemicalEnv == nil { // ChemicalEnv is accessed for logging
		return fmt.Errorf("cannot log network state: ChemicalEnv in CrowNet is nil")
	}
	if net.SimParams == nil { // SimParams is accessed for CortisolGlandPosition
		return fmt.Errorf("cannot log network state: SimParams in CrowNet is nil")
	}

	tx, err := sl.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin SQLite transaction: %w", err)
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
		return fmt.Errorf("failed to insert into NetworkSnapshots: %w", err)
	}
	snapshotID, err := snapshotRes.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get LastInsertId for snapshot: %w", err)
	}

	// SQL query for inserting neuron states. Position and Velocity are now single TEXT columns.
	neuronStateSQL := `INSERT INTO NeuronStates (
		SnapshotID, NeuronID, Position, Velocity,
		Type, CurrentState, AccumulatedPotential, BaseFiringThreshold,
		CurrentFiringThreshold, LastFiredCycle, CyclesInCurrentState
	) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	stmt, err := tx.Prepare(neuronStateSQL)
	if err != nil {
		return fmt.Errorf("failed to prepare statement for NeuronStates: %w", err)
	}
	defer stmt.Close()

	for _, n := range net.Neurons {
		// Serialize Position and Velocity to JSON strings.
		posJSON, err := json.Marshal(n.Position)
		if err != nil {
			return fmt.Errorf("failed to serialize Position to JSON for neuron %d: %w", n.ID, err)
		}
		velJSON, err := json.Marshal(n.Velocity)
		if err != nil {
			return fmt.Errorf("failed to serialize Velocity to JSON for neuron %d: %w", n.ID, err)
		}

		_, err = stmt.Exec(
			snapshotID,
			n.ID,
			string(posJSON), // Store as JSON string
			string(velJSON), // Store as JSON string
			int(n.Type),
			int(n.CurrentState),
			float64(n.AccumulatedPotential),
			float64(n.BaseFiringThreshold),
			float64(n.CurrentFiringThreshold),
			int(n.LastFiredCycle),
			int(n.CyclesInCurrentState),
		)
		if err != nil {
			return fmt.Errorf("failed to insert state for neuron %d: %w", n.ID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit SQLite transaction: %w", err)
	}
	return nil
}

// Close closes the underlying SQLite database connection.
// It's important to call this when the logger is no longer needed to free resources.
// Returns an error if closing the database fails. Sets sl.db to nil on successful close.
func (sl *SQLiteLogger) Close() error {
	if sl.db != nil {
		err := sl.db.Close()
		sl.db = nil // Set to nil even if close fails, to prevent further use of potentially bad connection
		if err != nil {
			return fmt.Errorf("failed to close SQLite database: %w", err)
		}
		return nil
	}
	return nil // No error if db is already nil
}
