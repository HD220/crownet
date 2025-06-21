package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"
	"math"
	"math/rand"

	"github.com/user/crownet/src/core"
	"github.com/user/crownet/src/mnist"
	"github.com/user/crownet/src/storage"
)

// TrainingConfig define os parâmetros para o processo de treinamento.
type TrainingConfig struct {
	Epochs         int
	LearningRate   float64
	BatchSize      int
	NumInputNeurons  int
	NumOutputNeurons int
	InputEncodingStrength float64
	RewardAmountPositive float64
	RewardAmountNegative float64
}

// DefaultTrainingConfig retorna uma configuração de treinamento padrão.
func DefaultTrainingConfig(netCfg core.NetworkConfig) TrainingConfig {
	numInput := 0
	numOutput := 0
	for nType, percentage := range netCfg.NeuronDistribution {
		count := int(float64(netCfg.NumNeurons) * percentage)
		if nType == core.InputNeuron {
			numInput = count
		} else if nType == core.OutputNeuron {
			numOutput = count
		}
	}
	return TrainingConfig{
		Epochs:         10,
		LearningRate:   0.01,
		BatchSize:      1,
		NumInputNeurons:  numInput,
		NumOutputNeurons: numOutput,
		InputEncodingStrength: 1.5,
		RewardAmountPositive: 0.5,
		RewardAmountNegative: -0.3,
	}
}

func encodeInput(normalizedPixels []float64, numInputNeurons int, strength float64) []float64 {
	inputSignal := make([]float64, numInputNeurons)
	if numInputNeurons == 0 { return inputSignal }
	numPixels := len(normalizedPixels)
	if numPixels == 0 { return inputSignal }
	if numInputNeurons >= numPixels {
		for i := 0; i < numPixels; i++ { inputSignal[i] = normalizedPixels[i] * strength }
	} else {
		pixelsPerNeuron := float64(numPixels) / float64(numInputNeurons)
		for i := 0; i < numInputNeurons; i++ {
			startPixel := int(math.Floor(float64(i) * pixelsPerNeuron))
			endPixel := int(math.Floor(float64(i+1) * pixelsPerNeuron))
			if endPixel > numPixels { endPixel = numPixels }
			if startPixel >= endPixel {
				if startPixel < numPixels { inputSignal[i] = normalizedPixels[startPixel] * strength }
				continue
			}
			sum := 0.0
			count := 0
			for j := startPixel; j < endPixel; j++ { sum += normalizedPixels[j]; count++ }
			if count > 0 { inputSignal[i] = (sum / float64(count)) * strength }
		}
	}
	return inputSignal
}

func decodeOutput(outputActivations []float64) (predictedLabel int) {
	if len(outputActivations) == 0 { return -1 }
	maxActivation := -1.0
	predictedLabel = -1
	numClassesToCheck := len(outputActivations)
	if numClassesToCheck > 10 { numClassesToCheck = 10 }
	for i := 0; i < numClassesToCheck; i++ {
		activation := outputActivations[i]
		if activation > maxActivation { maxActivation = activation; predictedLabel = i }
	}
	return predictedLabel
}

