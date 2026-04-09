# 基础骨架与共享契约说明

本文档记录 `ProjCompiler` 第一批可编译骨架代码的范围与约束，方便后续并行 worker 接入扫描模块、ADK 模块和 TUI 模块。

## 本次骨架只做了什么

- 建立了 Go module 和最小可运行入口。
- 引入了 `google.golang.org/adk` 与 `github.com/charmbracelet/bubbletea` 依赖。
- 定义了共享数据结构：
  - `RepoFacts`
  - `ProjectSpec`
  - `PromptBundle`
  - `Observed / Inferred / Missing` 对应的信号结构
- 定义了最小服务接口：
  - `Scanner`
  - `SpecBuilder`
  - `PromptCompiler`
  - `Exporter`
  - `AgentAssembler`
- 建立了 `.env` 配置加载能力，当前至少支持：
  - `baseurl`
  - `apikey`
  - `model`
  - `output language`
- 建立了应用装配层和一个最小可运行的 Bubble Tea 程序骨架。

## 本次骨架刻意没做什么

- 没有实现项目扫描逻辑。
- 没有实现真实的 ADK agent 调用。
- 没有实现真实的 TUI 流程页。
- 没有实现 `prompt.txt` 的真实导出流程。
- 没有实现任何向量库、图数据库、tree-sitter 或知识图谱能力。

## 与其他模块的契约假设

### 扫描模块

- 扫描模块后续只需要实现 `Scanner` 接口即可接入。
- `RepoFacts` 保持“轻扫描结果”定位，不应演变成复杂图结构。
- 入口点、依赖提示、关键片段等可以继续补字段，但不应破坏现有 `Observed / Inferred / Missing` 信号语义。

### 规格与提示词模块

- `SpecBuilder` 负责把 `RepoFacts` 压缩成最小 `ProjectSpec`。
- `PromptCompiler` 负责把 `ProjectSpec` 编译成英文纯提示词。
- `PromptBundle` 对外主产物仍然只有 `PromptText`，其余字段作为内部辅助信息存在。

### ADK 装配模块

- `AgentAssembler` 是 ADK 接入点保留位。
- 当前假设后续实现会使用 ADK 的固定顺序工作流，而不是复杂的自主调度系统。
- 模型配置由 `internal/config` 提供，后续多提供商适配层可在此基础上实现。

### TUI 模块

- 当前 Bubble Tea 只承担“程序可启动”的最低骨架职责。
- 真正的页面状态机可以后续替换 `rootModel`，不需要改动共享契约。

## 设计来源

- 项目扫描层保留了 DeepWiki-open 的“先结构、后内容”和 GitNexus 的“分阶段抽取 + 置信度”思想，但首版只做轻量本地扫描。
- `ProjectSpec -> PromptBundle` 的收敛方式借鉴了 `claudecode` 的 `/init` 思路：只保留模型最容易做错的事实，不把整仓库重新讲一遍。
- 当前骨架层只把这些思想反映为类型与接口，不提前实现复杂功能。

## 下一步推荐

- 扫描 worker 实现 `Scanner`
- ADK worker 实现 `SpecBuilder`、`PromptCompiler`、`AgentAssembler`
- TUI worker 替换当前占位 `rootModel`
- 导出 worker 实现 `Exporter`，首版只写 `prompt.txt`
