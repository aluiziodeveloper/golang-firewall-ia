#!/usr/bin/env bash
# Laboratório local de DNS rebinding do Capítulo 7.
#
# Sobe o servidor MCP do laboratório e a Guarita em loopback e envia os
# pedidos que um navegador envia depois do rebinding. A troca de DNS é
# simulada por curl --resolve: evil.test passa a apontar para 127.0.0.1,
# e o pedido sai com o Host e a Origin do atacante, como no navegador.
# Nada sai da máquina.
set -euo pipefail

# #region deploy-rebinding-lab
GUARITA=127.0.0.1:18080
SERVER=127.0.0.1:19081
EVIL=evil.test:18080

work=$(mktemp -d)
trap 'kill $(jobs -p) 2>/dev/null || true; rm -rf "$work"' EXIT
go build -o "$work/guarita" ./cmd/guarita
go build -o "$work/lab-mcp" ./cmd/lab-poisoned-mcp

MCP_ADDR=$SERVER "$work/lab-mcp" -mode legitimate 2>/dev/null &
"$work/guarita" -allow read_file -listen "$GUARITA" \
	-upstream "http://$SERVER/mcp" 2>/dev/null &
for _ in $(seq 50); do
	curl -s -o /dev/null "http://$GUARITA/mcp" && break
	sleep 0.1
done
# #endregion

# #region deploy-rebinding-probe
init='{"jsonrpc":"2.0","id":1,"method":"initialize","params":{'
init+='"protocolVersion":"2025-06-18","capabilities":{},'
init+='"clientInfo":{"name":"lab","version":"0"}}}'

# probe nome método url [opções do curl]: imprime o código HTTP.
probe() {
	local name=$1 method=$2 url=$3
	shift 3
	local code
	code=$(curl -s -o /dev/null -w '%{http_code}' -X "$method" \
		-H 'Content-Type: application/json' \
		-H 'Accept: application/json, text/event-stream' \
		"$@" "$url")
	printf '%-44s %s\n' "$name" "$code"
}
# #endregion

# #region deploy-rebinding-probes
probe "agente local, sem Origin" POST "http://$GUARITA/mcp" \
	--data "$init"
probe "rebinding: POST com Host e Origin do atacante" POST \
	"http://$EVIL/mcp" --resolve "$EVIL:127.0.0.1" \
	-H "Origin: http://$EVIL" -H 'Sec-Fetch-Site: same-origin' \
	--data "$init"
probe "rebinding: GET do fluxo, sem Origin" GET \
	"http://$EVIL/mcp" --resolve "$EVIL:127.0.0.1" \
	-H 'Sec-Fetch-Site: same-origin' --max-time 2
probe "origem parecida: http://localhost.evil.test" POST \
	"http://$GUARITA/mcp" -H 'Origin: http://localhost.evil.test' \
	--data "$init"
# #endregion

# Contraste: o mesmo rebinding direto no servidor, sem a Guarita. O SDK
# Go do MCP, na v1.8.0, já recusa Host fora do loopback; outros não.
probe "sem Guarita: rebinding direto no servidor" POST \
	"http://evil.test:19081/mcp" --resolve "evil.test:19081:127.0.0.1" \
	-H 'Origin: http://evil.test:19081' --data "$init"
