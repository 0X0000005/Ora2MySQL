// ============ SQL 格式化器（本地实现）============
const SQLFormatter = {
    keywords: [
        'SELECT', 'FROM', 'WHERE', 'INSERT', 'UPDATE', 'DELETE', 'CREATE', 'DROP', 'ALTER',
        'TABLE', 'VIEW', 'INDEX', 'DATABASE', 'SCHEMA', 'PRIMARY', 'KEY', 'FOREIGN',
        'CONSTRAINT', 'UNIQUE', 'NOT', 'NULL', 'DEFAULT', 'AUTO_INCREMENT', 'CASCADE',
        'ON', 'REFERENCES', 'AS', 'JOIN', 'LEFT', 'RIGHT', 'INNER', 'OUTER', 'FULL',
        'UNION', 'ALL', 'DISTINCT', 'ORDER', 'BY', 'GROUP', 'HAVING', 'LIMIT', 'OFFSET',
        'INTO', 'VALUES', 'SET', 'AND', 'OR', 'IN', 'BETWEEN', 'LIKE', 'IS', 'EXISTS',
        'CASE', 'WHEN', 'THEN', 'ELSE', 'END', 'IF', 'BEGIN', 'COMMIT', 'ROLLBACK',
        'GRANT', 'REVOKE', 'WITH', 'RECURSIVE', 'ENGINE', 'CHARSET', 'COLLATE', 'COMMENT',
        'ADD', 'MODIFY', 'CHANGE', 'RENAME', 'TRUNCATE', 'REPLACE', 'CHECK', 'TINYINT',
        'SMALLINT', 'INT', 'BIGINT', 'DECIMAL', 'VARCHAR', 'CHAR', 'TEXT', 'LONGTEXT',
        'DATETIME', 'TIMESTAMP', 'BLOB', 'LONGBLOB', 'VARBINARY'
    ],

    functions: [
        'COUNT', 'SUM', 'AVG', 'MAX', 'MIN', 'UPPER', 'LOWER', 'LENGTH', 'SUBSTRING',
        'CONCAT', 'NOW', 'CURRENT_TIMESTAMP', 'DATE', 'TIME', 'YEAR', 'MONTH', 'DAY',
        'IFNULL', 'COALESCE', 'CAST', 'CONVERT', 'ROUND', 'FLOOR', 'CEIL', 'ABS'
    ],

    format(sql, indent = '  ') {
        if (!sql || !sql.trim()) return '';

        let formatted = sql;

        // 移除多余空格
        formatted = formatted.replace(/\s+/g, ' ').trim();

        // 在主要关键字前添加换行
        const majorKeywords = ['SELECT', 'FROM', 'WHERE', 'JOIN', 'LEFT JOIN', 'RIGHT JOIN',
            'INNER JOIN', 'GROUP BY', 'ORDER BY', 'HAVING', 'LIMIT'];

        majorKeywords.forEach(kw => {
            const regex = new RegExp('\\b' + kw + '\\b', 'gi');
            formatted = formatted.replace(regex, '\n' + kw);
        });

        // 处理 CREATE TABLE 语句
        formatted = formatted.replace(/CREATE\s+TABLE/gi, 'CREATE TABLE');
        formatted = formatted.replace(/\(\s*/g, ' (\n' + indent);
        formatted = formatted.replace(/\s*\)/g, '\n)');
        formatted = formatted.replace(/,\s*/g, ',\n' + indent);

        // 关键字大写
        this.keywords.forEach(kw => {
            const regex = new RegExp('\\b' + kw + '\\b', 'gi');
            formatted = formatted.replace(regex, kw.toUpperCase());
        });

        // 清理多余空行
        formatted = formatted.replace(/\n\s*\n/g, '\n');
        formatted = formatted.trim();

        return formatted;
    }
};

// ============ SQL 语法高亮器（本地实现）============
const SQLHighlighter = {
    keywords: SQLFormatter.keywords,
    functions: SQLFormatter.functions,

    highlight(sql) {
        if (!sql) return '';

        let highlighted = this.escapeHtml(sql);

        // 高亮注释（-- 和 /* */）
        highlighted = highlighted.replace(/--([^\n]*)/g, '<span class="sql-comment">--$1</span>');
        highlighted = highlighted.replace(/\/\*([^*]|\*(?!\/))*\*\//g, '<span class="sql-comment">$&</span>');

        // 高亮字符串
        highlighted = highlighted.replace(/'([^']|'')*'/g, '<span class="sql-string">$&</span>');

        // 高亮数字
        highlighted = highlighted.replace(/\b\d+(\.\d+)?\b/g, '<span class="sql-number">$&</span>');

        // 高亮关键字
        this.keywords.forEach(kw => {
            const regex = new RegExp('\\b(' + kw + ')\\b', 'gi');
            highlighted = highlighted.replace(regex, '<span class="sql-keyword">$1</span>');
        });

        // 高亮函数
        this.functions.forEach(fn => {
            const regex = new RegExp('\\b(' + fn + ')\\b', 'gi');
            highlighted = highlighted.replace(regex, '<span class="sql-function">$1</span>');
        });

        return highlighted;
    },

    escapeHtml(text) {
        const div = document.createElement('div');
        div.textContent = text;
        return div.innerHTML;
    }
};

// ============ 格式化和高亮功能 ============

function formatInput() {
    const textarea = document.getElementById('oracleInput');
    const sql = textarea.value.trim();

    if (!sql) {
        showAlert('请先输入 SQL 语句', 'error');
        return;
    }

    try {
        const formatted = SQLFormatter.format(sql);
        textarea.value = formatted;
        updateLineNumbers();
        showAlert('SQL 格式化成功！', 'success');
    } catch (error) {
        showAlert('格式化失败: ' + error.message, 'error');
    }
}

function formatOutput() {
    const outputElement = document.getElementById('mysqlOutput');
    const sql = outputElement.textContent.trim();

    if (!sql) {
        showAlert('没有可格式化的内容', 'error');
        return;
    }

    try {
        const formatted = SQLFormatter.format(sql);
        const lines = formatted.split('\n');
        updateResultLineNumbers(lines.length);
        // 格式化后自动应用高亮
        const highlighted = SQLHighlighter.highlight(formatted);
        outputElement.innerHTML = highlighted;
        showAlert('SQL 格式化并高亮成功！', 'success');
    } catch (error) {
        showAlert('格式化失败: ' + error.message, 'error');
    }
}

function highlightOutput() {
    const outputElement = document.getElementById('mysqlOutput');
    const sql = outputElement.textContent.trim();

    if (!sql) {
        showAlert('没有可高亮的内容', 'error');
        return;
    }

    try {
        const highlighted = SQLHighlighter.highlight(sql);
        outputElement.innerHTML = highlighted;
        showAlert('SQL 高亮成功！', 'success');
    } catch (error) {
        showAlert('高亮失败: ' + error.message, 'error');
    }
}
