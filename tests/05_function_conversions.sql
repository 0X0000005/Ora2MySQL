-- =====================================================
-- 测试用例：Oracle 函数转换
-- 文件：05_function_conversions.sql
-- 描述：测试各类 Oracle 函数到 MySQL 的转换
-- =====================================================

-- 一、字符串函数
SELECT SUBSTR('Hello World', 1, 5) FROM DUAL;
SELECT INSTR('Hello World', 'World') FROM DUAL;
SELECT LENGTH('Hello') FROM DUAL;
SELECT LPAD('123', 5, '0') FROM DUAL;
SELECT RPAD('123', 5, '0') FROM DUAL;
SELECT TRIM(' Hello ') FROM DUAL;
SELECT 'Hello' || ' ' || 'World' FROM DUAL;

-- 二、日期/时间函数
SELECT SYSDATE FROM DUAL;
SELECT SYSTIMESTAMP FROM DUAL;
SELECT TRUNC(SYSDATE) FROM DUAL;
SELECT ADD_MONTHS(SYSDATE, 3) FROM DUAL;
SELECT MONTHS_BETWEEN(SYSDATE, ADD_MONTHS(SYSDATE, -6)) FROM DUAL;
SELECT LAST_DAY(SYSDATE) FROM DUAL;

-- 三、空值/判断函数
SELECT NVL(NULL, 'default') FROM DUAL;
SELECT NVL2(email, '有邮箱', '无邮箱') FROM users;
SELECT DECODE(status, 1, '正常', 2, '停用', '未知') FROM users;

-- 四、系统函数
SELECT USER FROM DUAL;
SELECT UID FROM DUAL;
SELECT SYS_GUID() FROM DUAL;

-- 五、聚合函数
SELECT dept_id, LISTAGG(emp_name, ', ') WITHIN GROUP (ORDER BY emp_name) 
FROM employees GROUP BY dept_id;

-- 六、数据类型转换
SELECT TO_CHAR(SYSDATE, 'YYYY-MM-DD') FROM DUAL;
SELECT TO_DATE('2024-01-01', 'YYYY-MM-DD') FROM DUAL;
