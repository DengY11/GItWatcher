# Git Watcher

简体中文 | English

---

## 简介（中文）
Git Watcher 是一个使用 Go 编写的 Git 仓库统计工具，支持递归扫描目录下的所有 Git 仓库，并提供作者统计、时间分布、深夜与周末提交、每小时活跃度等信息。除了命令行输出外，还提供终端 UI（TUI）以更直观地查看与操作。

### 功能
- 递归扫描目录中的 Git 仓库
- 作者提交统计、最新提交信息
- 深夜提交（23:00-06:00）、周末提交统计
- 每小时活跃度柱形图（TUI）
- 支持输出 `json` 与 `text`
- 终端 UI（TUI）操作与导出统计为 JSON 文件

### 安装
```bash
go mod tidy
go build -o git-watcher
```

### 使用
- 命令行（CLI）：
```bash
# 扫描当前目录（默认输出 json）
git-watcher -p .

# 文本输出
git-watcher -p /path/to/directory -o text
```

- 终端 UI（TUI）：
```bash
git-watcher tui -p /path/to/directory
```
键位：
- r：在左侧仓库列表与右侧内容之间切换焦点
- R：刷新并显示分析进度
- j/k：在左侧移动仓库选择；在右侧滚动内容
- Up/Down：滚动右侧当前视图
- 1/2/3/4：切换 Overview/Commits/Authors/Timeline
- e：导出统计为 JSON（可输入保存路径，回车确认）
- a：启用/关闭自动刷新（30 秒）
- q：退出

### 输出示例
文本输出（英文）：
```
=== Repository: /path/to/repo ===
Total commits: 150
Latest commit: abc123 by Alice at 2024-01-15 14:30:00

Author statistics:
  Alice: 80
  Bob: 45
  Carol: 25

Late-night commits (23:00-06:00): 12
Late-night authors:
  Alice: 8
  Bob: 4
```

### 扩展统计类型
实现 `Statistics` 接口并注册到 `StatsCalculator`：
```go
type Statistics interface {
    Calculate(commits []analyzer.CommitInfo) interface{}
    Name() string
}
```

### 计划
- [x] 更多统计维度（周末、代码行数等）
- [x] 终端 UI（TUI）
- [ ] 导出为 CSV
- [ ] GitHub API 集成
- [ ] 配置文件支持

### 许可证
MIT License

---

## Overview (English)
Git Watcher is a Git repository statistics tool written in Go. It recursively scans directories for repositories and reports author stats, latest commit, late-night and weekend commits, and hourly activity. It offers both CLI output and a terminal UI (TUI) for interactive exploration.

### Features
- Recursive repository discovery
- Author commit statistics and latest commit
- Late-night (23:00–06:00) and weekend commit statistics
- Hourly activity bar chart in TUI
- `json` and `text` outputs
- TUI operations with JSON export

### Install
```bash
go mod tidy
go build -o git-watcher
```

### Usage
- CLI:
```bash
# Scan current directory (json by default)
git-watcher -p .

# Text output
git-watcher -p /path/to/directory -o text
```

- TUI:
```bash
git-watcher tui -p /path/to/directory
```
Key bindings:
- r: toggle focus between repos and content
- R: refresh with progress
- j/k: move repo selection (left) or scroll content (right)
- Up/Down: scroll content view
- 1/2/3/4: Overview/Commits/Authors/Timeline
- e: export statistics as JSON (enter a save path, press Enter)
- a: auto refresh (30s)
- q: quit

### Output Example
Text output:
```
=== Repository: /path/to/repo ===
Total commits: 150
Latest commit: abc123 by Alice at 2024-01-15 14:30:00

Author statistics:
  Alice: 80
  Bob: 45
  Carol: 25

Late-night commits (23:00-06:00): 12
Late-night authors:
  Alice: 8
  Bob: 4
```

### Extend Statistics
Implement the `Statistics` interface and register to `StatsCalculator`:
```go
type Statistics interface {
    Calculate(commits []analyzer.CommitInfo) interface{}
    Name() string
}
```

### Roadmap
- [x] More dimensions (weekend, lines changed)
- [x] Terminal UI (TUI)
- [ ] CSV export
- [ ] GitHub API integration
- [ ] Config file support

### License
MIT License
