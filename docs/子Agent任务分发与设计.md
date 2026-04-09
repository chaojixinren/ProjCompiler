# 子 Agent 任务分发与设计

本文档记录本轮 `ProjCompiler` 的子 agent 分工、借鉴来源、实际实现结果和当前尚未收尾的部分。

目标是把“谁负责了什么、现在代码已经到哪一步”写清楚，方便后续 worker 继续接力。

## 本轮子 agent 分工

### 子 agent 一：扫描模块

负责范围：

- `internal/scan`
- 只做本地轻扫描
- 不接触 TUI、ADK、prompt 渲染

本轮结果：

- 已实现 `RepoTreeTool`
- 已实现 `ManifestProbeTool`
- 已实现 `DocsProbeTool`
- 已实现 `EntrypointProbeTool`
- 已实现 `SnippetSelectTool`
- 已实现 `ScanOrchestrator`

明确收敛：

- 采用 `5` 个只读工具加 `1` 个协调器
- 输出对齐 `internal/spec` 中的 `RepoFacts`
- 结果显式区分 `observed / inferred / missing`

### 子 agent 二：ADK 编排与模型适配

负责范围：

- `internal/adkflow`
- OpenAI-compatible 模型接入
- ADK agent 组装

本轮结果：

- 已实现 OpenAI-compatible completion client
- 已实现 ADK `model.LLM` 适配层
- 已实现 `SpecBuilder`
- 已实现 `Assembler`

明确收敛：

- 采用固定顺序工作流思路
- 按 `SequentialAgent + 2 个 LlmAgent` 组织
- 通过 `.env` 读取模型配置

### 子 agent 三：Prompt 编译

负责范围：

- `internal/prompt`
- 只做 prompt 过滤和渲染

本轮结果：

- 已实现 `Compiler`
- 已实现 `/init` 风格的删减过滤
- 已实现最终英文 prompt 渲染
- 已补最小测试文件

明确收敛：

- 最终只输出英文纯提示词
- 不生成任何工具命令包装
- 有阻塞问题时停止输出最终 prompt

### 子 agent 四：Bubble Tea TUI

负责范围：

- `internal/tui`
- 只做人机交互与状态推进

本轮结果：

- 已实现 `PathInput -> Scanning -> SpecSummary -> PromptPreview -> Done`
- 已实现 `Cmd / Msg` 异步命令模式
- 已实现文本视图和日志缓冲

明确收敛：

- 不在 UI 层直接实现扫描逻辑
- 不在 UI 层直接实现 LLM 请求

### 子 agent 五：文档与设计收敛

负责范围：

- README
- 架构文档
- 模块任务文档
- 子 agent 分工文档

本轮结果：

- 已把设计阶段文档更新为“设计 + 当前实现状态”
- 已补充实现说明文档

## 借鉴来源与当前吸收情况

### DeepWiki-open

当前已吸收：

- 先看仓库结构和 README
- 先结构后内容
- 控制扫描预算

当前未吸收：

- wiki 结构生成
- 文档生成
- RAG

### GitNexus

当前已吸收：

- 分阶段处理流水线
- 入口与流程的轻量推断
- 给推断附带理由和证据

当前未吸收：

- tree-sitter
- 图数据库
- impact analysis

### ClaudeCode `/init` 与工具编排

当前已吸收：

- `/init` 风格的 prompt 删减规则
- 工具与协调器分离
- 预算和读取边界控制

当前参考文件：

- `/Users/chaoji_xinren/project/ClaudeCode/src/commands/init.ts`
- `/Users/chaoji_xinren/project/ClaudeCode/src/projectOnboardingState.ts`
- `/Users/chaoji_xinren/project/ClaudeCode/src/services/tools/toolOrchestration.ts`
- `/Users/chaoji_xinren/project/ClaudeCode/src/services/tools/toolExecution.ts`
- `/Users/chaoji_xinren/project/ClaudeCode/src/tools/FileReadTool/prompt.ts`
- `/Users/chaoji_xinren/project/ClaudeCode/src/tools/GrepTool/prompt.ts`
- `/Users/chaoji_xinren/project/ClaudeCode/src/tools/ToolSearchTool/prompt.ts`

### Halley

当前状态：

- 本轮在父目录 `ClaudeCode` 中没有检索到 `Halley` 相关代码或命名
- 因此还不能把 `Halley` 列为已吸收参考源
- 如果后续补到明确路径，应再补一轮文档和实现对照

## 当前目录结构

```text
cmd/projcompiler/main.go
internal/adkflow/
internal/app/
internal/config/
internal/prompt/
internal/scan/
internal/spec/
internal/tui/
```

## 当前实现现状

已经落代码的部分：

- 扫描模块已经实现 `5` 个只读工具加 `1` 个协调器
- ADK/Prompt 模块已经实现 OpenAI-compatible client、ADK model adapter、`SpecBuilder`、`PromptCompiler`
- TUI 模块已经实现 Bubble Tea 状态机
- app 装配层已经开始 wiring

当前仍未完成或待验证：

- 端到端运行尚未正式验证通过
- 当前 `go test ./...` 仍失败
- 失败原因当前可见为：
  - `go.sum` 缺少 ADK 与 Bubble Tea 的部分传递依赖校验项
  - 默认 Go cache 路径在当前环境下有权限限制
- `Halley` 参考仍待补充
- 集成测试和样本仓库回归测试仍未建立

## 后续建议

- 先补齐依赖校验，再做一次完整编译验证
- 以真实本地项目样本验证扫描质量和 prompt 质量
- 在 `internal/app` 层补充端到端集成测试
- 如果后续确认 `Halley` 路径，再追加对 spec / prompt 化策略的对照修订
