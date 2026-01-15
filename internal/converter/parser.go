package converter

import (
	"fmt"
	"log"
	"regexp"
	"strings"
)

// ParseDDL 解析 Oracle DDL 语句，支持多个语句
func ParseDDL(ddl string) ([]TableDef, []ViewDef, []IndexDef, map[string]string, map[string]string, map[string]map[string]string, error) {
	tables := []TableDef{}
	views := []ViewDef{}
	indexes := []IndexDef{}
	tableComments := make(map[string]string)             // 表注释映射
	viewComments := make(map[string]string)              // 视图注释映射
	columnComments := make(map[string]map[string]string) // 列注释映射 [表名][列名]

	// 按分号分割多个语句
	statements := splitStatements(ddl)

	for _, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}

		// 判断语句类型
		stmtUpper := strings.ToUpper(stmt)

		if strings.HasPrefix(stmtUpper, "CREATE TABLE") {
			// 解析 CREATE TABLE
			table, err := parseCreateTable(stmt)
			if err != nil {
				return nil, nil, nil, nil, nil, nil, err
			}
			tables = append(tables, table)
		} else if strings.HasPrefix(stmtUpper, "CREATE OR REPLACE VIEW") || strings.HasPrefix(stmtUpper, "CREATE VIEW") {
			// 解析 CREATE VIEW
			view, err := parseCreateView(stmt)
			if err != nil {
				return nil, nil, nil, nil, nil, nil, err
			}
			views = append(views, view)
		} else if strings.HasPrefix(stmtUpper, "CREATE INDEX") || strings.HasPrefix(stmtUpper, "CREATE UNIQUE INDEX") {
			// 解析 CREATE INDEX
			index, err := parseCreateIndex(stmt)
			if err != nil {
				return nil, nil, nil, nil, nil, nil, err
			}
			indexes = append(indexes, index)
		} else if strings.HasPrefix(stmtUpper, "COMMENT ON TABLE") {
			// 解析表注释
			tableName, comment := parseTableComment(stmt)
			if tableName != "" {
				tableComments[tableName] = comment
			}
		} else if strings.HasPrefix(stmtUpper, "COMMENT ON VIEW") {
			// 解析视图注释
			viewName, comment := parseViewComment(stmt)
			if viewName != "" {
				viewComments[viewName] = comment
			}
		} else if strings.HasPrefix(stmtUpper, "COMMENT ON COLUMN") {
			// 解析列注释
			tableName, columnName, comment := parseColumnComment(stmt)
			if tableName != "" && columnName != "" {
				if columnComments[tableName] == nil {
					columnComments[tableName] = make(map[string]string)
				}
				columnComments[tableName][columnName] = comment
			}
		} else if strings.HasPrefix(stmtUpper, "ALTER TABLE") {
			// 解析 ALTER TABLE
			tableName, constraint, index, isIndex, modifiedCols := parseAlterTable(stmt)
			if tableName != "" {
				if isIndex {
					indexes = append(indexes, index)
				} else {
					// 尝试找到已有的表并添加约束或修改列
					found := false
					for i := range tables {
						if tables[i].Name == tableName {
							if constraint.Type != "" {
								tables[i].Constraints = append(tables[i].Constraints, constraint)
							}

							// 更新列定义（如果是 MODIFY）
							if len(modifiedCols) > 0 {
								for _, modCol := range modifiedCols {
									for j := range tables[i].Columns {
										if tables[i].Columns[j].Name == modCol.Name {
											// 更新 NOT NULL
											if modCol.NotNull {
												tables[i].Columns[j].NotNull = true
											}
											// 更新类型（如果有）
											if modCol.DataType != "" {
												tables[i].Columns[j].DataType = modCol.DataType
												tables[i].Columns[j].Length = modCol.Length
												tables[i].Columns[j].Precision = modCol.Precision
												tables[i].Columns[j].Scale = modCol.Scale
											}
											// 更新默认值（如果有）
											if modCol.DefaultValue != "" {
												tables[i].Columns[j].DefaultValue = modCol.DefaultValue
											}
										}
									}
								}
							}

							found = true
							break
						}
					}
					// 如果没找到表，暂时不支持生成独立的 ALTER TABLE (MySQL 语法不同)
					// 但为了兼容，可以记录下来或后续处理
					if !found {
						log.Printf("警告: 发现 ALTER TABLE 约束但未找到对应的 CREATE TABLE: %s", tableName)
					}
				}
			}
		}
	}

	return tables, views, indexes, tableComments, viewComments, columnComments, nil
}

