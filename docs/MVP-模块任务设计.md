# MVP 模块任务设计

本文档用于记录 `ProjCompiler` 当前 MVP 的模块边界，以及这些模块在本轮子 agent 并行实现后的落地情况。

文档目标不是再展开大设计，而是把三件事写清楚：

- 每个模块现在负责什么
- 这轮已经落了哪些代码
- 还有哪些地方仍待补充或待验证

## 设计原则

- 首版只做本地轻扫描，不做远程仓库拉取。
- 首版只输出英文纯提示词，不生成任何特定工具命令。
- 首版不做 tree-sitter、向量库、图数据库、知识图谱。
- 首版不做自动编译、自动修复、自动回灌闭环。
- 所有重要事实都要区分 `observed`、`inferred`、`missing`。
- Prompt 编译只保留“模型不写明就容易做错”的信息。

## 借鉴来源

### GitNexus

借鉴：

- 分阶段处理流水线
- 对入口和流程做轻量推断
- 推断结果保留理由和证据

不借鉴：

- tree-sitter
- 知识图谱
- impact analysis
- hybrid search

### DeepWiki-open

借鉴：

- 先看仓库结构和 README
- 先结构后内容
- 对扫描预算和过滤保持克制

不借鉴：

- wiki 页面生成
- 文档分页生成
- embedding / RAG

### ClaudeCode `/init` 与工具设计

借鉴：

- 先抽候选事实，再删减，再编译 prompt
- 只保留模型会做错的事实
- 工具与协调器职责分离
- 对读取范围和预算保持保守

当前仓库内已明确参考的文件：

- `/Users/chaoji_xinren/project/ClaudeCode/src/commands/init.ts`
- `/Users/chaoji_xinren/project/ClaudeCode/src/projectOnboardingState.ts`
- `/Users/chaoji_xinren/project/ClaudeCode/src/services/tools/toolOrchestration.ts`
- `/Users/chaoji_xinren/project/ClaudeCode/src/services/tools/toolExecution.ts`
- `/Users/chaoji_xinren/project/ClaudeCode/src/tools/FileReadTool/prompt.ts`
- `/Users/chaoji_xinren/project/ClaudeCode/src/tools/GrepTool/prompt.ts`
- `/Users/chaoji_xinren/project/ClaudeCode/src/tools/ToolSearchTool/prompt.ts`

补充说明：

- 本轮没有在父目录 `ClaudeCode` 中检索到 `Halley` 相关代码，因此暂时无法把它列为已经落地吸收的参考源。

## 模块拆分

当前按五块推进：

- 扫描模块：`internal/scan`
- 规格契约模块：`internal/spec`
- ADK 编排模块：`internal/adkflow`
- Prompt 编译模块：`internal/prompt`
- TUI 模块：`internal/tui`

应用装配放在：

- `internal/app`
- `cmd/projcompiler`

## 模块一：扫描模块

位置：

- `internal/scan`

### 当前职责

- 从本地目录读取高信号文件
- 生成 `RepoFacts`
- 在扫描阶段维持确定性，不依赖 LLM

### 当前实现状态

已经实现 `5` 个只读工具加 `1` 个协调器：

- `RepoTreeTool`
- `ManifestProbeTool`
- `DocsProbeTool`
- `EntrypointProbeTool`
- `SnippetSelectTool`
- `ScanOrchestrator`

### 当前编排

1. `RepoTreeTool`
2. `ManifestProbeTool` 与 `DocsProbeTool` 并发
3. `EntrypointProbeTool`
4. `SnippetSelectTool`
5. 汇总为 `RepoFacts`

### 当前实现要点

- 支持 `.projcompilerignore`、`.gitignore` 和内建忽略目录
- 支持最大深度、最大文件数、最大片段数、最大文件大小等预算
- 输出文档事实、构建信号、入口候选、关键片段、约束和信号
- 扫描结果统一挂接 `observed / inferred / missing`

### 当前待补充

- ignore 语义当前是 MVP 轻量版，不是完整 gitignore 兼容
- 入口推断仍是启发式，不是语言级解析
- 还缺针对真实仓库样本的回归测试

## 模块二：规格契约模块

位置：

- `internal/spec`

### 当前职责

