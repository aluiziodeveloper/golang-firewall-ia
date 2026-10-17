# ADR 0005: O protocolo MCP fica numa camada anticorrupção

- **Estado:** aceita
- **Data:** 2026-10-17
- **Contexto:** Capítulo 6; detalha o ADR 0001 para o transporte stdio

## Contexto

Até o Capítulo 5, a Guarita repassava bytes sem interpretá-los. Uma linha com duas mensagens coladas passava como uma, e uma mensagem acima do limite do `bufio.Scanner` derrubava a sessão inteira. Para decidir sobre `tools/call`, a Guarita precisa ler as mensagens; para continuar testável e sobreviver a mudanças do MCP, o domínio não pode conhecê-las.

Uma porta que decide pelo primeiro valor da linha e repassa a linha inteira deixa passar a mensagem contrabandeada; o exemplo `examples/firstvalue` mostra o servidor executando o `delete_file` escondido atrás de um `tools/list`.

## Decisão

1. Todo o conhecimento do JSON-RPC e do MCP fica em `internal/adapters/mcpproxy`. O adaptador traduz `tools/call` para `domain.Call` e entrega a decisão a `InterceptCall.Authorize`; o domínio e o caso de uso não conhecem `jsonrpc`, `id` nem `params`.
2. O parser aceita exatamente uma mensagem JSON-RPC 2.0 por linha: UTF-8 válido, sem dados depois do valor, sem chave repetida, sem lote, com aninhamento até 64 níveis e até 4 MiB. Uma mensagem fora dessas regras é recusada inteira, e o agente recebe um erro JSON-RPC.
3. `tools/call` sem `id` é recusado, porque escaparia da política e não teria como receber a negação.
4. O que a política nega volta ao agente como erro `-32001`, sem chegar ao servidor.
5. Uma mensagem acima do limite é descartada sozinha; a sessão continua.
6. O parser tem arquivos golden e fuzzing, e a CI roda o fuzzing por 20 segundos a cada push.

## Alternativas consideradas

1. **Usar o SDK MCP como parser no proxy.** Adiada: o SDK é escrito para ser cliente ou servidor, não intermediário, e esconderia as regras de recusa que a Guarita precisa controlar e testar.
2. **Ler o primeiro valor e repassar a linha.** Rejeitada pelo contrabando.
3. **Aceitar chave repetida ficando com o último valor.** Rejeitada: outro parser pode ficar com o primeiro, e a decisão e a execução leriam ferramentas diferentes.

## Consequências

- Recusar chaves repetidas e lotes pode rejeitar clientes que dependam deles; a especificação removeu os lotes na revisão 2025-06-18.
- As respostas do servidor ainda passam sem interpretação; aplicar o parser a elas é o desafio do Capítulo 6, e a inspeção de conteúdo, o Capítulo 14.
- Ampliar o limite de 4 MiB ou o de aninhamento é mudança de configuração que precisa de revisão.