// splitStatements 按分号分割 SQL 语句，支持多行语句
func splitStatements(ddl string) []string {
	// 先移除块注释
	ddl = removeBlockComments(ddl)

	result := []string{}
	var current strings.Builder
	lines := strings.Split(ddl, "\n")

	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)

		// 跳过空行和注释
		if trimmedLine == "" || strings.HasPrefix(trimmedLine, "--") {
			continue
		}

		// 添加到当前语句
		if current.Len() > 0 {
			current.WriteString(" ")
		}
		current.WriteString(trimmedLine)

		// 如果行以分号结束，这是一个完整的语句
		if strings.HasSuffix(trimmedLine, ";") {
			stmt := current.String()
			// 移除末尾的分号
			stmt = strings.TrimSuffix(stmt, ";")
			stmt = strings.TrimSpace(stmt)
			if stmt != "" {
				result = append(result, stmt)
			}
			current.Reset()
		}
	}

	// 处理最后一个没有分号的语句
	if current.Len() > 0 {
		stmt := strings.TrimSpace(current.String())
		if stmt != "" {
			result = append(result, stmt)
		}
	}

	return result
}

// removeBlockComments 移除 SQL 中的块注释 /* ... */
func removeBlockComments(sql string) string {
	var sb strings.Builder
	n := len(sql)
	i := 0

	for i < n {
		// 检查是否是块注释开始 /*
		if i+1 < n && sql[i] == '/' && sql[i+1] == '*' {
			// 找到注释结束 */
			end := -1
			for j := i + 2; j < n-1; j++ {
				if sql[j] == '*' && sql[j+1] == '/' {
					end = j + 2
					break
				}
			}

			if end != -1 {
				// 替换为空格，避免粘连
				sb.WriteByte(' ')
				i = end
				continue
			}
		}

		sb.WriteByte(sql[i])
		i++
	}

	return sb.String()
}

// parseCreateTable 解析 CREATE TABLE 语句
func parseCreateTable(stmt string) (TableDef, error) {
	table := TableDef{}

	// 提取表名
	re := regexp.MustCompile(`(?i)CREATE\s+TABLE\s+([^\s(]+)`)
	matches := re.FindStringSubmatch(stmt)
	if len(matches) < 2 {
		return table, fmt.Errorf("无法解析表名")
	}
	table.Name = strings.TrimSpace(matches[1])

	// 提取列定义和约束
	// 找到括号内的内容
	startIdx := strings.Index(stmt, "(")
	endIdx := strings.LastIndex(stmt, ")")
	if startIdx == -1 || endIdx == -1 {
		return table, fmt.Errorf("无法找到表定义的括号")
	}

	content := stmt[startIdx+1 : endIdx]

	// 分割列定义（需要处理嵌套括号和逗号）
	parts := smartSplit(content, ',')

	for _, part := range parts {
		part = strings.TrimSpace(part)
		partUpper := strings.ToUpper(part)

		if strings.HasPrefix(partUpper, "CONSTRAINT") || strings.HasPrefix(partUpper, "PRIMARY KEY") ||
			strings.HasPrefix(partUpper, "FOREIGN KEY") || strings.HasPrefix(partUpper, "UNIQUE") ||
			strings.HasPrefix(partUpper, "CHECK") {
			// 这是约束定义
			constraint := parseConstraint(part)
			table.Constraints = append(table.Constraints, constraint)
		} else {
			// 这是列定义
			column, inlineConstraint := parseColumnDef(part)
			table.Columns = append(table.Columns, column)

			// 如果有内联约束（如列级 PRIMARY KEY），添加到约束列表
			if inlineConstraint.Type != "" {
				table.Constraints = append(table.Constraints, inlineConstraint)
			}
		}
	}

	return table, nil
}

