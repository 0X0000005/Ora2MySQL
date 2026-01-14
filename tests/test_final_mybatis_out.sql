SELECT 
  emp_id,
  emp_name,
  IFNULL(email, 'no-email@company.com') as email,
  DATE_FORMAT(hire_date, '%Y-%m-%d') as hire_date_str,
  SUBSTRING(emp_name, 1, 10) as short_name,
  CASE status WHEN 1 THEN 'Active' WHEN 0 THEN 'Inactive' ELSE 'Unknown' END as status_text
FROM employees
WHERE emp_name LIKE CONCAT('%', #{searchName}, '%')
  AND status = #{status}
  AND dept_id = ${deptId}
  AND created_date >= CURRENT_TIMESTAMP - #{days}
ORDER BY hire_date DESC;
