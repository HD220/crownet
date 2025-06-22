# Guia de Estilo de Código - CrowNet

## 1. Introdução

Este guia estabelece as convenções de estilo de código, formatação e princípios de qualidade que devem ser seguidos no desenvolvimento do projeto CrowNet. A adesão a este guia visa garantir que o código-fonte seja legível, manutenível, consistente e de alta qualidade.

Este documento é uma referência viva e pode ser atualizado conforme o projeto evolui.

## 2. Formatação Automática

Todo o código Go no projeto **deve** ser formatado usando as ferramentas padrão da linguagem Go.

*   **`gofmt`**: É a ferramenta primária para formatação de código Go. Garante um estilo visual uniforme.
*   **`goimports`**: É uma extensão do `gofmt` que, além de formatar o código, também gerencia automaticamente as declarações de `import`, adicionando as faltantes e removendo as não utilizadas. Recomenda-se o uso do `goimports`.

Muitos editores e IDEs podem ser configurados para executar `goimports` (ou `gofmt`) automaticamente ao salvar um arquivo. Isso é altamente encorajado.

**Exemplo de uso manual:**
```bash
goimports -w .
# ou
gofmt -w .
```
Execute na raiz do projeto ou em diretórios específicos para formatar os arquivos.

## 3. Convenções de Nomenclatura

Siga as convenções de nomenclatura padrão da comunidade Go, conforme descrito em [Effective Go](https://go.dev/doc/effective_go#names). Alguns pontos chave:

*   **Pacotes:** Nomes curtos, concisos, em minúsculas e de uma única palavra (ex: `network`, `neuron`). Evite sublinhados ou `mixedCaps`.
*   **Variáveis, Funções, Métodos e Tipos:** Use `camelCase` (ex: `myVariable`, `calculateValue`).
*   **Exportados vs. Não Exportados:** Nomes que começam com letra maiúscula são exportados (públicos). Nomes que começam com letra minúscula não são exportados (privados ao pacote).
*   **Interfaces:** Geralmente não possuem prefixo `I` ou sufixo `Interface`. Frequentemente, são nomeadas com o sufixo "-er" (ex: `Reader`, `Writer`) ou descrevem o comportamento.
*   **Constantes:** Use `camelCase` como outras variáveis. Se forem exportadas, a primeira letra é maiúscula.
*   **Acrônimos:** Devem ser todos em maiúsculas (ex: `ServeHTTP`, `UserID`) ou todos em minúsculas se não exportados e no início de um nome (ex: `userID`).

## 4. Princípios de Código Limpo

O projeto CrowNet adere aos princípios de Código Limpo, conforme popularizado por Robert C. Martin. Os desenvolvedores devem se esforçar para:

*   **Nomes Significativos:** Escolha nomes claros e autoexplicativos para variáveis, funções, classes, etc., que revelem a intenção.
*   **Funções Pequenas e Focadas:** Funções devem ser curtas e fazer apenas uma coisa. Devem ter poucos argumentos.
*   **Evitar Duplicação (DRY - Don't Repeat Yourself):** Generalize e consolide código comum para evitar redundância.
*   **Manter a Simplicidade (KISS - Keep It Simple, Stupid):** Prefira soluções simples e diretas em vez de complexidade desnecessária.
*   **Baixo Acoplamento, Alta Coesão:** Módulos devem ser independentes e componentes dentro de um módulo devem ter responsabilidades relacionadas e bem definidas.
*   **Lei de Demeter (Princípio do Menor Conhecimento):** Um módulo não deve saber sobre os detalhes internos dos objetos que manipula.

## 5. Object Calisthenics

Conforme especificado nos requisitos do projeto (RNF-MAINT-004), o código deve, na medida do possível e pragmaticamente, seguir as regras do Object Calisthenics. Para Go, algumas dessas regras são interpretadas da seguinte forma:

1.  **Um nível de indentação por método/função:** Encoraja funções menores e extração de lógica para outras funções.
2.  **Não use a palavra-chave `else`:** Use retornos antecipados (early returns/guard clauses) ou polimorfismo (interfaces) para evitar blocos `else`.
3.  **Encapsule todos os primitivos e strings:** Use tipos específicos (value objects) em vez de tipos primitivos soltos quando eles representarem um conceito do domínio (ex: `type NeuronID int` em vez de `int` para um ID de neurônio).
4.  **Coleções de primeira classe:** Se um pacote manipula uma coleção de forma significativa, essa coleção deve ser encapsulada em seu próprio tipo com comportamentos associados.
5.  **Um ponto por linha:** Refere-se a não encadear múltiplas chamadas de método/acessos a campos em uma única linha (menos aplicável diretamente a Go, mas o espírito é manter expressões simples).
6.  **Não abrevie:** Use nomes completos e descritivos. Exceções podem ser feitas para variáveis de escopo muito curto e bem compreendido (ex: `i` em loops).
7.  **Mantenha todas as entidades pequenas:** Funções devem ser pequenas (ver Código Limpo). Structs/tipos devem ter poucas instâncias de variáveis. Pacotes devem ser coesos e não excessivamente grandes.
8.  **Não mais que duas variáveis de instância por tipo (struct):** Esta é uma das regras mais desafiadoras e pode precisar de interpretação pragmática em Go, mas o objetivo é promover alta coesão e responsabilidade única.
9.  **Não use getters e setters (propriedades):** Em Go, isso se traduz em ter cuidado com a exportação excessiva de campos de structs. Se um campo é exportado, ele é acessível diretamente. Se não, o acesso é controlado por métodos (que não devem ser apenas getters/setters triviais, mas sim representar comportamentos).

A aplicação dessas regras deve ser feita com bom senso, visando melhorar a qualidade do código sem levar a uma complexidade artificial.

## 6. Comentários

Comentários são importantes, mas o melhor código é aquele que se auto-documenta através de nomes claros e boa estrutura.

*   **Comentários de Pacote:** Todo pacote deve ter um comentário de pacote (`// package <nome> ...` ou `/* package <nome> ... */`) explicando seu propósito e responsabilidade.
*   **Comentários em Funções/Métodos Exportados:** Todas as funções, métodos e tipos exportados devem ter um comentário explicando seu propósito, parâmetros e valores de retorno (seguindo o formato Godoc).
*   **Comentários Explicativos:** Use comentários para explicar lógica complexa, decisões de design não óbvias ou para alertar sobre possíveis armadilhas.
*   **Evite Comentários Óbvios:** Não comente código que já é claro por si só.
*   **TODOs e FIXMEs:** Use `// TODO:` para tarefas pendentes e `// FIXME:` para problemas conhecidos que precisam ser corrigidos. Inclua uma breve explicação.

## 7. Linters

O uso de um linter abrangente como `golangci-lint` é altamente recomendado para identificar problemas de estilo, bugs potenciais e inconsistências no código.

Atualmente, o projeto não possui um arquivo de configuração dedicado para `golangci-lint` (ex: `.golangci.yml`) versionado no repositório. A execução do `golangci-lint` usaria suas configurações padrão ou aquelas definidas localmente pelo desenvolvedor. Considera-se para o futuro a adição de um arquivo `.golangci.yml` padronizado para o projeto.

**Exemplo de uso (com configuração padrão ou local):**
```bash
golangci-lint run ./...
```

## 8. Referências Adicionais

*   [Effective Go](https://go.dev/doc/effective_go)
*   [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
*   Princípios de Código Limpo (Clean Code - Robert C. Martin)
*   Object Calisthenics (Jeff Bay)

Este guia ajudará a manter a base de código do CrowNet consistente e de alta qualidade. Contribuições para melhorar este guia são bem-vindas.
