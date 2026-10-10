package diagnostics

import (
	"encoding/xml"
	"io"
	"regexp"
	"strconv"
	"strings"
)

var powerShellEscape = regexp.MustCompile(`_x([0-9a-fA-F]{4})_`)

// PowerShell extracts readable error text from PowerShell CLIXML.
func PowerShell(output []byte) string {
	text := strings.TrimSpace(string(output))
	if !strings.HasPrefix(text, "#< CLIXML") {
		return text
	}
	decoder := xml.NewDecoder(strings.NewReader(strings.TrimSpace(strings.TrimPrefix(text, "#< CLIXML"))))
	messages := []string{}
	for {
		token, err := decoder.Token()
		if err == io.EOF {
			break
		}
		if err != nil {
			return text
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "S" {
			continue
		}
		isError := false
		for _, attr := range start.Attr {
			if attr.Name.Local == "S" && attr.Value == "Error" {
				isError = true
			}
		}
		var value string
		if err := decoder.DecodeElement(&value, &start); err != nil {
			return text
		}
		if !isError {
			continue
		}
		value = powerShellEscape.ReplaceAllStringFunc(value, func(encoded string) string {
			code, _ := strconv.ParseUint(encoded[2:6], 16, 16)
			return string(rune(code))
		})
		if value = strings.TrimSpace(value); value != "" {
			messages = append(messages, value)
		}
	}
	if len(messages) == 0 {
		return text
	}
	return strings.Join(messages, "\n")
}