func Train(network *core.Network, dataset *mnist.Dataset, trainCfg TrainingConfig, netCfg core.NetworkConfig, modelPath string) error {
	if network == nil || dataset == nil { return fmt.Errorf("network or dataset is nil") }
	if trainCfg.NumOutputNeurons < 10 && netCfg.NumNeurons > 0 {
		fmt.Printf("Warning: MNIST requires at least 10 output neurons. Configured with %d.\n", trainCfg.NumOutputNeurons)
	}
	shuffleRand := rand.New(rand.NewSource(time.Now().UnixNano()))
	fmt.Printf("Starting training...\n  Epochs: %d\n", trainCfg.Epochs)
	for epoch := 0; epoch < trainCfg.Epochs; epoch++ {
		fmt.Printf("Epoch %d/%d\n", epoch+1, trainCfg.Epochs)
		shuffledTrainImages := make([]mnist.Image, len(dataset.TrainImages))
		copy(shuffledTrainImages, dataset.TrainImages)
		shuffleRand.Shuffle(len(shuffledTrainImages), func(i, j int) {
			shuffledTrainImages[i], shuffledTrainImages[j] = shuffledTrainImages[j], shuffledTrainImages[i]
		})
		correctPredictionsEpoch, totalProcessedEpoch, batchCorrect, batchTotal := 0, 0, 0, 0
		for i, image := range shuffledTrainImages {
			network.ResetNetworkState()
			normalizedPixels := mnist.NormalizePixels(image.Pixels)
			inputSignal := encodeInput(normalizedPixels, trainCfg.NumInputNeurons, trainCfg.InputEncodingStrength)
			network.SetInput(inputSignal)
			for cycle := 0; cycle < netCfg.MaxCycles; cycle++ { network.SimulateCycle() }
			outputActivations := network.GetOutput()
			predictedLabel := decodeOutput(outputActivations)
			trueLabel := int(image.Label)
			reward := 0.0
			if predictedLabel == trueLabel {
				correctPredictionsEpoch++; batchCorrect++; reward = trainCfg.RewardAmountPositive
			} else {
				reward = trainCfg.RewardAmountNegative
			}
			core.ApplyRewardSignal(network, reward)
			totalProcessedEpoch++; batchTotal++
			if (i+1)%1000 == 0 {
				accuracyBatch := 0.0
				if batchTotal > 0 { accuracyBatch = float64(batchCorrect) / float64(batchTotal) }
				fmt.Printf("  Epoch %d, Image %d/%d: Acc (batch): %.2f%% (P:%d, T:%d, R:%.2f, C:%.3f, D:%.3f)\n",
					epoch+1, i+1, len(shuffledTrainImages), accuracyBatch*100,
					predictedLabel, trueLabel, reward, network.CortisolLevel, network.DopamineLevel)
				batchCorrect, batchTotal = 0, 0
			}
		}
		epochAccuracy := 0.0
		if totalProcessedEpoch > 0 { epochAccuracy = float64(correctPredictionsEpoch) / float64(totalProcessedEpoch) }
		fmt.Printf("Epoch %d finished. Training Accuracy: %.2f%%\n", epoch+1, epochAccuracy*100)
		if modelPath != "" {
			fmt.Printf("Saving model to %s...\n", modelPath)
			if err := storage.SaveNetwork(network, modelPath); err != nil {
				fmt.Printf("Error saving model: %v\n", err)
			}
		}
	}
	fmt.Println("Training complete.")
	return nil
}

func Evaluate(network *core.Network, dataset *mnist.Dataset, trainCfg TrainingConfig, netCfg core.NetworkConfig) (float64, error) {
	if network == nil || dataset == nil { return 0, fmt.Errorf("network or dataset is nil") }
	fmt.Println("Starting evaluation on test set...")
	correctPredictions := 0
	for i, image := range dataset.TestImages {
		network.ResetNetworkState()
		normalizedPixels := mnist.NormalizePixels(image.Pixels)
		inputSignal := encodeInput(normalizedPixels, trainCfg.NumInputNeurons, trainCfg.InputEncodingStrength)
		network.SetInput(inputSignal)
		for cycle := 0; cycle < netCfg.MaxCycles; cycle++ { network.SimulateCycle() }
		outputActivations := network.GetOutput()
		predictedLabel := decodeOutput(outputActivations)
		if predictedLabel == int(image.Label) { correctPredictions++ }
		if (i+1)%1000 == 0 { fmt.Printf("  Evaluated %d/%d images...\n", i+1, len(dataset.TestImages)) }
	}
	accuracy := 0.0
	if len(dataset.TestImages) > 0 { accuracy = float64(correctPredictions) / float64(len(dataset.TestImages)) }
	fmt.Printf("Evaluation complete. Test Accuracy: %.2f%%\n", accuracy*100)
	return accuracy, nil
}


