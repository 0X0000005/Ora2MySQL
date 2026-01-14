package converter

import (
	"fmt"
	"regexp"
	"strings"
)

// ConvertSQLStatement 转换 SQL 语句（保留 MyBatis 语法）
func ConvertSQLStatement(sql string) string {
	// 保护 MyBatis 语法标记
	protected, placeholders := protectMyBatisSyntax(sql)

	// 转换 Oracle 函数
	converted := convertOracleFunctions(protected)

	// 转换连接运算符
	converted = convertConcatenation(converted)

	// 转换 DUAL 表
	converted = convertDualTable(converted)

	// 转换外连接语法
	converted = convertOracleJoins(converted)

	// 恢复 MyBatis 语法
	result := restoreMyBatisSyntax(converted, placeholders)

	return result
}

// protectMyBatisSyntax 保护 MyBatis 语法（#{}, ${}, XML 标签）
func protectMyBatisSyntax(sql string) (string, map[string]string) {
	placeholders := make(map[string]string)
	result := sql
	counter := 0

	// 保护 #{...}
	re1 := regexp.MustCompile(`#\{[^}]+\}`)
	result = re1.ReplaceAllStringFunc(result, func(match string) string {
		placeholder := fmt.Sprintf("___MYBATIS_PARAM_%d___", counter)
		placeholders[placeholder] = match
		counter++
		return placeholder
	})

	// 保护 ${...}
	re2 := regexp.MustCompile(`\$\{[^}]+\}`)
	result = re2.ReplaceAllStringFunc(result, func(match string) string {
		placeholder := fmt.Sprintf("___MYBATIS_EXPR_%d___", counter)
		placeholders[placeholder] = match
		counter++
		return placeholder
	})

	// 保护 XML 标签（如 <if>, <where>, <foreach> 等）
	re3 := regexp.MustCompile(`<[^>]+>`)
	result = re3.ReplaceAllStringFunc(result, func(match string) string {
		// 只保护 MyBatis 标签，不保护普通文本中的 < >
		tagLower := strings.ToLower(match) // Convert to lowercase for case-insensitive matching
		if strings.Contains(tagLower, "if ") || strings.Contains(tagLower, "/if") ||
			strings.Contains(tagLower, "where") || strings.Contains(tagLower, "foreach") ||
			strings.Contains(tagLower, "choose") || strings.Contains(tagLower, "when") ||
			strings.Contains(tagLower, "otherwise") || strings.Contains(tagLower, "set") ||
			strings.Contains(tagLower, "trim") || strings.Contains(tagLower, "bind") {
			placeholder := fmt.Sprintf("___MYBATIS_TAG_%d___", counter)
			placeholders[placeholder] = match
			counter++
			return placeholder
		}
		return match
	})

	return result, placeholders
}

// restoreMyBatisSyntax 恢复 MyBatis 语法
func restoreMyBatisSyntax(sql string, placeholders map[string]string) string {
	result := sql
	for placeholder, original := range placeholders {
		result = strings.ReplaceAll(result, placeholder, original)
	}
	return result
}

