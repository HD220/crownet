# Documentação de Funcionalidades do CrowNet

Este diretório contém documentos que descrevem as principais funcionalidades do simulador de rede neural CrowNet, detalhando **o que** o sistema faz do ponto de vista do usuário ou do sistema para cada capacidade chave.

## Documentos Principais de Funcionalidades:

*   **[01-inicializacao-rede.md](./01-inicializacao-rede.md):** Descreve o processo de configuração e inicialização da rede neural antes de qualquer simulação.
*   **[02-ciclo-simulacao-aprendizado.md](./02-ciclo-simulacao-aprendizado.md):** Detalha as dinâmicas centrais que ocorrem a cada ciclo de simulação, incluindo atualização neuronal, propagação de pulso, aprendizado, sinaptogênese e neuromodulação.
*   **[03-entrada-saida-dados.md](./03-entrada-saida-dados.md):** Explica como os dados de entrada são preparados e apresentados à rede, como a resposta da rede é representada e que feedback o sistema fornece ao usuário.
*   **[04-modos-operacao.md](./04-modos-operacao.md):** Fornece um sumário dos diferentes modos de operação da CLI (`expose`, `observe`, `sim`), suas configurações chave e links para casos de uso detalhados.
*   **[05-persistencia-dados.md](./05-persistencia-dados.md):** Descreve os mecanismos de persistência de dados, incluindo o salvamento/carregamento de pesos sinápticos (JSON) e o logging opcional do estado da rede (SQLite).

## Casos de Uso Detalhados:

O subdiretório [**casos-de-uso/**](./casos-de-uso/) contém descrições detalhadas dos fluxos de interação para cada modo de operação principal da CLI:

*   `uc-expose.md`: Treinamento da rede.
*   `uc-observe.md`: Observação da resposta da rede.
*   `uc-sim.md`: Execução de simulação geral.

Estes documentos visam fornecer uma compreensão clara das capacidades do CrowNet. Para informações sobre a arquitetura de software, consulte `../02_arquitetura.md`, e para guias práticos (configuração, estilo de código), veja o diretório `../03_guias/`.
