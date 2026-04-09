package tui

// Language identifies the display language for all UI strings.
type Language string

const (
	LangEN Language = "en"
	LangZH Language = "zh"
)

// Strings holds every translatable string rendered by the TUI.
type Strings struct {
	// Header
	Subtitle    string
	StageInput  string
	StageScan   string
	StageSpec   string
	StagePrompt string
	StageDone   string
	StateMeta   string

	// State titles
	TitlePathInput     string
	TitleScanning      string
	TitleSpecSummary   string
	TitlePromptPreview string
	TitleDone          string
	TitleConfigEdit    string
	TitleUnknown       string

	// Path input view
	PathIntro      string
	PathPanelInfo  string
	PathPanelInput string
	PathExample    string
	PathFieldLabel string
	PathFieldHint  string
	PathDefaultExp string

	// Scanning view
	ScanQueued      string
	ScanRunning     string
	ScanPhaseMain   string
	ScanPhaseFinal  string
	ScanPanelStatus string
	ScanPanelNotes  string
	ScanLabelPath   string
	ScanLabelStatus string
	ScanLabelPhase  string
	ScanLabelCancel string
	ScanCancelHint  string
	ScanNote1       string
	ScanNote2       string
	ScanNote3       string

	// Spec summary view
	SpecPanelGoal         string
	SpecPanelTech         string
	SpecPanelModules      string
	SpecPanelFlows        string
	SpecPanelConstraints  string
	SpecPanelChecks       string
	SpecPanelEvidence     string
	SpecPanelQuestions    string
	SpecGoalFallback      string
	SpecNone              string
	SpecUnknown           string
	SpecBlockingPrefix    string
	SpecUnnamedModule     string
	SpecNoSummary         string
	SpecUnnamedFlow       string
	SpecUnnamedConstraint string
	SpecUnnamedCheck      string
	TechPlatform          string
	TechLanguages         string
	TechFrameworks        string
	TechBuildSystem       string
	TechPackageManager    string
	TechIntegrations      string

	// Prompt preview view
	PromptStatusLabel    string
	PromptCompiling      string
	PromptWaiting        string
	PromptPanelCompiler  string
	PromptPanelStatus    string
	PromptNotReady       string
	PromptBlockingQTitle string
	PromptLangLabel      string
	PromptExportLabel    string
	PromptVisibleLabel   string
	PromptScrollLabel    string
	PromptSectionsLabel  string
	PromptPanelPreview   string
	PromptPanelMeta      string

	// Done view
	DonePanelTitle  string
	DoneOutputPath  string
	DonePromptLen   string
	DoneNextStep    string
	DoneNextStepVal string

	// Config edit view
	ConfigPanelTitle    string
	ConfigBaseURLLabel  string
	ConfigAPIKeyLabel   string
	ConfigModelLabel    string
	ConfigPanelHint     string
	ConfigSaved         string
	ConfigSaveFailed    string

	// Key hints
	HintStartScan     string
	HintQuit          string
	HintCancelRun     string
	HintCompilePrompt string
	HintRestart       string
	HintScroll        string
	HintExport        string
	HintBack          string
	HintNewRun        string
	HintToggleLang    string
	HintConfigEdit    string
	HintConfigSwitch  string
	HintConfigSave    string
	HintConfigCancel  string

	// Shared / misc
	LogPanelTitle     string
	UnknownPanelTitle string
	UnknownPanelBody  string
	ErrorPanelTitle   string
	NotSet            string

	// Fact stats labels
	StatRootPath    string
	StatDocuments   string
	StatEntryPoints string
	StatSnippets    string
	StatConstraints string
	StatTruncated   string
	StatLastExport  string
	StatPrompt      string

	// Format strings (contain printf verbs)
	FmtPromptChars  string // e.g. "%d chars"
	FmtPromptLen    string // e.g. "%d characters"
	FmtVisibleLines string // e.g. "%d-%d of %d"
	FmtScrollPos    string // e.g. "%d / %d"

	// User-visible error messages
	ErrPathEmpty       string
	ErrScanFailed      string
	ErrSpecFailed      string
	ErrPromptFmt       string // contains %v
	ErrExportFmt       string // contains %v
	ErrConfigIncomplete string
	ErrConfigSaveFmt   string // contains %v

	// Log messages (shown in Recent Activity panel)
	LogReady           string
	LogScanningFmt     string // contains %s
	LogScanDoneFmt     string // contains %d, %d
	LogBuildingSpec    string
	LogSpecReady       string
	LogCompilingPrompt string
	LogPromptReady     string
	LogExportingFmt    string // contains %s
	LogExportedFmt     string // contains %s
	LogCancelled       string
	LogConfigSaving    string
	LogConfigSaved     string
}