// convertOracleFunctions 转换 Oracle 函数为 MySQL 函数
func convertOracleFunctions(sql string) string {
	result := sql

	// SYSDATE → CURRENT_TIMESTAMP 或 NOW()
	result = regexp.MustCompile(`(?i)\bSYSDATE\b`).ReplaceAllString(result, "CURRENT_TIMESTAMP")

	// SYSTIMESTAMP → CURRENT_TIMESTAMP
	result = regexp.MustCompile(`(?i)\bSYSTIMESTAMP\b`).ReplaceAllString(result, "CURRENT_TIMESTAMP")

	// NVL(a, b) → IFNULL(a, b)
	result = regexp.MustCompile(`(?i)\bNVL\s*\(`).ReplaceAllString(result, "IFNULL(")

	// NVL2(expr, val1, val2) → IF(expr IS NOT NULL, val1, val2)
	result = convertNVL2(result)

	// TO_CHAR(date, format) → DATE_FORMAT(date, format)
	// 需要转换日期格式字符串
	result = convertToChar(result)

	// TO_DATE(str, format) → STR_TO_DATE(str, format)
	result = convertToDate(result)

	// SUBSTR → SUBSTRING
	result = regexp.MustCompile(`(?i)\bSUBSTR\s*\(`).ReplaceAllString(result, "SUBSTRING(")

	// INSTR(str, substr) → LOCATE(substr, str) - 注意参数顺序相反
	result = convertInstr(result)

	// LENGTH → CHAR_LENGTH
	result = regexp.MustCompile(`(?i)\bLENGTH\s*\(`).ReplaceAllString(result, "CHAR_LENGTH(")

	// TRUNC(date) → DATE(date)
	result = regexp.MustCompile(`(?i)\bTRUNC\s*\(\s*([^,)]+)\s*\)`).ReplaceAllString(result, "DATE($1)")

	// DECODE → CASE WHEN
	result = convertDecode(result)

	// MONTHS_BETWEEN → TIMESTAMPDIFF(MONTH, ...)
	result = regexp.MustCompile(`(?i)\bMONTHS_BETWEEN\s*\(\s*([^,]+),\s*([^)]+)\)`).
		ReplaceAllString(result, "TIMESTAMPDIFF(MONTH, $2, $1)")

	// ADD_MONTHS(date, n) → DATE_ADD(date, INTERVAL n MONTH)
	result = regexp.MustCompile(`(?i)\bADD_MONTHS\s*\(\s*([^,]+),\s*([^)]+)\)`).
		ReplaceAllString(result, "DATE_ADD($1, INTERVAL $2 MONTH)")

	// System functions
	// USER → CURRENT_USER()
	result = regexp.MustCompile(`(?i)\bUSER\b`).ReplaceAllString(result, "CURRENT_USER()")

	// UID → USER()
	result = regexp.MustCompile(`(?i)\bUID\b`).ReplaceAllString(result, "USER()")

	// SYS_GUID() → UUID()
	result = regexp.MustCompile(`(?i)\bSYS_GUID\s*\(\s*\)`).ReplaceAllString(result, "UUID()")

	// Aggregation functions
	// LISTAGG → GROUP_CONCAT
	result = convertListagg(result)

	return result
}

// convertToChar 转换 TO_CHAR 函数
func convertToChar(sql string) string {
	// TO_CHAR(date, 'YYYY-MM-DD') → DATE_FORMAT(date, '%Y-%m-%d')
	re := regexp.MustCompile(`(?i)TO_CHAR\s*\(\s*([^,]+),\s*'([^']+)'\s*\)`)
	return re.ReplaceAllStringFunc(sql, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) >= 3 {
			expr := parts[1]
			oracleFormat := parts[2]
			mysqlFormat := convertDateFormat(oracleFormat)
			return "DATE_FORMAT(" + expr + ", '" + mysqlFormat + "')"
		}
		return match
	})
}

// convertToDate 转换 TO_DATE 函数
func convertToDate(sql string) string {
	// TO_DATE(str, 'YYYY-MM-DD') → STR_TO_DATE(str, '%Y-%m-%d')
	re := regexp.MustCompile(`(?i)TO_DATE\s*\(\s*([^,]+),\s*'([^']+)'\s*\)`)
	return re.ReplaceAllStringFunc(sql, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) >= 3 {
			expr := parts[1]
			oracleFormat := parts[2]
			mysqlFormat := convertDateFormat(oracleFormat)
			return "STR_TO_DATE(" + expr + ", '" + mysqlFormat + "')"
		}
		return match
	})
}

// convertDateFormat 转换日期格式字符串（Oracle → MySQL）
func convertDateFormat(oracleFormat string) string {
	mysqlFormat := oracleFormat

	// Oracle → MySQL 格式映射
	replacements := map[string]string{
		"YYYY": "%Y", "YY": "%y",
		"MM": "%m", "MON": "%b", "MONTH": "%M",
		"DD": "%d", "DY": "%a", "DAY": "%W",
		"HH24": "%H", "HH12": "%h", "HH": "%h",
		"MI": "%i", "SS": "%s",
		"AM": "%p", "PM": "%p",
	}

	for oracle, mysql := range replacements {
		mysqlFormat = strings.ReplaceAll(mysqlFormat, oracle, mysql)
	}

	return mysqlFormat
}

// convertInstr 转换 INSTR 函数（参数顺序不同）
func convertInstr(sql string) string {
	// INSTR(string, substring) → LOCATE(substring, string)
	re := regexp.MustCompile(`(?i)INSTR\s*\(\s*([^,]+),\s*([^)]+)\)`)
	return re.ReplaceAllString(sql, "LOCATE($2, $1)")
}

