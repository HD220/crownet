package cli

import (
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
}

// NewOrchestrator cria um novo orquestrador.
func NewOrchestrator(appCfg *config.AppConfig) *Orchestrator {
	// rand.Seed(time.Now().UnixNano()) // Seed global, se necessário para aleatoriedade consistente
	return &Orchestrator{
		AppCfg: appCfg,
	}
}

// Run inicia a execução do modo selecionado.
func (o *Orchestrator) Run() {
	fmt.Println("CrowNet Inicializando...")
	fmt.Printf("Modo Selecionado: %s\n", o.AppCfg.Cli.Mode)
	fmt.Printf("Configuração Base: Neurônios=%d, ArquivoDePesos='%s'\n",
		o.AppCfg.Cli.TotalNeurons, o.AppCfg.Cli.WeightsFile)

	// Imprimir informações específicas do modo
	switch o.AppCfg.Cli.Mode {
	case "expose":
		fmt.Printf("  expose: Épocas=%d, TaxaAprendizadoBase=%.4f, CiclosPorPadrão=%d\n",
			o.AppCfg.Cli.Epochs, o.AppCfg.Cli.BaseLearningRate, o.AppCfg.Cli.CyclesPerPattern)
	case "observe":
		fmt.Printf("  observe: Dígito=%d, CiclosParaAcomodar=%d\n",
			o.AppCfg.Cli.Digit, o.AppCfg.Cli.CyclesToSettle)
	case "sim":
		fmt.Printf("  sim: TotalCiclos=%d, CaminhoDB='%s', IntervaloSaveDB=%d\n",
			o.AppCfg.Cli.Cycles, o.AppCfg.Cli.DbPath, o.AppCfg.Cli.SaveInterval)
		if o.AppCfg.Cli.StimInputFreqHz > 0 && o.AppCfg.Cli.StimInputID != -2 {
			fmt.Printf("  sim: EstímuloGeral: InputID=%d a %.1f Hz\n",
				o.AppCfg.Cli.StimInputID, o.AppCfg.Cli.StimInputFreqHz)
		}
	}

	// Inicializar logger SQLite se dbPath for fornecido e o modo o utilizar
	if o.AppCfg.Cli.DbPath != "" && (o.AppCfg.Cli.Mode == "sim" || (o.AppCfg.Cli.Mode == "expose" && o.AppCfg.Cli.SaveInterval > 0)) {
		var err error
		o.Logger, err = storage.NewSQLiteLogger(o.AppCfg.Cli.DbPath)
		if err != nil {
			log.Fatalf("Falha ao inicializar logger SQLite: %v", err)
		}
		defer func() {
			if errClose := o.Logger.Close(); errClose != nil {
				log.Printf("Erro ao fechar logger SQLite: %v", errClose)
			}
		}()
		fmt.Printf("Logging SQLite ativado para: %s\n", o.AppCfg.Cli.DbPath)
	}


	// Criar a rede
	// A configuração da CLI (como BaseLearningRate) já está no AppCfg que NewCrowNet recebe.
	o.Net = network.NewCrowNet(o.AppCfg)
	fmt.Printf("Rede criada: %d neurônios. IDs Input: %v..., IDs Output: %v...\n",
		len(o.Net.Neurons), o.Net.InputNeuronIDs[:min(5, len(o.Net.InputNeuronIDs))], o.Net.OutputNeuronIDs[:min(10, len(o.Net.OutputNeuronIDs))])
	fmt.Printf("Estado Inicial: Cortisol=%.3f, Dopamina=%.3f\n",
		o.Net.ChemicalEnv.CortisolLevel, o.Net.ChemicalEnv.DopamineLevel)


	// Executar o modo selecionado
	startTime := time.Now()
	switch o.AppCfg.Cli.Mode {
	case "sim":
		o.runSimMode()
	case "expose":
		o.runExposeMode()
	case "observe":
		o.runObserveMode()
	default:
		log.Fatalf("Modo desconhecido: %s. Escolha 'sim', 'expose', ou 'observe'.", o.AppCfg.Cli.Mode)
	}
	duration := time.Since(startTime)
	fmt.Printf("\nSessão CrowNet finalizada. Duração total: %s.\n", duration)
}

