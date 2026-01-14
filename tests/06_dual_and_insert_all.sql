-- =====================================================
-- 测试用例：DUAL 表和 INSERT ALL
-- 文件：06_dual_and_insert_all.sql
-- 描述：测试 DUAL 表移除和 INSERT ALL 语法转换
-- =====================================================

-- 测试 DUAL 表移除
SELECT 1 FROM DUAL;
SELECT SYSDATE FROM DUAL;
SELECT 'Hello' || ' World' FROM DUAL;

-- 测试 INSERT ALL (Oracle 批量插入)
INSERT ALL
    INTO users (user_id, username, email) VALUES (1, 'user1', 'user1@example.com')
    INTO users (user_id, username, email) VALUES (2, 'user2', 'user2@example.com')
    INTO users (user_id, username, email) VALUES (3, 'user3', 'user3@example.com')
SELECT * FROM DUAL;