- 定义扫描、规格和 prompt 的共享数据结构
- 为模块之间提供统一的信号表达

### 当前实现状态

已落地的数据结构：

- `RepoFacts`
- `ProjectSpec`
- `PromptBundle`
- `FactSignal`
- `EvidenceRef`

### 当前实现要点

- `RepoFacts` 聚焦树摘要、文档、构建、入口、片段、约束、信号
- `ProjectSpec` 聚焦 `goal`、`tech_profile`、`implementation_shape`、`critical_constraints`、`acceptance_checks`、`open_questions`
- `PromptBundle` 聚焦最终文本、分段、阻塞问题和说明

### 当前待补充

- 目前仍以最小共享契约为主，还没有补复杂 schema 校验器
- 还没有系统化的 JSON roundtrip 或兼容性测试

## 模块三：ADK 编排模块

位置：

- `internal/adkflow`

### 当前职责

- 提供模型接入
- 把 `RepoFacts -> ProjectSpec` 的智能体步骤接起来
- 为后续固定流水线保留 ADK agent 组装能力

### 当前实现状态

已经落地：

- OpenAI-compatible completion client
- ADK `model.LLM` 适配层
- `SpecBuilder`
- `Assembler`

### 当前实现要点

- `.env` 读取 `baseurl`、`apikey`、`model`
- `SpecBuilder` 先走 deterministic fallback，再在模型可用时用 LLM refine
- `Assembler` 按 `SequentialAgent + 2 个 LlmAgent` 的方向组装固定顺序工作流

### 当前待补充或待验证

- 真实 ADK 端到端链路尚未在当前仓库内完成正式验证
- 当前 `go test ./...` 会被 ADK 相关依赖校验阻塞
- 还没有稳定的 telemetry、重试策略和 token 统计出口

## 模块四：Prompt 编译模块

位置：

- `internal/prompt`

### 当前职责

- 把 `ProjectSpec` 编译成英文纯提示词
- 执行 `/init` 风格的删减规则

### 当前实现状态

已经落地：

- `Compiler`
- `filterPromptSections`
- `renderPrompt`
- 最小测试文件

### 当前实现要点

- 只输出纯 prompt，不输出命令包装
- 先检查 blocking questions
- 再过滤 Goal、Constraints、Implementation Shape、Acceptance 四类信息
- 去掉 generic advice 和明显可推断内容

### 当前待补充或待验证

- 目前测试覆盖仍然偏少
- 真实 prompt 质量还需要在真实项目样本上继续打磨

## 模块五：TUI 模块

位置：

- `internal/tui`

### 当前职责

- 接收路径输入
- 展示扫描和生成进度
- 展示规格摘要和 prompt 预览
- 触发导出

### 当前实现状态

已经落地：

- `types.go`
- `messages.go`
- `commands.go`
- `model.go`
- `view.go`

### 当前状态机

- `PathInput`
- `Scanning`
- `SpecSummary`
- `PromptPreview`
- `Done`

### 当前实现要点

- 使用 Bubble Tea `Init / Update / View`
- 通过 `Cmd / Msg` 驱动异步扫描、规格构建、prompt 生成和导出
- TUI 不直接承担扫描和 LLM 逻辑

### 当前待补充

- 还没有更细的输入编辑体验
- 还没有集成测试验证完整界面流程

## 应用装配层

位置：

- `internal/app`
- `cmd/projcompiler/main.go`

### 当前实现状态

装配已经开始：

- `main.go` 已读取配置并创建应用
- `SpecBuilder` 已接 `internal/adkflow`
- `PromptCompiler` 已接 `internal/prompt`
- `AgentAssembler` 已接 `internal/adkflow`
- `Scanner` 通过默认 `scan.NewDefaultOrchestrator()` 接入
- `Exporter` 通过默认 `FileExporter` 接入
- `internal/app/model.go` 已把应用服务适配进 TUI

### 当前待补充或待验证

- 端到端运行结果仍待正式验证
- 部分服务当前仍缺完整集成测试

## 当前主要风险

- 依赖校验尚未收齐，当前 `go test ./...` 失败
- 模型链路虽已实现，但还没有在当前仓库里完成正式连通性验证
- 扫描和 prompt 质量还缺真实项目样本回归
