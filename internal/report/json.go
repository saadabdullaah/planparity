// Package report renders sanitized domain models.
package report

import (
	"encoding/json"
	"io"
)

// JSON writes stable, indented JSON without HTML escaping.
func JSON(writer io.Writer, value any) error {
	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "  ")
	return encoder.Encode(value)
}
