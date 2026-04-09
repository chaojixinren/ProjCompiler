# MVP 实现说明

本文档记录 `ProjCompiler` 当前 MVP 的真实实现情况。内容基于当前仓库代码目录和本轮子 agent 的实现结果整理，不把尚未验证通过的部分写成已完成。

## 本轮实现目标

本轮的重点不是继续扩设计，而是先把 MVP 的最短链路落成代码骨架：

- 本地项目扫描
- 规格结构定义
- 规格构建
- Prompt 编译
- TUI 状态推进
- 应用装配与导出

## 当前已实现模块

## 一、扫描模块

位置：

- `internal/scan`

当前已经实现：

- `RepoTreeTool`
- `ManifestProbeTool`
- `DocsProbeTool`
- `EntrypointProbeTool`
- `SnippetSelectTool`
- `ScanOrchestrator`

当前实现能力：

- 从本地目录读取文件树和候选文件
- 支持 `.projcompilerignore`、`.gitignore` 和内建忽略项
- 支持预算控制，包括深度、文件数、片段数、文件大小
- 提取文档事实、构建信号、入口候选和关键片段
- 汇总为 `RepoFacts`

当前限制：

- ignore 规则是轻量实现
- 入口推断是启发式
- 还没有回归测试集

## 二、规格结构模块

位置：

- `internal/spec`

当前已经实现：

- `RepoFacts`
- `ProjectSpec`
- `PromptBundle`
- `FactSignal`
- `EvidenceRef`

当前实现能力：

- 在模块间共享统一的数据结构
- 用 `observed / inferred / missing` 表达事实来源和置信度

当前限制：

- 还没有更严格的 schema 校验或兼容性测试

## 三、ADK 与模型接入模块

位置：

- `internal/adkflow`

当前已经实现：

- OpenAI-compatible completion client
- ADK `model.LLM` 适配层
- `SpecBuilder`
- `Assembler`

当前实现能力：

- 从 `.env` 读取 `baseurl`、`apikey`、`model`
- 把 OpenAI-compatible 接口包装成 ADK 可用的 `model.LLM`
- 先走 deterministic fallback，再尝试用模型 refine `ProjectSpec`
- 预留固定顺序工作流装配能力

当前限制：

- 还没有在当前仓库里完成正式的 ADK 端到端验证
- 还缺更完整的错误恢复与运行指标

## 四、Prompt 编译模块

位置：

- `internal/prompt`

当前已经实现：

- `Compiler`
- `filterPromptSections`
- `renderPrompt`
- 最小测试文件

当前实现能力：

- 只输出英文纯提示词
- 先过滤 blocking questions
- 按 Goal、Constraints、Implementation Shape、Acceptance 四段生成 prompt
- 应用 `/init` 风格的删减规则

当前限制：

- 测试覆盖仍然有限
- 仍需用真实项目样本继续调整 prompt 质量

## 五、TUI 模块

位置：

- `internal/tui`

当前已经实现：

- `types.go`
- `messages.go`
- `commands.go`
- `model.go`
- `view.go`

当前实现能力：

- Bubble Tea MVU 状态机
- 支持路径输入
- 支持扫描、规格构建、prompt 编译、导出四段异步命令
- 支持规格摘要与 prompt 预览

当前状态机：

- `PathInput`
- `Scanning`
- `SpecSummary`
- `PromptPreview`
- `Done`

当前限制：

- 还没有更细的交互控件
- 还没有完整的 TUI 集成测试

## 六、应用装配层

位置：

- `internal/app`
- `cmd/projcompiler/main.go`

当前已经实现：

- `app.New`
- `newProgram`
- TUI service adapter
- `FileExporter`
- `main.go` 配置读取和应用启动

当前 wiring 状态：

- `SpecBuilder` 已接 `internal/adkflow`
- `PromptCompiler` 已接 `internal/prompt`
- `AgentAssembler` 已接 `internal/adkflow`
- `Scanner` 通过默认装配接入 `internal/scan`
- `Exporter` 通过默认装配接入 `FileExporter`

说明：

- 这意味着 app 装配层已经开始工作，但不代表当前端到端运行已正式验证通过

## 当前验证情况

我在当前仓库里额外检查到的状态如下：

- `go test ./...` 当前没有通过
- 当前可见阻塞主要有两类：
  - `go.sum` 缺少 ADK 与 Bubble Tea 的部分传递依赖校验项
  - 当前环境默认 Go cache 路径存在权限限制

因此，下面这些能力只能写成“已落代码，待完整验证”：

- `cmd/projcompiler` 启动链路
- `internal/adkflow` 的真实模型调用链路
- `internal/tui` 的完整交互链路

## 当前已确认的借鉴来源

- `DeepWiki-open`
  - 借仓库结构优先和克制扫描
- `GitNexus`
  - 借分阶段流水线和轻量推断
- `ClaudeCode /init` 与工具编排
  - 借 prompt 删减策略和工具边界

本轮未确认的来源：

- `Halley`
  - 当前没有在父目录 `ClaudeCode` 中检索到明确实现路径
  - 暂时只能标记为待补充参考

## 当前仍待补充的工作

- 补齐依赖校验并重新做完整编译验证
- 建立至少一个端到端样本测试
- 增加扫描模块和 prompt 模块的回归测试
- 确认 `Halley` 是否存在对应实现路径，并补充参考对照

## 当前结论

`ProjCompiler` 现在已经从纯设计阶段进入“有实际模块代码、但还没有完成完整验证”的阶段。

更准确地说：

- 主要模块都已经落代码
- 模块边界已经比较清晰
- 应用装配已开始
- 但端到端可运行性和依赖完整性仍待收尾
