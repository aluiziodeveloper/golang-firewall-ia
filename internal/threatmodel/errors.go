package threatmodel

import "errors"

var (
	// ErrInvalidTrustZone identifica uma zona fora do conjunto declarado.
	ErrInvalidTrustZone = errors.New("trust zone is not supported")
	// ErrInvalidTrustBoundary identifica uma fronteira sem duas zonas
	// válidas e distintas.
	ErrInvalidTrustBoundary = errors.New("trust boundary is not valid")
	// ErrInvalidSTRIDECategory identifica uma categoria fora do STRIDE.
	ErrInvalidSTRIDECategory = errors.New("STRIDE category not supported")
	// ErrInvalidThreatStatus identifica um estado fora do conjunto.
	ErrInvalidThreatStatus = errors.New("threat status is not supported")
)
