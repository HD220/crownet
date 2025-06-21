package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"

	"github.com/user/crownet/src/core"
	_ "github.com/mattn/go-sqlite3"
)

const schemaVersion = 1

// createSchema garante que as tabelas necessárias existam no banco de dados.
func createSchema(db *sql.DB) error {
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS network_metadata (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			config_json TEXT,
			current_cycle INTEGER,
			cortisol_level REAL,
			dopamine_level REAL,
			random_seed INTEGER,
			schema_version INTEGER
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create network_metadata table: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS neurons (
			neuron_id INTEGER PRIMARY KEY,
			type INTEGER,
			state INTEGER,
			position_json TEXT,
			current_potential REAL,
			firing_threshold REAL,
			base_firing_threshold REAL,
			last_firing_cycle INTEGER,
			refractory_cycles INTEGER,
			cycles_in_rest INTEGER,
			cycles_in_firing INTEGER,
			cycles_in_refractory INTEGER,
			refractory_period_absolute INTEGER,
			refractory_period_relative INTEGER
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create neurons table: %w", err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS gland_metadata (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			position_json TEXT
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create gland_metadata table: %w", err)
	}
	return nil
}

// SaveNetwork salva o estado completo da rede em um arquivo SQLite.
func SaveNetwork(network *core.Network, filePath string) error {
	_ = os.Remove(filePath)

	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return fmt.Errorf("failed to open database at %s: %w", filePath, err)
	}
	defer db.Close()

	if err_schema := createSchema(db); err_schema != nil { // Renomeado
		return fmt.Errorf("failed to create schema: %w", err_schema)
	}

	tx, err_tx := db.Begin() // Renomeado
	if err_tx != nil {
		return fmt.Errorf("failed to begin transaction: %w", err_tx)
	}
	defer tx.Rollback()

	configJSON, err_marshal_cfg := json.Marshal(network.Config) // Renomeado
	if err_marshal_cfg != nil {
		return fmt.Errorf("failed to marshal network config: %w", err_marshal_cfg)
	}

	_, err_exec_meta := tx.Exec(`
		INSERT OR REPLACE INTO network_metadata
		(id, config_json, current_cycle, cortisol_level, dopamine_level, random_seed, schema_version)
		VALUES (1, ?, ?, ?, ?, ?, ?);
	`, string(configJSON), network.CurrentCycle, network.CortisolLevel, network.DopamineLevel, network.Config.RandomSeed, schemaVersion) // Renomeado
	if err_exec_meta != nil {
		return fmt.Errorf("failed to insert network metadata: %w", err_exec_meta)
	}

	glandPosJSON, err_marshal_gland := json.Marshal(network.Gland.Position) // Renomeado
	if err_marshal_gland != nil {
		return fmt.Errorf("failed to marshal gland position: %w", err_marshal_gland)
	}
	_, err_exec_gland := tx.Exec(`INSERT OR REPLACE INTO gland_metadata (id, position_json) VALUES (1, ?);`, string(glandPosJSON)) // Renomeado
	if err_exec_gland != nil {
		return fmt.Errorf("failed to save gland metadata: %w", err_exec_gland)
	}

	stmt, err_prepare := tx.Prepare(`
		INSERT INTO neurons (
			neuron_id, type, state, position_json, current_potential, firing_threshold,
			base_firing_threshold, last_firing_cycle, refractory_cycles, cycles_in_rest,
			cycles_in_firing, cycles_in_refractory, refractory_period_absolute, refractory_period_relative
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`) // Renomeado
	if err_prepare != nil {
		return fmt.Errorf("failed to prepare neuron insert statement: %w", err_prepare)
	}
	defer stmt.Close()

	for _, neuron := range network.Neurons {
		if neuron == nil { continue }
		posJSON, err_marshal_pos := json.Marshal(neuron.Position)
		if err_marshal_pos != nil {
			return fmt.Errorf("failed to marshal neuron %d position: %w", neuron.ID, err_marshal_pos)
		}
		_, err_exec_neuron := stmt.Exec(
			neuron.ID, neuron.Type, neuron.State, string(posJSON),
			neuron.CurrentPotential, neuron.FiringThreshold, neuron.BaseFiringThreshold,
			neuron.LastFiringCycle, neuron.RefractoryCycles, neuron.CyclesInRest,
			neuron.CyclesInFiring, neuron.CyclesInRefractory,
			neuron.RefractoryPeriodAbsolute, neuron.RefractoryPeriodRelative,
		)
		if err_exec_neuron != nil {
			return fmt.Errorf("failed to insert neuron %d: %w", neuron.ID, err_exec_neuron)
		}
	}

	if err_commit := tx.Commit(); err_commit != nil {
		return fmt.Errorf("failed to commit transaction: %w", err_commit)
	}
	return nil
}

