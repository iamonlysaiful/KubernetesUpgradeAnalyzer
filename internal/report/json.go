package report

import "encoding/json"

// RenderJSON returns canonical JSON report output.
func RenderJSON(doc Document) ([]byte, error) {
	b, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(b, '\n'), nil
}
