package cli

import (
	"crownet/common"
	"crownet/config"
	"crownet/datagen"
	"crownet/network"
	"crownet/storage" // Para persistência JSON e SQLite
	"fmt"
	"log"
	"os"
	"time"
	// "math/rand" // Para seed, se necessário
)

// Orchestrator gerencia a execução da simulação com base na configuração da CLI.
type Orchestrator struct {
	AppCfg *config.AppConfig
	Net    *network.CrowNet
	Logger *storage.SQLiteLogger

	// Funções para dependências externas, facilitando o mock em testes.
	loadWeightsFn func(filepath string) (synaptic.NetworkWeights, error)
	saveWeightsFn func(weights synaptic.NetworkWeights, filepath string) error
}

// NewOrchestrator cria um novo orquestrador.
func NewOrchestrator(appCfg *config.AppConfig) *Orchestrator {
	return &Orchestrator{
		AppCfg: appCfg,
		// Inicializa com as funções reais do pacote storage
		loadWeightsFn: storage.LoadNetworkWeightsFromJSON,
		saveWeightsFn: storage.SaveNetworkWeightsToJSON,
	}
}

func (o *Orchestrator) initializeLogger() error {
	if o.AppCfg.Cli.DbPath != "" &&
		(o.AppCfg.Cli.Mode == config.ModeSim ||
			(o.AppCfg.Cli.Mode == config.ModeExpose && o.AppCfg.Cli.SaveInterval > 0)) {
		var err error
		o.Logger, err = storage.NewSQLiteLogger(o.AppCfg.Cli.DbPath)
		if err != nil {
			return fmt.Errorf("falha ao inicializar logger SQLite: %w", err)
		}
		fmt.Printf("Logging SQLite ativado para: %s\n", o.AppCfg.Cli.DbPath)
	}
	return nil
}

func (o *Orchestrator) createNetwork() {
	// Extrair os parâmetros necessários do AppCfg
	totalNeurons := o.AppCfg.Cli.TotalNeurons
	baseLearningRate := common.Rate(o.AppCfg.Cli.BaseLearningRate) // Converter para common.Rate
	simParams := &o.AppCfg.SimParams
	seed := o.AppCfg.Cli.Seed

	// Chamar NewCrowNet com os parâmetros individualizados, incluindo a semente
	o.Net = network.NewCrowNet(totalNeurons, baseLearningRate, simParams, seed)

	fmt.Printf("Rede criada: %d neurônios. IDs Input: %v..., IDs Output: %v...\n",
		len(o.Net.Neurons), o.Net.InputNeuronIDs[:min(5, len(o.Net.InputNeuronIDs))], o.Net.OutputNeuronIDs[:min(10, len(o.Net.OutputNeuronIDs))])
	fmt.Printf("Estado Inicial: Cortisol=%.3f, Dopamina=%.3f\n",
		o.Net.ChemicalEnv.CortisolLevel, o.Net.ChemicalEnv.DopamineLevel)
}

func (o *Orchestrator) loadWeights(filepath string) error {
	if _, err := os.Stat(filepath); err == nil { // os.Stat ainda é usado para verificar a existência do arquivo
		loadedWeights, errLoad := o.loadWeightsFn(filepath) // Usa a função injetada/mockável
		if errLoad == nil {
			o.Net.SynapticWeights = loadedWeights
			fmt.Printf("Pesos existentes carregados de %s\n", filepath)
			return nil
		}
		// Se o arquivo existe mas o carregamento falha, isso é um erro.
		return fmt.Errorf("falha ao carregar pesos de %s: %w", filepath, errLoad)
	}
	// Se o arquivo não existe, não é necessariamente um erro para todos os modos (ex: expose pode começar do zero)
	// A função chamadora decidirá se isso é um erro fatal.
	return fmt.Errorf("arquivo de pesos %s não encontrado", filepath) // Mensagem mais simples
}