// LoadNetwork carrega o estado da rede de um arquivo SQLite.
func LoadNetwork(filePath string) (*core.Network, error) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database at %s: %w", filePath, err)
	}
	defer db.Close()

	var configJSON string
	var currentCycle int
	var cortisolLevel, dopamineLevel float64
	var randomSeed int64
	var sVersion int

	err_scan_meta := db.QueryRow(`
		SELECT config_json, current_cycle, cortisol_level, dopamine_level, random_seed, schema_version
		FROM network_metadata WHERE id = 1;
	`).Scan(&configJSON, &currentCycle, &cortisolLevel, &dopamineLevel, &randomSeed, &sVersion) // Renomeado
	if err_scan_meta != nil {
		return nil, fmt.Errorf("failed to load network metadata: %w", err_scan_meta)
	}

	if sVersion > schemaVersion {
		return nil, fmt.Errorf("database schema version (%d) is newer than supported version (%d)", sVersion, schemaVersion)
	}

	var config core.NetworkConfig
	if err_unmarshal_cfg := json.Unmarshal([]byte(configJSON), &config); err_unmarshal_cfg != nil {
		return nil, fmt.Errorf("failed to unmarshal network config: %w", err_unmarshal_cfg)
	}

	config.RandomSeed = randomSeed
	network := core.NewNetwork(config)
	network.CurrentCycle = currentCycle
	network.CortisolLevel = cortisolLevel
	network.DopamineLevel = dopamineLevel

	var glandPosJSON string
	err_gland_scan := db.QueryRow(`SELECT position_json FROM gland_metadata WHERE id = 1;`).Scan(&glandPosJSON)
	if err_gland_scan != nil {
		if err_gland_scan != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to load gland metadata: %w", err_gland_scan)
		}
	} else {
		var glandPos [core.SpaceDimensions]float64
		if err_unmarshal_gland_pos := json.Unmarshal([]byte(glandPosJSON), &glandPos); err_unmarshal_gland_pos != nil {
			return nil, fmt.Errorf("failed to unmarshal gland position: %w", err_unmarshal_gland_pos)
		}
		network.Gland.Position = glandPos
	}

	rows, err_query_neurons := db.Query(`
		SELECT neuron_id, type, state, position_json, current_potential, firing_threshold,
		       base_firing_threshold, last_firing_cycle, refractory_cycles, cycles_in_rest,
		       cycles_in_firing, cycles_in_refractory, refractory_period_absolute, refractory_period_relative
		FROM neurons ORDER BY neuron_id ASC;
	`)
	if err_query_neurons != nil {
		return nil, fmt.Errorf("failed to query neurons: %w", err_query_neurons)
	}
	defer rows.Close()

	loadedNeurons := make([]*core.Neuron, 0, config.NumNeurons)
	for rows.Next() {
		neuron := &core.Neuron{}
		var posJSON string

		err_scan_neuron := rows.Scan(
			&neuron.ID, &neuron.Type, &neuron.State, &posJSON,
			&neuron.CurrentPotential, &neuron.FiringThreshold, &neuron.BaseFiringThreshold,
			&neuron.LastFiringCycle, &neuron.RefractoryCycles, &neuron.CyclesInRest,
			&neuron.CyclesInFiring, &neuron.CyclesInRefractory,
			&neuron.RefractoryPeriodAbsolute, &neuron.RefractoryPeriodRelative,
		)
		if err_scan_neuron != nil {
			return nil, fmt.Errorf("failed to scan neuron row: %w", err_scan_neuron)
		}

		var position [core.SpaceDimensions]float64
		if err_unmarshal_pos := json.Unmarshal([]byte(posJSON), &position); err_unmarshal_pos != nil {
			return nil, fmt.Errorf("failed to unmarshal position for neuron %d: %w", neuron.ID, err_unmarshal_pos)
		}
		neuron.Position = position
		loadedNeurons = append(loadedNeurons, neuron)
	}
	if err_rows_iter := rows.Err(); err_rows_iter != nil { // Renomeado
		return nil, fmt.Errorf("error iterating neuron rows: %w", err_rows_iter)
	}

	network.Neurons = loadedNeurons
	if len(network.Neurons) != config.NumNeurons {
		fmt.Printf("Warning: Loaded %d neurons, but config expected %d.\n", len(network.Neurons), config.NumNeurons)
	}
	return network, nil
}
