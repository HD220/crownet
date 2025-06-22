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
*   **`cli/`:** Lógica de orquestração da interface de linha de comando (`orchestrator.go`).
*   **`common/`:** Tipos de dados comuns e utilitários básicos.
*   **`config/`:** Definição e carregamento de configurações da aplicação e simulação (`config.go`).
*   **`datagen/`:** Geração de dados de entrada, como os padrões de dígitos (`digits.go`).
*   **`docs/`:** Documentação do projeto (visão geral, arquitetura, requisitos, guias, etc.).
*   **`network/`:** Lógica central da rede neural, incluindo o ciclo de simulação, aprendizado e sinaptogênese (`network.go`).
*   **`neurochemical/`:** Simulação dos neuroquímicos (cortisol, dopamina) e seus efeitos (`neurochemicals.go`).
*   **`neuron/`:** Definição da estrutura do neurônio, seus estados e comportamento individual (`neuron.go`, `enums.go`).
*   **`pulse/`:** Gerenciamento da propagação de pulsos na rede (`pulse.go`).
*   **`space/`:** Funções relacionadas ao espaço 16D, como cálculos de distância (`geometry.go`).
*   **`storage/`:** Lógica para persistência de dados, como salvar/carregar pesos sinápticos (JSON) e logs de simulação (SQLite) (`json_persistence.go`, `sqlite_logger.go`).
*   **`synaptic/`:** Gerenciamento dos pesos sinápticos (`weights.go`).

Arquivos com sufixo `_test.go` contêm testes unitários para os respectivos pacotes.

## Tecnologias Chave

*   **Linguagem Principal:** Go (versão indicada no `go.mod`, atualmente `1.24.3`).
*   **Gerenciamento de Dependências:** Go Modules (`go.mod`, `go.sum`).
*   **Persistência de Pesos:** Arquivos JSON.
*   **Logging Detalhado (Opcional):** SQLite (usando a biblioteca `github.com/mattn/go-sqlite3`).
*   **Testes:** Testes unitários padrão do Go (pacote `testing`).

## Fluxo de Trabalho de Desenvolvimento

*   **Branching:** Siga um modelo de branching consistente (ex: GitFlow adaptado). Crie branches a partir da `main` (ou `develop` se adotado) para novas features (`feature/nome-da-feature`), correções (`fix/nome-da-correcao`), etc.
*   **Commits:**
    *   Escreva mensagens de commit claras e descritivas, seguindo o padrão [Conventional Commits](https://www.conventionalcommits.org/).
    *   Faça commits pequenos e atômicos, focados em uma única mudança lógica.
*   **Testes (TDD/BDD):**
    *   Escreva ou atualize os testes unitários relevantes (`_test.go` files) para as alterações realizadas, especialmente para a lógica nos pacotes `network`, `neuron`, `neurochemical`, `pulse`, `synaptic`.
    *   Execute os testes (`go test ./...`) para garantir que todas as verificações passem antes de submeter.
*   **Revisões de Código (Pull Requests):**
    *   Ao submeter seu trabalho (simulado via `submit()`), o código passará por um processo de revisão.
    *   Esteja preparado para explicar suas decisões de design e implementação.
*   **Atualização da Documentação:** Se suas alterações impactarem a documentação (incluindo este `AGENTS.md`, o `README.md`, ou arquivos em `/docs`), atualize-os como parte da sua tarefa.

## Comunicação

*   **Progresso:** Utilize a ferramenta `plan_step_complete()` para marcar o progresso em seu plano. Ao concluir uma tarefa, informe o usuário com `message_user`.
*   **Dúvidas e Bloqueios:** Se encontrar ambiguidades nos requisitos, ou se estiver bloqueado por um problema técnico, não hesite em pedir ajuda usando `request_user_input`. Forneça contexto claro sobre o problema.
*   **Sugestões:** Se tiver sugestões para melhorar o projeto ou o processo de desenvolvimento, sinta-se à vontade para comunicá-las.

## Considerações Específicas

*   **Estilo de Código Go:** Siga as convenções padrão do Go (veja `go fmt`, `go vet`, [Effective Go](https://go.dev/doc/effective_go)). O arquivo `/docs/03_guias/guia_estilo_codigo.md` pode conter diretrizes adicionais.
*   **Tratamento de Erros:** Utilize o tratamento de erros idiomático do Go, retornando erros onde apropriado.
*   **Logging:** Para logs informativos durante a execução, utilize o pacote `fmt` ou `log` do Go. Para persistência de dados da simulação, o pacote `storage` com SQLite é utilizado.
*   **Reprodutibilidade:** A simulação visa ser determinística se a semente do gerador de números aleatórios (`-seed` flag) for fixada. Mantenha isso em mente ao introduzir nova aleatoriedade.
*   **Modulação de Limiares Neuronais por Neuroquímicos:**
    *   O limiar de disparo de um neurônio (`CurrentFiringThreshold`) é dinamicamente ajustado com base nos níveis de cortisol e dopamina. Este valor é derivado do `BaseFiringThreshold` do neurônio.
    *   O ajuste é feito pela função `ApplyEffectsToNeurons` no pacote `neurochemical`.
    *   **Cortisol:** Modifica o limiar multiplicativamente usando o parâmetro `simParams.FiringThresholdIncreaseOnCort`. Um valor positivo deste parâmetro faz com que o cortisol aumente o limiar.
    *   **Dopamina:** Similarmente, modifica o limiar (após o efeito do cortisol) multiplicativamente usando `simParams.FiringThresholdIncreaseOnDopa`.
    *   Ambos os efeitos são diretos e proporcionais aos níveis normalizados dos respectivos neuroquímicos. Notavelmente, qualquer documentação anterior sobre um efeito em "U" do cortisol nos limiares não reflete a implementação atual.
*   **Documentação Interna:** Comente o código de forma clara, especialmente para lógica complexa nos algoritmos de simulação.

Lembre-se que este documento é vivo e pode ser atualizado conforme o projeto evolui. Consulte-o regularmente. Boa codificação!
