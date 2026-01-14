# Ora2MySQL 测试用例说明

本目录包含完整的测试用例集，用于验证 Oracle 到 MySQL 的转换功能。

## 测试文件说明

### DDL 测试
- **01_ddl_basic.sql** - 基本 DDL 测试（CREATE TABLE, 约束, 索引, 注释）
- **02_ddl_alter_modify.sql** - ALTER TABLE MODIFY 测试（列修改, 约束添加）
- **03_view_conversion.sql** - 视图转换测试

### DML 和 MyBatis 测试
- **04_mybatis_basic.sql** - MyBatis 基本语法保留测试
- **test_mybatis.sql** - MyBatis 复杂场景测试
- **UserMapper.xml** - 真实 MyBatis Mapper 文件示例
- **CheckinMapper.xml** - MyBatis 动态 SQL 示例

### 函数转换测试
- **05_function_conversions.sql** - Oracle 函数转换测试（全面覆盖）
- **06_dual_and_insert_all.sql** - DUAL 表和 INSERT ALL 测试

### 综合测试
- **1.sql** - 综合场景测试
- **test_o.sql** - 原始 Oracle SQL 示例

## 运行测试

### 单个测试
```bash
o2m.exe -i tests/01_ddl_basic.sql -o output/01_result.sql
```

### 批量测试 (PowerShell)
```powershell
Get-ChildItem tests\*.sql | ForEach-Object {
    $outputFile = "output\$($_.BaseName)_output.sql"
    .\o2m.exe -i $_.FullName -o $outputFile
}
```

### 批量测试 (Bash)
```bash
for file in tests/*.sql; do
    output="output/$(basename ${file%.sql})_output.sql"
    ./o2m-linux -i "$file" -o "$output"
done
```

## 预期结果

所有测试文件转换后应符合：
1. ✅ MySQL 语法规范
2. ✅ 保留 MyBatis 标签和参数
3. ✅ Oracle 函数正确转换为 MySQL 函数
4. ✅ DUAL 表被正确移除
5. ✅ 注释和约束完整保留
