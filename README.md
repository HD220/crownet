# CrowNet: A Biologically Inspired Neural Network Simulator

**CrowNet** is a command-line application written in Go that simulates a computational neural network model. It draws inspiration from biological processes, featuring neurons interacting in a 16-dimensional vector space, synaptogenesis (activity-dependent neuron movement), and neuromodulation by simulated cortisol and dopamine.

The current Minimum Viable Product (MVP) focuses on demonstrating **neuromodulated Hebbian self-learning**. The network is exposed to simple digit patterns (0-9) and aims to self-organize its synaptic weights to form distinct internal representations (activity patterns across designated output neurons) for these different inputs. The learning process (plasticity) itself is influenced by the simulated chemical environment.

## Core Concepts Implemented

*   **16-Dimensional Space:** Neurons exist and move within a 16D vector space.
*   **Neuron Types:** The network consists of Excitatory, Inhibitory, Dopaminergic, Input, and Output neurons, each with specific roles and distribution patterns.
*   **Pulse Propagation:** A simplified spherical expansion model where pulses travel at a fixed speed (0.6 units/cycle). The original, more complex 10-step propagation model described below is currently deferred.
*   **Synaptic Weights:** Explicit synaptic weights between neurons are implemented and initialized randomly. These weights determine the strength of influence when a pulse travels from one neuron to another.
*   **Neuromodulated Hebbian Learning:**
    *   **Hebbian Plasticity:** Synaptic weights are adjusted based on the principle of "neurons that fire together, wire together." Co-activation of pre- and post-synaptic neurons (within a defined time window) leads to strengthening of their connection.
    *   **Neuromodulation:** The base learning rate for Hebbian updates is modulated by global levels of:
        *   **Dopamine:** Produced by firing dopaminergic neurons. Higher dopamine levels enhance the learning rate, promoting plasticity.
        *   **Cortisol:** Produced by excitatory pulses near a central "gland." High cortisol levels suppress the learning rate, reducing plasticity.
    *   Firing thresholds of neurons are also modulated by these chemicals.
*   **Synaptogenesis:** Neurons can move in the 16D space. Their movement is influenced by the activity of nearby neurons (attraction to active, repulsion from resting) and modulated by chemical levels (high cortisol reduces movement). This dynamic spatial arrangement can influence connectivity over time.
*   **Input Encoding:** Simple 5x7 binary patterns for digits 0-9 are used as input. These patterns activate a designated set of 35 input neurons.
*   **Output Representation:** The network has 10 designated output neurons. The goal of the self-learning process is for these neurons to develop distinct and consistent patterns of activity in response to different input digits. The MVP focuses on observing these patterns rather than achieving perfect classification to digit labels.

## Modes of Operation (Command Line)

The application supports three main modes:

1.  **`expose` Mode (`-mode expose`):**
    *   Presents the predefined digit patterns (0-9) to the network repeatedly over a specified number of epochs.
    *   During this phase, all dynamics are active: Hebbian learning updates weights, chemical levels modulate learning rates and thresholds, and synaptogenesis allows neurons to move.
    *   This mode is for allowing the network to self-organize based on input experience.
    *   Learned synaptic weights can be saved to a file.
    *   Key flags: `-epochs`, `-lrBase` (base learning rate), `-cyclesPerPattern`, `-weightsFile`.

2.  **`observe` Mode (`-mode observe`):**
    *   Loads a previously saved set of synaptic weights.
    *   Presents a specified digit pattern to the network.
    *   Runs the network for a few "settling" cycles (with Hebbian learning, synaptogenesis, and dynamic chemical changes temporarily disabled for a clean feed-forward pass).
    *   Outputs the activation pattern (e.g., `AccumulatedPulse` values) of the 10 output neurons, allowing the user to see how the trained network represents the input digit.
    *   Key flags: `-digit <0-9>`, `-weightsFile`, `-cyclesToSettle`.

3.  **`sim` Mode (`-mode sim`):**
    *   Runs a general simulation with all dynamics (synaptogenesis, chemical modulation, Hebbian learning if weights are present) enabled.
    *   Allows for continuous input stimulus to a specified input neuron at a given frequency.
    *   Can log full network snapshots to an SQLite database for detailed analysis.
    *   Useful for observing the original CrowNet dynamic behaviors over longer periods without the specific constraints of the digit exposure/observation tasks.
    *   Key flags: `-cycles`, `-stimInputID`, `-stimInputFreqHz`, `-monitorOutputID`, `-dbPath`, `-saveInterval`, `-debugChem`.

## Technologies Utilized (MVP)

*   **Go:** Implementation language.
*   **SQLite:** For saving detailed simulation snapshots (primarily in "sim" mode or for detailed analysis of "expose" mode).
*   **JSON:** For saving and loading learned synaptic weights.

### Deferred/Future Technologies (from original README)
*   **ArrayFire (GPU Acceleration):** Not implemented in MVP due to environmental constraints.
*   **Robotgo (Visualization):** Not implemented in MVP.
*   **OpenNoise (Procedural Generation):** Neuron placement currently uses random distribution within radial constraints; a sophisticated noise generator is deferred.

## Further Documentation

For more detailed information, please refer to:
*   `PROJECT_REQUIREMENTS.md`: Goals, scope, and features of the current MVP.
*   `USE_CASES.md`: Scenarios for using the different modes of the application.
*   `TECHNICAL_DESIGN.md`: Overview of the architecture, data structures, and core algorithms.

## Original README Concepts (Preserved for Context)

(The following sections are largely from the original README, providing context on the initial biological inspirations. Some aspects, like the detailed 10-step pulse propagation, are simplified in the current MVP.)

### Neuron Cycles & Firing
(Original section 4, 10, 11 from README can be kept as they generally apply)
O comportamento dos neurônios é modelado em 4 ciclos principais: Repouso, Disparo, Refratário Absoluto, Refratário. Cada neurônio mantém um registro do último ciclo em que disparou. Os neurônios disparam quando a soma dos pulsos recebidos (modulated by weights) excede o limiar de disparo. Quando não recebem pulsos, a soma diminui gradativamente.

### Pulse Propagation (Original Specification Detail)
(Original section 5, 6, and the 10-step pseudocode from README can be kept here, with a note that the MVP uses a simplified spherical expansion model currently.)
A propagação de pulso entre os neurônios é baseada na distância percorrida (velocidade: 0.6 unidades/ciclo). A distância é Euclidiana em 16D.
*Original 10-step pulse processing pseudocode can be included here, marked as 'deferred in current MVP implementation'.*

### Synaptogenesis (Neuron Movement - As Implemented)
A sinapogênese é a taxa de movimentação dos neurônios no espaço, ajustada após a propagação de pulsos e modulada por químicos:
- Neurônios se aproximam daqueles que dispararam ou estavam em período refratário.
- Neurônios se afastam daqueles que estavam em repouso.

### Cortisol e Dopamina (Effects on Thresholds & Plasticity - As Implemented)
- **Cortisol**: Produção afetada por pulsos excitatórios na glândula central. Modula limiar de disparo (U-shaped) e sinapogênese/taxa de aprendizado (supressão em níveis altos). Decai com o tempo.
- **Dopamina**: Gerada por neurônios dopaminérgicos. Aumenta limiar de disparo. Aumenta sinapogênese/taxa de aprendizado. Decai mais rapidamente que o cortisol.

---
This updated README should provide a good overview of the current project state.