// parseColumnDef 解析列定义，返回列定义和可能的内联约束（如列级 PRIMARY KEY）
func parseColumnDef(def string) (ColumnDef, Constraint) {
	column := ColumnDef{}
	constraint := Constraint{}
	def = strings.TrimSpace(def)

	// 分割列名和类型
	parts := strings.Fields(def)
	if len(parts) < 2 {
		return column, constraint
	}

	column.Name = parts[0]

	// 解析数据类型
	typeStr := parts[1]

	// 处理带括号的类型，如 VARCHAR2(100) 或 NUMBER(10,2) 或 VARCHAR2 (100)
	// 注意：类型名和括号之间可能有空格
	if strings.Contains(typeStr, "(") {
		// 类型名后直接跟括号，如 VARCHAR2(100)
		re := regexp.MustCompile(`([A-Z0-9_]+)\(([^)]+)\)`)
		matches := re.FindStringSubmatch(strings.ToUpper(typeStr))
		if len(matches) >= 3 {
			column.DataType = matches[1]
			params := strings.Split(matches[2], ",")
			if len(params) >= 1 {
				column.Length = strings.TrimSpace(params[0])
				column.Precision = column.Length
			}
			if len(params) >= 2 {
				column.Scale = strings.TrimSpace(params[1])
			}
		}
	} else if len(parts) > 2 && strings.HasPrefix(parts[2], "(") {
		// 类型名和括号之间有空格，如 VARCHAR2 (100)
		// 提取数据类型
		column.DataType = strings.ToUpper(typeStr)

		// 查找括号内的参数
		re := regexp.MustCompile(`\(([^)]+)\)`)
		matches := re.FindStringSubmatch(def)
		if len(matches) >= 2 {
			params := strings.Split(matches[1], ",")
			if len(params) >= 1 {
				column.Length = strings.TrimSpace(params[0])
				column.Precision = column.Length
			}
			if len(params) >= 2 {
				column.Scale = strings.TrimSpace(params[1])
			}
		}
	} else {
		upperType := strings.ToUpper(typeStr)
		// 检查是否为关键字，如果是，则不视为数据类型
		reservedWords := map[string]bool{
			"NOT": true, "NULL": true, "DEFAULT": true, "PRIMARY": true,
			"KEY": true, "UNIQUE": true, "CHECK": true, "CONSTRAINT": true,
		}
		if !reservedWords[upperType] {
			column.DataType = upperType
		}
	}

	// 检查 NOT NULL
	defUpper := strings.ToUpper(def)
	if strings.Contains(defUpper, "NOT NULL") {
		column.NotNull = true
	}

	// 检查列级 PRIMARY KEY
	if strings.Contains(defUpper, "PRIMARY KEY") {
		constraint.Type = "PRIMARY KEY"
		constraint.Columns = []string{column.Name}
		// 移除 NOT NULL，因为主键隐含非空
		column.NotNull = true
	}

	// 提取默认值
	defaultRe := regexp.MustCompile(`(?i)DEFAULT\s+([^\s,]+)`)
	defaultMatches := defaultRe.FindStringSubmatch(def)
	if len(defaultMatches) >= 2 {
		column.DefaultValue = strings.TrimSpace(defaultMatches[1])
	}

	return column, constraint
}

