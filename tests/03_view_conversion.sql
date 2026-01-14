-- 测试 CREATE VIEW 和 SQL 语句转换

-- 基本视图
CREATE VIEW v_active_employees AS
SELECT emp_id, emp_name, hire_date
FROM employees
WHERE status = 1;

COMMENT ON VIEW v_active_employees IS '活跃员工视图';

-- 带列名列表的视图
CREATE OR REPLACE VIEW v_employee_summary (dept_id, emp_count, avg_salary) AS
SELECT dept_id, COUNT(*) as emp_count, AVG(salary) as avg_salary
FROM employees
WHERE status = 1
GROUP BY dept_id;

-- 复杂函数转换的视图
CREATE VIEW v_employee_info AS
SELECT 
  emp_id,
  emp_name,
  NVL(email, 'no-email') as email,
  TO_CHAR(hire_date, 'YYYY-MM-DD') as hire_date_str,
  SUBSTR(emp_name, 1, 10) as short_name,
  SYSDATE as current_time,
  TRUNC(hire_date) as hire_date_only,
  MONTHS_BETWEEN(SYSDATE, hire_date) as months_worked
FROM employees;

-- MyBatis 风格的 SQL（模拟）
-- SELECT emp_id, emp_name, 
--   NVL(bonus, 0) as bonus,
--   TO_CHAR(created_date,  'YYYY-MM-DD HH24:MI:SS') as created_str
-- FROM employees
-- WHERE 1=1
-- <if test="name != null">
--   AND emp_name LIKE CONCAT('%', #{name}, '%')
-- </if>
-- AND created_date >= SYSDATE - 30

-- 字符串连接测试
CREATE VIEW v_full_name AS
SELECT emp_id, emp_name || ' - ' || dept_name as full_info
FROM employees e, departments d
WHERE e.dept_id = d.dept_id;

-- DECODE 函数转换
CREATE VIEW v_employee_status AS
SELECT 
  emp_id,
  emp_name,
  DECODE(status, 1, 'Active', 0, 'Inactive', 'Unknown') as status_text
FROM employees;
