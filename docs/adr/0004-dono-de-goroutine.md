# ADR 0004: Toda goroutine tem dono e termina com o contexto

- **Estado:** aceita
- **Data:** 2026-10-17
- **Contexto:** Capítulo 5; primeiro adaptador de entrada da Guarita

## Contexto

O proxy stdio repassa mensagens entre o agente e o servidor MCP. A forma mais curta de escrevê-lo cria uma goroutine por mensagem. Contra um servidor que para de ler, essa forma acumulou mil goroutines presas para mil mensagens, e um contador compartilhado entre elas produziu uma corrida acusada pelo detector. Goroutines não são recolhidas pelo coletor de lixo: precisam terminar por conta própria.

Uma goroutine parada em `Read` ou `Write` também não percebe o cancelamento de um contexto: ela só volta quando a operação de E/S volta.

## Decisão

1. Cada goroutine da Guarita tem um dono, que a inicia e espera o fim dela antes de retornar. No repasse, o dono é um `errgroup.Group` dentro de `mcpproxy.Relay`.
2. O repasse usa uma goroutine por sentido, com escrita síncrona: se o destino não lê, o remetente espera. Não há fila sem limite nem goroutine por mensagem.
3. Quando o contexto termina, uma goroutine do grupo fecha os quatro lados do repasse; é o fechamento que desbloqueia leituras e escritas paradas.
4. Contadores pertencem a uma única goroutine e só são lidos depois de `Wait`.
5. Todo componente com goroutines tem um teste que confere o retorno ao número inicial de goroutines depois do cancelamento.

## Alternativas consideradas

1. **Uma goroutine por mensagem.** Rejeitada pelos números do contexto.
2. **Fila com canal de capacidade fixa.** Adia o problema sem resolvê-lo: quando a fila enche, o remetente espera do mesmo jeito, e a fila guarda mensagens que talvez nunca sejam entregues.
3. **Biblioteca de detecção de vazamento.** Adiada: o verificador de dez linhas com `runtime.NumGoroutine` basta enquanto a Guarita tiver poucos componentes concorrentes.

## Consequências

- Um servidor lento deixa o agente lento, e não a Guarita sem memória. Limitar quanto tempo esperar é papel do Capítulo 18.
- Ao cancelar, o repasse fecha a saída padrão do agente; mensagens em trânsito podem ser perdidas, o que é aceitável no encerramento.
- O repasse ainda é transparente: o Capítulo 6 passa a interpretar as mensagens e a chamar `InterceptCall`.
