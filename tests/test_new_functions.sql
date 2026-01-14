-- 测试新增函数转换

-- NVL2 测试
SELECT NVL2(email, '有邮箱', '无邮箱') FROM users;

-- LENGTH 测试
SELECT LENGTH(username) FROM users;

-- 系统函数测试
SELECT USER, UID, SYS_GUID() FROM DUAL;

-- LISTAGG 测试
SELECT dept_id, LISTAGG(emp_name, ', ') FROM employees GROUP BY dept_id;
