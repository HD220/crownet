# Project Requirements Document: CrowNet - Neuromodulated Self-Learning MVP

## 1. Introduction

This document outlines the requirements for a Minimum Viable Product (MVP) of the CrowNet simulation. The project has pivoted to focus on demonstrating **neuromodulated Hebbian self-learning**, where the network, inspired by the general concepts in the original CrowNet README (16D space, neuron types, activity-dependent dynamics), aims to self-organize in response to input patterns (simple digits 0-9). The key is to show that the network can learn to form distinct internal representations (activity patterns in output neurons) for different inputs, with its plasticity (learning rate) influenced by simulated cortisol and dopamine levels.

## 2. Goals

*   **Primary Goal:** Create an MVP demonstrating that the CrowNet model can exhibit self-organizing behavior through neuromodulated Hebbian plasticity. This means showing that:
    *   Synaptic weights evolve based on correlated neuron activity.
    *   The rate of this evolution is influenced by simulated cortisol and dopamine levels.
    *   Different input digit patterns lead to distinguishable activity patterns across a set of output neurons after a period of exposure/learning.
*   **Secondary Goal:** Maintain a configurable command-line interface to run the simulation, expose it to patterns, observe its state, and save/load learned synaptic weights.
*   **Target User:** Individuals interested in computational neuroscience, models of learning and memory, neuromodulation, and complex adaptive systems.

## 3. Scope

### 3.1. MVP Features

*   **F1: Network Initialization:**
    *   F1.1: Initialize a network with a user-configurable total number of neurons.
    *   F1.2: Ensure at least 35 Input neurons (for 5x7 digit patterns) and 10 Output neurons are created. Other neurons (Excitatory, Inhibitory, Dopaminergic) are distributed based on remaining count and configured percentages.
    *   F1.3: Neurons are positioned in a 16D space (initially random within type-specific radial constraints).
    *   F1.4: Initialize neuron properties (ID, type, state, base thresholds).
    *   F1.5: Initialize explicit `SynapticWeights` (map `from_neuron_id` to `to_neuron_id` to `weight`) between all non-input neurons and from input neurons to other neurons. Initial weights can be small random values. Self-connections are weighted zero.
*   **F2: Core Simulation Cycle (with Learning):**
    *   F2.1: Discrete time step simulation (`RunCycle`).
    *   F2.2: Neuron state updates (resting, firing, refractory periods), pulse accumulation (`AccumulatedPulse += BaseSignal * Weight`), and decay.
    *   F2.3: Neuron firing based on `CurrentFiringThreshold`.
    *   F2.4: Basic pulse propagation (spherical expansion). `Pulse.Value` carries `BaseSignal` (+1.0 for excitatory, -1.0 for inhibitory).
    *   F2.5: **Neuromodulated Hebbian Plasticity (`ApplyHebbianPlasticity` called in `RunCycle`):**
        *   Identify pairs of neurons (pre, post) that have shown recent correlated activity (e.g., both fired).
        *   Calculate an `effective_learning_rate` based on `BaseLearningRate` and current `DopamineLevel` (enhances plasticity) and `CortisolLevel` (high levels suppress plasticity).
        *   Update `SynapticWeights[pre][post]` using a Hebbian rule: `Δw = effective_learning_rate * pre_activity * post_activity`.
        *   Implement basic weight bounds or decay if necessary to prevent runaway weights.
    *   F2.6: **Synaptogenesis (Neuron Movement):** Remains active as per README, potentially influencing connectivity over longer timescales by changing neuron proximities. Its rate is modulated by chemical levels.
    *   F2.7: **Chemical Modulation:** Cortisol and Dopamine production and decay mechanisms remain active. Their primary role in *this MVP's learning context* is to modulate neuron excitability (firing thresholds) and the *rate of Hebbian plasticity*.
*   **F3: Digit Representation & Input Encoding:**
    *   F3.1: Predefined 5x7 binary patterns for digits 0-9 provided by a `datagen` package.
    *   F3.2: `PresentPattern` function in `CrowNet` to activate the first 35 input neurons based on a flattened digit pattern. Active input neurons fire once per presentation.
*   **F4: Output Interpretation:**
    *   F4.1: The state of the 10 designated Output neurons (e.g., their `AccumulatedPulse` or short-term firing rate after pattern presentation and settling) constitutes the network's response pattern.
    *   F4.2: The MVP aims to show that these output patterns become distinct for different input digits after a period of self-learning/exposure. Explicit digit label prediction by the network is a post-MVP analysis/interpretation step.
*   **F5: Modes of Operation & Configuration:**
    *   F5.1: Command-line application (`main.go`).
    *   F5.2: CLI flags for:
        *   `-mode`: "expose" (repeatedly present digit patterns to allow learning), "observe" (present a specific digit and view output pattern), "sim" (general simulation with original dynamics and optional continuous input).
        *   `-neurons`, `-cycles` (total for simulation/exposure).
        *   `-epochs` (for "expose" mode).
        *   `-lrBase` (base learning rate for Hebbian rule).
        *   `-weightsFile`: Path to save/load synaptic weights (JSON format).
        *   `-digit`: Digit to present in "observe" mode.
        *   `-stimInputID`, `-stimInputFreqHz` for "sim" mode's continuous stimulus.
        *   `-monitorOutputID` for frequency reporting in "sim" mode.
        *   `-dbPath`, `-saveInterval` for SQLite logging (primarily for "sim" mode or detailed analysis of "expose" mode).
        *   `-debugChem` for chemical production/level debug prints.
*   **F6: Persistence:**
    *   F6.1: Save and load synaptic weights to/from a JSON file.
    *   F6.2: Optionally save full network state snapshots to SQLite during any mode.
*   **F7: MVP Output & Observation:**
    *   "expose" mode: Periodic updates on epoch number, chemical levels.
    *   "observe" mode: The input digit pattern and the resulting activation vector of the 10 output neurons.
    *   "sim" mode: General simulation stats.

### 3.2. Out of Scope for MVP

*   **Supervised Classification to Labels:** The network will not be trained to map output neuron activity to specific digit labels (0-9) using error-correction backpropagation or similar. The focus is on self-organized representation.
*   **Advanced Learning Algorithms:** Reinforcement learning with explicit reward signals, backpropagation, etc.
*   **Complex Input Data/Preprocessing:** MNIST, etc.
*   **Advanced Pulse Propagation:** The 10-step model from README.
*   **GPU Acceleration (ArrayFire), Graphical Visualization (Robotgo).**

## 4. Target Users

*   Individuals interested in computational models of self-organization, Hebbian learning, neuromodulation, and biologically-inspired neural dynamics.

## 5. Success Metrics for MVP

*   The simulation runs stably with all dynamics (Hebbian learning, synaptogenesis, chemical modulation) enabled.
*   Synaptic weights demonstrably change over time during exposure to digit patterns.
*   The rate/nature of weight changes and/or network activity patterns show influence from varying Cortisol/Dopamine levels.
*   After sufficient exposure, presenting different digit patterns results in observably different and reasonably consistent activation patterns across the 10 output neurons.
*   The CLI is usable for configuring and running "expose" and "observe" modes.
*   Weights can be saved and loaded.
*   Code and documentation clearly reflect the self-learning paradigm.
---
This PRD now focuses on the self-learning aspect.
