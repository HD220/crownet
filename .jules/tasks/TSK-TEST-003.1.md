# Tarefa: TEST-003.1 - Implementar testes de integração para CLI modo `expose`.

**ID da Tarefa:** TEST-003.1
**Título Breve:** Testes de integração para CLI `expose` mode.
**Descrição Completa:** Desenvolver testes de integração que validem o funcionamento de ponta a ponta do modo `expose` da CLI. Isso inclui verificar se a execução com parâmetros válidos resulta na criação ou atualização correta de arquivos de pesos sinápticos, e se o processo de treinamento (simulação de épocas e apresentação de padrões) ocorre conforme esperado.
**Status:** Pendente
**Dependências (IDs):** TEST-003, TEST-002
**Complexidade (1-5):** 2
**Prioridade (P0-P4):** P2
**Responsável:** AgenteJules
**Data de Criação:** 2024-07-28
**Data de Conclusão (Estimada/Real):** AAAA-MM-DD
**Branch Git Proposta:** test/integration-expose-mode
**Critérios de Aceitação:**
- Um teste de integração é implementado que executa o comando `crownet expose` com um conjunto mínimo de parâmetros válidos.
- O teste verifica se um arquivo de pesos é criado (se não existir) ou modificado (se existir) após a execução.
- O teste pode opcionalmente verificar se o conteúdo do arquivo de pesos parece razoável (ex: não está vazio, formato JSON válido), sem necessidade de validar a lógica de aprendizado em si.
- O teste utiliza configurações de simulação simplificadas para rodar rapidamente (poucas épocas, poucos ciclos por padrão).
- O teste limpa quaisquer artefatos criados (ex: arquivos de pesos temporários) após sua execução.
**Notas/Decisões:**
- Considerar o uso de um diretório temporário para os artefatos de teste.
- Pode ser necessário mockar ou controlar o gerador de dados de padrão (`datagen`) para ter inputs consistentes.
- Este teste foca na integração da CLI e na mecânica de arquivos, não na validação profunda do algoritmo de aprendizado.
