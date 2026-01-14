CREATE OR REPLACE VIEW v_search_employees AS
SELECT emp_id, emp_name, IFNULL(email, 'no-email') as email, DATE_FORMAT(hire_date, '%y%y-%m-%d') as hire_date_str, CURRENT_TIMESTAMP as query_time FROM employees WHERE 1=1 AND emp_name LIKE CONCAT('%', 'test', '%') AND created_date >= CURRENT_TIMESTAMP - 30;

