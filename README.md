# Oracle 转 MySQL DDL 工具

这是一个用 Go 语言编写的命令行工具，用于将 Oracle DDL 语句转换为 MySQL 兼容的 DDL 语句。

## 功能特性

- ✅ 支持 CREATE TABLE 语句转换
- ✅ 数据类型自动转换（VARCHAR2、NUMBER、DATE、CLOB、BLOB 等）
- ✅ 约束转换（主键、外键、唯一键、检查约束）
- ✅ 索引转换（普通索引和唯一索引）
- ✅ 表注释和列注释转换
- ✅ 默认值转换（包括 SYSDATE 等 Oracle 函数）
- ✅ 支持块注释（`/* ... */`）的正确解析
- ✅ 支持 ALTER TABLE MODIFY 语句（修改列约束）
- ✅ 智能输入验证（拒绝非法输入，兼容 MyBatis 文件）
- ✅ 全中文注释

## 数据类型转换对照表

| Oracle 类型 | MySQL 类型 | 说明 |
|------------|-----------|------|
| VARCHAR2(n) | VARCHAR(n) | 变长字符串 |
| CHAR(n) | CHAR(n) | 定长字符串 |
| NUMBER | DECIMAL(10,0) | 默认数值类型 |
| NUMBER(p) | INT/BIGINT | 根据精度选择 |
| NUMBER(p,s) | DECIMAL(p,s) | 带小数的数值 |
| DATE | DATETIME | 日期时间 |
| TIMESTAMP | DATETIME | 时间戳 |
| CLOB | LONGTEXT | 大文本 |
| BLOB | LONGBLOB | 大二进制对象 |

## 编译

### 快速编译（推荐）

**Windows 系统：**
```bash
build.bat
```

**Linux/macOS 系统：**
```bash
chmod +x build.sh
./build.sh
```

编译脚本会自动：
1. 编译 Windows 和 Linux 版本（macOS 脚本还会编译 macOS 版本）
2. 使用 `-ldflags="-s -w"` 减小二进制文件大小
3. 使用 UPX 进一步压缩（如果已安装）
4. 显示生成的文件大小

### 手动编译

**编译当前平台版本：**
```bash
go build -o o2m.exe
```

**交叉编译 Linux 版本（在 Windows 上）：**
```bash
set CGO_ENABLED=0
set GOOS=linux
set GOARCH=amd64
go build -ldflags="-s -w" -o o2m-linux
```

**交叉编译 Windows 版本（在 Linux/macOS 上）：**
```bash
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o o2m.exe
```

**使用 UPX 压缩（可选）：**
```bash
upx -9 o2m.exe       # Windows 版本
upx -9 o2m-linux     # Linux 版本
```

### 安装 UPX（可选）

**Windows：**
从 https://upx.github.io/ 下载并添加到 PATH

**Linux：**
```bash
sudo apt-get install upx      # Ubuntu/Debian
sudo yum install upx          # CentOS/RHEL
```

**macOS：**
```bash
brew install upx
```

## 使用方法

### 方式一：命令行模式

#### 基本用法

```bash
# 转换并输出到文件
o2m.exe -i oracle.sql -o mysql.sql

# 转换并输出到标准输出（Linux）
./o2m-linux -i oracle.sql

# 显示帮助
o2m.exe -h
```

#### 参数说明

- `-i` : 输入文件路径（必需），包含 Oracle DDL 语句
- `-o` : 输出文件路径（可选），不指定则输出到标准输出
- `-h` : 显示帮助信息

### 方式二：Web 模式

#### 启动 Web 服务器

```bash
# Windows - 使用默认端口 8080
o2m.exe -web

# Windows - 指定端口
o2m.exe -web -port 9000

# Linux - 使用默认端口
./o2m-linux -web

# Linux - 后台运行
nohup ./o2m-linux -web > o2m.log 2>&1 &
```

#### 访问 Web 界面

启动后在浏览器中打开：
```
http://localhost:8080
```

#### Web 功能特性

1. **文本转换**
   - 直接在文本框中输入 Oracle DDL
   - 点击"开始转换"按钮
   - 查看转换结果
   - 支持复制或下载结果

2. **文件上传**
   - 点击上传区域选择文件
   - 或直接拖拽 .sql 文件到上传区域
   - 自动转换并下载 MySQL DDL 文件

#### 参数说明

- `-web` : 启动 Web 服务器模式
- `-port` : Web 服务器端口（默认 8080）

### Linux 服务器部署

#### 使用 systemd 服务

创建服务文件 `/etc/systemd/system/o2m.service`：

```ini
[Unit]
Description=Oracle to MySQL DDL Converter
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/o2m
ExecStart=/opt/o2m/o2m-linux -web -port 8080
Restart=always
RestartSec=10

[Install]
WantedBy=multi-user.target
```

启动服务：
```bash
sudo systemctl daemon-reload
sudo systemctl enable o2m
sudo systemctl start o2m
sudo systemctl status o2m
```

#### 使用 Nginx 反向代理

