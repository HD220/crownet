# Tarefa: REFACTOR-004 - Refatorar comportamento neuronal para usar interfaces (Extensibilidade).

**ID da Tarefa:** REFACTOR-004
**Título Breve:** Refatorar neurônio para usar interfaces (Extensibilidade).
**Descrição Completa:** Refatorar o pacote `neuron` e suas interações com `network` para utilizar interfaces para comportamentos neuronais chave. O objetivo é aumentar a extensibilidade do sistema, facilitando a introdução de novos tipos de neurônios com comportamentos de disparo, integração de potencial ou efeitos de pulso distintos, sem modificar extensivamente o código central da simulação.
**Status:** Pendente
**Dependências (IDs):** -
**Complexidade (1-5):** 4
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2025-06-24
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** refactor/neuron-interfaces
**Critérios de Aceitação:**
- Interfaces como `FiringBehavior` (contendo `ShouldFire(n *Neuron) bool`), `PotentialIntegrationBehavior` (contendo `Integrate(n *Neuron, potential PulseValue) bool`), e `PulseEmissionBehavior` (contendo `EmitPulse(n *Neuron) *pulse.Pulse`) são definidas.
- A struct `Neuron` é modificada para conter instâncias dessas interfaces, ou o pacote `network` utiliza essas interfaces ao interagir com neurônios.
- A lógica atual de disparo, integração e emissão de pulso é movida para implementações concretas dessas interfaces para os tipos de neurônios existentes.
- O impacto nas chamadas existentes (ex: em `PulseList.ProcessCycle`, `CrowNet.RunCycle`) é gerenciado.
- A documentação de arquitetura (`docs/02_arquitetura.md`) é atualizada para refletir o novo design baseado em interfaces.
- Testes unitários são atualizados ou criados para as novas interfaces e suas implementações.
**Notas/Decisões:**
- Esta é uma refatoração significativa que pode impactar várias partes do sistema.
- Avaliar se uma única interface `NeuronalBehavior` com múltiplos métodos é melhor do que interfaces menores e mais focadas.
- Considerar o impacto no desempenho; chamadas de interface podem ter um pequeno overhead, mas a flexibilidade pode compensar.
- Esta refatoração alinha-se com RNF-EXT-001.
