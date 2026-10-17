# ADR 0006: Transporte HTTP com proxy reverso, guarda de Host e prazos

- **Estado:** aceita
- **Data:** 2026-10-17
- **Contexto:** Capítulo 7; estende o ADR 0005 ao transporte Streamable HTTP

## Contexto

Servidores MCP remotos falam o transporte Streamable HTTP: um único endpoint que aceita POST com uma mensagem JSON-RPC, responde em JSON ou num fluxo de eventos (SSE), e aceita GET para um fluxo iniciado pelo servidor. Para protegê-los, a Guarita precisa escutar na rede. Isso traz um atacante que o stdio não tinha: qualquer página aberta no navegador do usuário.

Depois de um DNS rebinding, o navegador manda ao endereço local um pedido de mesma origem, com `Host` e `Origin` do domínio do atacante. `http.CrossOriginProtection` deixa passar esse pedido, porque ele não é entre origens (`examples/csrfonly`). Um proxy escrito com `io.Copy` segura os eventos do fluxo até o servidor terminar (`examples/bufferedproxy`).

A especificação exige validar `Origin` em todas as conexões, responder 403 quando ele é inválido e recomenda escutar só em loopback quando o servidor é local.

## Decisão

1. O adaptador `mcphttp` usa `httputil.ReverseProxy` com `Rewrite`, que remove os cabeçalhos de salto a salto, não repassa os `X-Forwarded-*` recebidos e libera cada escrita de uma resposta `text/event-stream` assim que ela acontece.
2. Todo POST é lido inteiro, até 4 MiB e dentro de um prazo de leitura, e passa pela mesma porta do ADR 0005 antes de seguir. O servidor recebe os bytes que a porta leu. GET e DELETE seguem direto; outros métodos recebem 405.
3. A guarda confere o `Host` de todo pedido contra uma lista e o `Origin`, quando vem, contra uma lista exata. Sem `-host`, a lista é derivada do endereço em que a Guarita escuta, e só um endereço de loopback é aceito. A recusa é 403 com erro JSON-RPC sem `id`.
4. O padrão de escuta é `127.0.0.1:8080`.
5. Cada POST tem prazo (`-call-timeout`, 30 s por padrão), que inclui a resposta em fluxo. O servidor HTTP tem `ReadHeaderTimeout`, `IdleTimeout` e `MaxHeaderBytes`, e não tem `WriteTimeout`, que cortaria o fluxo longo do GET.
6. O encerramento usa `Shutdown` com prazo de 5 s e `Close` para o que sobrar.

## Alternativas consideradas

1. **Só `http.CrossOriginProtection`.** Rejeitada: protege contra CSRF, não contra rebinding.
2. **Só conferir `Origin`.** Rejeitada: o GET de mesma origem não leva `Origin`, e o fluxo de eventos ficaria aberto à página.
3. **Proxy escrito à mão com `http.Client` e `io.Copy`.** Rejeitada: segura o fluxo e exige reimplementar o tratamento de cabeçalhos de salto a salto.
4. **`WriteTimeout` no servidor.** Rejeitada pelo item 5; o prazo fica por chamada, no handler.

## Consequências

- Clientes de navegador legítimos precisam ser listados em `-origin`.
- Fora do loopback, a proteção depende de uma lista correta em `-host`; a Guarita se recusa a iniciar sem ela.
- O cabeçalho `Authorization` do agente ainda chega ao servidor. Separar a identidade do agente da credencial da Guarita é o ADR do Capítulo 8.
- Um fluxo GET aberto não tem prazo além do encerramento; limites de fluxos por agente ficam para o Capítulo 18.
