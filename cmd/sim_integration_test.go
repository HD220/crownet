package cmd

import (
	"testing"
	"time"
	// "path/filepath" // Not needed if not creating temp files for this basic test
	// "os" // Not needed for this basic test

	"crownet/cli"
	"crownet/config"
	"crownet/common" // For common.Rate if setting BaseLearningRate explicitly
	"database/sql"
	"fmt" // For Sprintf in SQLite row count query
	"os"
	"path/filepath"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

// Helper function to create a minimal AppConfig for sim tests
func newTestSimAppConfig(cycles int, totalNeurons int, dbPath string, saveInterval int) *config.AppConfig {
	// For basic sim run, many SimParams can be default.
	// Key CLIConfig fields are Mode, Cycles, TotalNeurons.
	return &config.AppConfig{
		SimParams: config.DefaultSimulationParameters(),
		Cli: config.CLIConfig{
			Mode:             config.ModeSim,
			TotalNeurons:     totalNeurons,
			Seed:             time.Now().UnixNano(),
			Cycles:           cycles,
			DbPath:           dbPath,
			SaveInterval:     saveInterval,
			BaseLearningRate: common.Rate(0.01), // Default, but explicit
			MonitorOutputID:  -2, // Explicitly disable output monitoring for basic test
			// Other sim-specific flags like StimInputID, DebugChem
			// can be left as default (0 or false).
		},
	}
}

func TestSimCommand_BasicRun(t *testing.T) {
	// 1. Construct an AppConfig for a minimal sim run
	// For this basic test, we are not testing SQLite logging, so DbPath can be empty.
	// Output monitoring is explicitly disabled by MonitorOutputID: -2 in newTestSimAppConfig.
	// Using a small number of cycles and neurons for speed.
	appCfg := newTestSimAppConfig(10, 50, "", 0) // 10 cycles, 50 neurons, no DB path, saveInterval 0

	// Validate the constructed AppConfig
	if err := appCfg.Validate(); err != nil {
		t.Fatalf("Constructed AppConfig is invalid: %v", err)
	}

	// 2. Create an orchestrator
	orchestrator := cli.NewOrchestrator(appCfg)

	// 3. Run the orchestrator
	// We are primarily checking if the sim mode runs for the specified cycles
	// without panic/error.
	err := orchestrator.Run()

	// 4. Assert that no error is returned
	if err != nil {
		t.Fatalf("Orchestrator.Run() for sim mode failed: %v", err)
	}

	// Optional: Check for basic console output (e.g., "Cycle X/Y completed...")
	// This would require stdout capture, similar to TestObserveCommand_BasicRun.
	// For TSK-TEST-003.3.1, just ensuring it runs without error is the primary goal.
	// t.Log("Sim mode basic run completed successfully.")
}

func TestSimCommand_SQLiteLogging(t *testing.T) {
	tempDir := t.TempDir()
	dbFileName := "test_sim_log.db"
	tempDbPath := filepath.Join(tempDir, dbFileName)

	// Configure for a short run with logging enabled and frequent saves
	cycles := 5
	saveInterval := 1 // Save every cycle
	appCfg := newTestSimAppConfig(cycles, 50, tempDbPath, saveInterval)

	if err := appCfg.Validate(); err != nil {
		t.Fatalf("Constructed AppConfig for SQLite logging test is invalid: %v", err)
	}

	orchestrator := cli.NewOrchestrator(appCfg)
	err := orchestrator.Run()
	if err != nil {
		t.Fatalf("Orchestrator.Run() for sim mode with SQLite logging failed: %v", err)
	}

	// 1. Verify DB file creation
	fileInfo, errStat := os.Stat(tempDbPath)
	if os.IsNotExist(errStat) {
		t.Fatalf("Expected SQLite DB file '%s' to be created, but it was not found.", tempDbPath)
	}
	if errStat != nil {
		t.Fatalf("Error stating SQLite DB file '%s': %v", tempDbPath, errStat)
	}

	// 2. Verify DB file not empty
	if fileInfo.Size() == 0 {
		t.Errorf("Expected SQLite DB file '%s' to be non-empty, but it was empty.", tempDbPath)
	}

	// 3. Connect to DB and check for tables
	db, errDbOpen := sql.Open("sqlite3", tempDbPath)
	if errDbOpen != nil {
		t.Fatalf("Failed to open created SQLite DB '%s': %v", tempDbPath, errDbOpen)
	}
	defer db.Close()

	tablesToVerify := []string{"NetworkSnapshots", "NeuronStates"}
	for _, tableName := range tablesToVerify {
		var name string
		query := "SELECT name FROM sqlite_master WHERE type='table' AND name=?;"
		errQuery := db.QueryRow(query, tableName).Scan(&name)
		if errQuery == sql.ErrNoRows {
			t.Errorf("Expected table '%s' to exist in SQLite DB, but it was not found.", tableName)
			continue
		}
		if errQuery != nil {
			t.Errorf("Error querying for table '%s': %v", tableName, errQuery)
			continue
		}

		// 4. Optionally, check for some rows
		var rowCount int
		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s;", tableName)
		errCount := db.QueryRow(countQuery).Scan(&rowCount)
		if errCount != nil {
			t.Errorf("Error counting rows in table '%s': %v", tableName, errCount)
		} else if rowCount == 0 {
			// With SaveInterval=1 and Cycles=5, we expect snapshots.
			// NetworkSnapshots should have at least 1 (final) or more (if periodic saves also happened).
			// NeuronStates should have TotalNeurons * number of snapshots.
			t.Errorf("Expected table '%s' to have some rows, but found 0.", tableName)
		} else {
			t.Logf("Table '%s' found with %d rows.", tableName, rowCount)
		}
	}
}
