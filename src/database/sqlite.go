package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"crownet/src/core" // Ajuste o caminho se necessário
	_ "github.com/mattn/go-sqlite3"
)

const dbFileName = "data/crownet.db"

// DB representa a conexão com o banco de dados.
var DB *sql.DB

// InitDB inicializa a conexão com o banco de dados SQLite e cria as tabelas se não existirem.
func InitDB() error {
	// Garante que o diretório 'data' exista
	if err := os.MkdirAll(filepath.Dir(dbFileName), 0755); err != nil {
		return fmt.Errorf("falha ao criar diretório data: %w", err)
	}

	var err error
	DB, err = sql.Open("sqlite3", dbFileName)
	if err != nil {
		return fmt.Errorf("falha ao abrir banco de dados: %w", err)
	}

	if err = DB.Ping(); err != nil {
		return fmt.Errorf("falha ao conectar ao banco de dados: %w", err)
	}

	return createTables()
}

// createTables cria as tabelas necessárias no banco de dados.
func createTables() error {
	neuronTableSQL := `
	CREATE TABLE IF NOT EXISTS neurons (
		id INTEGER PRIMARY KEY,
		neuron_type INTEGER,
		pos_x0 REAL, pos_x1 REAL, pos_x2 REAL, pos_x3 REAL,
		pos_x4 REAL, pos_x5 REAL, pos_x6 REAL, pos_x7 REAL,
		pos_x8 REAL, pos_x9 REAL, pos_x10 REAL, pos_x11 REAL,
		pos_x12 REAL, pos_x13 REAL, pos_x14 REAL, pos_x15 REAL
	);`

	// Tabela para registrar o estado dos neurônios em cada ciclo de simulação
	neuronStateLogSQL := `
	CREATE TABLE IF NOT EXISTS neuron_state_log (
		cycle INTEGER,
		neuron_id INTEGER,
		state INTEGER,
		current_potential REAL,
		firing_threshold REAL,
		last_firing_cycle INTEGER,
		pos_x0 REAL, pos_x1 REAL, pos_x2 REAL, pos_x3 REAL,
		pos_x4 REAL, pos_x5 REAL, pos_x6 REAL, pos_x7 REAL,
		pos_x8 REAL, pos_x9 REAL, pos_x10 REAL, pos_x11 REAL,
		pos_x12 REAL, pos_x13 REAL, pos_x14 REAL, pos_x15 REAL,
		PRIMARY KEY (cycle, neuron_id),
		FOREIGN KEY (neuron_id) REFERENCES neurons(id)
	);`

	// Tabela para registrar pulsos (pode gerar muitos dados, usar com cautela)
	pulseLogSQL := `
	CREATE TABLE IF NOT EXISTS pulse_log (
		pulse_id INTEGER PRIMARY KEY AUTOINCREMENT,
		source_neuron_id INTEGER,
		emitted_cycle INTEGER,
		strength REAL,
		arrival_time INTEGER,
		target_neuron_id INTEGER, -- Opcional, se o pulso for direcionado
		FOREIGN KEY (source_neuron_id) REFERENCES neurons(id)
	);`

	// Tabela para registrar níveis de cortisol e dopamina (global ou por região/neurônio)
	modulationLogSQL := `
	CREATE TABLE IF NOT EXISTS modulation_log (
		cycle INTEGER PRIMARY KEY,
		cortisol_level REAL
		-- Dopamine pode ser mais complexo, talvez uma tabela separada se for por neurônio
		-- ou uma média global aqui. Por simplicidade, começamos com cortisol.
	);`
	// Tabela para registrar níveis de dopamina por neurônio em cada ciclo
	dopamineLogSQL := `
	CREATE TABLE IF NOT EXISTS dopamine_log (
		cycle INTEGER,
		neuron_id INTEGER,
		dopamine_level REAL,
		PRIMARY KEY (cycle, neuron_id),
		FOREIGN KEY (neuron_id) REFERENCES neurons(id)
	);`


	tables := []string{neuronTableSQL, neuronStateLogSQL, pulseLogSQL, modulationLogSQL, dopamineLogSQL}
	for _, tableSQL := range tables {
		_, err := DB.Exec(tableSQL)
		if err != nil {
			return fmt.Errorf("falha ao criar tabela: %w\nSQL:\n%s", err, tableSQL)
		}
	}
	log.Println("Tabelas do banco de dados verificadas/criadas com sucesso.")
	return nil
}

