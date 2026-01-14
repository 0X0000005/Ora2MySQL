package converter

import (
	"strings"
)

// convertView 转换视图定义
func convertView(view ViewDef) string {
	var sb strings.Builder

	// CREATE OR REPLACE VIEW
	sb.WriteString("CREATE OR REPLACE VIEW ")
	sb.WriteString(view.Name)
	
	// 如果有指定列名
	if len(view.Columns) > 0 {
		sb.WriteString(" (")
		sb.WriteString(strings.Join(view.Columns, ", "))
		sb.WriteString(")")
	}
	
	sb.WriteString(" AS\n")
	
	// 转换 SELECT 语句中的 Oracle 函数
	mysqlSelect := ConvertSQLStatement(view.SelectSQL)
	sb.WriteString(mysqlSelect)
	
	sb.WriteString(";")
	
	// MySQL 不支持视图的 COMMENT 语法，注释会以注释形式添加
	if view.Comment != "" {
		sb.WriteString("\n-- ")
		sb.WriteString(view.Comment)
	}
	
	return sb.String()
}
