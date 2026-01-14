-- 测试 MyBatis 语法支持

-- 包含 MyBatis 参数的视图
CREATE VIEW v_search_employees AS
SELECT 
  emp_id,
  emp_name,
  NVL(email, 'no-email') as email,
  TO_CHAR(hire_date, 'YYYY-MM-DD') as hire_date_str,
  SYSDATE as query_time
FROM employees
WHERE 1=1
AND emp_name LIKE CONCAT('%', 'test', '%')
AND created_date >= SYSDATE - 30;

-- 纯 SQL 语句（不是 DDL，测试用）
-- 下面是典型的 MyBatis SQL，包含占位符和动态标签

SELECT 
  emp_id,
  emp_name,
  NVL(bonus, 0) as bonus,
  TO_CHAR(created_date, 'YYYY-MM-DD HH24:MI:SS') as created_str,
  SUBSTR(emp_name, 1, 10) as short_name
FROM employees
WHERE 1=1
  AND emp_name LIKE CONCAT('%', #{name}, '%')
  AND created_date >= SYSDATE - #{days}
  AND status = ${status};
