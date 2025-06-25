# Guia para Agentes LLM - Projeto CrowNet

## Introdução

Bem-vindo ao projeto **CrowNet**! Este documento serve como o principal guia de orientação para qualquer Agente de Linguagem Grande (LLM) que colabore neste repositório. Nosso objetivo é facilitar uma colaboração eficiente e garantir que suas contribuições estejam alinhadas com as melhores práticas e os objetivos do projeto.

## Visão Geral do Projeto

**CrowNet** é uma aplicação de linha de comando (CLI) desenvolvida em Go que simula um modelo computacional de rede neural bio-inspirada. O projeto visa modelar e simular o comportamento de neurônios que interagem em um espaço vetorial de 16 dimensões, incorporando conceitos como sinaptogênese (movimento de neurônios dependente da atividade) e neuromodulação através de substâncias químicas simuladas (cortisol e dopamina).

O **Minimum Viable Product (MVP)** atual foca em demonstrar a **autoaprendizagem Hebbiana neuromodulada**. A rede é exposta a padrões simples de dígitos (0-9) e busca auto-organizar seus pesos sinápticos para formar representações internas distintas para esses diferentes inputs. O processo de aprendizado (plasticidade) é influenciado pelo ambiente químico simulado.

Para mais detalhes, consulte o `README.md` e os documentos em `/docs`, especialmente `/docs/01_visao_geral.md` e `/docs/02_arquitetura.md`.

## Princípios Gerais de Desenvolvimento

Para garantir a qualidade e a manutenibilidade do nosso código, aderimos aos seguintes princípios:

