-- 测试 ROWNUM 转换功能

-- 测试1: 子查询 + ROWNUM = 1（用户提供的例子）
<!-- Oracle ROWNUM 示例：获取第一个用户记录 -->
SELECT 
    USER_ID, USERNAME, CREATED_DATE
FROM (
    SELECT 
        USER_ID, USERNAME, CREATED_DATE
    FROM T_USER
    ORDER BY CREATED_DATE DESC  -- 按创建日期排序获取最新用户
)
WHERE ROWNUM = 1;

-- 测试2: 子查询 + ROWNUM <= N
SELECT * FROM (
    SELECT * FROM T_USER ORDER BY ID
) WHERE ROWNUM <= 10;

-- 测试3: 简单 ROWNUM
SELECT * FROM T_USER WHERE ROWNUM <= 5;

-- 测试4: MyBatis 中的 ROWNUM
<select id="getFirstUser" resultMap="BaseResultMap">
    SELECT 
        <include refid="Base_Column_List"/>
    FROM (
        SELECT 
            <include refid="Base_Column_List"/>
        FROM T_USER
        ORDER BY CREATED_DATE DESC
    )
    WHERE ROWNUM = 1
</select>

-- 测试5: 获取TOP 3
SELECT * FROM (
    SELECT EMPLOYEE_ID, NAME, SALARY 
    FROM EMP 
    ORDER BY SALARY DESC
) WHERE ROWNUM <= 3;
