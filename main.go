package main

import (
	"crownet/datagen"
	"crownet/network"
	"crownet/neuron"
	"crownet/storage"
	"flag"
	"fmt"
	"log"
	// "math"
)

var (
	totalNeuronsOpt     *int
	cyclesOpt           *int
	dbPathOpt           *string
	saveIntervalOpt     *int
	stimInputIDOpt      *int
	stimInputFreqHzOpt  *float64
	monitorOutputIDOpt  *int
	debugChemOpt        *bool
	modeOpt             *string
	epochsOpt           *int
	weightsFileOpt      *string
	digitOpt            *int
	baseLearningRateOpt *float64
	cyclesPerPatternOpt *int
	cyclesToSettleOpt   *int
)

func main() {
	totalNeuronsOpt = flag.Int("neurons", 100, "Total number of neurons in the network.")
	cyclesOpt = flag.Int("cycles", 100, "Total simulation cycles for 'sim' mode.")
	stimInputIDOpt = flag.Int("stimInputID", -1, "ID of an input neuron for general continuous stimulus in 'sim' mode (-1 for first available, -2 to disable).")
	stimInputFreqHzOpt = flag.Float64("stimInputFreqHz", 0.0, "Frequency (Hz) for general stimulus in 'sim' mode (0.0 to disable).")
	monitorOutputIDOpt = flag.Int("monitorOutputID", -1, "ID of an output neuron to monitor for frequency reporting in 'sim' mode (-1 for first available).")
	dbPathOpt = flag.String("dbPath", "crownet_data.db", "Path for the SQLite database file.")
	saveIntervalOpt = flag.Int("saveInterval", 0, "Cycle interval for saving to DB (0 to disable periodic saves, only final if any).")
	debugChemOpt = flag.Bool("debugChem", false, "Enable debug prints for chemical production.")

	modeOpt = flag.String("mode", "sim", "Operation mode: 'sim', 'expose', or 'observe'.")
	epochsOpt = flag.Int("epochs", 50, "Number of exposure epochs (for 'expose' mode).") // Default lowered for faster iteration
	weightsFileOpt = flag.String("weightsFile", "crownet_digit_weights.json", "File to save/load synaptic weights.")
	digitOpt = flag.Int("digit", 0, "Digit (0-9) to present (for 'observe' mode).")
	baseLearningRateOpt = flag.Float64("lrBase", 0.005, "Base learning rate for Hebbian plasticity.") // Defaulted from previous run
	cyclesPerPatternOpt = flag.Int("cyclesPerPattern", 5, "Number of cycles to run per pattern presentation during 'expose' mode.")
	cyclesToSettleOpt = flag.Int("cyclesToSettle", 5, "Number of cycles to run for network settling during 'observe' mode.")

	flag.Parse()

	fmt.Println("CrowNet Initializing...")
	fmt.Printf("Selected Mode: %s\n", *modeOpt)
	fmt.Printf("Base Configuration: Neurons=%d, WeightsFile='%s'\n",
		*totalNeuronsOpt, *weightsFileOpt)

	if *modeOpt == "expose" {
		fmt.Printf("  expose: Epochs=%d, BaseLR=%.4f, CyclesPerPattern=%d\n", *epochsOpt, *baseLearningRateOpt, *cyclesPerPatternOpt)
	} else if *modeOpt == "observe" {
		fmt.Printf("  observe: Digit=%d, CyclesToSettle=%d\n", *digitOpt, *cyclesToSettleOpt)
	} else {
		fmt.Printf("  sim: TotalCycles=%d, DBPath='%s', SaveInterval=%d\n", *cyclesOpt, *dbPathOpt, *saveIntervalOpt)
		if *stimInputFreqHzOpt > 0 && *stimInputIDOpt != -2 {
			fmt.Printf("  sim: GeneralStimulus: InputID=%d at %.1f Hz\n", *stimInputIDOpt, *stimInputFreqHzOpt)
		}
	}

	if *debugChemOpt {
		network.DebugCortisolHit = true
	}

	if *modeOpt == "sim" || (*modeOpt == "expose" && *saveIntervalOpt > 0) {
		if err := storage.InitDB(*dbPathOpt); err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
		defer storage.CloseDB()
	}

	net := network.NewCrowNet(*totalNeuronsOpt)
	net.BaseLearningRate = *baseLearningRateOpt

	fmt.Printf("Network created: %d neurons. Input IDs: %s..., Output IDs: %s...\n", len(net.Neurons), net.InputNeuronIDs_MVP_Preview(5), net.OutputNeuronIDs_MVP_Preview(10))
	fmt.Printf("Initial State: Cortisol=%.3f, Dopamine=%.3f\n", net.CortisolLevel, net.DopamineLevel)

	switch *modeOpt {
	case "sim":
		runGeneralSimulation(net)
	case "expose":
		setupDopamineStimulationForExpose(net) // Add specific setup for "expose"
		runExposure(net)
	case "observe":
		runObservation(net)
	default:
		log.Fatalf("Unknown mode: %s. Choose 'sim', 'expose', or 'observe'.", *modeOpt)
	}

	fmt.Println("\nCrowNet session finished.")
}