func (o *Orchestrator) saveWeights(filepath string) error {
	if err := o.saveWeightsFn(o.Net.SynapticWeights, filepath); err != nil { // Usa a função injetada/mockável
		return fmt.Errorf("falha ao salvar pesos treinados em %s: %w", filepath, err)
	}
	fmt.Printf("Pesos treinados salvos em %s\n", filepath)
	return nil
}

func (o *Orchestrator) printModeSpecificConfig() {
	switch o.AppCfg.Cli.Mode {
	case config.ModeExpose:
		fmt.Printf("  %s: Épocas=%d, TaxaAprendizadoBase=%.4f, CiclosPorPadrão=%d\n",
			config.ModeExpose, o.AppCfg.Cli.Epochs, o.AppCfg.Cli.BaseLearningRate, o.AppCfg.Cli.CyclesPerPattern)
	case config.ModeObserve:
		fmt.Printf("  %s: Dígito=%d, CiclosParaAcomodar=%d\n",
			config.ModeObserve, o.AppCfg.Cli.Digit, o.AppCfg.Cli.CyclesToSettle)
	case config.ModeSim:
		fmt.Printf("  %s: TotalCiclos=%d, CaminhoDB='%s', IntervaloSaveDB=%d\n",
			config.ModeSim, o.AppCfg.Cli.Cycles, o.AppCfg.Cli.DbPath, o.AppCfg.Cli.SaveInterval)
		if o.AppCfg.Cli.StimInputFreqHz > 0 && o.AppCfg.Cli.StimInputID != -2 {
			fmt.Printf("  %s: EstímuloGeral: InputID=%d a %.1f Hz\n",
				config.ModeSim, o.AppCfg.Cli.StimInputID, o.AppCfg.Cli.StimInputFreqHz)
		}
	}
}

// Run inicia a execução do modo selecionado.
func (o *Orchestrator) Run() {
	fmt.Println("CrowNet Inicializando...")
	fmt.Printf("Modo Selecionado: %s\n", o.AppCfg.Cli.Mode)
	fmt.Printf("Configuração Base: Neurônios=%d, ArquivoDePesos='%s'\n",
		o.AppCfg.Cli.TotalNeurons, o.AppCfg.Cli.WeightsFile)

	o.printModeSpecificConfig()

	if err := o.initializeLogger(); err != nil {
		log.Fatalf("Erro na inicialização: %v", err)
	}
	if o.Logger != nil {
		defer func() {
			if errClose := o.Logger.Close(); errClose != nil {
				log.Printf("Erro ao fechar logger SQLite: %v", errClose)
			}
		}()
	}

	o.createNetwork()

	startTime := time.Now()
	switch o.AppCfg.Cli.Mode {
	case config.ModeSim:
		o.runSimMode()
	case config.ModeExpose:
		o.runExposeMode()
	case config.ModeObserve:
		o.runObserveMode()
	default:
		// A validação em config.NewAppConfig() deve pegar isso, mas um fallback é bom.
	log.Fatalf("Modo desconhecido ou não suportado encontrado em Orchestrator.Run: %s.", o.AppCfg.Cli.Mode)
	}

	if errRun != nil {
		// Usar log.Fatalf aqui ainda encerra, mas centraliza o tratamento do erro retornado pelos run...Mode.
		// Alternativamente, Run() poderia retornar o erro para main.go tratar (o que main.go já faz para config).
		log.Fatalf("Erro durante a execução do modo '%s': %v", o.AppCfg.Cli.Mode, errRun)
	}

	duration := time.Since(startTime)
	fmt.Printf("\nSessão CrowNet finalizada. Duração total: %s.\n", duration)
}

