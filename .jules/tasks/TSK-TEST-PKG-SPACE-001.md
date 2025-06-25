# Tarefa: TEST-PKG-SPACE-001 - Testes unitários para `space/geometry.go`

**ID da Tarefa:** TEST-PKG-SPACE-001
**Título Breve:** Testes unitários para `space/geometry.go`
**Descrição Completa:** Escrever testes unitários abrangentes para as funções no pacote `space/geometry.go`. O foco principal deve ser a função `ClampToHyperSphere`, cobrindo diversos casos de borda, como ponto dentro da esfera, ponto fora, ponto exatamente na superfície, raio zero, ponto na origem, e diferentes dimensionalidades (embora o código atual use `common.PointDimension` fixo). Outras funções como `EuclideanDistance`, `IsWithinRadius`, e `GenerateRandomPositionInHyperSphere` também devem ser testadas para garantir sua corretude.
**Status:** Pendente
**Dependências (IDs):** TEST-002 (para permitir a execução e validação dos testes)
**Complexidade (1-5):** 2
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-27
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** test/space-geometry-tests
**Critérios de Aceitação:**
- Testes unitários são criados para `EuclideanDistance`.
- Testes unitários são criados para `IsWithinRadius`, cobrindo casos de raio positivo, zero e negativo.
- Testes unitários são criados para `ClampToHyperSphere`, cobrindo:
    - Ponto dentro da esfera (não deve ser alterado).
    - Ponto fora da esfera (deve ser projetado para a superfície).
    - Ponto na superfície da esfera (não deve ser alterado).
    - Ponto na origem.
    - Raio zero.
    - Raio negativo (não deve prender).
- Testes unitários (ou pelo menos verificações de sanidade) são criados para `GenerateRandomPositionInHyperSphere` (ex: verificar se os pontos gerados estão dentro do raio especificado, se a distribuição parece razoável para N pequeno, ou se não retorna erros).
- Todos os novos testes passam quando o ambiente de teste estiver funcional.
**Notas/Decisões:**
- Funções geométricas são fundamentais para a simulação espacial, e sua corretude é importante.
- A testabilidade de `GenerateRandomPositionInHyperSphere` pode ser desafiadora para provar uniformidade estatística, mas verificações básicas de contorno são possíveis.
- Esta tarefa está atualmente bloqueada por `TEST-002`.
