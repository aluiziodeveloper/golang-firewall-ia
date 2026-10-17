# ADR 0003: Portas no consumidor e regra de dependência verificada

- **Estado:** aceita
- **Data:** 2026-10-17
- **Contexto:** Capítulo 4; detalha a arquitetura hexagonal decidida no ADR 0001

## Contexto

O ADR 0001 decidiu que o domínio e os casos de uso não dependem do transporte, e que o MCP fica nos adaptadores. Até o Capítulo 3, essa regra existia só no texto. Nada impedia o domínio de importar um SDK de infraestrutura, e um import desses compila e passa nos testes.

A Guarita também precisava do ponto que o Capítulo 2 deixou pendente: o lugar por onde toda chamada passa, valida e recebe uma decisão antes de chegar a um servidor.

## Decisão

1. O caso de uso `InterceptCall`, em `internal/app`, valida a chamada, consulta a política e só repassa ao servidor com `DecisionAllow` validado. Erro da política e decisão desconhecida negam a chamada.
2. As portas `PolicyEvaluator` e `Upstream` são declaradas em `internal/app`, onde são consumidas; os adaptadores as implementam e serão escolhidos em `cmd/guarita`, a raiz de composição, quando o proxy existir (Capítulo 5).
3. O domínio fica em `internal/domain`.
4. O `depguard` usa lista de permissão: `internal/domain` só importa a biblioteca padrão, sem `net` nem `os`; `internal/app` só importa a biblioteca padrão e o domínio. Os arquivos de teste ficam fora das regras.

## Alternativas consideradas

1. **Confiar na revisão de código.** Rejeitada: a regra some na primeira pressa, e o Capítulo 4 reproduz exatamente esse import.
2. **Lista de bloqueio no `depguard`.** Adotada e revertida no mesmo capítulo: ela barra só os pacotes citados, e o SDK MCP passou por ela.
3. **Interfaces declaradas pelos adaptadores.** Rejeitada: o adaptador definiria o contrato do caso de uso, e cada mudança no adaptador mudaria o núcleo.
4. **Repassar tudo que não for negado.** Rejeitada: falha aberta; o exemplo `examples/failopen` mostra a decisão desconhecida chegando ao servidor.

## Consequências

- O caso de uso é testado sem rede, com dublês das portas.
- Cada adaptador novo precisa caber numa porta existente ou justificar uma porta nova.
- A lista de permissão pode precisar crescer quando o domínio usar um pacote novo; essa mudança passa por revisão, porque altera a regra.
- A regra cobre imports, não comportamento: um adaptador mal escrito continua possível, e é papel dos testes de cada adaptador.
