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
- Proposta de uma ou mais interfaces (ex: `PulsePropagationStrategy`, `PulseEffectApplicator`) que poderiam desacoplar a mecânica de propagação e aplicação de efeitos.
- Documentação da proposta, incluindo como as implementações concretas (como a atual baseada em expansão esférica e grid espacial) se encaixariam.
- Consideração dos impactos no desempenho e na complexidade.
- (Opcional, se a proposta for simples) Implementação de um protótipo da refatoração.
**Notas/Decisões:**
- O objetivo principal é o design e a proposta, não necessariamente a implementação completa nesta tarefa, a menos que seja trivial.
- Alinha-se com o RNF-EXT-001 (Extensibilidade).
- A otimização de `PERF-002` (SpatialGrid) deve ser considerada no design das interfaces.
- Diferentes modelos poderiam incluir, por exemplo, propagação com atrasos variáveis, atenuação de sinal baseada em distância de forma mais complexa, ou diferentes formas de área de efeito.