// setupContinuousInputStimulus configura o estímulo de input para o modo sim.
// Retorna erro se a configuração falhar de forma crítica (ex: ID de neurônio inválido).
func (o *Orchestrator) setupContinuousInputStimulus() error {
	if o.AppCfg.Cli.StimInputFreqHz > 0.0 && o.AppCfg.Cli.StimInputID != -2 && len(o.Net.InputNeuronIDs) > 0 {
		stimID := o.AppCfg.Cli.StimInputID
		if stimID == -1 && len(o.Net.InputNeuronIDs) > 0 { // -1 para primeiro disponível
			stimID = int(o.Net.InputNeuronIDs[0])
		}

		isValidStimID := false
		for _, id := range o.Net.InputNeuronIDs {
			if int(id) == stimID {
				isValidStimID = true
				break
			}
		}
		if isValidStimID {
			if err := o.Net.ConfigureFrequencyInput(common.NeuronID(stimID), o.AppCfg.Cli.StimInputFreqHz); err != nil {
				return fmt.Errorf("falha ao configurar estímulo de input para neurônio %d a %.1f Hz: %w", stimID, o.AppCfg.Cli.StimInputFreqHz, err)
			}
			fmt.Printf("Estímulo geral: Neurônio de Input %d a %.1f Hz.\n", stimID, o.AppCfg.Cli.StimInputFreqHz)
		} else {
			return fmt.Errorf("ID do neurônio de input para estímulo geral (%d) não encontrado ou inválido", stimID)
		}
	}
	return nil
}

func (o *Orchestrator) runSimulationLoop() { // Esta função não parece precisar retornar erro, pois os logs são informativos/avisos.
	for i := 0; i < o.AppCfg.Cli.Cycles; i++ {
		o.Net.RunCycle()
		if i%10 == 0 || i == o.AppCfg.Cli.Cycles-1 { // Log a cada 10 ciclos e no último
			fmt.Printf("Ciclo %d/%d: C:%.3f D:%.3f LRMod:%.3f SynMod:%.3f Pulsos:%d\n",
				o.Net.CycleCount-1, o.AppCfg.Cli.Cycles,
				o.Net.ChemicalEnv.CortisolLevel, o.Net.ChemicalEnv.DopamineLevel,
				o.Net.ChemicalEnv.LearningRateModulationFactor, o.Net.ChemicalEnv.SynaptogenesisModulationFactor,
				len(o.Net.ActivePulses))
		}
		if o.Logger != nil && o.AppCfg.Cli.SaveInterval > 0 && o.Net.CycleCount > 0 && int(o.Net.CycleCount)%o.AppCfg.Cli.SaveInterval == 0 {
			if err := o.Logger.LogNetworkState(o.Net); err != nil {
				log.Printf("Aviso durante salvamento periódico no DB: %v", err)
			}
		}
	}

	// Log final se não coincidir com o intervalo de salvamento
	if o.Logger != nil && (o.AppCfg.Cli.SaveInterval == 0 || (o.AppCfg.Cli.Cycles > 0 && o.AppCfg.Cli.Cycles%o.AppCfg.Cli.SaveInterval != 0)) {
		if o.AppCfg.Cli.Cycles > 0 { // Apenas salvar se algum ciclo rodou
			if err := o.Logger.LogNetworkState(o.Net); err != nil {
				log.Printf("Aviso durante salvamento final no DB: %v", err)
			}
		}
	}
}

func (o *Orchestrator) reportMonitoredOutputFrequency() {
	if o.AppCfg.Cli.MonitorOutputID != -2 && len(o.Net.OutputNeuronIDs) > 0 {
		monitorID := o.AppCfg.Cli.MonitorOutputID
		if monitorID == -1 && len(o.Net.OutputNeuronIDs) > 0 { // -1 para primeiro disponível
			monitorID = int(o.Net.OutputNeuronIDs[0])
		}

		// Validar monitorID
		isValidMonitorID := false
		for _, outID := range o.Net.OutputNeuronIDs {
			if int(outID) == monitorID {
				isValidMonitorID = true
				break
			}
		}

		if isValidMonitorID {
			freq, err := o.Net.GetOutputFrequency(common.NeuronID(monitorID))
			if err == nil {
				fmt.Printf("Frequência para Neurônio de Output %d: %.2f Hz (sobre os últimos %.0f ciclos).\n",
					monitorID, freq, o.AppCfg.SimParams.OutputFrequencyWindowCycles)
			} else {
				fmt.Printf("Aviso ao obter frequência para Neurônio de Output %d: %v\n", monitorID, err)
			}
		} else {
			fmt.Printf("Aviso: ID do neurônio de output para monitoramento (%d) não encontrado ou inválido.\n", monitorID)
		}
	}
}

