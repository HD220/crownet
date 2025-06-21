package storage

import (
	"crownet/network" // To access CrowNet type for saving
	"crownet/neuron"  // To access Neuron type
	"database/sql"
	"fmt"
	"os"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

var db *sql.DB

// InitDB initializes the SQLite database connection and creates tables if they don't exist.
func InitDB(dataSourceName string) error {
	var err error
	// Ensure the directory for the database file exists
	// For now, assume dataSourceName is just a filename like "crownet.db"
	// In a real app, parse DSN or ensure path exists.

	// Remove existing database file to start fresh each time for this example
	// In a real application, you might want to append or manage existing data.
	os.Remove(dataSourceName)

	db, err = sql.Open("sqlite3", dataSourceName)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	// Create tables
	if err = createTables(); err != nil {
		return fmt.Errorf("failed to create tables: %w", err)
	}
	fmt.Println("Database initialized and tables created successfully.")
	return nil
}

// createTables defines and executes SQL for table creation.
func createTables() error {
	networkSnapshotsTableSQL := `
    CREATE TABLE IF NOT EXISTS NetworkSnapshots (
        SnapshotID INTEGER PRIMARY KEY AUTOINCREMENT,
        CycleCount INTEGER NOT NULL,
        Timestamp DATETIME DEFAULT CURRENT_TIMESTAMP,
        CortisolLevel REAL,
        DopamineLevel REAL
    );`

	_, err := db.Exec(networkSnapshotsTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create NetworkSnapshots table: %w", err)
	}

	// NeuronStates table with 16 columns for Position and 16 for Velocity
	posCols := make([]string, 16)
	velCols := make([]string, 16)
	for i := 0; i < 16; i++ {
		posCols[i] = fmt.Sprintf("Position%d REAL", i)
		velCols[i] = fmt.Sprintf("Velocity%d REAL", i)
	}
	positionColumnsSQL := strings.Join(posCols, ", ")
	velocityColumnsSQL := strings.Join(velCols, ", ")

	neuronStatesTableSQL := fmt.Sprintf(`
    CREATE TABLE IF NOT EXISTS NeuronStates (
        StateID INTEGER PRIMARY KEY AUTOINCREMENT,
        SnapshotID INTEGER NOT NULL,
        NeuronID INTEGER NOT NULL,
        %s,
        %s,
        Type INTEGER,
        State INTEGER,
        AccumulatedPulse REAL,
        BaseFiringThreshold REAL,
        CurrentFiringThreshold REAL,
        LastFiredCycle INTEGER,
        CyclesInCurrentState INTEGER,
        FOREIGN KEY (SnapshotID) REFERENCES NetworkSnapshots (SnapshotID)
    );`, positionColumnsSQL, velocityColumnsSQL)

	_, err = db.Exec(neuronStatesTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create NeuronStates table: %w", err)
	}
	return nil
}

// SaveNetworkState saves the current state of the network to the database.
func SaveNetworkState(net *network.CrowNet) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback if not committed

	// Insert into NetworkSnapshots
	snapshotRes, err := tx.Exec(`INSERT INTO NetworkSnapshots (CycleCount, Timestamp, CortisolLevel, DopamineLevel)
                                 VALUES (?, ?, ?, ?)`,
		net.CycleCount, time.Now(), net.CortisolLevel, net.DopamineLevel)
	if err != nil {
		return fmt.Errorf("failed to insert into NetworkSnapshots: %w", err)
	}
	snapshotID, err := snapshotRes.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID for snapshot: %w", err)
	}

	// Prepare NeuronStates insert statement
	// Dynamically build column names and placeholders for position and velocity
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
                                SnapshotID, NeuronID,
                                %s, %s,
                                Type, State, AccumulatedPulse, BaseFiringThreshold,
                                CurrentFiringThreshold, LastFiredCycle, CyclesInCurrentState
                             ) VALUES (?, ?, %s, %s, ?, ?, ?, ?, ?, ?, ?)`,
		strings.Join(posColNames, ", "), strings.Join(velColNames, ", "),
		strings.Join(posPlaceholders, ", "), strings.Join(velPlaceholders, ", "))

	stmt, err := tx.Prepare(sqlQuery)
	if err != nil {
		return fmt.Errorf("failed to prepare NeuronStates insert statement: %w", err)
	}
	defer stmt.Close()

	for _, n := range net.Neurons {
		args := make([]interface{}, 0, 2+32+7) // SnapshotID, NeuronID + 16 Pos + 16 Vel + 7 other fields
		args = append(args, snapshotID, n.ID)
		for i := 0; i < 16; i++ {
			args = append(args, n.Position[i])
		}
		for i := 0; i < 16; i++ {
			args = append(args, n.Velocity[i])
		}
		args = append(args, int(n.Type), int(n.State), n.AccumulatedPulse, n.BaseFiringThreshold,
			n.CurrentFiringThreshold, n.LastFiredCycle, n.CyclesInCurrentState)

		_, err = stmt.Exec(args...)
		if err != nil {
			return fmt.Errorf("failed to insert neuron state for neuron %d: %w", n.ID, err)
		}
	}

	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}
	fmt.Printf("Network state for cycle %d saved to database (SnapshotID: %d).\n", net.CycleCount, snapshotID)
	return nil
}

// CloseDB closes the database connection.
func CloseDB() {
	if db != nil {
		db.Close()
		fmt.Println("Database connection closed.")
	}
}

// GetNeuronByID retrieves a neuron's state from the database for a specific snapshot.
// This is an example and might not be directly used by the simulation but for analysis.
func GetNeuronByID(snapshotID int64, neuronID int) (*neuron.Neuron, error) {
	// This function would query NeuronStates table. Implementation deferred.
	return nil, fmt.Errorf("GetNeuronByID not fully implemented")
}
