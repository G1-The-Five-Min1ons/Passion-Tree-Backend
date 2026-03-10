package migration

import "strings"

// splitByGO splits SQL script by GO keyword (case-insensitive, must be on its own line)
// Handles both Unix (\n) and Windows (\r\n) line endings
func splitByGO(sql string) []string {
	sql = strings.ReplaceAll(sql, "\r\n", "\n")
	lines := strings.Split(sql, "\n")
	var batches []string
	var currentBatch strings.Builder

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.EqualFold(trimmed, "GO") {
			if currentBatch.Len() > 0 {
				batches = append(batches, currentBatch.String())
				currentBatch.Reset()
			}
		} else {
			currentBatch.WriteString(line)
			currentBatch.WriteString("\n")
		}
	}

	if currentBatch.Len() > 0 {
		batches = append(batches, currentBatch.String())
	}

	return batches
}

// isCommentOnly checks if a batch contains only comments
func isCommentOnly(batch string) bool {
	lines := strings.Split(batch, "\n")
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if !strings.HasPrefix(trimmed, "--") && !strings.HasPrefix(trimmed, "/*") {
			return false
		}
	}
	return true
}

// truncateSQL truncates SQL for logging preview
func truncateSQL(sql string, maxLen int) string {
	sql = strings.ReplaceAll(sql, "\n", " ")
	sql = strings.ReplaceAll(sql, "\t", " ")
	if len(sql) > maxLen {
		return sql[:maxLen] + "..."
	}
	return sql
}
