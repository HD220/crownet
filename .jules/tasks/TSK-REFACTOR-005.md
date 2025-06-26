# Tarefa: REFACTOR-005 - Avaliar e propor interfaces para estratégias de propagação de pulso.

**ID da Tarefa:** REFACTOR-005
**Título Breve:** Avaliar/propor interfaces para propagação de pulso.
**Descrição Completa:** Avaliar a atual implementação da lógica de propagação de pulso em `pulse/pulse.go` (especialmente `PulseList.ProcessCycle` e `Pulse.Propagate`) com o objetivo de propor refatorações baseadas em interfaces que aumentariam a extensibilidade. Isso permitiria a introdução de diferentes modelos ou estratégias de propagação de pulso no futuro com menor impacto no código existente.
**Status:** Pendente
**Dependências (IDs):** PERF-002
**Complexidade (1-5):** 3
**Prioridade (P0-P4):** P3
**Responsável:** AgenteJules
**Data de Criação:** 2025-06-24
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** refactor/pulse-propagation-strategy
**Critérios de Aceitação:**
- Análise da lógica atual de propagação de pulso em `pulse.go` concluída.
- Definição e implementação de interfaces para desacoplar a mecânica de propagação e aplicação de efeitos.
- Implementações padrão para as novas interfaces que replicam o comportamento original.
- `PulseList.ProcessCycle` refatorado para usar as novas interfaces.
- Documentação das interfaces e do design no código (GoDocs) e neste arquivo de tarefa.
**Notas/Decisões:**
- **Interfaces Implementadas:**
    - `PulsePropagator`: Controla como um pulso individual se move e se `isActive`. Impl. Padrão: `DefaultPulsePropagator`.
    - `PulseEffectZoneProvider`: Determina a "casca" esférica de efeito do pulso. Impl. Padrão: `DefaultPulseEffectZoneProvider`.
    - `PulseTargetSelector`: Seleciona neurônios candidatos dentro da zona de efeito (usando `SpatialGrid` por padrão). Impl. Padrão: `DefaultPulseTargetSelector`.
    - `PulseImpactCalculator`: Calcula o efeito do pulso em um neurônio alvo, incluindo a verificação precisa de distância e a geração de novo pulso se o alvo disparar. Impl. Padrão: `DefaultPulseImpactCalculator`.
- **Refatoração de `PulseList`:**
    - `PulseList` agora contém instâncias dessas quatro interfaces, inicializadas com as implementações padrão em `NewPulseList()`.
    - Um método `SetStrategies(...)` foi adicionado para permitir a injeção de diferentes implementações (útil para testes ou futuras extensões).
    - `PulseList.ProcessCycle` foi reescrito para orquestrar chamadas para os métodos das interfaces injetadas. A função auxiliar `processSinglePulseOnTargetNeuron` foi removida, sua lógica agora está em `DefaultPulseImpactCalculator`.
- **Considerações de `SpatialGrid`:** A interface `PulseTargetSelector` recebe `spatialGrid *space.SpatialGrid` como parâmetro, permitindo que implementações o utilizem (como faz a `DefaultPulseTargetSelector`) ou o ignorem se tiverem outra estratégia de seleção de alvos. `ProcessCycle` também recebe `allNeurons []*neuron.Neuron` para passar ao seletor de alvos, caso ele não use o grid.
- **Impacto no Desempenho:** Espera-se que o impacto no desempenho seja mínimo, pois a sobrecarga de chamadas de interface é geralmente pequena em comparação com os cálculos envolvidos.
- **Extensibilidade:** O design agora permite substituir qualquer uma das quatro etapas do processamento de pulso independentemente, implementando as respectivas interfaces.
- **Constante `defaultPulseMaxTravelRadiusFactor`:** Esta constante, usada na criação de novos pulsos, ainda é definida no pacote `pulse`. Para maior configurabilidade, poderia ser movida para `config.SimulationParameters.General` no futuro.
- Alinha-se com o RNF-EXT-001 (Extensibilidade).
- A otimização de `PERF-002` (SpatialGrid) é mantida na implementação padrão de `PulseTargetSelector`.
- Diferentes modelos (atrasos variáveis, atenuação complexa, etc.) podem agora ser implementados criando novas structs que satisfaçam estas interfaces.
