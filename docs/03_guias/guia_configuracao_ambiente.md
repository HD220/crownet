# Guia de Configuração do Ambiente de Desenvolvimento - CrowNet

## 1. Introdução

Este guia descreve os passos necessários para configurar o ambiente de desenvolvimento para o projeto CrowNet em sua máquina local. Isso permitirá que você compile, execute e modifique o código-fonte do simulador.

## 2. Pré-requisitos

Para compilar e executar o CrowNet, você precisará ter o seguinte software instalado:

*   **Go:** Versão 1.18 ou superior. O CrowNet é escrito em Go e utiliza Go Modules para gerenciamento de dependências. Você pode baixar o Go em [golang.org/dl](https://golang.org/dl/).

Verifique sua instalação do Go com:
```bash
go version
```

## 3. Obtendo o Código-Fonte

A maneira mais comum de obter o código-fonte é clonando o repositório Git:

```bash
# Se você estiver usando HTTPS
git clone <URL_DO_REPOSITORIO_HTTPS> crownet
# Ou se você estiver usando SSH
# git clone <URL_DO_REPOSITORIO_SSH> crownet

cd crownet
```
*Nota: Substitua `<URL_DO_REPOSITORIO_...>` pela URL correta do repositório.*

Se você já possui o código-fonte (por exemplo, baixado como um arquivo ZIP), apenas navegue até o diretório raiz do projeto.

## 4. Instalando Dependências

O CrowNet utiliza Go Modules para gerenciar suas dependências. As dependências são listadas no arquivo `go.mod`.

Para baixar e instalar as dependências, execute o seguinte comando na raiz do projeto:
```bash
go mod download
```
Alternativamente, o comando `go build` (no próximo passo) também cuidará de baixar as dependências necessárias se elas ainda não estiverem presentes.

## 5. Construindo o Projeto (Compilação)

Para compilar o projeto e criar o executável `crownet` (ou `crownet.exe` no Windows), execute o seguinte comando na raiz do projeto:

```bash
go build .
```
Isso irá gerar o binário executável no mesmo diretório.

## 6. Executando o CrowNet

Após a compilação bem-sucedida, você pode executar o CrowNet diretamente do terminal. A aplicação é controlada por flags de linha de comando.

Consulte o `guia_interface_linha_comando.md` para uma lista completa de modos e flags.

**Exemplo de execução do modo `expose` (treinamento):**
```bash
./crownet -mode expose -neurons 150 -epochs 20 -lrBase 0.005 -cyclesPerPattern 5 -weightsFile my_digit_weights.json
```

**Exemplo de execução do modo `observe` (teste):**
```bash
./crownet -mode observe -digit 7 -weightsFile my_digit_weights.json -cyclesToSettle 5
```

## 7. Observações Adicionais

*   **IDE/Editor:** Para desenvolvimento em Go, você pode usar qualquer editor de texto ou IDE de sua preferência. Algumas opções populares com bom suporte para Go incluem:
    *   Visual Studio Code com a extensão Go.
    *   GoLand (IDE comercial da JetBrains).
*   **Formatação de Código:** O projeto deve seguir as convenções padrão de formatação do Go (geralmente aplicadas com `gofmt` ou `goimports`). Consulte o `guia_estilo_codigo.md` para mais detalhes.
*   **Versão do Go:** Embora o guia mencione Go 1.18+, o projeto está atualmente configurado para `go 1.24.3` (conforme `go.mod`). Recomenda-se usar esta versão ou superior para garantir compatibilidade.
*   **URL do Repositório:** No passo de clonagem, substitua `<URL_DO_REPOSITORIO_HTTPS>` ou `<URL_DO_REPOSITORIO_SSH>` pela URL correta fornecida para o projeto.
*   **Variáveis de Ambiente:** Atualmente, o MVP do CrowNet não depende de variáveis de ambiente específicas para sua execução básica, mas isso pode mudar em versões futuras.

Se encontrar problemas durante a configuração, verifique se o Go está corretamente instalado e configurado em seu `PATH`.