// StateTitle returns the display title for the given state in this language.
func (s Strings) StateTitle(state State) string {
	switch state {
	case StatePathInput:
		return s.TitlePathInput
	case StateScanning:
		return s.TitleScanning
	case StateSpecSummary:
		return s.TitleSpecSummary
	case StatePromptPreview:
		return s.TitlePromptPreview
	case StateDone:
		return s.TitleDone
	case StateConfigEdit:
		return s.TitleConfigEdit
	default:
		return s.TitleUnknown
	}
}

// stringsFor returns the Strings for the given language, defaulting to English.
func stringsFor(lang Language) Strings {
	if lang == LangZH {
		return stringsZH()
	}
	return stringsEN()
}

func stringsEN() Strings {
	return Strings{
		Subtitle:    "Local repository analysis -> project specification -> final build prompt",
		StageInput:  "1 Input",
		StageScan:   "2 Scan",
		StageSpec:   "3 Spec",
		StagePrompt: "4 Prompt",
		StageDone:   "5 Done",
		StateMeta:   "State",

		TitlePathInput:     "Path Input",
		TitleScanning:      "Scanning",
		TitleSpecSummary:   "Spec Summary",
		TitlePromptPreview: "Prompt Preview",
		TitleDone:          "Done",
			TitleConfigEdit:    "Config Edit",
		TitleUnknown:       "Unknown",

		PathIntro:      "Enter a local repository path. The CLI will scan the repo, extract a compact project spec, and compile a production-oriented prompt.",
		PathPanelInfo:  "What This Run Does",
		PathPanelInput: "Input",
		PathExample:    "Example",
		PathFieldLabel: "Repository path",
		PathFieldHint:  "The path is submitted exactly as typed after trimming leading and trailing whitespace.",
		PathDefaultExp: "Default export",

		ScanQueued:      "Queued",
		ScanRunning:     "Running repository scan and spec extraction",
		ScanPhaseMain:   "Repository scan -> document probe -> spec build",
		ScanPhaseFinal:  "Finalizing project specification",
		ScanPanelStatus: "Pipeline Status",
		ScanPanelNotes:  "Notes",
		ScanLabelPath:   "Project path",
		ScanLabelStatus: "Status",
		ScanLabelPhase:  "Current phase",
		ScanLabelCancel: "Cancellation",
		ScanCancelHint:  "Press Esc to abandon this run and return to input",
		ScanNote1:       "Scan results are assembled from deterministic local reads.",
		ScanNote2:       "Prompt compilation starts only after the project spec is ready.",
		ScanNote3:       "No code is generated during this stage.",

		SpecPanelGoal:         "Goal",
		SpecPanelTech:         "Tech Profile",
		SpecPanelModules:      "Core Modules",
		SpecPanelFlows:        "Key Flows",
		SpecPanelConstraints:  "Critical Constraints",
		SpecPanelChecks:       "Acceptance Checks",
		SpecPanelEvidence:     "Scan Evidence",
		SpecPanelQuestions:    "Open Questions",
		SpecGoalFallback:      "Not available",
		SpecNone:              "None",
		SpecUnknown:           "Unknown",
		SpecBlockingPrefix:    "[blocking]",
		SpecUnnamedModule:     "Unnamed module",
		SpecNoSummary:         "No summary",
		SpecUnnamedFlow:       "Unnamed flow",
		SpecUnnamedConstraint: "Unnamed constraint",
		SpecUnnamedCheck:      "Unnamed acceptance check",
		TechPlatform:          "Platform",
		TechLanguages:         "Languages",
		TechFrameworks:        "Frameworks",
		TechBuildSystem:       "Build system",
		TechPackageManager:    "Package manager",
		TechIntegrations:      "Integrations",

		PromptStatusLabel:    "Status",
		PromptCompiling:      "Compiling the final prompt",
		PromptWaiting:        "The prompt will appear here as soon as compilation finishes.",
		PromptPanelCompiler:  "Prompt Compiler",
		PromptPanelStatus:    "Prompt Status",
		PromptNotReady:       "Prompt output is not ready yet.",
		PromptBlockingQTitle: "Blocking Questions",
		PromptLangLabel:      "Output language",
		PromptExportLabel:    "Export path",
		PromptVisibleLabel:   "Visible lines",
		PromptScrollLabel:    "Scroll",
		PromptSectionsLabel:  "Sections",
		PromptPanelPreview:   "Prompt Preview",
		PromptPanelMeta:      "Preview Metadata",

		DonePanelTitle:  "Export Complete",
		DoneOutputPath:  "Output path",
		DonePromptLen:   "Prompt length",
		DoneNextStep:    "Next step",
		DoneNextStepVal: "Use the exported prompt as the build brief for your code generation workflow",

			ConfigPanelTitle:    "Model Configuration",
			ConfigBaseURLLabel:  "Base URL",
			ConfigAPIKeyLabel:   "API Key",
			ConfigModelLabel:    "Model Name",
			ConfigPanelHint:     "Configure your model provider settings. These values will be saved to .env file.",
			ConfigSaved:         "Configuration saved successfully.",
			ConfigSaveFailed:    "Failed to save configuration.",

		HintStartScan:     "start scan",
		HintQuit:          "quit",
		HintCancelRun:     "cancel run",
		HintCompilePrompt: "compile prompt",
		HintRestart:       "restart",
		HintScroll:        "scroll",
		HintExport:        "export",
		HintBack:          "back",
		HintNewRun:        "new run",
		HintToggleLang:    "toggle language",
			HintConfigEdit:    "edit config",
			HintConfigSwitch:  "switch field",
			HintConfigSave:    "save",
			HintConfigCancel:  "cancel",

		LogPanelTitle:     "Recent Activity",
		UnknownPanelTitle: "Unknown State",
		UnknownPanelBody:  "The TUI entered an unknown state.",
		ErrorPanelTitle:   "Error",
		NotSet:            "not set",

		StatRootPath:    "Root path",
		StatDocuments:   "Documents",
		StatEntryPoints: "Entry points",
		StatSnippets:    "Snippets",
		StatConstraints: "Constraints",
		StatTruncated:   "Scan output was truncated to stay within the current budget.",
		StatLastExport:  "Last export",
		StatPrompt:      "Prompt",

		FmtPromptChars:  "%d chars",
		FmtPromptLen:    "%d characters",
		FmtVisibleLines: "%d-%d of %d",
		FmtScrollPos:    "%d / %d",

		ErrPathEmpty:  "Project path cannot be empty.",
		ErrScanFailed: "Scan failed",
		ErrSpecFailed: "Specification build failed",
		ErrPromptFmt:  "Prompt compilation failed: %v",
		ErrExportFmt:  "Export failed: %v",
			ErrConfigIncomplete: "Model configuration incomplete. Press 'c' to configure BaseURL, APIKey, and Model.",
			ErrConfigSaveFmt:   "Failed to save configuration: %v",

		LogReady:           "Ready for a local project path.",
		LogScanningFmt:     "Scanning project: %s",
		LogScanDoneFmt:     "Scan complete. %d docs, %d entry points.",
		LogBuildingSpec:    "Building project specification.",
		LogSpecReady:       "Project specification ready.",
		LogCompilingPrompt: "Compiling final prompt.",
		LogPromptReady:     "Prompt is ready for review.",
		LogExportingFmt:    "Exporting prompt to %s",
		LogExportedFmt:     "Prompt exported to %s",
		LogCancelled:       "Cancelled current run.",
			LogConfigSaving:    "Saving configuration...",
			LogConfigSaved:     "Configuration saved.",
	}
}

