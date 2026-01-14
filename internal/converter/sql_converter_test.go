package converter

import (
	"strings"
	"testing"
)

func TestConvertOracleFunctions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "SYSDATE conversion",
			input:    "SELECT SYSDATE FROM dual",
			expected: "SELECT CURRENT_TIMESTAMP FROM dual",
		},
		{
			name:     "NVL conversion",
			input:    "SELECT NVL(col, 0) FROM table1",
			expected: "SELECT IFNULL(col, 0) FROM table1",
		},
		{
			name:     "LENGTH conversion",
			input:    "SELECT LENGTH(name) FROM users",
			expected: "SELECT CHAR_LENGTH(name) FROM users",
		},
		{
			name:     "SUBSTR conversion",
			input:    "SELECT SUBSTR(name, 1, 5) FROM users",
			expected: "SELECT SUBSTRING(name, 1, 5) FROM users",
		},
		{
			name:     "TRUNC conversion",
			input:    "SELECT TRUNC(created_date) FROM orders",
			expected: "SELECT DATE(created_date) FROM orders",
		},
		{
			name:     "ADD_MONTHS conversion",
			input:    "SELECT ADD_MONTHS(SYSDATE, 3) FROM dual",
			expected: "SELECT DATE_ADD(CURRENT_TIMESTAMP, INTERVAL 3 MONTH) FROM dual",
		},
		{
			name:     "MONTHS_BETWEEN conversion",
			input:    "SELECT MONTHS_BETWEEN(date1, date2) FROM table1",
			expected: "SELECT TIMESTAMPDIFF(MONTH, date2, date1) FROM table1",
		},
		{
			name:     "USER function conversion",
			input:    "SELECT USER FROM dual",
			expected: "SELECT CURRENT_USER() FROM dual",
		},
		{
			name:     "UID function conversion",
			input:    "SELECT UID FROM dual",
			expected: "SELECT USER() FROM dual",
		},
		{
			name:     "SYS_GUID conversion",
			input:    "SELECT SYS_GUID() FROM dual",
			expected: "SELECT UUID() FROM dual",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertOracleFunctions(tt.input)
			if result != tt.expected {
				t.Errorf("convertOracleFunctions() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertNVL2(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic NVL2",
			input:    "SELECT NVL2(email, 'Has Email', 'No Email') FROM users",
			expected: "SELECT IF(email IS NOT NULL, 'Has Email', 'No Email') FROM users",
		},
		{
			name:     "NVL2 with column refs",
			input:    "SELECT NVL2(col1, col2, col3) FROM table1",
			expected: "SELECT IF(col1 IS NOT NULL, col2, col3) FROM table1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertNVL2(tt.input)
			if result != tt.expected {
				t.Errorf("convertNVL2() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertListagg(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Basic LISTAGG",
			input:    "SELECT LISTAGG(name, ', ') FROM table1",
			expected: "SELECT GROUP_CONCAT(name SEPARATOR ', ') FROM table1",
		},
		{
			name:     "LISTAGG with different separator",
			input:    "SELECT LISTAGG(col, ';') FROM table1",
			expected: "SELECT GROUP_CONCAT(col SEPARATOR ';') FROM table1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertListagg(tt.input)
			if result != tt.expected {
				t.Errorf("convertListagg() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertNullsOrdering(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "NULLS LAST with DESC",
			input:    "ORDER BY score DESC NULLS LAST",
			expected: "ORDER BY CASE WHEN score IS NULL THEN 1 ELSE 0 END, score DESC",
		},
		{
			name:     "NULLS FIRST with DESC",
			input:    "ORDER BY score DESC NULLS FIRST",
			expected: "ORDER BY CASE WHEN score IS NULL THEN 0 ELSE 1 END, score DESC",
		},
		{
			name:     "NULLS LAST with ASC",
			input:    "ORDER BY score ASC NULLS LAST",
			expected: "ORDER BY CASE WHEN score IS NULL THEN 1 ELSE 0 END, score ASC",
		},
		{
			name:     "NULLS FIRST with ASC",
			input:    "ORDER BY score ASC NULLS FIRST",
			expected: "ORDER BY score ASC",
		},
		{
			name:     "NULLS LAST without direction",
			input:    "ORDER BY score NULLS LAST",
			expected: "ORDER BY CASE WHEN score IS NULL THEN 1 ELSE 0 END, score",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertNullsOrdering(tt.input)
			if result != tt.expected {
				t.Errorf("convertNullsOrdering() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestConvertDualTable(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple SELECT FROM DUAL",
			input:    "SELECT 1 FROM DUAL",
			expected: "SELECT 1",
		},
		{
			name:     "SYSDATE FROM DUAL",
			input:    "SELECT SYSDATE FROM DUAL",
			expected: "SELECT SYSDATE",
		},
		{
			name:     "INSERT ALL with DUAL",
			input:    "INSERT ALL INTO table1 VALUES (1, 2) SELECT * FROM DUAL",
			expected: "INSERT ALL INTO table1 VALUES (1, 2)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertDualTable(tt.input)
			result = strings.TrimSpace(result)
			expected := strings.TrimSpace(tt.expected)
			if result != expected {
				t.Errorf("convertDualTable() = %v, want %v", result, expected)
			}
		})
	}
}

func TestConvertConcatenation(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "Simple concatenation",
			input: "SELECT 'Hello' || ' ' || 'World' FROM dual",
			want:  "SELECT CONCAT(CONCAT('Hello', ' '), 'World') FROM dual",
		},
		{
			name:  "Column concatenation",
			input: "SELECT first_name || ' ' || last_name FROM users",
			want:  "SELECT CONCAT(CONCAT(first_name, ' '), last_name) FROM users",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertConcatenation(tt.input)
			if result != tt.want {
				t.Errorf("convertConcatenation() = %v, want %v", result, tt.want)
			}
		})
	}
}

func TestConvertSQLStatement(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string // Check if result contains these strings
	}{
		{
			name:  "Complete SQL with multiple functions",
			input: "SELECT NVL(col1, 0), LENGTH(col2), SYSDATE FROM table1 WHERE id = 1",
			contains: []string{
				"IFNULL",
				"CHAR_LENGTH",
				"CURRENT_TIMESTAMP",
			},
		},
		{
			name:  "SQL with DUAL",
			input: "SELECT 1 FROM DUAL",
			contains: []string{
				"SELECT 1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ConvertSQLStatement(tt.input)
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("ConvertSQLStatement() result = %v, should contain %v", result, substr)
				}
			}
		})
	}
}
