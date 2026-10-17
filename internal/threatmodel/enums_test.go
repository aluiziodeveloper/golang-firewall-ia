package threatmodel

import (
	"errors"
	"strings"
	"testing"
)

// #region threatmodel-trust-zone-test
func TestTrustZone(t *testing.T) {
	tests := []struct {
		zone    TrustZone
		want    string
		wantErr bool
	}{
		{
			zone: TrustZoneUnspecified, want: "unspecified",
			wantErr: true,
		},
		{zone: TrustZoneAgent, want: "agent"},
		{zone: TrustZoneGuarita, want: "guarita"},
		{zone: TrustZoneMCPServer, want: "mcp-server"},
		{zone: TrustZoneOperator, want: "operator"},
		{zone: TrustZoneLocalModel, want: "local-model"},
		{zone: TrustZoneStorage, want: "storage"},
		{
			zone: TrustZoneStorage + 1, want: "TrustZone(7)",
			wantErr: true,
		},
		{zone: -1, want: "TrustZone(-1)", wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			if got := test.zone.String(); got != test.want {
				t.Errorf("String() = %q; want %q", got, test.want)
			}
			assertEnumValidation(t, test.zone.Validate(), test.wantErr,
				ErrInvalidTrustZone, "trust_zone")
		})
	}
}

// #endregion

func TestTrustBoundary(t *testing.T) {
	tests := []struct {
		name     string
		boundary TrustBoundary
		want     string
		wantErr  bool
	}{
		{
			name:     "agent to guarita",
			boundary: TrustBoundary{TrustZoneAgent, TrustZoneGuarita},
			want:     "agent->guarita",
		},
		{
			name:     "same zone",
			boundary: TrustBoundary{TrustZoneGuarita, TrustZoneGuarita},
			want:     "guarita->guarita",
			wantErr:  true,
		},
		{
			name:     "zero value",
			boundary: TrustBoundary{},
			want:     "unspecified->unspecified",
			wantErr:  true,
		},
		{
			name:     "unknown destination",
			boundary: TrustBoundary{TrustZoneAgent, TrustZone(99)},
			want:     "agent->TrustZone(99)",
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if got := test.boundary.String(); got != test.want {
				t.Errorf("String() = %q; want %q", got, test.want)
			}
			assertEnumValidation(t, test.boundary.Validate(), test.wantErr,
				ErrInvalidTrustBoundary, "trust_boundary")
		})
	}
}

func TestSTRIDECategory(t *testing.T) {
	tests := []struct {
		category STRIDECategory
		want     string
		wantErr  bool
	}{
		{category: STRIDEUnspecified, want: "unspecified", wantErr: true},
		{category: STRIDESpoofing, want: "spoofing"},
		{category: STRIDETampering, want: "tampering"},
		{category: STRIDERepudiation, want: "repudiation"},
		{
			category: STRIDEInformationDisclosure,
			want:     "information-disclosure",
		},
		{category: STRIDEDenialOfService, want: "denial-of-service"},
		{
			category: STRIDEElevationOfPrivilege,
			want:     "elevation-of-privilege",
		},
		{
			category: STRIDEElevationOfPrivilege + 1,
			want:     "STRIDECategory(7)",
			wantErr:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			if got := test.category.String(); got != test.want {
				t.Errorf("String() = %q; want %q", got, test.want)
			}
			assertEnumValidation(t, test.category.Validate(), test.wantErr,
				ErrInvalidSTRIDECategory, "stride_category")
		})
	}
}

func TestThreatStatus(t *testing.T) {
	tests := []struct {
		status  ThreatStatus
		want    string
		wantErr bool
	}{
		{status: ThreatStatusUnspecified, want: "unspecified", wantErr: true},
		{status: ThreatStatusOpen, want: "open"},
		{
			status: ThreatStatusPartiallyMitigated,
			want:   "partially-mitigated",
		},
		{status: ThreatStatusMitigated, want: "mitigated"},
		{status: ThreatStatusAccepted, want: "accepted"},
		{
			status:  ThreatStatusAccepted + 1,
			want:    "ThreatStatus(5)",
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.want, func(t *testing.T) {
			if got := test.status.String(); got != test.want {
				t.Errorf("String() = %q; want %q", got, test.want)
			}
			assertEnumValidation(t, test.status.Validate(), test.wantErr,
				ErrInvalidThreatStatus, "threat_status")
		})
	}
}

// #region threatmodel-enum-validation-helper
func assertEnumValidation(
	t *testing.T, err error, wantErr bool, kind error, field string,
) {
	t.Helper()

	if !wantErr {
		if err != nil {
			t.Fatalf("Validate() returned an error: %v", err)
		}
		return
	}

	if !errors.Is(err, kind) {
		t.Fatalf("errors.Is(%v, %v) = false", err, kind)
	}
	if !strings.HasPrefix(err.Error(), field+": ") {
		t.Errorf("Error() = %q; want prefix %q", err, field+": ")
	}
}

// #endregion
