package converter

import (
	"fmt"
	"regexp"
	"strings"
)

// parseCreateView 解析 CREATE VIEW 语句
func parseCreateView(stmt string) (ViewDef, error) {
	view := ViewDef{}

	// 移除 OR REPLACE
	stmt = regexp.MustCompile(`(?i)OR\s+REPLACE\s+`).ReplaceAllString(stmt, "")

	// 提取视图名和可选的列名列表
	// CREATE VIEW view_name [(col1, col2, ...)] AS SELECT ...
	re := regexp.MustCompile(`(?i)CREATE\s+VIEW\s+([^\s(]+)(?:\s*\(([^)]+)\))?\s+AS\s+(.+)`)
	matches := re.FindStringSubmatch(stmt)
	
	if len(matches) < 4 {
		return view, fmt.Errorf("无法解析视图定义")
	}

	view.Name = strings.TrimSpace(matches[1])
	
	// 解析列名（如果有）
	if matches[2] != "" {
		columns := strings.Split(matches[2], ",")
		for _, col := range columns {
			view.Columns = append(view.Columns, strings.TrimSpace(col))
		}
	}
	
	// SELECT 语句
	view.SelectSQL = strings.TrimSpace(matches[3])

	return view, nil
}

// parseViewComment 解析视图注释
func parseViewComment(stmt string) (string, string) {
	re := regexp.MustCompile(`(?i)COMMENT\s+ON\s+VIEW\s+([^\s]+)\s+IS\s+'([^']*)'`)
	matches := re.FindStringSubmatch(stmt)
	if len(matches) >= 3 {
		return strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2])
	}
	return "", ""
}
