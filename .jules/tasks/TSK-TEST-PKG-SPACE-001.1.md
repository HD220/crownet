# Tarefa: TEST-PKG-SPACE-001.1 - Testes unitários para `EuclideanDistance` e `IsWithinRadius`.

**ID da Tarefa:** TEST-PKG-SPACE-001.1
**Título Breve:** Testes para `EuclideanDistance` e `IsWithinRadius`.
**Descrição Completa:** Implementar testes unitários para as funções `EuclideanDistance` e `IsWithinRadius` no arquivo `space/geometry.go`. Os testes devem cobrir casos típicos e de borda para garantir a corretude dessas funções geométricas básicas.
**Status:** Pendente
**Dependências (IDs):** TEST-PKG-SPACE-001, TEST-002
**Complexidade (1-5):** 1
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-28
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** test/space-geometry-dist-radius
**Critérios de Aceitação:**
- Testes de tabela (table-driven tests) são implementados para `EuclideanDistance`.
    - Casos de teste incluem: distância zero (mesmo ponto), casos simples 2D e 1D (dentro do array 16D), e pontos com coordenadas negativas.
    - As comparações de float usam uma tolerância (epsilon) para imprecisões de ponto flutuante.
- Testes de tabela são implementados para `IsWithinRadius`.
    - Casos de teste incluem: ponto dentro do raio, ponto na fronteira, ponto fora do raio.
    - Casos de borda como raio zero (ponto no centro vs. fora do centro) e raio negativo são testados.
- Todos os novos testes passam.
**Notas/Decisões:**
- As funções já existem e têm testes em `space/geometry_test.go`. Esta tarefa é para verificar e garantir que esses testes existentes são abrangentes e corretos conforme os critérios de aceitação da tarefa pai `TEST-PKG-SPACE-001`. Se os testes existentes já cumprem, esta tarefa pode ser marcada como concluída rapidamente após revisão.
- O foco é validar a corretude dos cálculos de distância e checagem de raio.
