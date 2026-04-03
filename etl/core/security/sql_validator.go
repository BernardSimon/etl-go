package security

import (
	"fmt"
	"strings"
)

// dangerousKeywords are SQL statements that could cause data loss or schema changes.
var dangerousKeywords = []string{
	"DROP ", "TRUNCATE ", "ALTER ", "CREATE ", "GRANT ", "REVOKE ",
}

// ValidateSourceSQL ensures a SQL query used in a Source component is a SELECT statement.
func ValidateSourceSQL(query string) error {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return fmt.Errorf("sql validation: query is empty")
	}
	upper := strings.ToUpper(trimmed)
	if !strings.HasPrefix(upper, "SELECT") && !strings.HasPrefix(upper, "WITH") {
		return fmt.Errorf("sql validation: source query must start with SELECT or WITH, got: %.30s", trimmed)
	}
	return nil
}

// ValidateExecutorSQL checks if an executor SQL contains dangerous statements.
// If allowDangerous is true, dangerous statements are permitted.
func ValidateExecutorSQL(query string, allowDangerous bool) error {
	trimmed := strings.TrimSpace(query)
	if trimmed == "" {
		return fmt.Errorf("sql validation: query is empty")
	}
	if allowDangerous {
		return nil
	}
	upper := strings.ToUpper(trimmed)
	for _, keyword := range dangerousKeywords {
		if strings.Contains(upper, keyword) {
			return fmt.Errorf("sql validation: query contains dangerous keyword %q. Set allow_dangerous=true to permit this", strings.TrimSpace(keyword))
		}
	}
	return nil
}
