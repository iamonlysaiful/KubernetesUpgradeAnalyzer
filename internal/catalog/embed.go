package catalog

import _ "embed"

//go:embed testdata/embedded-catalog.json
var embeddedCatalog []byte

func LoadEmbedded() (Bundle, error) {
	return LoadBytes(embeddedCatalog, SourceEmbedded)
}