func stringsZH() Strings {
	return Strings{
		Subtitle:    "本地仓库分析 -> 项目规格提取 -> 最终提示词编译",
		StageInput:  "1 输入",
		StageScan:   "2 扫描",
		StageSpec:   "3 规格",
		StagePrompt: "4 提示词",
		StageDone:   "5 完成",
		StateMeta:   "状态",

		TitlePathInput:     "路径输入",
		TitleScanning:      "扫描中",
		TitleSpecSummary:   "规格摘要",
		TitlePromptPreview: "提示词预览",
		TitleDone:          "已完成",
			TitleConfigEdit:    "配置编辑",
		TitleUnknown:       "未知",

		PathIntro:      "输入本地仓库路径。工具将扫描仓库、提取精简的项目规格，并编译面向生产的提示词。",
		PathPanelInfo:  "本次运行说明",
		PathPanelInput: "输入",
		PathExample:    "示例",
		PathFieldLabel: "仓库路径",
		PathFieldHint:  "路径在去除首尾空白后原样提交。",
		PathDefaultExp: "默认导出路径",

		ScanQueued:      "排队中",
		ScanRunning:     "正在扫描仓库并提取规格",
		ScanPhaseMain:   "仓库扫描 -> 文档探测 -> 规格构建",
		ScanPhaseFinal:  "正在完成项目规格",
		ScanPanelStatus: "流水线状态",
		ScanPanelNotes:  "备注",
		ScanLabelPath:   "项目路径",
		ScanLabelStatus: "状态",
		ScanLabelPhase:  "当前阶段",
		ScanLabelCancel: "取消",
		ScanCancelHint:  "按 Esc 放弃本次运行并返回输入",
		ScanNote1:       "扫描结果来自确定性的本地读取。",
		ScanNote2:       "只有项目规格就绪后才会开始编译提示词。",
		ScanNote3:       "此阶段不生成任何代码。",

		SpecPanelGoal:         "目标",
		SpecPanelTech:         "技术档案",
		SpecPanelModules:      "核心模块",
		SpecPanelFlows:        "关键流程",
		SpecPanelConstraints:  "关键约束",
		SpecPanelChecks:       "验收检查",
		SpecPanelEvidence:     "扫描证据",
		SpecPanelQuestions:    "待解问题",
		SpecGoalFallback:      "暂无",
		SpecNone:              "无",
		SpecUnknown:           "未知",
		SpecBlockingPrefix:    "[阻塞]",
		SpecUnnamedModule:     "未命名模块",
		SpecNoSummary:         "无摘要",
		SpecUnnamedFlow:       "未命名流程",
		SpecUnnamedConstraint: "未命名约束",
		SpecUnnamedCheck:      "未命名验收检查",
		TechPlatform:          "平台",
		TechLanguages:         "编程语言",
		TechFrameworks:        "框架",
		TechBuildSystem:       "构建系统",
		TechPackageManager:    "包管理器",
		TechIntegrations:      "外部集成",

		PromptStatusLabel:    "状态",
		PromptCompiling:      "正在编译最终提示词",
		PromptWaiting:        "提示词编译完成后将在此显示。",
		PromptPanelCompiler:  "提示词编译器",
		PromptPanelStatus:    "提示词状态",
		PromptNotReady:       "提示词尚未就绪。",
		PromptBlockingQTitle: "阻塞问题",
		PromptLangLabel:      "输出语言",
		PromptExportLabel:    "导出路径",
		PromptVisibleLabel:   "可见行",
		PromptScrollLabel:    "滚动位置",
		PromptSectionsLabel:  "章节",
		PromptPanelPreview:   "提示词预览",
		PromptPanelMeta:      "预览元数据",

		DonePanelTitle:  "导出完成",
		DoneOutputPath:  "输出路径",
		DonePromptLen:   "提示词长度",
		DoneNextStep:    "下一步",
		DoneNextStepVal: "将导出的提示词作为代码生成工作流的构建说明",

			ConfigPanelTitle:    "模型配置",
			ConfigBaseURLLabel:  "基础URL",
			ConfigAPIKeyLabel:   "API密钥",
			ConfigModelLabel:    "模型名称",
			ConfigPanelHint:     "配置您的模型提供商设置。这些值将保存到.env文件。",
			ConfigSaved:         "配置已成功保存。",
			ConfigSaveFailed:    "配置保存失败。",

		HintStartScan:     "开始扫描",
		HintQuit:          "退出",
		HintCancelRun:     "取消运行",
		HintCompilePrompt: "编译提示词",
		HintRestart:       "重新开始",
		HintScroll:        "滚动",
		HintExport:        "导出",
		HintBack:          "返回",
		HintNewRun:        "新建运行",
		HintToggleLang:    "切换语言",
			HintConfigEdit:    "编辑配置",
			HintConfigSwitch:  "切换字段",
			HintConfigSave:    "保存",
			HintConfigCancel:  "取消",

		LogPanelTitle:     "最近活动",
		UnknownPanelTitle: "未知状态",
		UnknownPanelBody:  "TUI 进入了未知状态。",
		ErrorPanelTitle:   "错误",
		NotSet:            "未设置",

		StatRootPath:    "根路径",
		StatDocuments:   "文档数",
		StatEntryPoints: "入口点",
		StatSnippets:    "代码片段",
		StatConstraints: "约束数",
		StatTruncated:   "扫描输出已截断以保持在当前预算内。",
		StatLastExport:  "上次导出",
		StatPrompt:      "提示词",

		FmtPromptChars:  "%d 字符",
		FmtPromptLen:    "%d 个字符",
		FmtVisibleLines: "%d-%d / %d",
		FmtScrollPos:    "%d / %d",

		ErrPathEmpty:  "项目路径不能为空。",
		ErrScanFailed: "扫描失败",
		ErrSpecFailed: "规格构建失败",
		ErrPromptFmt:  "提示词编译失败：%v",
		ErrExportFmt:  "导出失败：%v",
			ErrConfigIncomplete: "模型配置不完整。按 'c' 配置 BaseURL、APIKey 和 Model。",
			ErrConfigSaveFmt:   "配置保存失败：%v",

		LogReady:           "等待本地项目路径。",
		LogScanningFmt:     "正在扫描项目：%s",
		LogScanDoneFmt:     "扫描完成。%d 个文档，%d 个入口点。",
		LogBuildingSpec:    "正在构建项目规格。",
		LogSpecReady:       "项目规格已就绪。",
		LogCompilingPrompt: "正在编译最终提示词。",
		LogPromptReady:     "提示词已可供查看。",
		LogExportingFmt:    "正在导出提示词至 %s",
		LogExportedFmt:     "提示词已导出至 %s",
		LogCancelled:       "已取消当前运行。",
			LogConfigSaving:    "正在保存配置...",
			LogConfigSaved:     "配置已保存。",
	}
}
