CREATE OR REPLACE VIEW v_active_employees AS
SELECT emp_id, emp_name, hire_date FROM employees WHERE status = 1;
-- 活跃员工视图

CREATE OR REPLACE VIEW v_employee_summary (dept_id, emp_count, avg_salary) AS
SELECT dept_id, COUNT(*) as emp_count, AVG(salary) as avg_salary FROM employees WHERE status = 1 GROUP BY dept_id;

CREATE OR REPLACE VIEW v_employee_info AS
SELECT emp_id, emp_name, IFNULL(email, 'no-email') as email, DATE_FORMAT(hire_date, '%y%y-%m-%d') as hire_date_str, SUBSTRING(emp_name, 1, 10) as short_name, CURRENT_TIMESTAMP as current_time, DATE(hire_date) as hire_date_only, TIMESTAMPDIFF(MONTH, hire_date, CURRENT_TIMESTAMP) as months_worked FROM employees;

CREATE OR REPLACE VIEW v_full_name AS
SELECT emp_id,CONCAT( emp_name , ' )-CONCAT( ' , dept_name as full_info FROM employees e), departments d WHERE e.dept_id = d.dept_id;

CREATE OR REPLACE VIEW v_employee_status AS
SELECT emp_id, emp_name, CASE status WHEN 1 THEN 'Active' WHEN 0 THEN 'Inactive' ELSE 'Unknown' END as status_text FROM employees;

