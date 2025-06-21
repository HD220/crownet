package storage

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os" // Para remover o arquivo se o salvamento falhar parcialmente

	"github.com/user/crownet/src/core"
	_ "github.com/mattn/go-sqlite3" // Driver SQLite
)

const schemaVersion = 1

// createSchema garante que as tabelas necessárias existam no banco de dados.
func createSchema(db *sql.DB) error {
	// Tabela para metadados da rede (configuração, estado global)
	// Usaremos JSON para armazenar estruturas complexas como NetworkConfig.
	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS network_metadata (
			id INTEGER PRIMARY KEY CHECK (id = 1), -- Apenas uma linha para metadados
			config_json TEXT,
			current_cycle INTEGER,
			cortisol_level REAL,
			dopamine_level REAL,
			random_seed INTEGER, -- Para o RNG da rede, se precisarmos salvar/restaurar
			schema_version INTEGER
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create network_metadata table: %w", err)
	}

	// Tabela para neurônios
	// Posição será armazenada como JSON array.
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS neurons (
			neuron_id INTEGER PRIMARY KEY,
			type INTEGER,
			state INTEGER,
			position_json TEXT, -- [SpaceDimensions]float64
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

	// Tabela para a glândula (se tiver mais propriedades no futuro)
	// Por ora, sua posição é central e derivada da config, mas podemos querer salvá-la.
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS gland_metadata (
			id INTEGER PRIMARY KEY CHECK (id = 1),
			position_json TEXT
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create gland_metadata table: %w", err)
	}

	// Pulsos geralmente não são persistidos, pois são transitórios.
	// Se precisarmos, uma tabela de pulsos seria similar à de neurônios.

	return nil
}

// SaveNetwork salva o estado completo da rede em um arquivo SQLite.
func SaveNetwork(network *core.Network, filePath string) error {
	// Remover arquivo existente para garantir um estado limpo se o salvamento falhar
	// ou se o esquema mudou e queremos começar do zero.
	// No entanto, para um salvamento incremental, poderíamos querer atualizar.
	// Para MVP, vamos recriar.
	_ = os.Remove(filePath) // Ignora erro se o arquivo não existir

	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return fmt.Errorf("failed to open database at %s: %w", filePath, err)
	}
	defer db.Close()

	if err := createSchema(db); err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Rollback se não houver commit explícito

	// 1. Salvar Metadados da Rede
	configJSON, err := json.Marshal(network.Config)
	if err != nil {
		return fmt.Errorf("failed to marshal network config: %w", err)
	}

	// Usar INSERT OR REPLACE para a linha de metadados (id=1)
	_, err = tx.Exec(`
		INSERT OR REPLACE INTO network_metadata
		(id, config_json, current_cycle, cortisol_level, dopamine_level, random_seed, schema_version)
		VALUES (1, ?, ?, ?, ?, ?, ?);
	`, string(configJSON), network.CurrentCycle, network.CortisolLevel, network.DopamineLevel, network.Config.RandomSeed, schemaVersion)
	if err != nil {
		return fmt.Errorf("failed to insert network metadata: %w", err)
	}

	// 2. Salvar Glândula (se necessário, ou se sua posição puder mudar)
	// Por enquanto, a posição da glândula é derivada da config (centro do espaço).
	// Se ela pudesse se mover ou tivesse estado, salvaríamos aqui.
	glandPosJSON, err := json.Marshal(network.Gland.Position)
	if err != nil {
		return fmt.Errorf("failed to marshal gland position: %w", err)
	}
	_, err = tx.Exec(`INSERT OR REPLACE INTO gland_metadata (id, position_json) VALUES (1, ?);`, string(glandPosJSON))
	if err != nil {
		return fmt.Errorf("failed to save gland metadata: %w", err)
	}


	// 3. Salvar Neurônios
	// Limpar tabela de neurônios antes de inserir novos, para evitar duplicatas se chamarmos Save várias vezes no mesmo DB.
	// No entanto, como estamos recriando o arquivo, isso não é estritamente necessário aqui.
	// Mas é uma boa prática se o arquivo pudesse ser reutilizado.
	// _, err = tx.Exec("DELETE FROM neurons;")
	// if err != nil {
	// 	return fmt.Errorf("failed to clear neurons table: %w", err)
	// }

	stmt, err := tx.Prepare(`
		INSERT INTO neurons (
			neuron_id, type, state, position_json, current_potential, firing_threshold,
			base_firing_threshold, last_firing_cycle, refractory_cycles, cycles_in_rest,
			cycles_in_firing, cycles_in_refractory, refractory_period_absolute, refractory_period_relative
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);
	`)
	if err != nil {
		return fmt.Errorf("failed to prepare neuron insert statement: %w", err)
	}
	defer stmt.Close()

	for _, neuron := range network.Neurons {
		if neuron == nil { continue }
		posJSON, err := json.Marshal(neuron.Position)
		if err != nil {
			return fmt.Errorf("failed to marshal neuron %d position: %w", neuron.ID, err)
		}
		_, err = stmt.Exec(
			neuron.ID, neuron.Type, neuron.State, string(posJSON),
			neuron.CurrentPotential, neuron.FiringThreshold, neuron.BaseFiringThreshold,
			neuron.LastFiringCycle, neuron.RefractoryCycles, neuron.CyclesInRest,
			neuron.CyclesInFiring, neuron.CyclesInRefractory,
			neuron.RefractoryPeriodAbsolute, neuron.RefractoryPeriodRelative,
		)
		if err != nil {
			return fmt.Errorf("failed to insert neuron %d: %w", neuron.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// LoadNetwork carrega o estado da rede de um arquivo SQLite.
// Retorna a rede carregada e sua configuração (que também está dentro da rede).
func LoadNetwork(filePath string) (*core.Network, error) {
	db, err := sql.Open("sqlite3", filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database at %s: %w", filePath, err)
	}
	defer db.Close()

	// 1. Carregar Metadados da Rede
	var configJSON string
	var currentCycle int
	var cortisolLevel, dopamineLevel float64
	var randomSeed int64
	var sVersion int

	err = db.QueryRow(`
		SELECT config_json, current_cycle, cortisol_level, dopamine_level, random_seed, schema_version
		FROM network_metadata WHERE id = 1;
	`).Scan(&configJSON, &currentCycle, &cortisolLevel, &dopamineLevel, &randomSeed, &sVersion)
	if err != nil {
		return nil, fmt.Errorf("failed to load network metadata: %w", err)
	}

	if sVersion > schemaVersion {
		return nil, fmt.Errorf("database schema version (%d) is newer than supported version (%d)", sVersion, schemaVersion)
	}


	var config core.NetworkConfig
	if err := json.Unmarshal([]byte(configJSON), &config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal network config: %w", err)
	}

	// Criar uma nova rede com a configuração carregada, mas sem inicializar neurônios ainda.
	// Os neurônios serão carregados do DB.
	// A semente do RNG precisa ser passada para a config antes de criar a rede,
	// para que o RNG da rede seja inicializado corretamente se precisarmos dele antes dos neurônios.
	config.RandomSeed = randomSeed
	network := core.NewNetwork(config) // NewNetwork usará a config para definir tamanhos, etc.
	                                   // mas vamos sobrescrever os neurônios e outros estados.
	network.CurrentCycle = currentCycle
	network.CortisolLevel = cortisolLevel
	network.DopamineLevel = dopamineLevel
	// O RNG da rede já foi inicializado em NewNetwork com config.RandomSeed.

	// Carregar Posição da Glândula (se foi salva e se ela pode mudar)
	var glandPosJSON string
	err = db.QueryRow(`SELECT position_json FROM gland_metadata WHERE id = 1;`).Scan(&glandPosJSON)
	if err != nil {
		// Se não houver gland_metadata, podemos assumir que ela está no centro (como em NewNetwork)
		// ou retornar um erro se for esperado que esteja lá.
		// Para o MVP, NewNetwork já a coloca no centro, então podemos ignorar o erro se for sql.ErrNoRows.
		if err != sql.ErrNoRows {
			return nil, fmt.Errorf("failed to load gland metadata: %w", err)
		}
		// Se não houver linha, a glândula já foi inicializada por NewNetwork com base na config.
	} else {
		var glandPos [core.SpaceDimensions]float64
		if err := json.Unmarshal([]byte(glandPosJSON), &glandPos); err != nil {
			return nil, fmt.Errorf("failed to unmarshal gland position: %w", err)
		}
		network.Gland.Position = glandPos
	}


	// 2. Carregar Neurônios
	rows, err := db.Query(`
		SELECT neuron_id, type, state, position_json, current_potential, firing_threshold,
		       base_firing_threshold, last_firing_cycle, refractory_cycles, cycles_in_rest,
		       cycles_in_firing, cycles_in_refractory, refractory_period_absolute, refractory_period_relative
		FROM neurons ORDER BY neuron_id ASC;
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to query neurons: %w", err)
	}
	defer rows.Close()

	loadedNeurons := make([]*core.Neuron, 0, config.NumNeurons)
	for rows.Next() {
		neuron := &core.Neuron{} // Criar um neurônio vazio
		var posJSON string

		err := rows.Scan(
			&neuron.ID, &neuron.Type, &neuron.State, &posJSON,
			&neuron.CurrentPotential, &neuron.FiringThreshold, &neuron.BaseFiringThreshold,
			&neuron.LastFiringCycle, &neuron.RefractoryCycles, &neuron.CyclesInRest,
			&neuron.CyclesInFiring, &neuron.CyclesInRefractory,
			&neuron.RefractoryPeriodAbsolute, &neuron.RefractoryPeriodRelative,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan neuron row: %w", err)
		}

		var position [core.SpaceDimensions]float64
		if err := json.Unmarshal([]byte(posJSON), &position); err != nil {
			return nil, fmt.Errorf("failed to unmarshal position for neuron %d: %w", neuron.ID, err)
		}
		neuron.Position = position
		loadedNeurons = append(loadedNeurons, neuron)
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating neuron rows: %w", err)
	}

	// Substituir os neurônios inicializados por NewNetwork pelos carregados
	network.Neurons = loadedNeurons
	if len(network.Neurons) != config.NumNeurons {
		fmt.Printf("Warning: Loaded %d neurons, but config expected %d. The model file might be from a different configuration.\n", len(network.Neurons), config.NumNeurons)
		// Poderia ser um erro fatal dependendo da política.
	}


	return network, nil
}
