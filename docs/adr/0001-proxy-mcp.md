# ADR 0001: Guarita como proxy MCP

- **Estado:** aceita para a implementação didática
- **Data:** 2026-10-10
- **Contexto:** Capítulo 1 e decisão arquitetural da seção 13.2 do plano mestre

## Contexto

Agentes descobrem e chamam ferramentas oferecidas por servidores MCP, mas o protocolo não substitui uma fronteira de segurança que avalie essas chamadas. As proteções da Guarita precisam permanecer fora do modelo e ser exercidas no caminho entre o agente e os servidores MCP.

O repositório deve permitir testar o domínio sem rede e acomodar mudanças do protocolo sem espalhar dependências MCP pela aplicação.

## Decisão

A Guarita será implementada como um proxy MCP entre o agente e os servidores upstream. A arquitetura será hexagonal: o domínio e os casos de uso ficam independentes do transporte, e o protocolo MCP é implementado nos adaptadores por meio do SDK oficial para Go, cuja versão será fixada no capítulo que introduzir o SDK.

O binário inicial deste capítulo é apenas a fundação do executável. Ele não encaminha tráfego nem aplica defesas; a implementação dessas capacidades ocorre nos capítulos correspondentes.

## Alternativas consideradas

1. **Confiar no agente para aplicar as regras de segurança.** Rejeitada porque o agente e as instruções que recebe não constituem uma fronteira de aplicação confiável.
2. **Implementar controles específicos em cada agente ou aplicação cliente.** Rejeitada como arquitetura principal porque duplica políticas e deixa caminhos de chamada sem um ponto comum de aplicação.
3. **Acoplar a lógica de segurança diretamente ao SDK MCP.** Rejeitada porque torna o domínio dependente do transporte e dificulta testes isolados e evolução do protocolo.

## Consequências

- Há um ponto externo ao modelo para aplicar políticas às chamadas de ferramentas.
- O domínio pode ser testado sem rede; detalhes MCP ficam nos adaptadores.
- A versão do SDK e os limites de compatibilidade precisam ser revistos quando o protocolo ou o SDK evoluírem.
- Um proxy não torna uma decisão de segurança infalível; validação, isolamento e testes continuam necessários.
- A latência, as falhas de upstream e os limites do protocolo precisarão ser tratados e medidos nos capítulos de implementação.

## Referências

- Plano mestre, seção 13.2: arquitetura hexagonal em Go e SDK oficial MCP.
- Briefing aprovado do Capítulo 1: a demonstração do ataque é local e não apresenta a Guarita como defesa já implementada.