func (o *Orchestrator) runSimMode() {
	fmt.Printf("\nIniciando Simulação Geral por %d ciclos...\n", o.AppCfg.Cli.Cycles)

	// Configurar estímulo de input contínuo, se especificado
	if o.AppCfg.Cli.StimInputFreqHz > 0.0 && o.AppCfg.Cli.StimInputID != -2 && len(o.Net.InputNeuronIDs) > 0 {
		stimID := o.AppCfg.Cli.StimInputID
		if stimID == -1 && len(o.Net.InputNeuronIDs) > 0 { // -1 para primeiro disponível
			stimID = int(o.Net.InputNeuronIDs[0])
		}

		// Validar stimID (deve ser um InputNeuronID existente)
		isValidStimID := false
		for _, id := range o.Net.InputNeuronIDs {
			if int(id) == stimID {
				isValidStimID = true
				break
			}
		}
		if isValidStimID {
			// Esta função deve existir em `network.CrowNet` para configurar o estímulo.
			// o.Net.SetInputFrequency(common.NeuronID(stimID), o.AppCfg.Cli.StimInputFreqHz)
			// Por enquanto, vamos simular a configuração:
			o.Net.ConfigureFrequencyInput(common.NeuronID(stimID), o.AppCfg.Cli.StimInputFreqHz)

			fmt.Printf("Estímulo geral: Neurônio de Input %d a %.1f Hz.\n", stimID, o.AppCfg.Cli.StimInputFreqHz)
		} else {
			fmt.Printf("Aviso: ID do neurônio de input para estímulo geral (%d) não encontrado ou inválido.\n", stimID)
		}
	}

	o.Net.SetDynamicState(true, true, true) // Todas as dinâmicas ativas para 'sim'

	for i := 0; i < o.AppCfg.Cli.Cycles; i++ {
		o.Net.RunCycle()
		if i%10 == 0 || i == o.AppCfg.Cli.Cycles-1 {
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

	if o.Logger != nil && (o.AppCfg.Cli.SaveInterval == 0 || (o.AppCfg.Cli.Cycles > 0 && o.AppCfg.Cli.Cycles%o.AppCfg.Cli.SaveInterval != 0)) {
		if o.AppCfg.Cli.Cycles > 0 { // Apenas salvar se algum ciclo rodou
			if err := o.Logger.LogNetworkState(o.Net); err != nil {
				log.Printf("Aviso durante salvamento final no DB: %v", err)
			}
		}
	}

	// Reportar frequência do neurônio monitorado, se configurado
	if o.AppCfg.Cli.MonitorOutputID != -2 && len(o.Net.OutputNeuronIDs) > 0 {
		monitorID := o.AppCfg.Cli.MonitorOutputID
		if monitorID == -1 && len(o.Net.OutputNeuronIDs) > 0 { // -1 para primeiro disponível
			monitorID = int(o.Net.OutputNeuronIDs[0])
		}
		// freq, err := o.Net.GetOutputFrequency(common.NeuronID(monitorID))
		// Esta função precisa ser implementada em network.go
		// if err == nil {
		// 	fmt.Printf("Frequência para Neurônio de Output %d: %.2f Hz (sobre os últimos %.0f ciclos).\n",
		// 		monitorID, freq, o.AppCfg.SimParams.OutputFrequencyWindowCycles)
		// } else {
		// 	fmt.Printf("Aviso ao obter frequência para Neurônio de Output %d: %v\n", monitorID, err)
		// }
	}
	fmt.Printf("Estado Final: Cortisol=%.3f, Dopamina=%.3f\n", o.Net.ChemicalEnv.CortisolLevel, o.Net.ChemicalEnv.DopamineLevel)
}

func (o *Orchestrator) runExposeMode() {
	fmt.Printf("\nIniciando Fase de Exposição por %d épocas (TaxaAprendizadoBase: %.4f, CiclosPorPadrão: %d)...\n",
		o.AppCfg.Cli.Epochs, o.AppCfg.Cli.BaseLearningRate, o.AppCfg.Cli.CyclesPerPattern)

	// Tentar carregar pesos existentes
	if _, err := os.Stat(o.AppCfg.Cli.WeightsFile); err == nil {
		loadedWeights, errLoad := storage.LoadNetworkWeightsFromJSON(o.AppCfg.Cli.WeightsFile)
		if errLoad == nil {
			o.Net.SynapticWeights = loadedWeights
			fmt.Printf("Pesos existentes carregados de %s\n", o.AppCfg.Cli.WeightsFile)
		} else {
			fmt.Printf("Falha ao carregar pesos de %s (%v). Iniciando com pesos aleatórios iniciais.\n", o.AppCfg.Cli.WeightsFile, errLoad)
			// A rede já foi inicializada com pesos aleatórios se o carregamento falhar ou não for tentado.
		}
	} else {
		fmt.Printf("Arquivo de pesos %s não encontrado. Iniciando com pesos aleatórios iniciais.\n", o.AppCfg.Cli.WeightsFile)
	}

	o.Net.SetDynamicState(true, true, true) // Todas as dinâmicas ativas para 'expose'

	// (Opcional) Chamar setupDopamineStimulationForExpose se essa lógica for portada para network.CrowNet
	// o.Net.SetupDopamineStimulationForExpose()

	allPatterns, err := datagen.GetAllDigitPatterns(&o.AppCfg.SimParams)
	if err != nil {
		log.Fatalf("Falha ao carregar padrões de dígitos: %v", err)
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

			o.Net.ResetNetworkStateForNewPattern() // Limpa pulsos e potenciais
			if err := o.Net.PresentPattern(pattern); err != nil {
				log.Fatalf("Falha ao apresentar padrão para dígito %d na época %d: %v", digit, epoch+1, err)
			}

			for cycleInPattern := 0; cycleInPattern < o.AppCfg.Cli.CyclesPerPattern; cycleInPattern++ {
				o.Net.RunCycle()
				// Logging SQLite dentro do RunCycle se configurado
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

	fmt.Println("Fase de exposição concluída.")
	if err := storage.SaveNetworkWeightsToJSON(o.Net.SynapticWeights, o.AppCfg.Cli.WeightsFile); err != nil {
		log.Fatalf("Falha ao salvar pesos treinados em %s: %v", o.AppCfg.Cli.WeightsFile, err)
	} else {
		fmt.Printf("Pesos treinados salvos em %s\n", o.AppCfg.Cli.WeightsFile)
	}
}

func (o *Orchestrator) runObserveMode() {
	fmt.Printf("\nObservando Resposta da Rede para o dígito %d (%d ciclos de acomodação)...\n",
		o.AppCfg.Cli.Digit, o.AppCfg.Cli.CyclesToSettle)

	loadedWeights, err := storage.LoadNetworkWeightsFromJSON(o.AppCfg.Cli.WeightsFile)
	if err != nil {
		log.Fatalf("Falha ao carregar pesos de %s para observação: %v. Exponha a rede primeiro.", o.AppCfg.Cli.WeightsFile, err)
	}
	o.Net.SynapticWeights = loadedWeights
	fmt.Printf("Pesos carregados de %s para observação.\n", o.AppCfg.Cli.WeightsFile)

	// Desabilitar dinâmicas que alteram a rede ou introduzem variabilidade desnecessária para observação
	o.Net.SetDynamicState(false, false, false)

	patternToObserve, err := datagen.GetDigitPattern(o.AppCfg.Cli.Digit, &o.AppCfg.SimParams)
	if err != nil {
		log.Fatalf("Falha ao obter padrão para o dígito %d: %v", o.AppCfg.Cli.Digit, err)
	}

	o.Net.ResetNetworkStateForNewPattern()
	if err := o.Net.PresentPattern(patternToObserve); err != nil {
		log.Fatalf("Falha ao apresentar padrão para observação: %v", err)
	}

	for i := 0; i < o.AppCfg.Cli.CyclesToSettle; i++ {
		o.Net.RunCycle()
	}

	outputActivation, err := o.Net.GetOutputActivation()
	if err != nil {
		log.Fatalf("Falha ao obter ativação de saída: %v", err)
	}

	fmt.Printf("Dígito Apresentado: %d\n", o.AppCfg.Cli.Digit)
	fmt.Println("Padrão de Ativação dos Neurônios de Saída (Potencial Acumulado):")
	for i, val := range outputActivation {
		neuronIDStr := "N/A"
		if i < len(o.Net.OutputNeuronIDs) {
			neuronIDStr = fmt.Sprintf("%d", o.Net.OutputNeuronIDs[i])
		}
		fmt.Printf("  OutNeurônio[%d] (ID %s): %.4f\n", i, neuronIDStr, val)
	}

	// Restaurar estado dinâmico padrão se necessário (embora a aplicação termine aqui)
	o.Net.SetDynamicState(true, true, true)
}

// Helper para evitar pânico com slices vazias
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Adicionar a network.CrowNet:
// ConfigureFrequencyInput(id common.NeuronID, hz float64)
// GetOutputFrequency(id common.NeuronID) (float64, error)
// SetupDopamineStimulationForExpose() (opcional, ou lógica similar)
```
