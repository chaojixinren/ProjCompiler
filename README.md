# ProjCompiler

`ProjCompiler` 是一个本地 Go CLI 应用，目标是把一个现有中小型项目解析成结构化规格，再生成一段英文纯提示词，供通用大模型直接使用。

首版仍然坚持 MVP 边界：只做本地项目分析和提示词生成，不做远程仓库拉取，不做特定工具命令包装，不做向量库、知识图谱和自动修复闭环。

## 当前进展

当前仓库已经不再停留在纯设计阶段，已有真实代码落在以下主链路上：

- `internal/scan` 已实现扫描模块，采用 `5` 个只读工具加 `1` 个协调器。
- `internal/spec` 已定义 `RepoFacts`、`ProjectSpec`、`PromptBundle` 及 `observed / inferred / missing` 信号结构。
- `internal/adkflow` 已实现 OpenAI-compatible client、ADK model adapter、`SpecBuilder`、`Assembler`。
- `internal/prompt` 已实现 prompt 过滤和渲染，遵循 `/init` 风格的删减规则。
- `internal/tui` 已实现 Bubble Tea 状态机。
- `internal/app` 与 `cmd/projcompiler` 已开始 wiring，把扫描、规格构建、提示词编译、导出和 TUI 串起来。

## 项目目标

- 输入一个本地项目目录
- 扫描项目结构、依赖、入口文件、说明文档和关键源码片段
- 抽取统一的项目规格
- 生成一段英文纯文本提示词
- 在本地 TUI 中预览并导出 `prompt.txt`

## 首版范围

首版包含：

- 本地目录输入
- 终端交互界面
- 轻量项目扫描
- 项目规格抽取
- 英文纯文本提示词生成
- 结果预览与导出

首版不包含：

- 远程仓库拉取
- tree-sitter
- 向量检索
- 图数据库
- 自动编译与自动修复闭环
- 多仓库联合分析
- `claude`、`codex` 等命令格式包装

## 技术栈

- 语言：Go
- 智能体框架：`google.golang.org/adk`
- 终端界面：`github.com/charmbracelet/bubbletea`
- 模型接入：OpenAI-compatible HTTP API

## 当前实现结构

```text
ProjCompiler/
├── cmd/
│   └── projcompiler/
│       └── main.go
├── internal/
│   ├── adkflow/
│   │   ├── analyzer.go
│   │   ├── assembler.go
│   │   ├── client.go
│   │   ├── client_test.go
│   │   └── model_adapter.go
│   ├── app/
│   │   ├── app.go
│   │   ├── contracts.go
│   │   ├── exporter.go
│   │   ├── model.go
│   │   ├── README.md
│   │   └── services.go
│   ├── config/
│   │   ├── config.go
│   │   └── env.go
│   ├── prompt/
│   │   ├── compiler.go
│   │   ├── compiler_test.go
│   │   ├── filter.go
│   │   └── render.go
│   ├── scan/
│   │   ├── docs_probe_tool.go
│   │   ├── entrypoint_probe_tool.go
│   │   ├── ignore.go
│   │   ├── manifest_probe_tool.go
│   │   ├── orchestrator.go
│   │   ├── repo_tree_tool.go
│   │   ├── snippet_select_tool.go
│   │   ├── tool.go
│   │   └── types.go
│   ├── spec/
│   │   ├── facts.go
│   │   ├── project.go
│   │   ├── prompt.go
│   │   └── signals.go
│   └── tui/
│       ├── commands.go
│       ├── messages.go
│       ├── model.go
│       ├── types.go
│       └── view.go
├── MVP-架构图.md
├── MVP-模块任务设计.md
├── MVP-实现说明.md
└── 子Agent任务分发与设计.md
```

## 当前主链路

```text
本地目录 -> RepoFacts -> ProjectSpec -> PromptBundle -> prompt.txt
```

对应实现位置：

