package threatmodel

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
	"testing"
)

const modelDocument = "../../docs/threats/model.md"

// #region threatmodel-document-row
// | `id` | nome | categorias | fronteira | capítulo | estado |
var documentRow = regexp.MustCompile("^\\| `([a-z-]+)` " +
	`\| (.+) \| (.+) \| (.+) \| (\d+) \| (.+) \|$`)

type documentThreat struct {
	name, categories, boundary, status string
	chapter                            int
}

// #endregion

func readModelDocument(t *testing.T) map[string]documentThreat {
	t.Helper()
	file, err := os.Open(modelDocument)
	if err != nil {
		t.Fatalf("abrir %s: %v", modelDocument, err)
	}
	t.Cleanup(func() { _ = file.Close() })

	rows := make(map[string]documentThreat)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		match := documentRow.FindStringSubmatch(scanner.Text())
		if match == nil {
			continue
		}
		chapter, err := strconv.Atoi(match[5])
		if err != nil {
			t.Fatalf("capítulo inválido em %q: %v", match[1], err)
		}
		rows[match[1]] = documentThreat{
			name:       strings.ReplaceAll(match[2], "*", ""),
			categories: match[3],
			boundary:   match[4],
			chapter:    chapter,
			status:     match[6],
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("ler %s: %v", modelDocument, err)
	}
	return rows
}

// #region threatmodel-document-test
// O documento e o código descrevem o mesmo modelo: uma ameaça a mais ou
// a menos em qualquer um dos dois quebra o build.
func TestModelo_coincideComODocumento(t *testing.T) {
	rows := readModelDocument(t)
	for _, threat := range Model() {
		row, ok := rows[threat.ID]
		if !ok {
			t.Errorf("%s: ausente de %s", threat.ID, modelDocument)
			continue
		}
		delete(rows, threat.ID)
		want := documentThreat{
			name:       threat.Name,
			categories: joinCategories(threat.Categories),
			boundary:   threat.Boundary.String(),
			chapter:    threat.DefenseChapter,
			status:     threat.Status.String(),
		}
		if row != want {
			t.Errorf("%s: documento %+v; código %+v",
				threat.ID, row, want)
		}
	}
	for id := range rows {
		t.Errorf("%s: no documento, ausente de Model()", id)
	}
}

// #endregion

func joinCategories(categories []STRIDECategory) string {
	names := make([]string, len(categories))
	for i, category := range categories {
		names[i] = category.String()
	}
	return strings.Join(names, ", ")
}