func main() {
	// Definir subcomandos: train, evaluate
	if len(os.Args) < 2 {
		printTopLevelHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "train":
		handleTrainCommand()
	case "evaluate":
		handleEvaluateCommand()
	case "help":
		printTopLevelHelp()
	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printTopLevelHelp()
		os.Exit(1)
	}
}

func printTopLevelHelp() {
	fmt.Println("CrowNet MVP - Neural Network Simulator")
	fmt.Println("Usage: crownet-mvp <command> [options]")
	fmt.Println("\nCommands:")
	fmt.Println("  train      Train a new network or continue training an existing one.")
	fmt.Println("  evaluate   Evaluate a trained network.")
	fmt.Println("  help       Show this help message.")
	fmt.Println("\nUse 'crownet-mvp <command> -help' for more information on a specific command.")
}

func handleTrainCommand() {
	trainCmd := flag.NewFlagSet("train", flag.ExitOnError)
	dataDir := trainCmd.String("data", "./data/mnist_data", "Directory for MNIST dataset")
	modelPath := trainCmd.String("model", "crownet_model.db", "Path to save the trained model")
	loadModelPath := trainCmd.String("load", "", "Path to load an existing model to continue training")

	defaultNetCfg := core.DefaultNetworkConfig()
	numNeurons := trainCmd.Int("neurons", defaultNetCfg.NumNeurons, "Total number of neurons in the network")
	spaceSize := trainCmd.Float64("space", defaultNetCfg.SpaceSize, "Size of the 16D space for neurons")
	pulseSpeed := trainCmd.Float64("pulse-speed", defaultNetCfg.PulsePropagationSpeed, "Pulse propagation speed (units/cycle)")
	netMaxCycles := trainCmd.Int("net-cycles", defaultNetCfg.MaxCycles, "Max simulation cycles per input presentation")
	randSeed := trainCmd.Int64("seed", time.Now().UnixNano(), "Random seed for network initialization")

	tempNetCfgForTrainCfg := core.DefaultNetworkConfig()
	tempNetCfgForTrainCfg.NumNeurons = *numNeurons

	defaultTrainCfg := DefaultTrainingConfig(tempNetCfgForTrainCfg)
	epochs := trainCmd.Int("epochs", defaultTrainCfg.Epochs, "Number of training epochs")
	inputStrength := trainCmd.Float64("input-strength", defaultTrainCfg.InputEncodingStrength, "Strength of input signal encoding")
	rewardPos := trainCmd.Float64("reward-pos", defaultTrainCfg.RewardAmountPositive, "Positive reward amount")
	rewardNeg := trainCmd.Float64("reward-neg", defaultTrainCfg.RewardAmountNegative, "Negative reward amount (e.g. -0.3)")

	trainCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s train [options]:\n", os.Args[0])
	trainCmd.PrintDefaults()
	}
	trainCmd.Parse(os.Args[2:])

	fmt.Println("Starting training process...")

	// mnistData, err := mnist.Load(*dataDir) // Dependências não estarão disponíveis ainda
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Error loading MNIST dataset from %s: %v\n", *dataDir, err)
	// 	os.Exit(1)
	// }

	var network *core.Network
	var netCfg core.NetworkConfig

	// Como os pacotes core, mnist, storage ainda não foram criados neste fluxo de reset,
	// as chamadas a eles causarão erro de compilação.
	// Vou comentar as partes que dependem desses pacotes para que o main.go seja sintaticamente válido
	// e possamos testar a criação do arquivo e da estrutura de diretórios.
	// Depois que confirmarmos que os arquivos são criados, eu os preencherei com o código completo.

	if *loadModelPath != "" {
		fmt.Printf("Loading existing model from %s to continue training...\n", *loadModelPath)
		// loadedNet, err := storage.LoadNetwork(*loadModelPath)
		// if err != nil {
		// 	fmt.Fprintf(os.Stderr, "Error loading model from %s: %v\n", *loadModelPath, err)
		// 	os.Exit(1)
		// }
		// network = loadedNet
		// netCfg = network.Config
		fmt.Println("Model loading will be fully functional after core/storage are created.")
		netCfg = core.DefaultNetworkConfig() // Fallback
		network = core.NewNetwork(netCfg) // core. Pelo menos este pode existir se criarmos neuron.go
										  // No entanto, DefaultNetworkConfig e NewNetwork são de `core`.

	} else {
		fmt.Println("Initializing new network...")
		// netCfg = core.DefaultNetworkConfig() // Precisa do pacote core
		netCfg.NumNeurons = *numNeurons
		netCfg.SpaceSize = *spaceSize
		netCfg.PulsePropagationSpeed = *pulseSpeed
		netCfg.MaxCycles = *netMaxCycles
		netCfg.RandomSeed = *randSeed

		// network = core.NewNetwork(netCfg) // Precisa do pacote core
		fmt.Printf("New network initialized with %d neurons (Seed: %d).\n", netCfg.NumNeurons, netCfg.RandomSeed)
	}

	// trainCfg := DefaultTrainingConfig(netCfg) // Precisa do pacote core em netCfg
	trainCfg.Epochs = *epochs
	trainCfg.InputEncodingStrength = *inputStrength
	trainCfg.RewardAmountPositive = *rewardPos
	trainCfg.RewardAmountNegative = *rewardNeg

	outputModelDir := filepath.Dir(*modelPath)
	if err := os.MkdirAll(outputModelDir, os.ModePerm); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating directory for output model %s: %v\n", outputModelDir, err)
		os.Exit(1)
	}

	// err = Train(network, mnistData, trainCfg, netCfg, *modelPath)
	// if err != nil {
	// 	fmt.Fprintf(os.Stderr, "Training failed: %v\n", err)
	// 	os.Exit(1)
	// }
	fmt.Println("Training function call skipped as dependencies (mnist, storage, core) are not yet fully created/linked.")
	fmt.Println("Simulating a successful training run for placeholder purposes.")
	if *modelPath != "" {
		fmt.Printf("Simulating saving model to %s...\n", *modelPath)
		f, err := os.Create(*modelPath + ".dummy")
		if err != nil {
			fmt.Printf("Error creating dummy model file: %v\n", err)
		} else {
			f.Close()
			os.Remove(*modelPath + ".dummy")
		}
	}
	fmt.Println("Training finished (simulated).")
}