- 扫描：`internal/scan`
- 规格结构：`internal/spec`
- ADK 规格构建：`internal/adkflow`
- Prompt 编译：`internal/prompt`
- TUI：`internal/tui`
- 装配与导出：`internal/app`、`cmd/projcompiler`

## 借鉴来源

当前实现主要借了三类思路：

- `DeepWiki-open`
  - 借“先看仓库结构和 README，再进入下一层理解”
  - 不借 wiki 生成、RAG、embedding
- `GitNexus`
  - 借“分阶段处理流水线”和“推断结果要带理由/证据”
  - 不借 tree-sitter、图数据库、impact analysis
- `ClaudeCode /init` 与工具编排
  - 借“只保留模型会做错的事实”
  - 借“工具边界清晰、协调器只做编排”

当前仓库内已经明确参考过的本地文件主要包括：

- `/Users/chaoji_xinren/project/ClaudeCode/src/commands/init.ts`
- `/Users/chaoji_xinren/project/ClaudeCode/src/projectOnboardingState.ts`
- `/Users/chaoji_xinren/project/ClaudeCode/src/services/tools/toolOrchestration.ts`
- `/Users/chaoji_xinren/project/ClaudeCode/src/services/tools/toolExecution.ts`
- `/Users/chaoji_xinren/project/ClaudeCode/src/tools/FileReadTool/prompt.ts`
- `/Users/chaoji_xinren/project/ClaudeCode/src/tools/GrepTool/prompt.ts`
- `/Users/chaoji_xinren/project/ClaudeCode/src/tools/ToolSearchTool/prompt.ts`

补充说明：

- 父目录 `ClaudeCode` 中本轮未检索到 `Halley` 相关代码或命名，因此目前文档里仍将其视为待补充参考源。

## 当前状态

已经完成的实现：

- 扫描模块的 `5 + 1` 结构已落代码
- `.env` 配置读取已实现
- OpenAI-compatible completion client 已实现
- ADK model adapter 已实现
- `SpecBuilder` 已实现，包含 deterministic fallback 和 LLM refine 路径
- `PromptCompiler` 已实现，输出英文纯提示词
- Bubble Tea 状态机已实现
- 文件导出器已实现
- `cmd/projcompiler/main.go` 已开始装配应用

当前状态或待验证项：

- `internal/app` 的装配已经开始，但还没有把所有模块的端到端运行结果正式验证通过
- 当前仓库执行 `go test ./...` 仍失败，主要原因是 `go.sum` 中缺少 ADK 与 Bubble Tea 的部分传递依赖校验项，同时默认 Go cache 路径在当前沙箱环境下也有权限问题
- `internal/scan`、`internal/spec` 在当前仓库内可被 `go test` 识别到，但它们目前没有测试文件
- Prompt 生成虽然已有实现，但真实模型调用链路仍需要在补齐依赖校验后再做一次完整验证

## 配置

模型配置通过 `.env` 读取，当前实现支持这些键名：

- `PROJCOMPILER_BASEURL` 或 `BASEURL`
- `PROJCOMPILER_APIKEY` 或 `APIKEY`
- `PROJCOMPILER_MODEL` 或 `MODEL`
- `PROJCOMPILER_OUTPUT_LANGUAGE` 或 `OUTPUT_LANGUAGE`
- `PROJCOMPILER_PROMPT_FILE` 或 `PROMPT_FILE`

默认输出语言当前仍为英文。

## 相关文档

- [MVP-架构图.md](./MVP-%E6%9E%B6%E6%9E%84%E5%9B%BE.md)
- [MVP-模块任务设计.md](./MVP-%E6%A8%A1%E5%9D%97%E4%BB%BB%E5%8A%A1%E8%AE%BE%E8%AE%A1.md)
- [子Agent任务分发与设计.md](./%E5%AD%90Agent%E4%BB%BB%E5%8A%A1%E5%88%86%E5%8F%91%E4%B8%8E%E8%AE%BE%E8%AE%A1.md)
- [MVP-实现说明.md](./MVP-%E5%AE%9E%E7%8E%B0%E8%AF%B4%E6%98%8E.md)
