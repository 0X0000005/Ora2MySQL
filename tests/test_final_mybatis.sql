SELECT 
  emp_id,
  emp_name,
  NVL(email, 'no-email@company.com') as email,
  TO_CHAR(hire_date, 'YYYY-MM-DD') as hire_date_str,
  SUBSTR(emp_name, 1, 10) as short_name,
  DECODE(status, 1, 'Active', 0, 'Inactive', 'Unknown') as status_text
FROM employees
WHERE emp_name LIKE CONCAT('%', #{searchName}, '%')
  AND status = #{status}
  AND dept_id = ${deptId}
  AND created_date >= SYSDATE - #{days}
ORDER BY hire_date DESC;
