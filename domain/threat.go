package domain

import "strconv"

// #region domain-trust-zone
// TrustZone identifies a part of the system with one trust level.
type TrustZone int

const (
	// TrustZoneUnspecified is the zero value and is never valid.
	TrustZoneUnspecified TrustZone = iota
	TrustZoneAgent
	TrustZoneGuarita
	TrustZoneMCPServer
	TrustZoneOperator
	TrustZoneLocalModel
	TrustZoneStorage
)

// #endregion

// String returns the stable name used in the threat model document.
func (zone TrustZone) String() string {
	switch zone {
	case TrustZoneUnspecified:
		return "unspecified"
	case TrustZoneAgent:
		return "agent"
	case TrustZoneGuarita:
		return "guarita"
	case TrustZoneMCPServer:
		return "mcp-server"
	case TrustZoneOperator:
		return "operator"
	case TrustZoneLocalModel:
		return "local-model"
	case TrustZoneStorage:
		return "storage"
	default:
		return "TrustZone(" + strconv.Itoa(int(zone)) + ")"
	}
}

// #region domain-trust-zone-validate
// Validate rejects the zero value and values outside the declared set.
func (zone TrustZone) Validate() error {
	if zone > TrustZoneUnspecified && zone <= TrustZoneStorage {
		return nil
	}

	return &ValidationError{field: "trust_zone", kind: ErrInvalidTrustZone}
}

// #endregion

// TrustBoundary is the crossing point between two trust zones.
type TrustBoundary struct {
	From TrustZone
	To   TrustZone
}

// String returns the boundary as "from->to".
func (boundary TrustBoundary) String() string {
	return boundary.From.String() + "->" + boundary.To.String()
}

// Validate requires two valid and distinct zones.
func (boundary TrustBoundary) Validate() error {
	if boundary.From.Validate() != nil ||
		boundary.To.Validate() != nil ||
		boundary.From == boundary.To {
		return &ValidationError{
			field: "trust_boundary",
			kind:  ErrInvalidTrustBoundary,
		}
	}

	return nil
}

// STRIDECategory classifies a threat by the property it violates.
type STRIDECategory int

const (
	// STRIDEUnspecified is the zero value and is never valid.
	STRIDEUnspecified STRIDECategory = iota
	STRIDESpoofing
	STRIDETampering
	STRIDERepudiation
	STRIDEInformationDisclosure
	STRIDEDenialOfService
	STRIDEElevationOfPrivilege
)

// String returns the stable name used in the threat model document.
func (category STRIDECategory) String() string {
	switch category {
	case STRIDEUnspecified:
		return "unspecified"
	case STRIDESpoofing:
		return "spoofing"
	case STRIDETampering:
		return "tampering"
	case STRIDERepudiation:
		return "repudiation"
	case STRIDEInformationDisclosure:
		return "information-disclosure"
	case STRIDEDenialOfService:
		return "denial-of-service"
	case STRIDEElevationOfPrivilege:
		return "elevation-of-privilege"
	default:
		return "STRIDECategory(" + strconv.Itoa(int(category)) + ")"
	}
}

// Validate rejects the zero value and values outside the declared set.
func (category STRIDECategory) Validate() error {
	if category > STRIDEUnspecified &&
		category <= STRIDEElevationOfPrivilege {
		return nil
	}

	return &ValidationError{
		field: "stride_category",
		kind:  ErrInvalidSTRIDECategory,
	}
}

// ThreatStatus records how far the Guarita mitigates a threat.
type ThreatStatus int

const (
	// ThreatStatusUnspecified is the zero value and is never valid.
	ThreatStatusUnspecified ThreatStatus = iota
	ThreatStatusOpen
	ThreatStatusPartiallyMitigated
	ThreatStatusMitigated
	ThreatStatusAccepted
)

// String returns the stable name used in the threat model document.
func (status ThreatStatus) String() string {
	switch status {
	case ThreatStatusUnspecified:
		return "unspecified"
	case ThreatStatusOpen:
		return "open"
	case ThreatStatusPartiallyMitigated:
		return "partially-mitigated"
	case ThreatStatusMitigated:
		return "mitigated"
	case ThreatStatusAccepted:
		return "accepted"
	default:
		return "ThreatStatus(" + strconv.Itoa(int(status)) + ")"
	}
}

// Validate rejects the zero value and values outside the declared set.
func (status ThreatStatus) Validate() error {
	if status > ThreatStatusUnspecified && status <= ThreatStatusAccepted {
		return nil
	}

	return &ValidationError{
		field: "threat_status",
		kind:  ErrInvalidThreatStatus,
	}
}