// convertDecode 转换 DECODE 为 CASE WHEN
func convertDecode(sql string) string {
	// DECODE(col, val1, res1, val2, res2, default) → CASE col WHEN val1 THEN res1 WHEN val2 THEN res2 ELSE default END
	// 这个转换比较复杂，需要解析参数
	re := regexp.MustCompile(`(?i)DECODE\s*\(([^)]+)\)`)

	return re.ReplaceAllStringFunc(sql, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) < 2 {
			return match
		}

		args := smartSplit(parts[1], ',')
		if len(args) < 3 {
			return match // 参数不足
		}

		expr := strings.TrimSpace(args[0])
		caseStmt := "CASE " + expr

		// 处理成对的 value, result
		i := 1
		for i < len(args)-1 {
			value := strings.TrimSpace(args[i])
			result := strings.TrimSpace(args[i+1])
			caseStmt += " WHEN " + value + " THEN " + result
			i += 2
		}

		// 如果有剩余参数，作为 ELSE 子句
		if i < len(args) {
			defaultVal := strings.TrimSpace(args[i])
			caseStmt += " ELSE " + defaultVal
		}

		caseStmt += " END"
		return caseStmt
	})
}

// convertConcatenation 转换字符串连接运算符
func convertConcatenation(sql string) string {
	// Oracle 的 || 连接符 → MySQL 的 CONCAT()

	// 匹配 || 两边的操作数
	// 操作数可以是：'string', ___MYBATIS_PARAM_N___, ___MYBATIS_EXPR_N___, 普通列名，或者已经是 CONCAT(...) 的结果
	token := `(?:'[^']*'|___MYBATIS_[A-Z]+_\d+___|[a-zA-Z0-9_.]+|CONCAT\s*\(.*?\))`
	re := regexp.MustCompile(`(` + token + `)\s*\|\|\s*(` + token + `)`)

	// 多次替换，直到没有 || 为止
	for re.MatchString(sql) {
		sql = re.ReplaceAllString(sql, "CONCAT($1, $2)")
	}

	return sql
}

// convertDualTable 转换 DUAL 表
func convertDualTable(sql string) string {
	// 1. 移除 INSERT ALL 后面的 SELECT * FROM DUAL
	// 这是 Oracle 特有的批量插入语法，使用 (?s) 允许 . 匹配换行符
	sql = regexp.MustCompile(`(?is)SELECT\s+\*\s+FROM\s+DUAL`).ReplaceAllString(sql, "")

	// 2. 移除其他 FROM DUAL 引用（保留 SELECT 部分）
	// 例如: SELECT SYSDATE FROM DUAL -> SELECT SYSDATE
	sql = regexp.MustCompile(`(?i)\s+FROM\s+DUAL\b`).ReplaceAllString(sql, "")

	return sql
}

// convertOracleJoins 转换 Oracle 外连接语法
func convertOracleJoins(sql string) string {
	// Oracle: WHERE a.id = b.id(+) → MySQL: LEFT JOIN
	// Oracle: WHERE a.id(+) = b.id → MySQL: RIGHT JOIN

	// 这个转换非常复杂，需要重写 WHERE 子句为 JOIN 语法
	// 暂时保留原样，建议手动转换
	// TODO: 实现复杂的 JOIN 转换

	return sql
}

// convertNVL2 转换 NVL2 函数
func convertNVL2(sql string) string {
	// NVL2(expr, val1, val2) → IF(expr IS NOT NULL, val1, val2)
	re := regexp.MustCompile(`(?i)NVL2\s*\(\s*([^,]+),\s*([^,]+),\s*([^)]+)\)`)
	return re.ReplaceAllStringFunc(sql, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) >= 4 {
			expr := strings.TrimSpace(parts[1])
			val1 := strings.TrimSpace(parts[2])
			val2 := strings.TrimSpace(parts[3])
			return fmt.Sprintf("IF(%s IS NOT NULL, %s, %s)", expr, val1, val2)
		}
		return match
	})
}

// convertListagg 转换 LISTAGG 函数
func convertListagg(sql string) string {
	// LISTAGG(column, separator) WITHIN GROUP (ORDER BY ...) → GROUP_CONCAT(column ORDER BY ... SEPARATOR separator)
	// 简化版本：LISTAGG(column, separator) → GROUP_CONCAT(column SEPARATOR separator)
	re := regexp.MustCompile(`(?i)LISTAGG\s*\(\s*([^,]+),\s*([^)]+)\)`)
	return re.ReplaceAllStringFunc(sql, func(match string) string {
		parts := re.FindStringSubmatch(match)
		if len(parts) >= 3 {
			column := strings.TrimSpace(parts[1])
			separator := strings.TrimSpace(parts[2])
			return fmt.Sprintf("GROUP_CONCAT(%s SEPARATOR %s)", column, separator)
		}
		return match
	})
}