// parseConstraint 解析约束定义
func parseConstraint(def string) Constraint {
	constraint := Constraint{}
	defUpper := strings.ToUpper(def)

	// 解析约束名称
	nameRe := regexp.MustCompile(`(?i)CONSTRAINT\s+([^\s]+)`)
	nameMatches := nameRe.FindStringSubmatch(def)
	if len(nameMatches) >= 2 {
		constraint.Name = strings.TrimSpace(nameMatches[1])
	}

	// 判断约束类型
	if strings.Contains(defUpper, "PRIMARY KEY") {
		constraint.Type = "PRIMARY KEY"
		// 提取列名
		re := regexp.MustCompile(`(?i)PRIMARY\s+KEY\s*\(([^)]+)\)`)
		matches := re.FindStringSubmatch(def)
		if len(matches) >= 2 {
			columns := strings.Split(matches[1], ",")
			for _, col := range columns {
				constraint.Columns = append(constraint.Columns, strings.TrimSpace(col))
			}
		}
	} else if strings.Contains(defUpper, "FOREIGN KEY") {
		constraint.Type = "FOREIGN KEY"
		// 提取列名
		fkRe := regexp.MustCompile(`(?i)FOREIGN\s+KEY\s*\(([^)]+)\)`)
		fkMatches := fkRe.FindStringSubmatch(def)
		if len(fkMatches) >= 2 {
			columns := strings.Split(fkMatches[1], ",")
			for _, col := range columns {
				constraint.Columns = append(constraint.Columns, strings.TrimSpace(col))
			}
		}
		// 提取引用表和列
		refRe := regexp.MustCompile(`(?i)REFERENCES\s+([^\s(]+)\s*\(([^)]+)\)`)
		refMatches := refRe.FindStringSubmatch(def)
		if len(refMatches) >= 3 {
			constraint.RefTable = strings.TrimSpace(refMatches[1])
			refColumns := strings.Split(refMatches[2], ",")
			for _, col := range refColumns {
				constraint.RefColumns = append(constraint.RefColumns, strings.TrimSpace(col))
			}
		}
	} else if strings.Contains(defUpper, "UNIQUE") {
		constraint.Type = "UNIQUE"
		// 提取列名
		re := regexp.MustCompile(`(?i)UNIQUE\s*\(([^)]+)\)`)
		matches := re.FindStringSubmatch(def)
		if len(matches) >= 2 {
			columns := strings.Split(matches[1], ",")
			for _, col := range columns {
				constraint.Columns = append(constraint.Columns, strings.TrimSpace(col))
			}
		}
	} else if strings.Contains(defUpper, "CHECK") {
		constraint.Type = "CHECK"
		// 提取 CHECK 表达式
		re := regexp.MustCompile(`(?i)CHECK\s*\((.+)\)`)
		matches := re.FindStringSubmatch(def)
		if len(matches) >= 2 {
			constraint.CheckExpr = strings.TrimSpace(matches[1])
		}
	}

	return constraint
}

// parseCreateIndex 解析 CREATE INDEX 语句
func parseCreateIndex(stmt string) (IndexDef, error) {
	index := IndexDef{}

	// 检查是否为唯一索引
	stmtUpper := strings.ToUpper(stmt)
	index.Unique = strings.Contains(stmtUpper, "UNIQUE INDEX")

	// 提取索引名和表名
	var re *regexp.Regexp
	if index.Unique {
		re = regexp.MustCompile(`(?i)CREATE\s+UNIQUE\s+INDEX\s+([^\s]+)\s+ON\s+([^\s(]+)`)
	} else {
		re = regexp.MustCompile(`(?i)CREATE\s+INDEX\s+([^\s]+)\s+ON\s+([^\s(]+)`)
	}

	matches := re.FindStringSubmatch(stmt)
	if len(matches) < 3 {
		return index, fmt.Errorf("无法解析索引定义")
	}

	index.Name = strings.TrimSpace(matches[1])
	index.Table = strings.TrimSpace(matches[2])

	// 提取列名
	colRe := regexp.MustCompile(`\(([^)]+)\)`)
	colMatches := colRe.FindStringSubmatch(stmt)
	if len(colMatches) >= 2 {
		columns := strings.Split(colMatches[1], ",")
		for _, col := range columns {
			index.Columns = append(index.Columns, strings.TrimSpace(col))
		}
	}

	return index, nil
}

// parseTableComment 解析表注释
func parseTableComment(stmt string) (string, string) {
	re := regexp.MustCompile(`(?i)COMMENT\s+ON\s+TABLE\s+([^\s]+)\s+IS\s+'([^']*)'`)
	matches := re.FindStringSubmatch(stmt)
	if len(matches) >= 3 {
		return strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2])
	}
	return "", ""
}

