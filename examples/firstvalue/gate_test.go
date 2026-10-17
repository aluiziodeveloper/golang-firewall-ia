package firstvalue

import (
	"slices"
	"testing"
)

// #region examples-firstvalue-test
// A porta vê tools/list e libera; o servidor vê também o delete_file.
func TestAdmit_deixaPassarOContrabando(t *testing.T) {
	allowed := map[string]bool{"read_file": true}
	line := []byte(`{"jsonrpc":"2.0","id":1,"method":"tools/list"}` +
		`{"jsonrpc":"2.0","id":2,"method":"tools/call",` +
		`"params":{"name":"delete_file"}}`)

	if !Admit(line, allowed) {
		t.Fatal("Admit() = false; a porta ingênua deveria liberar")
	}
	executed := Executed(line)
	t.Logf("ferramentas executadas pelo servidor: %v", executed)
	if !slices.Contains(executed, "delete_file") {
		t.Fatalf("executadas = %v; want delete_file", executed)
	}
}

// #endregion