func (o *Orchestrator) runSimMode() error {
	fmt.Printf("\nIniciando Simulação Geral por %d ciclos...\n", o.AppCfg.Cli.Cycles)
	if err := o.setupContinuousInputStimulus(); err != nil {
		return fmt.Errorf("erro em setupContinuousInputStimulus: %w", err)
	}
	o.Net.SetDynamicState(true, true, true) // Todas as dinâmicas ativas para 'sim'
	o.runSimulationLoop() // runSimulationLoop não retorna erro atualmente, apenas loga avisos.
	o.reportMonitoredOutputFrequency()
	fmt.Printf("Estado Final: Cortisol=%.3f, Dopamina=%.3f\n", o.Net.ChemicalEnv.CortisolLevel, o.Net.ChemicalEnv.DopamineLevel)
	return nil
}

func (o *Orchestrator) runExposureEpochs() error {
	allPatterns, err := datagen.GetAllDigitPatterns(&o.AppCfg.SimParams)
	if err != nil {
		return fmt.Errorf("falha ao carregar padrões de dígitos: %w", err)
	}

	for epoch := 0; epoch < o.AppCfg.Cli.Epochs; epoch++ {
		fmt.Printf("Época %d/%d iniciando...\n", epoch+1, o.AppCfg.Cli.Epochs)
		patternsProcessedThisEpoch := 0
		for digit := 0; digit <= 9; digit++ { // Assumindo dígitos 0-9
			pattern, ok := allPatterns[digit]
			if !ok {
				log.Printf("Aviso: Padrão para o dígito %d não encontrado, pulando.", digit)
				continue
			}

			o.Net.ResetNetworkStateForNewPattern()
			if errPres := o.Net.PresentPattern(pattern); errPres != nil {
				return fmt.Errorf("falha ao apresentar padrão para dígito %d na época %d: %w", digit, epoch+1, errPres)
			}

			for cycleInPattern := 0; cycleInPattern < o.AppCfg.Cli.CyclesPerPattern; cycleInPattern++ {
				o.Net.RunCycle()
				if o.Logger != nil && o.AppCfg.Cli.SaveInterval > 0 && o.Net.CycleCount > 0 && int(o.Net.CycleCount)%o.AppCfg.Cli.SaveInterval == 0 {
					if errLog := o.Logger.LogNetworkState(o.Net); errLog != nil {
						log.Printf("Aviso durante salvamento periódico no DB (época %d, dígito %d): %v", epoch+1, digit, errLog)
					}
				}
			}
			patternsProcessedThisEpoch++
		}
		fmt.Printf("Época %d/%d concluída. Processados %d padrões. Cortisol: %.3f, Dopamina: %.3f, FatorLR Efetivo: %.4f\n",
			epoch+1, o.AppCfg.Cli.Epochs, patternsProcessedThisEpoch,
			o.Net.ChemicalEnv.CortisolLevel, o.Net.ChemicalEnv.DopamineLevel, o.Net.ChemicalEnv.LearningRateModulationFactor)
	}
	return nil
}

func (o *Orchestrator) runExposeMode() error {
	fmt.Printf("\nIniciando Fase de Exposição por %d épocas (TaxaAprendizadoBase: %.4f, CiclosPorPadrão: %d)...\n",
		o.AppCfg.Cli.Epochs, o.AppCfg.Cli.BaseLearningRate, o.AppCfg.Cli.CyclesPerPattern)

	if err := o.loadWeights(o.AppCfg.Cli.WeightsFile); err != nil {
		fmt.Println(err) // Loga o erro mas continua, a rede usa pesos aleatórios
	}

	o.Net.SetDynamicState(true, true, true) // Todas as dinâmicas ativas para 'expose'
	if err := o.runExposureEpochs(); err != nil {
		return err
	}

	fmt.Println("Fase de exposição concluída.")
	if err := o.saveWeights(o.AppCfg.Cli.WeightsFile); err != nil {
		return err // Retorna o erro em vez de log.Fatalf
	}
	return nil
}