```nginx
server {
    listen 80;
    server_name your-domain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}
```

## 使用示例

### 输入示例（Oracle DDL）

```sql
CREATE TABLE employees (
  emp_id NUMBER(10) PRIMARY KEY,
  emp_name VARCHAR2(100) NOT NULL,
  hire_date DATE DEFAULT SYSDATE,
  salary NUMBER(10,2),
  dept_id NUMBER(10),
  CONSTRAINT fk_dept FOREIGN KEY (dept_id) REFERENCES departments(dept_id)
);

COMMENT ON TABLE employees IS '员工表';
COMMENT ON COLUMN employees.emp_name IS '员工姓名';

CREATE INDEX idx_emp_name ON employees(emp_name);
```

### 输出示例（MySQL DDL）

```sql
CREATE TABLE employees (
  emp_id INT NOT NULL,
  emp_name VARCHAR(100) NOT NULL COMMENT '员工姓名',
  hire_date DATETIME DEFAULT CURRENT_TIMESTAMP,
  salary DECIMAL(10,2),
  dept_id INT,
  CONSTRAINT employees_PRIMARY PRIMARY KEY (emp_id),
  CONSTRAINT fk_dept FOREIGN KEY (dept_id) REFERENCES departments (dept_id)
) COMMENT='员工表' ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

CREATE INDEX idx_emp_name ON employees (emp_name);
```

## 支持的 Oracle DDL 元素

### 表定义
- ✅ CREATE TABLE
- ✅ 列定义（名称、类型、长度、精度）
- ✅ NOT NULL 约束
- ✅ DEFAULT 默认值

### 约束
- ✅ PRIMARY KEY（主键）
- ✅ FOREIGN KEY（外键）
- ✅ UNIQUE（唯一键）
- ✅ CHECK（检查约束，MySQL 8.0+）

### 索引
- ✅ CREATE INDEX（普通索引）
- ✅ CREATE UNIQUE INDEX（唯一索引）

### 注释
- ✅ COMMENT ON TABLE（表注释）
- ✅ COMMENT ON COLUMN（列注释）

## 注意事项

1. **CHECK 约束**：MySQL 8.0+ 才支持 CHECK 约束，如使用旧版本需手动移除
2. **默认字符集**：转换后的表使用 `utf8mb4` 字符集和 `utf8mb4_unicode_ci` 排序规则
3. **存储引擎**：默认使用 InnoDB 存储引擎
4. **NUMBER 类型**：根据精度自动选择合适的 MySQL 类型（TINYINT、INT、BIGINT、DECIMAL）
5. **DATE 类型**：Oracle DATE 包含时间部分，转换为 MySQL DATETIME

## 项目结构

```
o2m/
├── main.go           # 主程序入口和命令行参数处理
├── web.go            # Web 服务器实现
├── converter.go      # 核心转换逻辑
├── parser.go         # Oracle DDL 解析器
├── types.go          # 数据结构定义
├── go.mod            # Go 模块文件
├── build.bat         # Windows 编译脚本
├── build.sh          # Linux/macOS 编译脚本
├── static/
│   └── index.html    # Web 界面
├── README.md         # 使用说明
├── test_oracle.sql   # 测试输入文件
└── test_mysql.sql    # 测试输出文件
```

## 技术栈

- **后端**: Go 1.21+
- **Web 框架**: Go 标准库 `net/http`
- **前端**: HTML5 + CSS3 + JavaScript（原生，无框架依赖）
- **静态资源**: 使用 `embed` 包嵌入到二进制文件
- **跨平台**: 支持 Windows、Linux、macOS

## 性能优化

- 使用 `-ldflags="-s -w"` 编译参数减小二进制文件大小
- 使用 UPX 压缩可进一步减小 60-70% 文件大小
- 静态资源嵌入，无需额外文件
- 零外部依赖，单一可执行文件

## 常见问题

### 1. 如何在 Linux 服务器上运行？

```bash
# 上传文件到服务器
scp o2m-linux user@server:/opt/o2m/

# SSH 登录服务器
ssh user@server

# 设置执行权限
chmod +x /opt/o2m/o2m-linux

# 启动 Web 服务器
cd /opt/o2m
./o2m-linux -web -port 8080
```

### 2. 如何处理大文件？

Web 模式默认限制上传文件大小为 10MB。如需处理更大的文件，建议使用命令行模式：

```bash
o2m.exe -i large_oracle.sql -o large_mysql.sql
```

### 3. 转换结果需要手动调整吗？

大多数情况下转换结果可以直接使用。但个别情况可能需要微调：
- Oracle 特有的函数需要手动替换
- 某些复杂的 CHECK 约束可能需要调整
- 分区表和物化视图不支持自动转换

### 4. 支持哪些 Oracle 版本？

工具主要针对 Oracle 10g、11g、12c、19c 的标准 DDL 语法。

### 5. 支持哪些 MySQL 版本？

生成的 DDL 兼容 MySQL 5.7、8.0+。CHECK 约束需要 MySQL 8.0+。

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
