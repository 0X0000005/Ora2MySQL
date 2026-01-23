package converter

import (
	"fmt"
	"regexp"
	"strings"
)

// ConvertToMySQL 将 Oracle DDL 或 SQL 语句转换为 MySQL 格式
func ConvertToMySQL(oracleDDL string) (string, error) {
	// 解析 Oracle DDL
	tables, views, indexes, tableComments, viewComments, columnComments, err := ParseDDL(oracleDDL)
	if err != nil {
		return "", err
	}

	var result strings.Builder
	hasContent := false

	// 转换表定义
	for _, table := range tables {
		// 应用表注释（如果有）
		if comment, ok := tableComments[table.Name]; ok {
			table.Comment = comment
		}

		// 应用列注释（如果有）
		if colComments, ok := columnComments[table.Name]; ok {
			for i := range table.Columns {
				if comment, exists := colComments[table.Columns[i].Name]; exists {
					table.Columns[i].Comment = comment
				}
			}
		}

		mysqlDDL := convertTable(table)
		result.WriteString(mysqlDDL)
		result.WriteString("\n\n\n")
		hasContent = true
	}

	// 转换视图定义
	for _, view := range views {
		// 应用视图注释（如果有）
		if comment, ok := viewComments[view.Name]; ok {
			view.Comment = comment
		}

		mysqlView := convertView(view)
		result.WriteString(mysqlView)
		result.WriteString("\n\n\n")
		hasContent = true
	}

	// 转换索引定义
	for _, index := range indexes {
		mysqlIndex := convertIndex(index)
		result.WriteString(mysqlIndex)
		result.WriteString("\n\n\n")
		hasContent = true
	}

	// 如果没有解析到任何 DDL，说明可能是纯 SQL 语句或非法输入
	// 直接使用 SQL 转换器处理
	if !hasContent {
		if !isValidInput(oracleDDL) {
			return "", fmt.Errorf("illegal input: no valid SQL or MyBatis statement found")
		}
		converted := ConvertSQLStatement(oracleDDL)
		// 尝试合并 INSERT 语句
		converted = mergeInsertStatements(converted)
		return converted, nil
	}

	return result.String(), nil
}

// mergeInsertStatements 检测并合并同一张表的多个 INSERT 语句为批量插入
func mergeInsertStatements(sql string) string {
	// 分割语句
	statements := splitSQLStatements(sql)
	if len(statements) <= 1 {
		return sql
	}

	// 解析 INSERT 语句，按表名分组
	type insertInfo struct {
		tableName string
		columns   string // 列定义部分（可选）
		values    string // VALUES 部分
		original  string // 原始语句
	}

	var inserts []insertInfo
	var otherStatements []string

	// 匹配 INSERT INTO table [(columns)] VALUES (...)
	insertRe := regexp.MustCompile(`(?i)^\s*INSERT\s+INTO\s+(\S+)\s*(\([^)]*\))?\s*VALUES\s*(\(.*\))\s*;?\s*$`)

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		matches := insertRe.FindStringSubmatch(stmt)
		if matches != nil {
			inserts = append(inserts, insertInfo{
				tableName: strings.ToLower(matches[1]),
				columns:   matches[2],
				values:    matches[3],
				original:  stmt,
			})
		} else {
			otherStatements = append(otherStatements, stmt)
		}
	}

	// 如果没有足够的 INSERT 语句，或者有其他语句，不合并
	if len(inserts) < 2 || len(otherStatements) > 0 {
		return sql
	}

	// 检查是否所有 INSERT 都是同一张表
	firstTable := inserts[0].tableName
	firstColumns := inserts[0].columns
	allSameTable := true
	allSameColumns := true

	for _, ins := range inserts[1:] {
		if ins.tableName != firstTable {
			allSameTable = false
			break
		}
		if ins.columns != firstColumns {
			allSameColumns = false
		}
	}

	// 只有当所有语句都是同一张表且列结构相同时才合并
	if !allSameTable || !allSameColumns {
		return sql
	}

	// 合并为批量 INSERT
	var sb strings.Builder
	sb.WriteString("INSERT INTO ")
	sb.WriteString(inserts[0].tableName)
	if inserts[0].columns != "" {
		sb.WriteString(" ")
		sb.WriteString(inserts[0].columns)
	}
	sb.WriteString(" VALUES\n")

	for i, ins := range inserts {
		sb.WriteString("  ")
		sb.WriteString(ins.values)
		if i < len(inserts)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}
	sb.WriteString(";")

	return sb.String()
}

