# Technical Design Document: CrowNet - Neuromodulated Self-Learning MVP

## 1. Introduction

This document provides a technical overview of the CrowNet MVP, which has been pivoted to focus on demonstrating **neuromodulated Hebbian self-learning**. The system aims to show how a network, inspired by CrowNet's original concepts (16D space, neuron types), can self-organize to differentiate simple digit patterns (0-9) through activity-dependent plasticity modulated by simulated neurochemicals (cortisol and dopamine).

## 2. Architecture

The system remains a command-line application in Go, with the following package structure:

*   **`main`**: Entry point, CLI flag parsing, mode orchestration ("expose", "observe", "sim").
*   **`datagen`**: Provides predefined 5x7 binary patterns for digits 0-9.
*   **`network`**: Core `CrowNet` struct and its methods. This includes:
    *   Initialization, neuron distribution.
    *   `RunCycle`: The main simulation step orchestrator.
    *   Pulse mechanics (`Pulse` struct, propagation).
    *   Synaptogenesis (`applySynaptogenesis`).
    *   Chemical modulation (`updateCortisolLevel`, `applyCortisolEffects`, etc.).
    *   **Neuromodulated Hebbian Learning (`ApplyHebbianPlasticity`)**.
    *   Input/Output (`PresentPattern`, `GetOutputPatternForInput`, frequency I/O for "sim" mode).
    *   Synaptic weight management (`SynapticWeights` map, `InitializeSynapticWeights`, `SaveWeights`, `LoadWeights`).
*   **`neuron`**: `Neuron` struct, types, states, and basic functions (pulse accumulation, firing logic, `GetBasePulseSign`).
*   **`utils`**: Utility functions (e.g., `EuclideanDistance`).
*   **`storage`**: SQLite interaction for saving full simulation snapshots (primarily for "sim" mode or detailed analysis).

## 3. Key Data Structures

*   **`neuron.Point`, `neuron.Neuron`, `network.Pulse`**: Largely unchanged from previous design, but `Neuron.GetEffectivePulseValue` is replaced by `GetBasePulseSign` (+1.0/-1.0).
*   **`network.CrowNet`**:
    *   `SynapticWeights map[int]map[int]float64`: Stores `[from_neuron_ID][to_neuron_ID] -> weight`. Crucial for learned associations.
    *   `EnableSynaptogenesis`, `EnableChemicalModulation`: Boolean flags, always `true` for the self-learning MVP as these dynamics are part of the learning environment. (Previous plan to disable them for train/classify is revised; they now modulate learning).
    *   Other fields (Neurons, ActivePulses, ChemicalLevels, I/O maps for frequency input) remain.
*   **`datagen.DigitPatterns`**: Stores the 5x7 patterns.

## 4. Core Algorithms

### 4.1. Simulation Cycle (`CrowNet.RunCycle`)
The order of operations per cycle is critical for interaction between dynamics:
1.  **Input Processing (`processInputs`):** Handles *continuous, frequency-based* input for "sim" mode. For "expose" and "observe" modes, pattern presentation is done via `PresentPattern` *before* starting a series of `RunCycle` calls.
2.  **Neuron Updates:**
    *   `DecayPulseAccumulation()` for all neurons.
    *   `UpdateState()` for all neurons (refractory periods, etc.).
3.  **Pulse Propagation & Effects:**
    *   Pulses propagate, and their `CurrentDistance` is updated.
    *   For neurons hit by a pulse `p` (from `emitter_ID` to `receiver_ID`):
        *   `weight = cn.GetWeight(emitter_ID, receiver_ID)`.
        *   `effectiveValue = p.Value * weight` (where `p.Value` is base sign +/-1.0 from emitter).
        *   `receiver.ReceivePulse(effectiveValue)`.
        *   If `receiver` fires, a new pulse with `receiver.GetBasePulseSign()` is created. Output neuron firings are recorded (for frequency calculation in "sim" mode, or for potential alternative output metrics).
