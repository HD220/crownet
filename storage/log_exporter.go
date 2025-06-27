// Package storage provides functionalities for data persistence,
// including SQLite logging and log data exporting.
package storage

import (
	"database/sql"
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"

	// "strings" // Unused import
	// Import sqlite3 driver
	_ "github.com/mattn/go-sqlite3"

	"crownet/neuron" // For neuron.Type and neuron.State enums
)

// neuronTypeToString maps neuron.Type enum to its string representation
// for CSV export.
// var neuronTypeToString = map[neuron.Type]string{ // Replaced by using .String() method
// 	neuron.Excitatory:   "Excitatory",
// 	neuron.Inhibitory:   "Inhibitory",
// 	neuron.Dopaminergic: "Dopaminergic",
// 	neuron.Input:        "Input",
// 	neuron.Output:       "Output",
// 	// neuron.UnknownType:  "UnknownType", // This constant does not exist
// }

// neuronStateToString maps neuron.State enum to its string representation.
// var neuronStateToString = map[neuron.State]string{ // Replaced by using .String() method
// 	neuron.Resting:            "Resting",
// 	neuron.Firing:             "Firing",
// 	neuron.AbsoluteRefractory: "AbsoluteRefractory",
// 	neuron.RelativeRefractory: "RelativeRefractory",
// 	// neuron.UnknownState:       "UnknownState", // This constant does not exist
// }

// ExportLogData connects to an SQLite database specified by dbPath,
// reads data from the given tableName, and exports it in the specified format
// to outputPath. If outputPath is empty, data is written to os.Stdout.
// Currently, only "csv" format is supported, and valid tableNames are
// "NetworkSnapshots" and "NeuronStates".
func ExportLogData(dbPath, tableName, format, outputPath string) error {
	if format != "csv" {
		return fmt.Errorf("unsupported format '%s', only 'csv' is currently supported", format)
	}

	db, err := sql.Open("sqlite3", dbPath+"?mode=ro") // Open in read-only mode
	if err != nil {
		return fmt.Errorf("failed to open SQLite database at %s: %w", dbPath, err)
	}
	defer db.Close()

	if err = db.Ping(); err != nil {
		return fmt.Errorf("failed to ping SQLite database at %s: %w", dbPath, err)
	}

	var writer *csv.Writer
	var file *os.File
	var out io.Writer

	if outputPath != "" {
		// Validate output path (basic check, ensure parent dir exists for writing)
		// For simplicity, this validation is omitted here but would be good for production code.
		// We assume if a path is given, we try to write to it.
		file, err = os.Create(outputPath)
		if err != nil {
			return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
		}
		defer file.Close()
		out = file
	} else {
		out = os.Stdout
	}
	writer = csv.NewWriter(out)
	defer writer.Flush()

	switch tableName {
	case "NetworkSnapshots":
		return exportNetworkSnapshots(db, writer)
	case "NeuronStates":
		return exportNeuronStates(db, writer)
	default:
		return fmt.Errorf("unsupported table '%s'. Supported tables are 'NetworkSnapshots', 'NeuronStates'", tableName)
	}
}

// exportNetworkSnapshots exports the NetworkSnapshots table to CSV.
func exportNetworkSnapshots(db *sql.DB, writer *csv.Writer) error {
	headers := []string{
		"SnapshotID", "CycleCount", "Timestamp", "CortisolLevel",
		"DopamineLevel", "LearningRateModFactor", "SynaptogenesisModFactor",
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV headers for NetworkSnapshots: %w", err)
	}

	rows, err := db.Query("SELECT SnapshotID, CycleCount, Timestamp, CortisolLevel, DopamineLevel, LearningRateModFactor, SynaptogenesisModFactor FROM NetworkSnapshots ORDER BY SnapshotID")
	if err != nil {
		return fmt.Errorf("failed to query NetworkSnapshots: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r [7]sql.NullString // Use NullString to handle potential NULLs gracefully, then convert
		if err := rows.Scan(&r[0], &r[1], &r[2], &r[3], &r[4], &r[5], &r[6]); err != nil {
			return fmt.Errorf("failed to scan row from NetworkSnapshots: %w", err)
		}
		record := make([]string, len(r))
		for i, val := range r {
			if val.Valid {
				record[i] = val.String
			} else {
				record[i] = "" // Represent NULL as empty string in CSV
			}
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record for NetworkSnapshots: %w", err)
		}
	}
	return rows.Err() // Check for errors during iteration
}

// exportNeuronStates exports the NeuronStates table to CSV.
func exportNeuronStates(db *sql.DB, writer *csv.Writer) error {
	headers := []string{
		"StateID", "SnapshotID", "NeuronID", "Position", "Velocity", "Type", "CurrentState",
		"AccumulatedPotential", "BaseFiringThreshold", "CurrentFiringThreshold",
		"LastFiredCycle", "CyclesInCurrentState",
	}
	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV headers for NeuronStates: %w", err)
	}

	rows, err := db.Query(`SELECT StateID, SnapshotID, NeuronID, Position, Velocity, Type, CurrentState,
                                AccumulatedPotential, BaseFiringThreshold, CurrentFiringThreshold,
                                LastFiredCycle, CyclesInCurrentState
                         FROM NeuronStates ORDER BY StateID`)
	if err != nil {
		return fmt.Errorf("failed to query NeuronStates: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var stateID, snapshotID, neuronID, typeInt, currentStateInt, lastFiredCycle, cyclesInCurrentState sql.NullInt64
		var position, velocity sql.NullString
		var accPot, baseThr, currThr sql.NullFloat64

		if err := rows.Scan(
			&stateID, &snapshotID, &neuronID, &position, &velocity,
			&typeInt, &currentStateInt, &accPot, &baseThr, &currThr,
			&lastFiredCycle, &cyclesInCurrentState,
		); err != nil {
			return fmt.Errorf("failed to scan row from NeuronStates: %w", err)
		}

		typeStr := ""
		if typeInt.Valid {
			typeStr = neuron.Type(typeInt.Int64).String()
		}

		currentStateStr := ""
		if currentStateInt.Valid {
			currentStateStr = neuron.State(currentStateInt.Int64).String()
		}

		record := []string{
			intToString(stateID), intToString(snapshotID), intToString(neuronID),
			nullStringToString(position), nullStringToString(velocity),
			typeStr, currentStateStr,
			floatToString(accPot), floatToString(baseThr), floatToString(currThr),
			intToString(lastFiredCycle), intToString(cyclesInCurrentState),
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record for NeuronStates: %w", err)
		}
	}
	return rows.Err()
}

// Helper functions to convert sql.Null types to string for CSV
func nullStringToString(ns sql.NullString) string {
	if ns.Valid {
		return ns.String
	}
	return ""
}

func intToString(ni sql.NullInt64) string {
	if ni.Valid {
		return strconv.FormatInt(ni.Int64, 10)
	}
	return ""
}

func floatToString(nf sql.NullFloat64) string {
	if nf.Valid {
		return strconv.FormatFloat(nf.Float64, 'f', -1, 64) // Default precision
	}
	return ""
}

// Placeholder for main.go or cli.Orchestrator to call RunExport.
// This file assumes it's part of the 'storage' package.
// The actual call will be from cli.runLogUtilMode().
