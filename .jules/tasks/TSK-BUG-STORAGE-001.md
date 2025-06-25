# Tarefa: BUG-STORAGE-001 - Corrigir carregamento de pesos em `storage.LoadNetworkWeightsFromJSON`

**ID da Tarefa:** BUG-STORAGE-001
**Título Breve:** Corrigir instanciação de `NetworkWeights` em `LoadNetworkWeightsFromJSON`
**Descrição Completa:** A função `storage.LoadNetworkWeightsFromJSON` tentava criar `synaptic.NetworkWeights` usando `synaptic.NewNetworkWeights()` sem argumentos, o que é incorreto, pois o construtor real (`synaptic.NewNetworkWeights(simParams *config.SimulationParameters, rng *rand.Rand)`) requer `simParams` e `rng`. A função foi refatorada para retornar `map[common.NeuronID]synaptic.WeightMap`. O chamador (`cli.Orchestrator.loadWeights`) agora é responsável por criar a instância `NetworkWeights` e popular seus pesos usando o método `LoadWeights` da struct `NetworkWeights`. A função `storage.SaveNetworkWeightsToJSON` também foi ajustada para aceitar `*synaptic.NetworkWeights`.
**Status:** Concluído
**Dependências (IDs):** -
**Complexidade (1-5):** 3
**Prioridade (P0-P4):** P0
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-27
**Data de Conclusão (Estimada/Real):** 2024-07-27
**Branch Git Proposta:** fix/bug-storage-001-loadweights
**Critérios de Aceitação:**
- `storage.LoadNetworkWeightsFromJSON` retorna `map[common.NeuronID]synaptic.WeightMap` e não instancia `NetworkWeights`.
- `cli.Orchestrator.loadWeights` usa o mapa retornado para popular uma instância `NetworkWeights` existente.
- `storage.SaveNetworkWeightsToJSON` aceita `*synaptic.NetworkWeights`.
- O carregamento e salvamento de pesos funcionam corretamente após as modificações.
**Notas/Decisões:**
- Bug crítico que impedia o carregamento correto de pesos sinápticos.
- Teste manual bloqueado.
