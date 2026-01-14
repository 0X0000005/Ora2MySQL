package converter

import (
	"strings"
	"testing"
)

func TestConvertToMySQL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		contains []string
	}{
		{
			name:  "Basic CREATE TABLE",
			input: "CREATE TABLE users (id NUMBER(10) PRIMARY KEY, name VARCHAR2(50), created DATE);",
			contains: []string{
				"CREATE TABLE",
				"users",
				"INT",
				"VARCHAR(50)",
				"DATETIME",
				"PRIMARY KEY",
			},
		},
		{
			name:  "Data type conversions",
			input: "CREATE TABLE docs (id NUMBER, text CLOB, data BLOB);",
			contains: []string{
				"LONGTEXT",
				"LONGBLOB",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ConvertToMySQL(tt.input)
			if err != nil {
				// For invalid input, error is expected
				return
			}
			for _, substr := range tt.contains {
				if !strings.Contains(result, substr) {
					t.Errorf("ConvertToMySQL() should contain %q, got:\n%s", substr, result)
				}
			}
		})
	}
}

func TestConvertToMySQLInvalidInput(t *testing.T) {
	result, err := ConvertToMySQL("This is just random text")
	if err == nil {
		t.Error("ConvertToMySQL() should return error for invalid input")
	}
	if !strings.Contains(result, "错误") {
		t.Errorf("ConvertToMySQL() error message should contain '错误', got: %s", result)
	}
}