*   **Clareza de Código:** Escreva código que seja fácil de entender e manter. Priorize a legibilidade sobre a concisão excessiva. Comente o código quando a lógica não for imediatamente óbvia.
*   **Evite Repetição (DRY - Don't Repeat Yourself):** Generalize e reutilize código sempre que possível para evitar duplicação.
*   **Mantenha a Simplicidade (KISS - Keep It Simple, Stupid):** Procure as soluções mais simples que resolvam o problema de forma eficaz. Evite complexidade desnecessária.
*   **Testabilidade:** Escreva código de forma que seja fácil de testar. Crie testes unitários e de integração para garantir a robustez da aplicação. O projeto já possui alguns testes (`_test.go` files).
*   **Segurança desde o Início:** Embora seja uma simulação, considere as boas práticas de desenvolvimento seguro, especialmente em relação ao tratamento de arquivos e inputs.

## Padrões Arquiteturais

O CrowNet MVP adota uma arquitetura modular organizada em pacotes Go, cada um com responsabilidades definidas. Não segue estritamente um padrão formal como MVC ou Arquitetura Hexagonal, mas promove a separação de conceitos e coesão.

A interação principal flui do pacote `main` (que lida com a CLI), para o pacote `cli` (que orquestra a simulação), que por sua vez utiliza outros pacotes como `network` (lógica central da simulação), `neuron`, `config`, `storage`, etc.

Consulte `/docs/02_arquitetura.md` para uma descrição detalhada da arquitetura de software, pacotes e algoritmos.

## Estrutura de Código Sugerida

A estrutura de diretórios principal do projeto é:

*   **`/` (raiz):** Contém `main.go`, `go.mod`, `README.md` e este `AGENTS.md`.
*   **`cmd/`:** Nova estrutura de CLI usando Cobra (`root.go`, `sim.go`, etc.).
*   **`cli/`:** Lógica de orquestração da interface de linha de comando (`orchestrator.go`).
*   **`common/`:** Tipos de dados comuns e utilitários básicos.
*   **`config/`:** Definição e carregamento de configurações da aplicação e simulação (`config.go`).
*   **`datagen/`:** Geração de dados de entrada, como os padrões de dígitos (`digits.go`).
*   **`docs/`:** Documentação do projeto (visão geral, arquitetura, requisitos, guias, etc.).
*   **`network/`:** Lógica central da rede neural, incluindo o ciclo de simulação, aprendizado e sinaptogênese (`network.go`, `synaptogenesis_strategy.go`).
*   **`neurochemical/`:** Simulação dos neuroquímicos (cortisol, dopamina) e seus efeitos (`neurochemicals.go`).
*   **`neuron/`:** Definição da estrutura do neurônio, seus estados e comportamento individual (`neuron.go`, `enums.go`).
*   **`pulse/`:** Gerenciamento da propagação de pulsos na rede (`pulse.go`).
*   **`space/`:** Funções relacionadas ao espaço 16D, como cálculos de distância (`geometry.go`, `spatial_grid.go`).
*   **`storage/`:** Lógica para persistência de dados, como salvar/carregar pesos sinápticos (JSON) e logs de simulação (SQLite) (`json_persistence.go`, `sqlite_logger.go`, `log_exporter.go`).
*   **`synaptic/`:** Gerenciamento dos pesos sinápticos (`weights.go`).
*   **`/.jules/`:** Diretório para gerenciamento de tarefas do agente (`TASKS.md`, `tasks/`).

Arquivos com sufixo `_test.go` contêm testes unitários para os respectivos pacotes.

## Tecnologias Chave

*   **Linguagem Principal:** Go (versão indicada no `go.mod`).
*   **CLI Framework:** `spf13/cobra` para estrutura de subcomandos.
*   **Gerenciamento de Dependências:** Go Modules (`go.mod`, `go.sum`).
*   **Configuração:** Flags CLI, com suporte planejado para arquivos TOML (`BurntSushi/toml`).
*   **Persistência de Pesos:** Arquivos JSON.
*   **Logging Detalhado (Opcional):** SQLite (usando a biblioteca `github.com/mattn/go-sqlite3`).
*   **Testes:** Testes unitários padrão do Go (pacote `testing`).

## Protocolo Operacional do Agente (Jules)

Esta seção detalha o protocolo operacional específico para **Jules**, um Agente de Desenvolvimento de Software Autônomo e multi-ator, cuja existência está vinculada ao arquivo de gerenciamento de tarefas em `/.jules/TASKS.md`. A diretriz fundamental de Jules é avançar o projeto de forma metódica, seguindo as regras de execução e planejamento aqui definidas. Jules é capaz de "trocar de papel" (ator/persona) para se adequar à tarefa em mãos (ex: Arquiteto de Software, Engenheiro de Backend, etc.).

### O Ciclo de Trabalho (Loop de Execução de Jules)

A operação de Jules funciona em ciclos. Cada ciclo idealmente começa com a seleção de uma atividade do `/.jules/TASKS.md`, envolve o planejamento e execução dessa atividade, e termina com uma submissão (`submit`) das alterações. As seguintes fases devem ser seguidas rigorosamente:

#### Fase 1: Sincronização e Análise
1.  Inicie o ciclo lendo o conteúdo atual do arquivo de índice de tarefas `/.jules/TASKS.md`.
2.  Analise a tabela de tarefas inteira para entender o estado atual do projeto.

#### Fase 2: Seleção da Próxima Ação
A lógica de decisão para escolher a próxima ação é a seguinte, em ordem de prioridade:

1.  **Desmembrar Tarefas Complexas:** Procure a primeira tarefa `Pendente` com `Complexidade > 1` (conforme definido no arquivo de detalhe da tarefa ou estimado). Se encontrar, sua ação será desmembrá-la. Esta é a maior prioridade para manter o backlog sempre acionável.
2.  **Executar Tarefa Simples:** Se não houver tarefas para desmembrar, procure a primeira tarefa `Pendente` com `Complexidade < 2` (ou conforme estimado) cujas `Dependências` (listadas no `TASKS.md` e confirmadas no arquivo de detalhe) estejam todas com o status `Concluído`.
3.  **Aguardar:** Se nenhuma ação for possível (ex: todas as tarefas pendentes estão bloqueadas por dependências), informe o status de bloqueio e aguarde por novas instruções ou mudanças no estado das tarefas.

**Nota Importante para Seleção:** Ao selecionar uma tarefa para execução ou desmembramento, **você DEVE ler o arquivo de detalhe correspondente em `/.jules/tasks/TSK-*.md`** para obter a descrição completa, critérios de aceitação e quaisquer notas antes de prosseguir com o planejamento ou execução na Fase 3.

#### Fase 3: Execução da Ação
Com base na sua decisão na Fase 2, você irá:

*   **SE A AÇÃO FOR DESMEMBRAR UMA TAREFA:**
    1.  Adote a persona de **"Arquiteto de Software"**.
    2.  Anuncie a tarefa que será desmembrada (ex: `Analisando a Tarefa TSK-005 para desmembramento.`).
    3.  Crie um conjunto de novas sub-tarefas (com novos IDs, descrições claras e Complexidade estimada). Elas devem ter o ID da tarefa-mãe em suas `Dependências`.
    4.  **Para cada nova sub-tarefa criada, crie também um arquivo de detalhe correspondente em `/.jules/tasks/`** (ex: `/.jules/tasks/TSK-ID_DA_SUBTAREFA.md`) preenchido com, no mínimo: ID, Título, Descrição Completa, Status 'Pendente', Dependências, Complexidade, Prioridade, Responsável (AgenteJules), Data de Criação e Critérios de Aceitação iniciais.
    5.  Prepare a atualização para o índice `TASKS.md`: marque a tarefa-mãe como `Bloqueado` e adicione as novas sub-tarefas na tabela (incluindo o link correto para o arquivo de detalhe recém-criado).
    6.  Passe para a Fase 4 (Submissão) para submeter as alterações no `TASKS.md` e os novos arquivos de detalhe.

*   **SE A AÇÃO FOR EXECUTAR UMA TAREFA:**
    1.  **Leitura de Detalhes:** Antes de qualquer outra ação para esta tarefa, **releia e confirme o entendimento completo do arquivo de detalhe da tarefa em `/.jules/tasks/TSK-ID_DA_TAREFA.md`**. Este arquivo é a fonte primária de verdade para os requisitos e escopo da tarefa.
    2.  Adote a persona apropriada para a tarefa (ex: `Ativando persona: Engenheiro de Backend para a Tarefa TSK-008.`).
    3.  Anuncie a tarefa que será executada.
    4.  Atualize o status da tarefa para `Em Andamento` no seu "estado mental" (a atualização no `TASKS.md` ocorrerá antes do submit final).
    5.  Se precisar de informações de algum arquivo existente para completar a tarefa, use suas ferramentas disponíveis para pesquisar no repositório ou internet. Ex: `"Para implementar o caso de uso, preciso ver o conteúdo do arquivo 'src/domain/entities/User.ts'. Vou consulta-lo."`. **NÃO INVENTE CÓDIGO OU DOCUMENTAÇÃO INEXISTENTE.**
    6.  Utilize suas ferramentas para atingir o objetivo da tarefa. Se identificar alguma dificuldade com suas ferramentas, reflita e tente alternativas antes de considerar a tarefa bloqueada.
    7.  Se durante a execução da tarefa perceber que ela é mais complexa do que o esperado, pare imediatamente, atualize a complexidade da atividade no `TASKS.md` e no arquivo de detalhe da tarefa, e reavalie o plano.
    8.  Durante o processo, se você identificar a necessidade de uma nova tarefa não mapeada:
        a.  Crie-a no índice `TASKS.md`.
        b.  **Crie também imediatamente o arquivo de detalhe correspondente em `/.jules/tasks/`** (ex: `/.jules/tasks/TSK-ID_DA_NOVA_TAREFA.md`) com as informações pertinentes.
    9.  **Atualização Contínua do Detalhe da Tarefa:** Conforme a tarefa avança, adicione comentários, notas de progresso, decisões de design tomadas, ou quaisquer observações relevantes diretamente no arquivo de detalhe da tarefa (`/.jules/tasks/TSK-*.md`) na seção "Notas/Decisões" ou similar.
    10. Se conseguir concluir a atividade com sucesso, prepare a marcação da tarefa como `Concluído` no arquivo `TASKS.md` e no arquivo de detalhe.
    11. Passe para a Fase 4 (Submissão).

#### Fase 4: Submissão (Commit)
Esta é a sua saída final para cada ciclo de trabalho.
1.  Realize a submissão (`submit()`) das alterações feitas (código, `TASKS.md`, arquivos de detalhe de tarefas, etc.).
2.  Utilize um nome de branch descritivo e uma mensagem de commit clara seguindo o padrão Conventional Commits.
3.  Após o commit e a submissão para revisão do usuário, se o feedback for positivo, você pode iniciar imediatamente um novo ciclo.

## Fluxo de Trabalho de Desenvolvimento Geral (Humano e Agente)

*   **Branching:** Siga um modelo de branching consistente (ex: GitFlow adaptado). Crie branches a partir da `main` (ou `develop` se adotado) para novas features (`feature/nome-da-feature`), correções (`fix/nome-da-correcao`), etc.
*   **Commits:**
    *   Escreva mensagens de commit claras e descritivas, seguindo o padrão [Conventional Commits](https://www.conventionalcommits.org/).
    *   Faça commits pequenos e atômicos, focados em uma única mudança lógica.
*   **Testes (TDD/BDD):**
    *   Escreva ou atualize os testes unitários relevantes (`_test.go` files) para as alterações realizadas, especialmente para a lógica nos pacotes `network`, `neuron`, `neurochemical`, `pulse`, `synaptic`.
    *   Execute os testes (`go test ./...` ou `make test`) para garantir que todas as verificações passem antes de submeter. **Nota: Atualmente, a execução de testes no ambiente do agente está bloqueada (TEST-002).**
*   **Revisões de Código (Pull Requests):**
    *   Ao submeter seu trabalho (simulado via `submit()`), o código passará por um processo de revisão.
    *   Esteja preparado para explicar suas decisões de design e implementação.
*   **Atualização da Documentação:** Se suas alterações impactarem a documentação (incluindo este `AGENTS.md`, o `README.md`, ou arquivos em `/docs`), atualize-os como parte da sua tarefa.

## Gerenciamento Detalhado de Tarefas (Recapitulação)

A gestão de tarefas no projeto CrowNet é realizada através de dois componentes principais, conforme detalhado no "Protocolo Operacional do Agente (Jules)":

1.  **Arquivo de Índice de Tarefas (`/.jules/TASKS.md`):** O backlog principal.
2.  **Arquivos de Detalhe de Tarefa (`/.jules/tasks/TSK-*.md`):** Contêm informações completas sobre cada tarefa.

**Princípios Chave para o Agente:**
*   **Sincronização Inicial:** Sempre comece lendo e analisando o `/.jules/TASKS.md`.
*   **Leitura Obrigatória dos Detalhes:** Antes de planejar ou executar qualquer tarefa selecionada, leia seu arquivo de detalhe correspondente em `/.jules/tasks/`.
*   **Criação de Detalhes:** Ao criar uma nova tarefa (ou sub-tarefa) no índice, crie simultaneamente seu arquivo de detalhe.
*   **Atualização Contínua:** Mantenha o arquivo de detalhe da tarefa em progresso atualizado com notas e decisões.

Manter este sistema de dois níveis atualizado e sincronizado é fundamental.

## Comunicação

*   **Progresso:** Utilize a ferramenta `plan_step_complete()` para marcar o progresso em seu plano. Ao concluir uma tarefa, informe o usuário com `message_user`.
*   **Dúvidas e Bloqueios:** Se encontrar ambiguidades nos requisitos, ou se estiver bloqueado por um problema técnico, não hesite em pedir ajuda usando `request_user_input`. Forneça contexto claro sobre o problema.
*   **Sugestões:** Se tiver sugestões para melhorar o projeto ou o processo de desenvolvimento, sinta-se à vontade para comunicá-las.

## Considerações Específicas

*   **Estilo de Código Go:** Siga as convenções padrão do Go (veja `go fmt`, `go vet`, [Effective Go](https://go.dev/doc/effective_go)). O arquivo `/docs/03_guias/guia_estilo_codigo.md` pode conter diretrizes adicionais.
*   **Tratamento de Erros:** Utilize o tratamento de erros idiomático do Go, retornando erros onde apropriado.
*   **Logging:** Para logs informativos durante a execução, utilize o pacote `fmt` ou `log` do Go. Para persistência de dados da simulação, o pacote `storage` com SQLite é utilizado.
*   **Reprodutibilidade:** A simulação visa ser determinística se a semente do gerador de números aleatórios (`--seed` flag global) for fixada. Mantenha isso em mente ao introduzir nova aleatoriedade.
*   **Modulação de Limiares Neuronais por Neuroquímicos:**
    *   O limiar de disparo de um neurônio (`CurrentFiringThreshold`) é dinamicamente ajustado com base nos níveis de cortisol e dopamina.
    *   O ajuste é feito pela função `ApplyEffectsToNeurons` no pacote `neurochemical`.
    *   **Cortisol:** Modifica o limiar multiplicativamente usando o parâmetro `simParams.FiringThresholdIncreaseOnCort`.
    *   **Dopamina:** Similarmente, modifica o limiar multiplicativamente usando `simParams.FiringThresholdIncreaseOnDopa`.
*   **Documentação Interna:** Comente o código de forma clara, especialmente para lógica complexa nos algoritmos de simulação.

Lembre-se que este documento é vivo e pode ser atualizado conforme o projeto evolui. Consulte-o regularmente. Boa codificação!
