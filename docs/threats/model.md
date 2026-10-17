# Modelo de ameaças da Guarita — v1

- **Versão:** v1 com a ameaça do desafio do Capítulo 3 (*tag* `cap03-desafio`).
- **Método:** zonas e fronteiras de confiança, STRIDE por ameaça e árvores de ataque em [`trees.md`](trees.md).
- **Fonte de verdade:** a tabela de ameaças deste arquivo e a função `threatmodel.Model()` precisam coincidir; o teste `TestModelo_coincideComODocumento` quebra a CI quando divergem.

## Escopo

A Guarita é um proxy MCP entre um agente e os servidores MCP que ele usa. Na v1, ela ainda não intercepta nada: o modelo descreve o sistema que o livro vai construir, para que cada defesa tenha uma ameaça e um capítulo antes de existir código de transporte.

Fora do escopo: a segurança do modelo de linguagem em si, do sistema operacional do host e de serviços de terceiros alcançados pelos servidores MCP.

## Ativos

| Ativo | Por que importa |
|---|---|
| Credenciais da Guarita para os servidores | Quem as controla age em nome da Guarita |
| *Tokens* do usuário | Não podem sair da fronteira em que foram emitidos |
| Catálogo aprovado de ferramentas | Base das decisões de política |
| Argumentos e respostas das chamadas | Podem conter dados pessoais e segredos |
| Políticas | Decidem o que o agente pode fazer |
| Trilha de auditoria | Prova do que aconteceu |
| Disponibilidade e cota | Laços de chamadas consomem recursos e dinheiro |

## Zonas de confiança

O identificador de cada zona é o valor de `TrustZone.String()`.

| Zona | Identificador | Confiança |
|---|---|---|
| Agente | `agent` | Não confiável: segue o que o modelo propõe |
| Guarita | `guarita` | Confiável: é onde as decisões são aplicadas |
| Servidor MCP | `mcp-server` | Não confiável: escrito e operado por terceiros |
| Operador humano | `operator` | Confiável para aprovar, sujeito a fadiga e engano |
| Modelo local | `local-model` | Não confiável: usado como classificador a partir do Capítulo 15 |
| Armazenamento | `storage` | Confiável, mas precisa resistir a adulteração |

## Fronteiras de confiança

Uma fronteira é escrita `origem->destino`, como `TrustBoundary.String()`.

| Fronteira | O que atravessa |
|---|---|
| `agent->guarita` | Pedidos de chamada, argumentos gerados pelo modelo, mensagens do protocolo |
| `guarita->mcp-server` | Chamadas repassadas, credenciais e o processo isolado do servidor |
| `mcp-server->guarita` | Definições de ferramentas e respostas das chamadas |
| `operator->guarita` | Aprovações de chamadas pausadas |
| `operator->mcp-server` | Instalação e atualização de servidores |
| `guarita->storage` | Catálogo, políticas e entradas da trilha de auditoria |

## Ameaças

Estado `open` significa que a Guarita ainda não tem a defesa. A conferência opt-in de descrições do agente de laboratório (Capítulo 1) fica fora da Guarita e não muda nenhum estado.

| ID | Ameaça | STRIDE | Fronteira | Capítulo | Estado |
|---|---|---|---|---|---|
| `tool-poisoning` | Envenenamento de ferramenta | tampering, elevation-of-privilege | mcp-server->guarita | 10 | open |
| `rug-pull` | Troca silenciosa da definição aprovada | tampering | mcp-server->guarita | 9 | open |
| `tool-shadowing` | Sombreamento de ferramenta por nome repetido | spoofing | mcp-server->guarita | 10 | open |
| `indirect-prompt-injection` | Injeção de *prompt* indireta na resposta | tampering, elevation-of-privilege | mcp-server->guarita | 14 | open |
| `excessive-agency` | Agência excessiva | elevation-of-privilege | agent->guarita | 11 | open |
| `data-exfiltration` | Exfiltração de dados nos argumentos | information-disclosure | guarita->mcp-server | 13 | open |
| `sandbox-escape` | Fuga da sandbox do servidor | elevation-of-privilege | guarita->mcp-server | 17 | open |
| `credential-theft` | Repasse ou roubo de credenciais (*confused deputy*) | spoofing, information-disclosure | guarita->mcp-server | 8 | open |
| `supply-chain` | Pacote de servidor comprometido | tampering | operator->mcp-server | 20 | open |
| `resource-abuse` | Laços de chamadas que esgotam cota ou custo | denial-of-service | agent->guarita | 18 | open |
| `insecure-parsing` | Mensagem malformada ou ambígua no protocolo | tampering, denial-of-service | agent->guarita | 6 | open |
| `approval-forgery` | Aprovação forjada ou obtida por fadiga | spoofing, repudiation | operator->guarita | 12 | open |
| `audit-secret-leak` | Segredos gravados na trilha de auditoria | information-disclosure | guarita->storage | 19 | open |

## A cena do Capítulo 3: o *confused deputy*

A Guarita guarda credenciais próprias para falar com os servidores (fronteira `guarita->mcp-server`). O agente, que não é confiável, escolhe servidor, ferramenta e argumentos (fronteira `agent->guarita`). Sem uma decisão explícita no meio, o agente consegue fazer a Guarita usar o privilégio dela num pedido que a Guarita jamais faria por conta própria. A ameaça é `credential-theft`, e o desenho já mostra a defesa: credencial por servidor, nunca o *token* do usuário, e política entre a escolha do agente e o uso da credencial (Capítulos 8 e 11).

## Como atualizar

1. Toda ameaça nova entra na tabela e em `threatmodel.Model()` no mesmo commit.
2. O estado só muda para `partially-mitigated` ou `mitigated` com teste de ataque na suíte do capítulo que trouxe a defesa.
3. Mudanças de zona ou de fronteira exigem revisão das árvores em `trees.md`.