// setupDopamineStimulationForExpose is a helper to ensure dopaminergic neurons might fire during exposure
func setupDopamineStimulationForExpose(net *network.CrowNet) {
	fmt.Println("[SETUP-EXPOSE] Attempting to set up dopamine stimulation...")
	var dopaNeuron *neuron.Neuron
	for _, n := range net.Neurons {
		if n.Type == neuron.DopaminergicNeuron {
			dopaNeuron = n
			break
		}
	}

	if dopaNeuron == nil {
		fmt.Println("[SETUP-EXPOSE] No dopaminergic neurons found.")
		return
	}
	fmt.Printf("[SETUP-EXPOSE] Found DopaminergicNeuron ID %d.\n", dopaNeuron.ID)

	var stimulatorNeuron *neuron.Neuron
	// Find an input neuron, preferably not one of the first 35 (pattern neurons)
	// and not the dopa neuron itself.
	for i := len(net.InputNeuronIDs) - 1; i >= 0; i-- { // Iterate backwards to pick higher ID input neurons first
		inputID := net.InputNeuronIDs[i]
		if inputID != dopaNeuron.ID && inputID >= datagen.PatternSize { // Ensure it's beyond pattern neurons
			for _, n := range net.Neurons {
				if n.ID == inputID {
					stimulatorNeuron = n
					break
				}
			}
			if stimulatorNeuron != nil {
				break
			}
		}
	}
	// If no such neuron, try any input neuron not the dopa neuron
	if stimulatorNeuron == nil {
		for _, inputID := range net.InputNeuronIDs {
			if inputID != dopaNeuron.ID {
				for _, n := range net.Neurons {
					if n.ID == inputID {
						stimulatorNeuron = n
						break
					}
				}
				if stimulatorNeuron != nil {
					break
				}
			}
		}
	}

	if stimulatorNeuron != nil {
		fmt.Printf("[SETUP-EXPOSE] Using InputNeuron %d to stimulate DopaminergicNeuron %d.\n", stimulatorNeuron.ID, dopaNeuron.ID)
		// Position stimulator very close to the dopaminergic neuron
		for k := 0; k < 16; k++ {
			stimulatorNeuron.Position[k] = dopaNeuron.Position[k]
			if k == 0 {
				stimulatorNeuron.Position[k] += 0.01
			} // Offset slightly
		}
		// Set stimulator to fire very frequently
		stimFreq := network.CyclesPerSecond * 2 // e.g., 20Hz
		// Boost the weight from stimulator to dopaNeuron for reliable firing in test
		net.SetWeight(stimulatorNeuron.ID, dopaNeuron.ID, 1.0) // Strong excitatory connection
		fmt.Printf("[SETUP-EXPOSE] Set weight from Input %d to Dopa %d to 1.0\n", stimulatorNeuron.ID, dopaNeuron.ID)

		err := net.SetInputFrequency(stimulatorNeuron.ID, stimFreq)
		if err != nil {
			fmt.Printf("[SETUP-EXPOSE] Error setting high frequency for dopamine test stimulus neuron %d: %v\n", stimulatorNeuron.ID, err)
		} else {
			fmt.Printf("[SETUP-EXPOSE] Set neuron %d (Input for Dopamine Test) to fire at %.1f Hz.\n", stimulatorNeuron.ID, stimFreq)
		}
	} else {
		fmt.Println("[SETUP-EXPOSE] Could not find a suitable input neuron to stimulate dopaminergic neuron.")
	}
}

