# 最小可行性应用架构图

本文档描述 `ProjCompiler` 当前 MVP 的目标架构，并标注本轮代码已经落地到哪些包、哪些地方仍处于待验证状态。

## 架构说明

- 表现层负责接收项目路径、展示阶段状态、预览提示词并导出结果。
- 应用层负责装配扫描、规格构建、提示词编译与导出流程。
- 核心服务层分成扫描、规格、prompt 三块，尽量保持边界清晰。
- 智能体层使用 `google.golang.org/adk`，当前采用固定顺序工作流思路。
- 基础设施层负责 `.env` 配置读取、文件系统访问、HTTP 模型请求和导出文件。

## 首版输出约束

- 首版最终输出物只是一段英文纯提示词文本。
- 首版不生成任何特定工具的命令包装。
- 首版不绑定 `claude`、`codex` 或其他命令行工具格式。
- 首版不导出 `spec.json`，只导出提示词文本。

## 首版模型配置约束

- 模型配置通过 `.env` 文件提供。
- 当前实现读取 `baseurl`、`apikey`、`model` 这三类核心配置。
- 接入方式是 OpenAI-compatible HTTP API。
- 首版按“单次选择一个 provider”工作，不做多模型自动切换。

## 最小流程图

```mermaid
flowchart TD
    开始["开始"] --> 输入路径["用户输入本地项目路径"]
    输入路径 --> 路径校验{"路径是否有效"}
    路径校验 -- 否 --> 提示错误["提示路径错误并返回输入"]
    提示错误 --> 输入路径
    路径校验 -- 是 --> 扫描项目["ScanOrchestrator 扫描项目目录"]
    扫描项目 --> 提取事实["生成 RepoFacts"]
    提取事实 --> 构建规格["SpecBuilder 生成 ProjectSpec"]
    构建规格 --> 校验规格{"是否存在阻塞问题"}
    校验规格 -- 是 --> 展示问题["停止生成最终 prompt，展示 blocking questions"]
    展示问题 --> 结束失败["结束或等待补充"]
    校验规格 -- 否 --> 生成提示词["PromptCompiler 生成 PromptBundle"]
    生成提示词 --> 预览结果["Bubble Tea TUI 预览规格摘要和提示词"]
    预览结果 --> 导出文件["FileExporter 导出 prompt.txt"]
    导出文件 --> 结束成功["结束"]
```

## 最小数据流图

```mermaid
flowchart LR
    用户["用户"] -->|输入项目路径| TUI["Bubble Tea TUI"]
    项目目录["本地项目目录"] -->|文件与源码| 扫描器["internal/scan"]
    TUI -->|发起扫描| 扫描器
    扫描器 -->|输出| RepoFacts["RepoFacts"]
    RepoFacts -->|输入| 规格构建["internal/adkflow.SpecBuilder"]
    规格构建 -->|输出| ProjectSpec["ProjectSpec"]
    ProjectSpec -->|输入| Prompt编译["internal/prompt.Compiler"]
    Prompt编译 -->|输出| PromptBundle["PromptBundle"]
    ProjectSpec -->|摘要展示| TUI
    PromptBundle -->|预览与导出| TUI
    PromptBundle -->|写入| 导出器["internal/app.FileExporter"]
```

## 架构图

```mermaid
flowchart LR
    subgraph 表现层
        TUI["internal/tui<br/>Bubble Tea 状态机"]
    end

    subgraph 应用层
        APP["internal/app<br/>服务装配与适配"]
        CMD["cmd/projcompiler<br/>入口"]
    end

    subgraph 核心服务层
        SCAN["internal/scan<br/>5 个只读工具 + 1 个协调器"]
        SPEC["internal/spec<br/>共享数据结构与信号"]
        PROMPT["internal/prompt<br/>过滤与渲染"]
    end

    subgraph 智能体层
        ADKFLOW["internal/adkflow<br/>OpenAI-compatible client + ADK adapter + SpecBuilder + Assembler"]
    end

    subgraph 基础设施层
        CONFIG["internal/config<br/>.env 读取"]
        EXPORT["internal/app.FileExporter"]
        FS["文件系统"]
        HTTP["模型 HTTP API"]
    end

    CMD --> APP
    APP --> TUI

    TUI --> APP
    APP --> SCAN
    APP --> ADKFLOW
    APP --> PROMPT
    APP --> EXPORT

    SCAN --> SPEC
    ADKFLOW --> SPEC
    PROMPT --> SPEC

    SCAN --> FS
    ADKFLOW --> HTTP
    APP --> CONFIG
```

## 当前实现映射

已经落地到代码的模块：

- `internal/scan`
  - 已实现 `RepoTreeTool`
  - 已实现 `ManifestProbeTool`
  - 已实现 `DocsProbeTool`
  - 已实现 `EntrypointProbeTool`
  - 已实现 `SnippetSelectTool`
  - 已实现 `ScanOrchestrator`
- `internal/spec`
  - 已实现 `RepoFacts`
  - 已实现 `ProjectSpec`
  - 已实现 `PromptBundle`
  - 已实现 `FactSignal`
- `internal/adkflow`
  - 已实现 OpenAI-compatible client
  - 已实现 ADK `model.LLM` 适配层
  - 已实现 `SpecBuilder`
  - 已实现 `Assembler`
- `internal/prompt`
  - 已实现 prompt-worthiness filter
  - 已实现 prompt 渲染
- `internal/tui`
  - 已实现 `PathInput -> Scanning -> SpecSummary -> PromptPreview -> Done`
- `internal/app` 与 `cmd/projcompiler`
  - 已开始 wiring
  - 当前 `main.go` 已注入 `SpecBuilder`、`PromptCompiler`、`AgentAssembler`
  - `Scanner` 和 `Exporter` 通过默认装配接入

## 当前仍未完成或待验证

- ADK 流程虽然已落代码，但还没有在当前仓库里完成一次正式的端到端验证
- `go test ./...` 当前仍未通过
  - 当前可见原因是 `go.sum` 缺少部分传递依赖校验项
  - 当前环境的默认 Go cache 路径也有权限限制
- 当前尚未补充稳定的集成测试或端到端测试