// parseAlterTable 解析 ALTER TABLE 语句，提取约束或索引，或列修改
func parseAlterTable(stmt string) (string, Constraint, IndexDef, bool, []ColumnDef) {
	tableName := ""
	constraint := Constraint{}
	index := IndexDef{}
	var modifiedCols []ColumnDef

	// 提取表名
	re := regexp.MustCompile(`(?i)ALTER\s+TABLE\s+([^\s]+)`)
	matches := re.FindStringSubmatch(stmt)
	if len(matches) < 2 {
		return "", constraint, index, false, nil
	}
	tableName = strings.TrimSpace(matches[1])

	stmtUpper := strings.ToUpper(stmt)

	// 1. 处理 ADD CONSTRAINT
	if strings.Contains(stmtUpper, "ADD CONSTRAINT") || strings.Contains(stmtUpper, "ADD PRIMARY KEY") ||
		strings.Contains(stmtUpper, "ADD UNIQUE") || strings.Contains(stmtUpper, "ADD FOREIGN KEY") {

		// 找到 ADD 之后的内容
		addIdx := strings.Index(stmtUpper, "ADD")
		if addIdx != -1 {
			constraintDef := stmt[addIdx+4:]
			constraintDef = strings.TrimSpace(constraintDef)
			if !strings.HasPrefix(strings.ToUpper(constraintDef), "CONSTRAINT") {
				constraintDef = "CONSTRAINT " + constraintDef
			}
			constraint = parseConstraint(constraintDef)

			// 如果 parseConstraint 没解析到类型，尝试直接寻找类型
			if constraint.Type == "" {
				if strings.Contains(stmtUpper, "PRIMARY KEY") {
					constraint.Type = "PRIMARY KEY"
					rePK := regexp.MustCompile(`(?i)PRIMARY\s+KEY\s*\(([^)]+)\)`)
					mPK := rePK.FindStringSubmatch(stmt)
					if len(mPK) >= 2 {
						for _, col := range strings.Split(mPK[1], ",") {
							constraint.Columns = append(constraint.Columns, strings.TrimSpace(col))
						}
					}
				} else if strings.Contains(stmtUpper, "UNIQUE") {
					// Handle direct ADD UNIQUE (col)
					constraint.Type = "UNIQUE"
					reUnique := regexp.MustCompile(`(?i)UNIQUE\s*\(([^)]+)\)`)
					mUnique := reUnique.FindStringSubmatch(stmt)
					if len(mUnique) >= 2 {
						for _, col := range strings.Split(mUnique[1], ",") {
							constraint.Columns = append(constraint.Columns, strings.TrimSpace(col))
						}
					}
				}
			}
			return tableName, constraint, index, false, nil
		}
	} else if strings.Contains(stmtUpper, "MODIFY") {
		// 2. 处理 MODIFY
		// ALTER TABLE table MODIFY col type NOT NULL
		// ALTER TABLE table MODIFY (col type ...)

		// 简化处理：假设 MODIFY 后跟的是列定义
		modifyIdx := strings.Index(stmtUpper, "MODIFY")
		if modifyIdx != -1 {
			def := stmt[modifyIdx+6:]
			def = strings.TrimSpace(def)

			// 去掉可能的括号
			if strings.HasPrefix(def, "(") && strings.HasSuffix(def, ")") {
				def = def[1 : len(def)-1]
			}

			// 可能有多个列修改，用逗号分隔
			parts := smartSplit(def, ',')
			for _, part := range parts {
				col, _ := parseColumnDef(part)
				if col.Name != "" {
					modifiedCols = append(modifiedCols, col)
				}
			}

			return tableName, constraint, index, false, modifiedCols
		}
	}

	return "", constraint, index, false, nil
}

// parseColumnComment 解析列注释
func parseColumnComment(stmt string) (string, string, string) {
	re := regexp.MustCompile(`(?i)COMMENT\s+ON\s+COLUMN\s+([^.]+)\.([^\s]+)\s+IS\s+'([^']*)'`)
	matches := re.FindStringSubmatch(stmt)
	if len(matches) >= 4 {
		return strings.TrimSpace(matches[1]), strings.TrimSpace(matches[2]), strings.TrimSpace(matches[3])
	}
	return "", "", ""
}

// smartSplit 智能分割，考虑括号嵌套
func smartSplit(s string, sep rune) []string {
	var result []string
	var current strings.Builder
	var parenDepth int

	for _, ch := range s {
		if ch == '(' {
			parenDepth++
			current.WriteRune(ch)
		} else if ch == ')' {
			parenDepth--
			current.WriteRune(ch)
		} else if ch == sep && parenDepth == 0 {
			result = append(result, current.String())
			current.Reset()
		} else {
			current.WriteRune(ch)
		}
	}

	if current.Len() > 0 {
		result = append(result, current.String())
	}

	return result
}
