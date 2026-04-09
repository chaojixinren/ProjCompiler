# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在此仓库中工作时提供指导。

## 构建和测试命令

```bash
# 构建 CLI
go build -o projcompiler ./cmd/projcompiler

# 运行所有包的测试
go test ./...

# 运行特定包的测试
go test ./internal/scan/...
go test ./internal/prompt/...
go test ./internal/adkflow/...

# 运行单个测试文件
go test -run TestOrchestrator ./internal/scan/orchestrator_test.go

# 运行测试并显示详细输出
go test -v ./...

# 运行应用（需要配置 .env）
./projcompiler
# 或直接运行：
go run ./cmd/projcompiler
```

## 配置

应用从工作目录的 `.env` 文件读取配置。复制 `.env.example` 到 `.env` 并配置：

- `PROJCOMPILER_BASEURL` - 模型端点（程序会自动追加 `/chat/completions`)
- `PROJCOMPILER_APIKEY` - 模型提供商的 Bearer 令牌
- `PROJCOMPILER_MODEL` - 模型名称
- `PROJCOMPILER_OUTPUT_LANGUAGE` - 输出语言（默认：english）
- `PROJCOMPILER_PROMPT_FILE` - 输出文件名（默认：prompt.txt）

环境变量优先级高于文件值。旧版键名（`BASEURL`、`APIKEY`、`MODEL`) 也受支持。

## 架构概览

ProjCompiler 是一个 Go 工具，用于扫描仓库、合成紧凑的项目规格，并编译实现提示词。它遵循四阶段流水线：

**Scan → SpecBuilder → PromptCompiler → Exporter**

流水线通过以下共享数据结构流转（定义在 `internal/spec/`):

1. `RepoFacts` - 从仓库提取的原始事实（文件树、文档、构建信息、入口点、代码片段）
2. `ProjectSpec` - 从事实推导的紧凑规格（目标、技术画像、模块、约束）
3. `PromptBundle` - 最终编译的提示词文本，包含章节和备注

每个事实携带一个 `FactSignal`，包含置信度级别（`Observed`、`Inferred` 或 `Missing`) 和证据引用。

## 模块组织

```
cmd/projcompiler/     # 入口点，组装服务并启动 TUI
internal/
  app/                # 应用装配、服务契约、导出器
  config/             # 从 .env 加载配置
  scan/               # 仓库扫描（编排器 + 探测工具）
  spec/               # 共享数据结构（事实、信号、项目规格、提示词包）
  prompt/             # 提示词编译和过滤
  adkflow/            # LLM 驱动的规格合成和 ADK agent 装配
  tui/                # Bubble Tea 终端 UI（状态、模型、视图）
```

## 关键接口 (internal/app/contracts.go)

- `Scanner` - 从项目路径提取 `RepoFacts`
- `SpecBuilder` - 将 `RepoFacts` 转换为 `ProjectSpec`
- `PromptCompiler` - 将 `ProjectSpec` 转换为 `PromptBundle`
- `Exporter` - 将 `PromptBundle` 写入文件
- `AgentAssembler` - 构建 ADK 顺序工作流 agent

## 扫描工具 (internal/scan/)

`ScanOrchestrator` 并行/顺序运行五个探测工具：

1. `RepoTreeTool` - 遍历文件树，识别候选文档/清单/源文件
2. `ManifestProbeTool` - 解析构建清单（go.mod、package.json、Cargo.toml)
3. `DocsProbeTool` - 从 README 和文档中提取摘要
4. `EntrypointProbeTool` - 识别可能的入口点（main.go、main 函数）
5. `SnippetSelectTool` - 选择入口点附近的代表性代码片段

扫描请求支持限制（最大深度、最大文件数、最大片段数）和通过 `.gitignore` 风格匹配的忽略模式。

## 设计原则

- **基于信号的置信度**：每个事实都有一个 `FactSignal`（Observed=1.0、Inferred=<1.0、Missing=0)，附带证据引用。这让下游阶段能够判断事实的可靠性。
- **/init 剪枝规则**：提示词编译器遵循 ClaudeCode `/init` 理念：只包含模型在没有明确指导时可能出错的内容。舍弃通用建议和文件清单。
- **优雅降级**：当 LLM 不可用时，`SpecBuilder` 会回退到确定性合成。测试应能在没有模型配置的情况下运行。
- **未实现的服务**：缺失的服务会被替换为返回 `ErrServiceUnavailable` 的 stub 实现，让 TUI 能够显示状态。