func runGeneralSimulation(net *network.CrowNet) {
	fmt.Printf("\nRunning General Simulation for %d cycles...\n", *cyclesOpt)
	if *stimInputFreqHzOpt > 0.0 && *stimInputIDOpt != -2 && len(net.InputNeuronIDs) > 0 {
		stimID := *stimInputIDOpt
		if stimID == -1 {
			stimID = net.InputNeuronIDs[0]
		}

		isValidStimID := false
		for _, id := range net.InputNeuronIDs {
			if id == stimID {
				isValidStimID = true
				break
			}
		}

		if isValidStimID {
			net.SetInputFrequency(stimID, *stimInputFreqHzOpt)
			fmt.Printf("General stimulus: Input Neuron %d at %.1f Hz.\n", stimID, *stimInputFreqHzOpt)
		} else {
			fmt.Printf("Warning: General stimulus input neuron ID %d not found or invalid.\n", stimID)
		}
	}

	for i := 0; i < *cyclesOpt; i++ {
		net.RunCycle()
		if i%10 == 0 || i == *cyclesOpt-1 {
			fmt.Printf("Cycle %d/%d: C:%.3f D:%.3f SynModF:%.3f Pulses:%d\n",
				net.CycleCount-1, *cyclesOpt, net.CortisolLevel, net.DopamineLevel, net.GetSynaptogenesisModulationFactor(), len(net.ActivePulses))
		}
		if *saveIntervalOpt > 0 && net.CycleCount > 0 && net.CycleCount%*saveIntervalOpt == 0 {
			if err := storage.SaveNetworkState(net); err != nil {
				log.Printf("Warning during periodic save: %v", err)
			}
		}
	}
	if *saveIntervalOpt == 0 || (*cyclesOpt > 0 && *cyclesOpt%*saveIntervalOpt != 0) {
		if *cyclesOpt > 0 {
			if err := storage.SaveNetworkState(net); err != nil {
				log.Printf("Warning during final save: %v", err)
			}
		}
	}

	monitoredID := -1
	if len(net.OutputNeuronIDs) > 0 {
		if *monitorOutputIDOpt == -1 {
			monitoredID = net.OutputNeuronIDs[0]
		} else {
			for _, id := range net.OutputNeuronIDs {
				if id == *monitorOutputIDOpt {
					monitoredID = id
					break
				}
			}
		}
	}
	if monitoredID != -1 {
		freq, _ := net.GetOutputFrequency(monitoredID)
		fmt.Printf("Frequency for Output Neuron %d: %.2f Hz (over last %.0f cycles).\n", monitoredID, freq, network.OutputFrequencyWindowCycles)
	}
	fmt.Printf("Final State: Cortisol=%.3f, Dopamine=%.3f\n", net.CortisolLevel, net.DopamineLevel)
}

func runExposure(net *network.CrowNet) {
	fmt.Printf("\nStarting Exposure Phase for %d epochs (BaseLR: %.4f, CyclesPerPattern: %d)...\n", *epochsOpt, net.BaseLearningRate, *cyclesPerPatternOpt)

	if err := net.LoadWeights(*weightsFileOpt); err == nil {
		fmt.Printf("Loaded existing weights from %s\n", *weightsFileOpt)
	} else {
		fmt.Printf("Could not load weights from %s (%v). Starting with initial random weights.\n", *weightsFileOpt, err)
	}

	net.EnableSynaptogenesis = true
	net.EnableChemicalModulation = true

	err := net.ExposeToPatterns(datagen.GetDigitPattern, 10, *epochsOpt, *cyclesPerPatternOpt)
	if err != nil {
		log.Fatalf("Exposure phase failed: %v", err)
	}
	fmt.Println("Exposure phase completed.")

	if err := net.SaveWeights(*weightsFileOpt); err != nil {
		log.Fatalf("Failed to save weights to %s: %v", *weightsFileOpt, err)
	} else {
		fmt.Printf("Saved trained weights to %s\n", *weightsFileOpt)
	}
}

func runObservation(net *network.CrowNet) {
	fmt.Printf("\nObserving Network Response for digit %d (%d settle cycles)...\n", *digitOpt, *cyclesToSettleOpt)

	if err := net.LoadWeights(*weightsFileOpt); err != nil {
		log.Fatalf("Failed to load weights from %s for observation: %v. Expose the network first.", *weightsFileOpt, err)
	}
	fmt.Printf("Loaded weights from %s for observation.\n", *weightsFileOpt)

	pattern, err := datagen.GetDigitPattern(*digitOpt)
	if err != nil {
		log.Fatalf("Failed to get pattern for digit %d: %v", *digitOpt, err)
	}

	outputPattern, err := net.GetOutputPatternForInput(pattern, *cyclesToSettleOpt)
	if err != nil {
		log.Fatalf("Observation failed: %v", err)
	}
	fmt.Printf("Presented Digit: %d\n", *digitOpt)
	fmt.Printf("Output Neuron Activation Pattern (AccumulatedPulse):\n")
	for i, val := range outputPattern {
		if i < len(net.OutputNeuronIDs) { // Ensure we don't go out of bounds if outputPattern is shorter
			fmt.Printf("  OutNeuron[%d] (ID %d): %.4f\n", i, net.OutputNeuronIDs[i], val)
		} else {
			fmt.Printf("  OutPattern[%d]: %.4f (No corresponding OutputNeuronID index)\n", i, val)
		}
	}
}
