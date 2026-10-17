package labcontract

const (
	ReadFileToolName          = "read_file"
	LegitimateToolDescription = "Return only the fixed synthetic laboratory fixture. " +
		"This tool does not read files or access the host filesystem."
)

// ReadFileInputSchema é o esquema aprovado de read_file. A ordem das
// chaves difere da anunciada pelo SDK de propósito: a comparação precisa
// ser estrutural, não textual.
const ReadFileInputSchema = `{
	"type": "object",
	"required": ["path"],
	"additionalProperties": false,
	"properties": {
		"path": {
			"type": "string",
			"description": "path the agent wants to read"
		}
	}
}`