// SaveInitialNeurons salva o estado inicial (ID, tipo, posição) dos neurônios.
func SaveInitialNeurons(neurons map[int]*core.Neuron) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("falha ao iniciar transação: %w", err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO neurons (id, neuron_type,
			pos_x0, pos_x1, pos_x2, pos_x3, pos_x4, pos_x5, pos_x6, pos_x7,
			pos_x8, pos_x9, pos_x10, pos_x11, pos_x12, pos_x13, pos_x14, pos_x15)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("falha ao preparar statement para salvar neurônios: %w", err)
	}
	defer stmt.Close()

	for _, n := range neurons {
		args := []interface{}{n.ID, n.Type}
		for i := 0; i < 16; i++ {
			args = append(args, n.Position[i])
		}
		_, err := stmt.Exec(args...)
		if err != nil {
			tx.Rollback()
			// Consider logging the specific neuron that failed
			return fmt.Errorf("falha ao executar statement para salvar neurônio ID %d: %w", n.ID, err)
		}
	}

	return tx.Commit()
}

// LogNetworkState registra o estado atual da rede (neurônios, cortisol, dopamina) no banco de dados.
func LogNetworkState(cycle uint64, network *core.NeuralNetwork) error {
	tx, err := DB.Begin()
	if err != nil {
		return fmt.Errorf("falha ao iniciar transação para log: %w", err)
	}

	// Log Neuron States
	neuronStmt, err := tx.Prepare(`
		INSERT INTO neuron_state_log (cycle, neuron_id, state, current_potential, firing_threshold, last_firing_cycle,
			pos_x0, pos_x1, pos_x2, pos_x3, pos_x4, pos_x5, pos_x6, pos_x7,
			pos_x8, pos_x9, pos_x10, pos_x11, pos_x12, pos_x13, pos_x14, pos_x15)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("falha ao preparar statement para log de estado de neurônio: %w", err)
	}
	defer neuronStmt.Close()

	for _, n := range network.Neurons {
		args := []interface{}{cycle, n.ID, n.State, n.CurrentPotential, n.FiringThreshold, n.LastFiringCycle}
		for i := 0; i < 16; i++ {
			args = append(args, n.Position[i])
		}
		_, err = neuronStmt.Exec(args...)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("falha ao executar statement para log do neurônio ID %d no ciclo %d: %w", n.ID, cycle, err)
		}
	}

	// Log Cortisol Level
	_, err = tx.Exec("INSERT INTO modulation_log (cycle, cortisol_level) VALUES (?, ?)", cycle, network.CortisolGland.CortisolLevel)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("falha ao logar nível de cortisol no ciclo %d: %w", cycle, err)
	}

	// Log Dopamine Levels
	dopamineStmt, err := tx.Prepare("INSERT INTO dopamine_log (cycle, neuron_id, dopamine_level) VALUES (?, ?, ?)")
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("falha ao preparar statement para log de dopamina: %w", err)
	}
	defer dopamineStmt.Close()

	for neuronID, level := range network.DopamineLevels {
		_, err = dopamineStmt.Exec(cycle, neuronID, level)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("falha ao logar nível de dopamina para neurônio ID %d no ciclo %d: %w", neuronID, cycle, err)
		}
	}


	return tx.Commit()
}

// LogPulse registra um pulso no banco de dados.
func LogPulse(pulse *core.Pulse) error {
	_, err := DB.Exec(`
		INSERT INTO pulse_log (source_neuron_id, emitted_cycle, strength, arrival_time, target_neuron_id)
		VALUES (?, ?, ?, ?, ?)`,
		pulse.SourceNeuronID, pulse.EmittedCycle, pulse.Strength, pulse.ArrivalTime, sql.NullInt64{Int64: int64(pulse.TargetNeuronID), Valid: pulse.TargetNeuronID != 0}) // Assumindo que 0 não é um ID de neurônio válido ou indica broadcast
	if err != nil {
		return fmt.Errorf("falha ao logar pulso: %w", err)
	}
	return nil
}


// LoadNeurons carrega os neurônios do banco de dados para reconstruir um estado.
// Esta função é mais complexa se precisarmos carregar o estado exato de uma simulação anterior.
// Para o MVP, podemos focar em salvar. A carga pode ser uma adição futura.
func LoadNetworkState(cycle uint64) (*core.NeuralNetwork, error) {
	nn := &core.NeuralNetwork{
		Neurons:        make(map[int]*core.Neuron),
		DopamineLevels: make(map[int]float64),
		CurrentCycle:   cycle,
	}

	// Carregar neurônios e seu último estado logado no ciclo especificado
	rows, err := DB.Query(`
        SELECT
            n.id, n.neuron_type,
            nsl.state, nsl.current_potential, nsl.firing_threshold, nsl.last_firing_cycle,
            nsl.pos_x0, nsl.pos_x1, nsl.pos_x2, nsl.pos_x3, nsl.pos_x4, nsl.pos_x5, nsl.pos_x6, nsl.pos_x7,
            nsl.pos_x8, nsl.pos_x9, nsl.pos_x10, nsl.pos_x11, nsl.pos_x12, nsl.pos_x13, nsl.pos_x14, nsl.pos_x15
        FROM neurons n
        JOIN neuron_state_log nsl ON n.id = nsl.neuron_id
        WHERE nsl.cycle = ?
    `, cycle)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar estados dos neurônios no ciclo %d: %w", cycle, err)
	}
	defer rows.Close()

	for rows.Next() {
		n := &core.Neuron{}
		var posFields [16]interface{}
		for i := 0; i < 16; i++ {
			posFields[i] = &n.Position[i]
		}
		scanArgs := []interface{}{
			&n.ID, &n.Type, &n.State, &n.CurrentPotential, &n.FiringThreshold, &n.LastFiringCycle,
		}
		scanArgs = append(scanArgs, posFields[:]...)

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("falha ao escanear dados do neurônio: %w", err)
		}
		nn.Neurons[n.ID] = n
	}
	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("erro durante iteração de linhas de neurônios: %w", err)
	}

	// Carregar nível de cortisol
	cortisolRow := DB.QueryRow("SELECT cortisol_level FROM modulation_log WHERE cycle = ?", cycle)
	var cortisolLevel float64
	if err := cortisolRow.Scan(&cortisolLevel); err != nil {
		if err == sql.ErrNoRows {
			log.Printf("Nenhum registro de cortisol encontrado para o ciclo %d", cycle)
			// nn.CortisolGland pode precisar ser inicializado se não houver dados
			nn.CortisolGland = &core.Gland{CortisolLevel: 0.0} // Default ou valor de config
		} else {
			return nil, fmt.Errorf("falha ao escanear nível de cortisol no ciclo %d: %w", cycle, err)
		}
	} else {
		// Assumindo que a posição da glândula é fixa e definida na inicialização da rede.
		// Se a posição da glândula pudesse mudar e fosse logada, precisaríamos carregá-la também.
		// Por agora, apenas atualizamos o nível.
		// Precisamos de uma configuração inicial para a posição da glândula.
		// Esta função de Load pode precisar de um `core.Config` para recriar certos aspectos.
		nn.CortisolGland = &core.Gland{CortisolLevel: cortisolLevel, Position: core.GetDefaultConfig().CortisolGlandPosition}
	}


	// Carregar níveis de dopamina
	dopamineRows, err := DB.Query("SELECT neuron_id, dopamine_level FROM dopamine_log WHERE cycle = ?", cycle)
	if err != nil {
		return nil, fmt.Errorf("falha ao consultar níveis de dopamina no ciclo %d: %w", cycle, err)
	}
	defer dopamineRows.Close()

	for dopamineRows.Next() {
		var neuronID int
		var level float64
		if err := dopamineRows.Scan(&neuronID, &level); err != nil {
			return nil, fmt.Errorf("falha ao escanear nível de dopamina: %w", err)
		}
		nn.DopamineLevels[neuronID] = level
	}
	if err = dopamineRows.Err(); err != nil {
		return nil, fmt.Errorf("erro durante iteração de linhas de dopamina: %w", err)
	}

	// Carregar pulsos ativos (mais complexo, pois pulsos são transitórios)
	// Para um MVP, pode ser suficiente carregar o estado dos neurônios e modulações.
	// Recriar a lista exata de pulsos em voo exigiria salvar e carregar seus estados detalhados.

	log.Printf("Estado da rede carregado para o ciclo %d com %d neurônios.", cycle, len(nn.Neurons))
	return nn, nil
}


// Helper para converter string de vetor para core.Vector16D
func parseVector16D(s string) (core.Vector16D, error) {
	var v core.Vector16D
	parts := strings.Split(strings.Trim(s, "[]"), " ")
	if len(parts) != 16 {
		return v, fmt.Errorf("string de vetor inválida, esperado 16 componentes, obteve %d: %s", len(parts), s)
	}
	for i, part := range parts {
		val, err := strconv.ParseFloat(part, 64)
		if err != nil {
			return v, fmt.Errorf("falha ao converter componente do vetor '%s': %w", part, err)
		}
		v[i] = val
	}
	return v, nil
}

// Helper para formatar core.Vector16D para string (não usado diretamente no DB, mas útil para debug)
func formatVector16D(v core.Vector16D) string {
	parts := make([]string, 16)
	for i, val := range v {
		parts[i] = strconv.FormatFloat(val, 'f', -1, 64)
	}
	return "[" + strings.Join(parts, " ") + "]"
}

// CloseDB fecha a conexão com o banco de dados.
func CloseDB() {
	if DB != nil {
		DB.Close()
		log.Println("Conexão com o banco de dados fechada.")
	}
}