4.  **Neuromodulated Hebbian Plasticity (`ApplyHebbianPlasticity`):**
    *   Calculates `effectiveLearningRate = BaseLearningRate * f(DopamineLevel) * g(CortisolLevel)`.
        *   `f(DopamineLevel)`: Increases with dopamine (e.g., `1 + DopamineEffectOnLR * DopamineLevel`).
        *   `g(CortisolLevel)`: Decreases with high cortisol (e.g., `1 / (1 + CortisolEffectOnLR * HighCortisolConcentration)` or similar).
    *   Iterates through neuron pairs (or connections with non-zero weights).
    *   If pre-synaptic neuron `i` and post-synaptic neuron `j` were recently co-active (e.g., both fired in the current or previous cycle, or `i` fired and `j`'s `AccumulatedPulse` is high):
        *   `Δw_ij = effectiveLearningRate * activity_i * activity_j`. (`activity` could be 1 if fired, or normalized firing rate/pulse sum).
        *   `cn.SynapticWeights[i][j] += Δw_ij`.
        *   Apply weight bounds/normalization (e.g., clip weights to a max/min range like [-1, 1]).
5.  **Chemical Modulation (Updates levels and applies effects on thresholds):**
    *   `updateCortisolLevel()`: Based on excitatory pulses near gland.
    *   `updateDopamineLevel()`: Based on firing of dopaminergic neurons.
    *   `applyCortisolEffects()`: Modifies neuron firing thresholds and `currentSynaptogenesisModulationFactor`.
    *   `applyDopamineEffects()`: Further modifies thresholds and `currentSynaptogenesisModulationFactor`.
6.  **Synaptogenesis (`applySynaptogenesis`):** Neuron movement occurs, modulated by the (potentially chemically-altered) `currentSynaptogenesisModulationFactor`.
7.  Increment `CycleCount`.

### 4.2. Input Encoding for Digits
*   `datagen.GetDigitPattern()` provides a 35-element `[]float64` vector (0.0 or 1.0).
*   `CrowNet.PresentPattern(pattern)`: For each `1.0` in the pattern, the corresponding input neuron (from the first 35 in `InputNeuronIDs`) is made to fire once by setting its state to `FiringState` and creating a pulse with its `BaseSignal`.

### 4.3. Output Interpretation for Digits
*   The 10 designated `OutputNeuronIDs` provide the output.
*   After presenting an input pattern and running the network for `ClassificationCycles` (a small number of `RunCycle` calls, e.g., 3-5, to let activity settle), the vector of `AccumulatedPulse` values (or another activity measure like short-term firing rate) of these 10 neurons is taken as the network's representation of the input.
*   The MVP goal is for these 10-element vectors to be distinct for different input digits.

### 4.4. "Exposure" Loop (Training for Self-Organization)
*   `CrowNet.Train` (renamed conceptually to `ExposeToPatterns`):
    *   Iterates for a number of epochs.
    *   In each epoch, presents each of the 10 digit patterns sequentially.
    *   For each pattern:
        *   Calls `ResetNeuronActivations()` and clears `ActivePulses`.
        *   Calls `PresentPattern()`.
        *   Calls `RunCycle()` for `ClassificationCycles` (e.g., 3-5) to allow the pattern to be processed, weights to be updated via Hebbian rule, chemicals to change, and neurons to move.

### 4.5. Weight Persistence
*   `SaveWeights(filePath)`: Saves `cn.SynapticWeights` map to a JSON file.
*   `LoadWeights(filePath)`: Loads weights from JSON, populating `cn.SynapticWeights`.

## 5. Deferred/Simplified for MVP
*   **Direct Classification to Labels:** The network self-organizes representations; mapping these to specific digit labels (0-9) is an interpretation step, not an automated output of the MVP.
*   **Advanced Hebbian Variants:** Using a basic form of Hebbian learning. More complex variants (BCH, Oja's rule for normalization) are deferred.
*   **Detailed Reward Signals for Dopamine:** Dopamine release is tied to dopaminergic neuron activity based on their inputs, not an explicit external reward signal for "correctness."
*   The complex 10-step pulse propagation from the original README.

This design emphasizes "learning by use" where internal dynamics (including chemical feedback on plasticity and thresholds) and Hebbian weight changes drive the network's adaptation to input patterns.
---
The documentation has been updated to reflect the pivot to neuromodulated self-learning.
