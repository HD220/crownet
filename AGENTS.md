# Guia para Agentes LLM - Projeto [NOME_DO_PROJETO]

## Introdução

Bem-vindo ao projeto `[NOME_DO_PROJETO]`! Este documento serve como o principal guia de orientação para qualquer Agente de Linguagem Grande (LLM) que colabore neste repositório. Nosso objetivo é facilitar uma colaboração eficiente e garantir que suas contribuições estejam alinhadas com as melhores práticas e os objetivos do projeto.

## Visão Geral do Projeto (Placeholder)

`[Esta seção será preenchida com a descrição detalhada do projeto [NOME_DO_PROJETO], seus objetivos principais, público-alvo e o problema que visa resolver.]`

## Princípios Gerais de Desenvolvimento

Para garantir a qualidade e a manutenibilidade do nosso código, aderimos aos seguintes princípios:

*   **Clareza de Código:** Escreva código que seja fácil de entender e manter. Priorize a legibilidade sobre a concisão excessiva. Comente o código quando a lógica não for imediatamente óbvia.
*   **Evite Repetição (DRY - Don't Repeat Yourself):** Generalize e reutilize código sempre que possível para evitar duplicação.
*   **Mantenha a Simplicidade (KISS - Keep It Simple, Stupid):** Procure as soluções mais simples que resolvam o problema de forma eficaz. Evite complexidade desnecessária.
*   **Testabilidade:** Escreva código de forma que seja fácil de testar. Crie testes unitários, de integração e/ou de ponta a ponta para garantir a robustez da aplicação.
*   **Segurança desde o Início:** Considere as implicações de segurança de suas contribuições. Siga as melhores práticas para evitar vulnerabilidades comuns.

## Padrões Arquiteturais (Placeholder)

`[Esta seção descreverá a arquitetura principal adotada pelo projeto (ex: Arquitetura em Camadas, Microserviços, MVC, Event-Driven, etc.). Detalhes adicionais e diagramas podem ser encontrados em [LINK_PARA_DOCUMENTO_DE_ARQUITETURA].]`

## Estrutura de Código Sugerida (Placeholder)

`[Aqui será detalhada a organização das pastas e arquivos do projeto. Por exemplo:

/src
  /components         # Componentes de UI reutilizáveis
  /pages              # Páginas da aplicação (se aplicável)
  /services           # Lógica de negócio, integrações com APIs externas
  /utils              # Funções utilitárias
  /config             # Configurações da aplicação
  /core               # Lógica central do sistema, casos de uso (actions)
  /domain             # Entidades e regras de negócio centrais
  /infra              # Implementações de interfaces (ex: repositórios, gateways)
/tests
  /unit               # Testes unitários
  /integration        # Testes de integração
/docs                 # Documentação do projeto
...etc.
]`

## Tecnologias Chave (Placeholder)

`[Listar as principais linguagens de programação, frameworks, bibliotecas e ferramentas utilizadas no projeto. Exemplo:

*   **Linguagem Principal:** Python 3.10+ / TypeScript 5.x / Java 17
*   **Framework Backend:** FastAPI / NestJS / Spring Boot
*   **Framework Frontend:** React / Vue.js / Angular
*   **Banco de Dados:** PostgreSQL / MongoDB
*   **Testes:** Pytest / Jest / JUnit
*   **Containerização:** Docker
*   **CI/CD:** GitHub Actions
]`

## Fluxo de Trabalho de Desenvolvimento

*   **Branching:** Siga um modelo de branching consistente (ex: GitFlow adaptado). Crie branches a partir da `develop` (ou `main`) para novas features (`feature/nome-da-feature`), correções (`fix/nome-da-correcao`), etc.
*   **Commits:**
    *   Escreva mensagens de commit claras e descritivas, seguindo o padrão [Conventional Commits](https://www.conventionalcommits.org/).
    *   Faça commits pequenos e atômicos, focados em uma única mudança lógica.
*   **Testes (TDD/BDD):**
    *   Incentivamos a escrita de testes antes ou durante o desenvolvimento do código (Test-Driven Development ou Behavior-Driven Development).
    *   Garanta que todos os testes passem antes de submeter seu código.
    *   Para lógica de backend (casos de uso, actions), testes unitários e de integração são cruciais.
*   **Revisões de Código (Pull Requests):**
    *   Ao submeter seu trabalho (simulado via `submit()`), o código passará por um processo de revisão.
    *   Esteja preparado para explicar suas decisões de design e implementação.
*   **Atualização da Documentação:** Se suas alterações impactarem a documentação (incluindo este `AGENTS.md` ou o `TASKS.md`), atualize-os como parte da sua tarefa.

## Comunicação

*   **Progresso:** Utilize a ferramenta `plan_step_complete()` para marcar o progresso em seu plano. Ao concluir uma tarefa, informe o usuário com `message_user`.
*   **Dúvidas e Bloqueios:** Se encontrar ambiguidades nos requisitos, ou se estiver bloqueado por um problema técnico, não hesite em pedir ajuda usando `request_user_input`. Forneça contexto claro sobre o problema.
*   **Sugestões:** Se tiver sugestões para melhorar o projeto ou o processo de desenvolvimento, sinta-se à vontade para comunicá-las.

## Considerações Específicas (Placeholder)

`[Esta seção é reservada para quaisquer regras, convenções ou considerações únicas deste projeto que não se encaixam nas seções anteriores. Por exemplo:

*   **Padrões de Nomenclatura Específicos:** (ex: `camelCase` para variáveis, `PascalCase` para classes).
*   **Limites de Linha de Código:** (ex: máximo de 100 caracteres por linha).
*   **Configurações de Linters/Formatadores:** (Se houver ferramentas específicas como Prettier, ESLint, Black, Flake8, suas configurações devem ser respeitadas).
*   **Internacionalização (i18n):** (Se o projeto requer suporte a múltiplos idiomas).
]`

Lembre-se que este documento é vivo e pode ser atualizado conforme o projeto evolui. Consulte-o regularmente. Boa codificação!
