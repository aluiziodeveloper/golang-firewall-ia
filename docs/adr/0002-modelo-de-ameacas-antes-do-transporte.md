# ADR 0002: Modelo de ameaças antes do transporte

- **Estado:** aceita
- **Data:** 2026-10-17
- **Contexto:** Capítulo 3

## Contexto

A tentação de um proxy é começar pelo transporte: aceitar conexões, repassar mensagens e deixar a segurança para depois. Feito assim, cada controle acaba espalhado pelo adaptador que primeiro precisou dele, e ameaças que não tocam o transporte, como o repasse de credenciais ou a fadiga de aprovação, ficam sem dono.

Um modelo de ameaças só em documento envelhece calado. Ninguém percebe quando o código ganha um conceito que o documento não tem, ou quando o documento promete uma defesa que nenhum capítulo entrega.

## Decisão

1. O modelo de ameaças é escrito antes do transporte, em `docs/threats/model.md`, com zonas, fronteiras, ativos e uma tabela de ameaças classificadas por STRIDE, com a fronteira, o capítulo da defesa e o estado de cada uma. As árvores de ataque ficam em `docs/threats/trees.md`.
2. O mesmo modelo existe como dados em `internal/threatmodel`, e `TestModelo_coincideComODocumento` compara as duas versões linha a linha. Qualquer divergência quebra a CI.
3. Os conjuntos fechados do modelo usam `iota` com valor zero inválido, e o linter `exhaustive` exige que todo `switch` sobre eles trate cada valor.
4. Uma ameaça só muda de `open` para `partially-mitigated` ou `mitigated` com um teste de ataque na suíte do capítulo que trouxe a defesa.

## Alternativas consideradas

1. **Começar pelo transporte e modelar depois.** Rejeitada pelo motivo do contexto: controles espalhados e ameaças sem dono.
2. **Manter o modelo só em documento ou planilha.** Rejeitada porque nada avisa quando código e modelo se afastam.
3. **Pôr os tipos do modelo no pacote `domain`.** Foi o primeiro incremento do capítulo e foi revertido. O modelo documenta o projeto e não participa do caminho das chamadas; no domínio, ele ampliaria a API pública com tipos que nenhum caso de uso consome e acoplaria o teste do domínio a um arquivo de documentação.

## Consequências

- Cada capítulo que trouxer uma defesa atualiza a linha da ameaça, no documento e no código, no mesmo commit.
- O modelo é qualitativo: não há pontuação numérica de risco. Priorizar continua sendo uma decisão registrada em ADR.
- O teste de consistência depende do formato da tabela; mudar o formato exige mudar o teste.
- O modelo não prova que uma defesa funciona. Quem prova é o teste de ataque.
