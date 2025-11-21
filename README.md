# Git Watcher

一个用Go编写的Git仓库统计工具，可以递归扫描目录中的所有Git仓库，并提供丰富的统计信息。

## 功能特性

- 🔍 递归扫描指定目录下的所有Git仓库
- 📊 统计提交者信息
- 📅 分析提交时间模式
- 🌙 识别深夜提交（23:00-06:00）
- ⏰ 按小时统计提交活跃度
- 🎯 模块化设计，易于扩展新的统计类型
- 📱 支持JSON和文本格式输出

## 安装

```bash
go mod tidy
go build -o git-watcher
mv git-watcher /usr/local/bin/
```

## 使用方法

### 基本用法

```bash
# 扫描当前目录
git-watcher

# 扫描指定目录
git-watcher -p /path/to/directory

# 输出文本格式
git-watcher -p /path/to/directory -o text
```

### 命令行参数

- `-p, --path`: 要扫描的目录路径（默认为当前目录）
- `-o, --output`: 输出格式，支持 `json` 或 `text`（默认为json）

## 输出示例

### JSON格式输出
```json
{
  "/path/to/repo": {
    "total_commits": 150,
    "statistics": {
      "commit_count_by_author": {
        "张三": 80,
        "李四": 45,
        "王五": 25
      },
      "latest_commit": {
        "Author": "张三",
        "Email": "zhangsan@example.com",
        "Date": "2024-01-15T14:30:00Z",
        "Message": "修复bug",
        "Hash": "abc123..."
      },
      "late_night_commits": {
        "total": 12,
        "authors": {
          "张三": 8,
          "李四": 4
        }
      },
      "commit_activity_by_hour": {
        "9": 25,
        "10": 30,
        "14": 20,
        "23": 5
      }
    }
  }
}
```

### 文本格式输出
```
=== 仓库: /path/to/repo ===
总提交数: 150
最新提交: abc123 by 张三 at 2024-01-15 14:30:00

提交者统计:
  蒋波: 80次
  龙飞: 45次
  王: 25次

深夜提交 (23:00-06:00): 12次
深夜提交者:
  张三: 8次
  李四: 4次
```

## 扩展统计类型

工具采用模块化设计，可以轻松添加新的统计类型。只需实现 `Statistics` 接口：

```go
type Statistics interface {
    Calculate(commits []analyzer.CommitInfo) interface{}
    Name() string
}
```

然后将新的统计器添加到 `StatsCalculator` 中即可。

## 开发计划

- [x] 添加更多统计维度（工作日/周末提交、代码行数统计等）
- [ ] 支持导出为CSV格式
- [ ] 添加Web界面展示统计结果
- [ ] 支持GitHub API集成
- [ ] 添加配置文件支持

## 许可证

MIT License