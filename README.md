# ProjCompiler

ProjCompiler 是一个 Go CLI 应用，用于扫描本地项目仓库、合成项目规格并生成实现提示词。

## 功能特性

- **仓库扫描**：自动识别项目结构、构建清单、文档、入口点和关键代码片段
- **项目规格合成**：从原始事实提取紧凑的项目规格（目标、技术画像、模块、约束）
- **提示词编译**：遵循 `/init` 剪枝规则，只保留模型可能出错的关键信息
- **终端交互界面**：基于 Bubble Tea 的 TUI，支持中英文切换（Ctrl+L）

## 快速开始

### 构建

```bash
go build -o projcompiler ./cmd/projcompiler
```

### 配置

复制 `.env.example` 到 `.env` 并填写配置：

```bash
cp .env.example .env
```

配置项说明：

| 键名 | 说明 | 必需 |
|------|------|------|
| `PROJCOMPILER_BASEURL` | 模型端点（程序自动追加 `/chat/completions`) | 是 |
| `PROJCOMPILER_APIKEY` | API 密钥 | 是 |
| `PROJCOMPILER_MODEL` | 模型名称 | 是 |
| `PROJCOMPILER_OUTPUT_LANGUAGE` | 输出语言（默认：english) | 否 |
| `PROJCOMPILER_PROMPT_FILE` | 输出文件名（默认：prompt.txt) | 否 |

环境变量优先级高于文件值，旧版键名（`BASEURL`、`APIKEY`、`MODEL`) 也受支持。

### 运行

```bash
./projcompiler
# 或直接运行
go run ./cmd/projcompiler
```

启动后输入项目路径，程序依次执行扫描、规格合成、提示词编译，最终导出 `prompt.txt`。

## 架构

ProjCompiler 采用四阶段流水线：

```
Scan → SpecBuilder → PromptCompiler → Exporter
```

数据流转：

| 结构 | 来源 | 说明 |
|------|------|------|
| `RepoFacts` | Scanner | 原始事实（文件树、文档、构建信息、入口点、代码片段）|
| `ProjectSpec` | SpecBuilder | 紧凑规格（目标、技术画像、模块、约束）|
| `PromptBundle` | PromptCompiler | 最终提示词（章节、备注）|

每个事实携带 `FactSignal` 标注置信度：
- `Observed` (1.0) - 直接观测到的事实
- `Inferred` (<1.0) - 推断得出的事实  
- `Missing` (0) - 缺失的信息

## 项目结构

```
cmd/projcompiler/     # 入口点，组装服务并启动 TUI
internal/
  app/                # 应用装配、服务契约、导出器
  config/             # 配置加载（从 .env）
  scan/               # 仓库扫描（编排器 + 5 个探测工具）
  spec/               # 共享数据结构（事实、信号、项目规格、提示词包）
  prompt/             # 提示词编译和过滤
  adkflow/            # LLM 规格合成、ADK agent 装配
  tui/                # Bubble Tea 终端 UI（状态、模型、视图）
```

### 扫描工具

`ScanOrchestrator` 协调五个探测工具：

1. **RepoTreeTool** - 遍历文件树，识别候选文档/清单/源文件
2. **ManifestProbeTool** - 解析构建清单（go.mod、package.json、Cargo.toml)
3. **DocsProbeTool** - 提取 README 和文档摘要
4. **EntrypointProbeTool** - 识别入口点（main.go、main 函数)
5. **SnippetSelectTool** - 选择入口点附近的代表性代码片段

扫描支持限制（最大深度、文件数、片段数）和 `.gitignore` 风格的忽略模式。

## 技术栈

- Go 1.25
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - 终端 UI 框架
- [Google ADK](https://google.golang.org/adk) - Agent 开发框架

## 许可证

MIT
