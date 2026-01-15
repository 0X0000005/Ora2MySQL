---
description: 代码修改完整流程（分析-修改-测试-编译-提交）
---

# 代码修改完整工作流程

这是一个标准的代码修改流程，适用于需要修改核心代码、更新版本、编译和提交的场景。

## 流程概览

```
分析 → 规划 → 修改 → 测试 → 编译 → 提交
```

## 详细步骤

### 1. 代码分析阶段

**目标**：理解现有代码结构，定位需要修改的文件和函数

**操作**：
- 使用 `view_file_outline` 查看文件大纲
- 使用 `grep_search` 搜索相关关键字
- 使用 `view_file` 查看具体代码实现
- 查找版本号配置位置（通常在 `internal/server/web.go` 或 `cmd/*/main.go`）

**关键文件**：
- `internal/converter/converter.go` - 核心转换逻辑
- `internal/converter/parser.go` - SQL 解析逻辑
- `internal/server/web.go` - Web 服务和版本号
- `tests/*.sql` - 测试用例

### 2. 创建实施计划

**操作**：
- 创建 `task.md` - 任务清单
- 创建 `implementation_plan.md` - 实施计划
- 使用 `notify_user` 请求用户审核计划

**任务清单示例**：
```markdown
### 代码修改
- [ ] 定位需要修改的函数
- [ ] 实现核心逻辑
- [ ] 处理边界情况

### 版本管理
- [ ] 更新版本号

### 编译和验证
- [ ] 运行单元测试
- [ ] 编译二进制文件
- [ ] 手动测试

### Git 提交
- [ ] 提交代码更改
```

### 3. 代码修改阶段

**操作**：
- 使用 `multi_replace_file_content` 进行多处修改（同一文件多个位置）
- 使用 `replace_file_content` 进行单处修改
- 修改核心逻辑文件（如 `converter.go`）
- 更新版本号（如 `web.go` 中的 `Version` 常量）

**注意事项**：
- 保持代码格式一致
- 添加必要的注释
- 考虑向后兼容性

### 4. 测试验证阶段

**操作**：

#### 4.1 运行单元测试
```bash
go test ./internal/converter/... -v
```

#### 4.2 创建测试用例
在 `tests/` 目录创建测试 SQL 文件验证新功能

#### 4.3 编译项目
```bash
go build -ldflags="-s -w" -o o2m.exe ./cmd/o2m
```

#### 4.4 手动测试
```bash
# 测试转换功能
.\o2m.exe -i tests/test_input.sql -o tests/test_output.sql

# 查看输出文件验证结果
```

### 5. 完整编译阶段

**使用编译脚本**：

// turbo
```bash
# Windows 环境
cmd /c build.bat
```

编译脚本会自动：
- 编译 Windows 版本（`o2m.exe`）
- 编译 Linux 版本（`o2m-linux`）
- 使用 UPX 压缩二进制文件（如果可用）

**验证编译产物**：
- 检查文件是否生成
- 运行版本检查：`.\o2m.exe -v`

### 6. Git 提交阶段

**操作**：

#### 6.1 查看修改状态
```bash
git status
git diff
```

#### 6.2 添加文件
```bash
git add -A
```

#### 6.3 提交更改
```bash
git commit -m "feat: 功能描述

- 详细修改点1
- 详细修改点2
- 更新版本号至 vX.X.X"
```

**提交信息规范**：
- 使用规范的 commit message 格式（feat/fix/docs/refactor/test）
- 第一行简短描述（50字符内）
- 空一行后添加详细说明
- 列出主要修改点

#### 6.4 验证提交
```bash
git log -1
```

### 7. 创建完成报告

**操作**：
- 创建 `walkthrough.md` 文档
- 总结所有修改
- 记录测试结果
- 包含示例输入输出
- 标注 git 提交哈希

## 工作流检查清单

在执行完整流程时，确保：

- [ ] 已理解需求并定位相关代码
- [ ] 已创建实施计划并获得用户确认
- [ ] 代码修改完成且添加了必要注释
- [ ] 版本号已更新
- [ ] 单元测试通过
- [ ] 手动测试验证功能正确
- [ ] 使用 build.bat 完整编译
- [ ] git 提交包含清晰的说明
- [ ] 创建了完成报告

## 常见问题

### Q: 测试失败怎么办？
A: 检查错误信息，回退到代码修改阶段进行调整

### Q: 编译失败怎么办？
A: 检查语法错误，确保所有依赖正确，运行 `go mod tidy`

### Q: 如何跳过某些步骤？
A: 根据实际需求调整，但建议至少执行：修改 → 测试 → 提交

### Q: 版本号如何更新？
A: 
- 小改动：补丁版本 +1（v1.2.0 → v1.2.1）
- 新功能：次版本 +1（v1.2.0 → v1.3.0）
- 破坏性变更：主版本 +1（v1.2.0 → v2.0.0）

## 参考文件

- `build.bat` / `build.sh` - 编译脚本
- `tests/` - 测试用例目录
- `internal/converter/` - 核心转换逻辑
- `internal/server/web.go` - 版本号配置

## 下次使用

直接告诉 AI：
> "参考 code-modify-workflow 工作流，我需要修改..."

AI 将遵循这个标准流程执行任务。
