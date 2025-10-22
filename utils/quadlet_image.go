package utils

import (
	"fmt"
	"strings"
)

// ReplaceOrInsertImage replaces the Image= line inside the [Container] section with the provided image name,
// or inserts it if missing. If no [Container] section exists, attempts to insert one.
func ReplaceOrInsertImage(content, image string) (string, error) {
	if strings.TrimSpace(image) == "" {
		return "", fmt.Errorf("image cannot be empty")
	}
	lines := strings.Split(content, "\n")

	// Find [Container] section bounds
	start := -1
	end := len(lines)
	for i, l := range lines {
		if strings.TrimSpace(l) == "[Container]" {
			start = i
			break
		}
	}
	if start == -1 {
		// No [Container] section; append minimal section
		builder := strings.Builder{}
		builder.WriteString(strings.TrimRight(content, "\n"))
		if !strings.HasSuffix(content, "\n") { builder.WriteString("\n") }
		builder.WriteString("[Container]\n")
		builder.WriteString("Image=")
		builder.WriteString(image)
		builder.WriteString("\n")
		return builder.String(), nil
	}

	// find end: next section or EOF
	for i := start + 1; i < len(lines); i++ {
		trim := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trim, "[") && strings.HasSuffix(trim, "]") {
			end = i
			break
		}
	}

	// Now within [Container], replace or insert Image=
	found := false
	for i := start + 1; i < end; i++ {
		trim := strings.TrimSpace(lines[i])
		if strings.HasPrefix(trim, "Image=") {
			lines[i] = "Image=" + image
			found = true
			break
		}
	}
	if !found {
		// insert after [Container]
		insertAt := start + 1
		newLines := append([]string{}, lines[:insertAt]...)
		newLines = append(newLines, "Image="+image)
		newLines = append(newLines, lines[insertAt:]...)
		lines = newLines
	}

	return strings.Join(lines, "\n"), nil
}
