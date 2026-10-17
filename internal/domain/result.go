package domain

import "encoding/json"

// Result é o que o servidor devolveu para uma chamada permitida. O
// conteúdo segue codificado: o domínio não interpreta a resposta.
type Result struct {
	Content json.RawMessage
}
