# Tarefa: DOC-DESIGN-001 - Revisão de alto nível dos documentos de design

**ID da Tarefa:** DOC-DESIGN-001
**Título Breve:** Revisar documentos de design de alto nível
**Descrição Completa:** Agendar e realizar uma revisão mais ampla dos principais documentos de design do projeto, como `docs/02_arquitetura.md` e as descrições de funcionalidades em `docs/04_funcionalidades/`. O objetivo é verificar se esses documentos ainda refletem com precisão os mecanismos centrais, as decisões arquiteturais e o escopo funcional do código após as recentes evoluções e refatorações. Onde houver desvios significativos, os documentos devem ser atualizados ou notas devem ser adicionadas para indicar as mudanças.
**Status:** Pendente
**Dependências (IDs):** - (Idealmente após grandes refatorações ou ciclos de features)
**Complexidade (1-5):** 4
**Prioridade (P0-P4):** P3
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-27
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** docs/review-design-docs
**Critérios de Aceitação:**
- Os principais documentos de design (arquitetura, funcionalidades) são lidos e comparados com o estado atual do código.
- Áreas de divergência são identificadas.
- Um plano para atualizar esses documentos é proposto, ou atualizações menores são feitas diretamente.
- (Opcional) Diagramas de arquitetura (se existentes) são verificados quanto à precisão.
**Notas/Decisões:**
- Esta é uma tarefa importante para a manutenção de longo prazo da saúde do projeto e para garantir que novos desenvolvedores (ou o próprio agente no futuro) tenham uma compreensão precisa do sistema.
- Pode ser dividida em sub-tarefas menores por documento ou seção, se necessário.
- Requer um bom entendimento tanto da documentação quanto do código atual.

**Resultados da Revisão (Julho 2024):**

Foram revisados: `docs/02_arquitetura.md`, `docs/04_funcionalidades/*` (README, 01 a 05), e `docs/requisitos.md`.

**Principais Divergências Identificadas e Ações Tomadas/Propostas:**

1.  **CLI (Cobra vs. `-mode`):**
    *   **Divergência:** `02_arquitetura.md` e `04-modos-operacao.md` descreviam a CLI antiga baseada em `-mode`.
    *   **Ação:** Adicionada nota proeminente em `04-modos-operacao.md` sobre a desatualização e exemplos de comando atualizados. Em `02_arquitetura.md`, descrições dos pacotes `main`, `cmd`, `cli`, `config` foram atualizadas e uma nota adicionada ao diagrama de componentes.
    *   **Proposta Futura:** Reescrita completa das seções de CLI nesses documentos para refletir a estrutura Cobra (nova tarefa `DOC-CLI-001` poderia ser criada).

2.  **Estrutura de `config.SimulationParameters`:**
    *   **Divergência:** Documentos poderiam referenciar a estrutura plana antiga.
    *   **Ação:** `docs/02_arquitetura.md` (Seção 3, Estruturas) atualizado para mencionar que `SimulationParameters` é agrupada. `docs/requisitos.md` e `02-ciclo-simulacao-aprendizado.md` atualizados para usar os novos caminhos aninhados ao referenciar parâmetros específicos.
    *   **Proposta Futura:** Garantir que todos os documentos que detalham parâmetros específicos usem a nova nomenclatura aninhada.

3.  **Interfaces de Sinaptogênese:**
    *   **Divergência:** Documentos descreviam chamadas diretas de função.
    *   **Ação:** `docs/02_arquitetura.md` (Seção 4, Algoritmos) e `docs/04_funcionalidades/02-ciclo-simulacao-aprendizado.md` (Seção 5) atualizados para mencionar o uso de interfaces.
    *   **Proposta Futura:** Detalhar mais as interfaces e suas implementações padrão se necessário.

4.  **Carregamento de Configuração TOML:**
    *   **Divergência:** Anteriormente documentado como planejado.
    *   **Ação:** `docs/02_arquitetura.md` (Seção 6.1) e `docs/requisitos.md` (RNF-CONF-001) atualizados para refletir que está implementado.
    *   **Proposta Futura:** Seção sobre TOML no `guia_interface_linha_comando.md` já está atualizada.

5.  **Requisito `RF-PERSIST-005` (Recriação de BD SQLite):**
    *   **Divergência:** Requisito dizia que BD era recriado.
    *   **Ação:** `docs/requisitos.md` corrigido. `docs/04_funcionalidades/05-persistencia-dados.md` corrigido.

6.  **Requisito `RF-CHEM-005` (Efeito do Cortisol em "U"):**
    *   **Divergência:** Requisito mencionava efeito em "U".
    *   **Ação:** `docs/requisitos.md` corrigido para efeito multiplicativo direto. `docs/04_funcionalidades/02-ciclo-simulacao-aprendizado.md` e `docs/02_arquitetura.md` alinhados.

7.  **Esquema do BD SQLite (Tabela `NeuronStates`):**
    *   **Divergência:** Documentação antiga com `Position0..15`, etc.
    *   **Ação:** `docs/04_funcionalidades/05-persistencia-dados.md` corrigido para `Position TEXT (JSON)`, etc. `docs/guia_interface_linha_comando.md` já havia sido corrigido.

**Conclusão da Tarefa DOC-DESIGN-001:**
A revisão identificou várias áreas onde os documentos de design de alto nível divergiam do código atual, principalmente devido a refatorações e implementações recentes. Correções factuais e atualizações menores foram aplicadas diretamente. Áreas que necessitam de reescrita mais substancial (como a descrição completa da nova CLI nos documentos de arquitetura e funcionalidades) foram notadas e podem originar novas tarefas de documentação específicas.