// splitSQLStatements 按分号分割 SQL 语句（简单版本）
func splitSQLStatements(sql string) []string {
	var result []string
	var current strings.Builder
	inString := false
	stringChar := rune(0)

	for _, ch := range sql {
		if inString {
			current.WriteRune(ch)
			if ch == stringChar {
				inString = false
			}
		} else {
			if ch == '\'' || ch == '"' {
				inString = true
				stringChar = ch
				current.WriteRune(ch)
			} else if ch == ';' {
				stmt := strings.TrimSpace(current.String())
				if stmt != "" {
					result = append(result, stmt)
				}
				current.Reset()
			} else {
				current.WriteRune(ch)
			}
		}
	}

	// 处理最后一个语句（可能没有分号结尾）
	stmt := strings.TrimSpace(current.String())
	if stmt != "" {
		result = append(result, stmt)
	}

	return result
}

// convertTable 转换表定义
func convertTable(table TableDef) string {
	var sb strings.Builder

	// 检查是否有主键约束
	hasPrimaryKey := false
	for _, constraint := range table.Constraints {
		if constraint.Type == "PRIMARY KEY" {
			hasPrimaryKey = true
			break
		}
	}

	// 如果没有主键，自动生成一个自增主键
	if !hasPrimaryKey {
		pkName := generatePrimaryKeyName(table.Columns)
		// 创建自增主键列
		pkColumn := ColumnDef{
			Name:     pkName,
			DataType: "BIGINT",
			NotNull:  true,
		}
		// 将主键列插入到列列表的开头
		table.Columns = append([]ColumnDef{pkColumn}, table.Columns...)
		// 添加主键约束
		table.Constraints = append(table.Constraints, Constraint{
			Type:    "PRIMARY KEY",
			Columns: []string{pkName},
		})
	}

	sb.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s (\n", table.Name))

	// 转换列定义
	for i, col := range table.Columns {
		sb.WriteString("  ")
		// 检查是否是自动生成的主键列（第一列且无主键时添加的）
		if !hasPrimaryKey && i == 0 {
			sb.WriteString(fmt.Sprintf("%s BIGINT NOT NULL AUTO_INCREMENT", col.Name))
		} else {
			sb.WriteString(convertColumn(col))
		}

		// 如果不是最后一列，或者后面还有约束，添加逗号
		if i < len(table.Columns)-1 || len(table.Constraints) > 0 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	// 转换约束定义
	for i, constraint := range table.Constraints {
		sb.WriteString("  ")
		sb.WriteString(convertConstraint(constraint))

		if i < len(table.Constraints)-1 {
			sb.WriteString(",")
		}
		sb.WriteString("\n")
	}

	sb.WriteString(")")

	// 添加表注释
	if table.Comment != "" {
		sb.WriteString(fmt.Sprintf(" COMMENT='%s'", escapeSingleQuotes(table.Comment)))
	}

	// 添加引擎和字符集
	sb.WriteString(" ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_bin")

	sb.WriteString(";")

	return sb.String()
}

// generatePrimaryKeyName 生成主键列名，避免与现有列名冲突
func generatePrimaryKeyName(columns []ColumnDef) string {
	// 收集现有列名（转为小写进行比较）
	existingNames := make(map[string]bool)
	for _, col := range columns {
		existingNames[strings.ToLower(col.Name)] = true
	}

	// 按优先级尝试主键名
	candidates := []string{"id", "uid", "idx"}
	for _, name := range candidates {
		if !existingNames[name] {
			return name
		}
	}

	// 如果都冲突，使用带数字后缀的名称
	for i := 1; i <= 100; i++ {
		name := fmt.Sprintf("id_%d", i)
		if !existingNames[name] {
			return name
		}
	}

	return "auto_id"
}

// convertColumn 转换列定义
func convertColumn(col ColumnDef) string {
	var sb strings.Builder

	sb.WriteString(col.Name)
	sb.WriteString(" ")

	// 转换数据类型
	mysqlType := convertDataType(col.DataType, col.Length, col.Precision, col.Scale)
	sb.WriteString(mysqlType)

	// NOT NULL 约束
	if col.NotNull {
		sb.WriteString(" NOT NULL")
	}

	// 默认值
	if col.DefaultValue != "" {
		defaultVal := convertDefaultValue(col.DefaultValue)
		sb.WriteString(fmt.Sprintf(" DEFAULT %s", defaultVal))
	}

	// 列注释
	if col.Comment != "" {
		sb.WriteString(fmt.Sprintf(" COMMENT '%s'", escapeSingleQuotes(col.Comment)))
	}

	return sb.String()
}

// convertDataType 转换数据类型
func convertDataType(oracleType, length, precision, scale string) string {
	oracleType = strings.ToUpper(oracleType)

	switch oracleType {
	case "VARCHAR2", "NVARCHAR2":
		if length != "" {
			return fmt.Sprintf("VARCHAR(%s)", length)
		}
		return "VARCHAR(255)"

	case "CHAR", "NCHAR":
		if length != "" {
			return fmt.Sprintf("CHAR(%s)", length)
		}
		return "CHAR(1)"

	case "NUMBER":
		// NUMBER 类型需要根据精度和小数位数转换
		if scale != "" {
			// 有小数位，使用 DECIMAL
			if precision != "" {
				return fmt.Sprintf("DECIMAL(%s,%s)", precision, scale)
			}
			return fmt.Sprintf("DECIMAL(10,%s)", scale)
		} else if precision != "" {
			// 无小数位
			p := parseIntSafe(precision)
			if p <= 0 {
				return "DECIMAL(10,0)"
			} else if p <= 3 {
				return "TINYINT"
			} else if p <= 5 {
				return "SMALLINT"
			} else if p <= 9 {
				return "INT"
			} else if p <= 19 {
				return "BIGINT"
			} else {
				return fmt.Sprintf("DECIMAL(%s,0)", precision)
			}
		}
		// 默认 NUMBER 转为 DECIMAL(10,0)
		return "DECIMAL(10,0)"

	case "INTEGER", "INT":
		return "INT"

	case "SMALLINT":
		return "SMALLINT"

	case "FLOAT":
		return "FLOAT"

	case "DOUBLE":
		return "DOUBLE"

	case "DATE":
		// Oracle DATE 包含时间，转为 DATETIME
		return "DATETIME"

	case "TIMESTAMP":
		return "DATETIME"

	case "CLOB", "NCLOB":
		return "LONGTEXT"

	case "BLOB":
		return "LONGBLOB"

	case "RAW":
		if length != "" {
			return fmt.Sprintf("VARBINARY(%s)", length)
		}
		return "VARBINARY(255)"

	case "LONG":
		return "LONGTEXT"

	default:
		// 未知类型，保持原样
		if length != "" {
			return fmt.Sprintf("%s(%s)", oracleType, length)
		}
		return oracleType
	}
}

// convertDefaultValue 转换默认值
func convertDefaultValue(oracleDefault string) string {
	upper := strings.ToUpper(oracleDefault)

	// 转换 Oracle 特定函数
	switch upper {
	case "SYSDATE":
		return "CURRENT_TIMESTAMP"
	case "SYSTIMESTAMP":
		return "CURRENT_TIMESTAMP"
	case "USER":
		return "CURRENT_USER"
	case "NULL":
		return "NULL"
	default:
		// 如果是字符串，确保有引号
		if !strings.HasPrefix(oracleDefault, "'") && !isNumeric(oracleDefault) {
			return fmt.Sprintf("'%s'", oracleDefault)
		}
		return oracleDefault
	}
}

// convertConstraint 转换约束定义
func convertConstraint(constraint Constraint) string {
	var sb strings.Builder

	switch constraint.Type {
	case "PRIMARY KEY":
		if constraint.Name != "" {
			// 为主键添加 _pk 后缀（如果尚未包含）
			constraintName := ensureSuffix(constraint.Name, "_pk")
			sb.WriteString(fmt.Sprintf("CONSTRAINT %s ", constraintName))
		}
		sb.WriteString("PRIMARY KEY (")
		sb.WriteString(strings.Join(constraint.Columns, ", "))
		sb.WriteString(")")

	case "FOREIGN KEY":
		if constraint.Name != "" {
			sb.WriteString(fmt.Sprintf("CONSTRAINT %s ", constraint.Name))
		}
		sb.WriteString("FOREIGN KEY (")
		sb.WriteString(strings.Join(constraint.Columns, ", "))
		sb.WriteString(") REFERENCES ")
		sb.WriteString(constraint.RefTable)
		sb.WriteString(" (")
		sb.WriteString(strings.Join(constraint.RefColumns, ", "))
		sb.WriteString(")")

	case "UNIQUE":
		if constraint.Name != "" {
			// 为唯一索引添加 _uk 后缀（如果尚未包含）
			constraintName := ensureSuffix(constraint.Name, "_uk")
			sb.WriteString(fmt.Sprintf("CONSTRAINT %s ", constraintName))
		}
		sb.WriteString("UNIQUE (")
		sb.WriteString(strings.Join(constraint.Columns, ", "))
		sb.WriteString(")")

	case "CHECK":
		// MySQL 8.0+ 支持 CHECK 约束
		if constraint.Name != "" {
			sb.WriteString(fmt.Sprintf("CONSTRAINT %s ", constraint.Name))
		}
		sb.WriteString("CHECK (")
		sb.WriteString(constraint.CheckExpr)
		sb.WriteString(")")
	}

	return sb.String()
}

// convertIndex 转换索引定义
func convertIndex(index IndexDef) string {
	var sb strings.Builder

	// 根据索引类型添加相应后缀
	indexName := index.Name
	if index.Unique {
		// 唯一索引添加 _uk 后缀（如果尚未包含）
		indexName = ensureSuffix(indexName, "_uk")
		sb.WriteString("CREATE UNIQUE INDEX ")
	} else {
		// 普通索引添加 _idx 后缀（如果尚未包含）
		indexName = ensureSuffix(indexName, "_idx")
		sb.WriteString("CREATE INDEX ")
	}

	sb.WriteString(indexName)
	sb.WriteString(" ON ")
	sb.WriteString(index.Table)
	sb.WriteString(" (")
	sb.WriteString(strings.Join(index.Columns, ", "))
	sb.WriteString(");")

	return sb.String()
}

// ensureSuffix 确保字符串以指定后缀结尾（如果尚未包含）
func ensureSuffix(name, suffix string) string {
	if !strings.HasSuffix(name, suffix) {
		return name + suffix
	}
	return name
}

// 辅助函数：转义单引号
func escapeSingleQuotes(s string) string {
	return strings.ReplaceAll(s, "'", "\\'")
}

// 辅助函数：判断是否为数字
func isNumeric(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}

	for i, ch := range s {
		if i == 0 && (ch == '-' || ch == '+') {
			continue
		}
		if ch < '0' || ch > '9' {
			if ch != '.' {
				return false
			}
		}
	}
	return true
}

// 辅助函数：安全地解析整数
func parseIntSafe(s string) int {
	var result int
	fmt.Sscanf(s, "%d", &result)
	return result
}

// isValidInput 检查输入是否为合法的 SQL 或 MyBatis 内容
func isValidInput(input string) bool {
	upper := strings.ToUpper(input)

	// 1. 检查 SQL DML 关键字
	dmlKeywords := []string{"SELECT", "INSERT", "UPDATE", "DELETE", "MERGE", "WITH"}
	for _, kw := range dmlKeywords {
		if strings.Contains(upper, kw) {
			return true
		}
	}

	// 2. 检查 MyBatis 标签
	// 不区分大小写
	mybatisTags := []string{"<mapper", "<select", "<insert", "<update", "<delete", "<sql", "<resultMap", "<include"}
	for _, tag := range mybatisTags {
		if strings.Contains(strings.ToLower(input), tag) {
			return true
		}
	}

	// 3. 检查常见 DDL 关键字（可能未被 ParseDDL 不完全支持或者是片段）
	ddlKeywords := []string{"CREATE", "ALTER", "DROP", "TRUNCATE", "COMMENT", "GRANT", "REVOKE"}
	for _, kw := range ddlKeywords {
		if strings.Contains(upper, kw) {
			return true
		}
	}

	return false
}
