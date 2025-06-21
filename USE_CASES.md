# Use Cases: CrowNet - Neuromodulated Self-Learning MVP

This document describes key use cases for the CrowNet MVP, focusing on its capacity for self-organization and pattern differentiation via neuromodulated Hebbian learning.

## Actors

*   **User:** An individual (researcher, student, hobbyist) interacting with the CrowNet application via the command line.

## Self-Learning & Observation Use Cases

**UC-EXPOSE: Expose Network to Digit Patterns for Self-Learning**
*   **Actor:** User
*   **Description:** The user exposes the network to a sequence of digit patterns over many cycles/epochs, allowing it to self-organize its synaptic weights through neuromodulated Hebbian plasticity.
*   **Preconditions:** None (network can start with initial random weights if no weights file is specified or found).
*   **Steps:**
    1.  User executes CrowNet with ` -mode expose`.
    2.  Optional flags: `-neurons <N>`, `-epochs <E>`, `-lrBase <rate>`, `-weightsFile <path>`, `-cycles <C_per_pattern>`, `-debugChem <true/false>`, `-saveInterval <I_db>`.
    3.  System initializes/loads a CrowNet instance. If `<path>` for weights exists, weights are loaded. Synaptogenesis and Chemical Modulation are ENABLED.
    4.  System enters an exposure loop for `E` epochs:
        a.  For each digit pattern (0-9) from `datagen`:
            i.  `PresentPattern()`: The digit pattern activates corresponding input neurons (once).
            ii. The network runs for `C_per_pattern` cycles (e.g., 3-10 cycles, allowing pulses to propagate and activity to evolve). During these cycles:
                1.  `RunCycle()` is called repeatedly.
                2.  `ApplyHebbianPlasticity()` (within `RunCycle`) updates weights based on neuron co-activity and current chemical levels.
                3.  Chemical levels and neuron positions evolve.
            iii. Neuron activations are reset before the next pattern.
        b.  Periodic status (epoch number, chemical levels, perhaps a summary of weight changes) is printed.
        c.  If `saveInterval > 0`, full network state (including weights) is saved to SQLite periodically.
    5.  After all epochs, the final synaptic weights are saved to `<path>`.
*   **Postconditions:**
    *   Synaptic weights in `<path>` are updated/created, reflecting learned associations.
    *   Exposure process completes. Data may be saved to SQLite for detailed analysis of dynamics.

**UC-OBSERVE: Observe Network's Internal Representation of a Digit**
*   **Actor:** User
*   **Description:** After the network has been exposed to patterns (UC-EXPOSE), the user presents a specific digit to observe the resulting activation pattern across the output neurons. This helps assess if the network has formed distinct internal representations.
*   **Preconditions:** Preferably, a trained weights file (`<path>`) exists from a previous "expose" run.
*   **Steps:**
    1.  User executes CrowNet with ` -mode observe -digit <D>` (where D is 0-9).
    2.  Optional flags: `-weightsFile <path>`, `-neurons <N>`, `-cycles <C_settle>`.
    3.  System initializes a CrowNet instance and loads synaptic weights from `<path>`. Synaptogenesis and Chemical Modulation are ENABLED (to reflect the state during learning, or can be selectively disabled for pure feed-forward observation if desired via future flags).
    4.  System retrieves the pattern for digit `<D>` from `datagen`.
    5.  `ResetNeuronActivations()`.
    6.  `PresentPattern()`: The digit pattern activates input neurons.
    7.  The network runs for `C_settle` cycles (e.g., 3-5, `ClassificationCycles` from config).
    8.  System retrieves and prints the activation levels (e.g., `AccumulatedPulse`) of the 10 designated output neurons.
*   **Postconditions:**
    *   A vector of 10 floating-point values representing the output neurons' activities is displayed. The user can compare these vectors for different input digits to qualitatively assess if distinct representations have formed.

## General Simulation Use Case (Original CrowNet Dynamics)

**UC-SIM: Run a General Simulation with Full Dynamics**
*   **Actor:** User
*   **Description:** The user runs the simulation with all original CrowNet dynamics enabled (synaptogenesis, chemical modulation, frequency-based I/O) to observe general network behavior, not specifically tied to a digit task with Hebbian learning. Synaptic weights will still be present and could be initialized randomly or loaded, but the Hebbian `ApplyHebbianPlasticity` might be less relevant unless co-activity naturally occurs and is modulated.
*   **Preconditions:** None.
*   **Steps:**
    1.  User executes CrowNet with ` -mode sim`.
    2.  Optional flags: `-neurons <N>`, `-cycles <C>`, `-stimInputID <ID>`, `-stimInputFreqHz <Hz>`, `-monitorOutputID <ID>`, `-dbPath <dbfile>`, `-saveInterval <I>`, `-debugChem <true/false>`, `-weightsFile <path>` (to load pre-existing weights if desired).
    3.  System initializes/loads a CrowNet instance. Synaptogenesis and Chemical Modulation are ENABLED by default. Hebbian plasticity will also be active if implemented within `RunCycle`.
    4.  If `-stimInputFreqHz > 0`, the specified input neuron fires at that frequency.
    5.  System runs the simulation for `C` cycles, calling `RunCycle` which includes all dynamics.
    6.  System prints periodic status and saves to SQLite as configured.
    7.  At the end, reports final chemical levels and monitored output neuron frequency (if any).
*   **Postconditions:**
    *   Simulation completes.
    *   SQLite database contains snapshots for analysis of the full dynamic system.
---
These use cases now align with the self-learning paradigm using neuromodulated Hebbian plasticity.