func handleEvaluateCommand() {
	evalCmd := flag.NewFlagSet("evaluate", flag.ExitOnError)
	// dataDir := evalCmd.String("data", "./data/mnist_data", "Directory for MNIST dataset") // Duplicado
	// modelPath := evalCmd.String("model", "crownet_model.db", "Path to load the trained model for evaluation") // Duplicado
	_ = evalCmd.String("data", "./data/mnist_data", "Directory for MNIST dataset") // Evitar erro de não usado
    _ = evalCmd.String("model", "crownet_model.db", "Path to load the trained model for evaluation")


	evalCmd.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s evaluate [options]:\n", os.Args[0])
	evalCmd.PrintDefaults()
	}
	evalCmd.Parse(os.Args[2:])

	// if *modelPath == "" { // modelPath não está no escopo aqui se re-declarado dentro do if
	// 	fmt.Fprintln(os.Stderr, "Error: -model flag is required for evaluation.")
	// 	evalCmd.Usage()
	// 	os.Exit(1)
	// }
	// fmt.Printf("Starting evaluation process for model: %s\n", *modelPath)
	fmt.Println("Evaluation function call skipped as dependencies are not yet fully created/linked.")
	fmt.Println("Simulating successful evaluation for placeholder purposes. Accuracy: 10.00%")
}