func (o *Orchestrator) runObservationPattern() ([]float64, error) {
	patternToObserve, err := datagen.GetDigitPattern(o.AppCfg.Cli.Digit, &o.AppCfg.SimParams)
	if err != nil {
		return nil, fmt.Errorf("falha ao obter padrão para o dígito %d: %w", o.AppCfg.Cli.Digit, err)
	}

	o.Net.ResetNetworkStateForNewPattern()
	if errPres := o.Net.PresentPattern(patternToObserve); errPres != nil {
		return nil, fmt.Errorf("falha ao apresentar padrão para observação: %w", errPres)
	}

	for i := 0; i < o.AppCfg.Cli.CyclesToSettle; i++ {
		o.Net.RunCycle()
	}

	outputActivation, errAct := o.Net.GetOutputActivation()
	if errAct != nil {
		return nil, fmt.Errorf("falha ao obter ativação de saída: %w", errAct)
	}
	return outputActivation, nil
}

func (o *Orchestrator) displayOutputActivation(outputActivation []float64) {
	fmt.Printf("Dígito Apresentado: %d\n", o.AppCfg.Cli.Digit)
	fmt.Println("Padrão de Ativação dos Neurônios de Saída (Potencial Acumulado):")
	for i, val := range outputActivation {
		neuronIDStr := "N/A"
		if i < len(o.Net.OutputNeuronIDs) {
			neuronIDStr = fmt.Sprintf("%d", o.Net.OutputNeuronIDs[i])
		}
		fmt.Printf("  OutNeurônio[%d] (ID %s): %.4f\n", i, neuronIDStr, val)
	}
}

func (o *Orchestrator) runObserveMode() error {
	fmt.Printf("\nObservando Resposta da Rede para o dígito %d (%d ciclos de acomodação)...\n",
		o.AppCfg.Cli.Digit, o.AppCfg.Cli.CyclesToSettle)

	if err := o.loadWeights(o.AppCfg.Cli.WeightsFile); err != nil {
		return fmt.Errorf("para o modo %s: %w. Exponha a rede primeiro", config.ModeObserve, err)
	}

	o.Net.SetDynamicState(false, false, false) // Dinâmicas alteradoras desligadas para observação

	outputActivation, err := o.runObservationPattern()
	if err != nil {
		return fmt.Errorf("falha ao rodar padrão de observação: %w", err)
	}

	o.displayOutputActivation(outputActivation)
	o.Net.SetDynamicState(true, true, true) // Restaurar estado dinâmico padrão
	return nil
}

// Helper para evitar pânico com slices vazias
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Fim do arquivo

// --- Métodos de Teste (Exportados para uso no pacote _test) ---

// SetupContinuousInputStimulusForTest é um wrapper de teste para setupContinuousInputStimulus.
func (o *Orchestrator) SetupContinuousInputStimulusForTest() error {
	return o.setupContinuousInputStimulus()
}

// RunObserveModeForTest é um wrapper de teste para runObserveMode.
func (o *Orchestrator) RunObserveModeForTest() error {
	return o.runObserveMode()
}

// RunExposeModeForTest é um wrapper de teste para runExposeMode.
func (o *Orchestrator) RunExposeModeForTest() error {
	return o.runExposeMode()
}

// SetLoadWeightsFn é um setter de teste para a função de carregar pesos.
func (o *Orchestrator) SetLoadWeightsFn(fn func(filepath string) (synaptic.NetworkWeights, error)) {
	o.loadWeightsFn = fn
}

// SetSaveWeightsFn é um setter de teste para a função de salvar pesos.
func (o *Orchestrator) SetSaveWeightsFn(fn func(weights synaptic.NetworkWeights, filepath string) error) {
	o.saveWeightsFn = fn
}
```
