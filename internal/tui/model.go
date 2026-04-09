package tui

import (
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"projcompiler/internal/spec"
)

type Model struct {
	config   Config
	services Services

	state       State
	activeRunID int
	lang        Language

	pathInput  string
	outputPath string

	facts       spec.RepoFacts
	projectSpec spec.ProjectSpec
	bundle      spec.PromptBundle

	width  int
	height int

	loading          bool
	compilingPrompt  bool
	exporting        bool
	promptScroll     int
	lastError        string
	lastExportedPath string
	logs             []string
}

func (m Model) t() Strings {
	return stringsFor(m.lang)
}

func NewModel(services Services, options ...Option) Model {
	cfg := DefaultConfig()
	for _, option := range options {
		if option != nil {
			option(&cfg)
		}
	}

	model := Model{
		config:     cfg,
		services:   services,
		state:      StatePathInput,
		outputPath: cfg.DefaultPromptPath,
		lang:       LangEN,
	}
	model.appendLog(model.t().LogReady)
	return model
}

func NewProgram(services Services, options ...Option) *tea.Program {
	model := NewModel(services, options...)
	return tea.NewProgram(model, tea.WithAltScreen())
}

func (m Model) Init() tea.Cmd {
	return tea.WindowSize()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch typed := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = typed.Width
		m.height = typed.Height
		return m, nil
	case tea.KeyMsg:
		return m.updateKey(typed)
	case scanStartedMsg:
		if typed.RunID != m.activeRunID {
			return m, nil
		}
		m.loading = true
		m.lastError = ""
		m.appendLog(fmt.Sprintf(m.t().LogScanningFmt, typed.Path))
		return m, nil
	case scanFinishedMsg:
		if typed.RunID != m.activeRunID {
			return m, nil
		}
		m.loading = false
		if typed.Err != nil {
			m.failToInput(m.t().ErrScanFailed, typed.Err)
			return m, nil
		}
		m.facts = typed.Facts
		m.appendLog(fmt.Sprintf(m.t().LogScanDoneFmt,
			len(typed.Facts.Docs), len(typed.Facts.EntryPoints)))
		return m, startBuildSpecCmd(m.activeRunID, m.facts, m.services.SpecBuilder)
	case specStartedMsg:
		if typed.RunID != m.activeRunID {
			return m, nil
		}
		m.loading = true
		m.lastError = ""
		m.appendLog(m.t().LogBuildingSpec)
		return m, nil
	case specFinishedMsg:
		if typed.RunID != m.activeRunID {
			return m, nil
		}
		m.loading = false
		if typed.Err != nil {
			m.failToInput(m.t().ErrSpecFailed, typed.Err)
			return m, nil
		}
		m.projectSpec = typed.ProjectSpec
		m.state = StateSpecSummary
		m.appendLog(m.t().LogSpecReady)
		return m, nil
	case promptStartedMsg:
		if typed.RunID != m.activeRunID {
			return m, nil
		}
		m.state = StatePromptPreview
		m.compilingPrompt = true
		m.lastError = ""
		m.appendLog(m.t().LogCompilingPrompt)
		return m, nil
	case promptFinishedMsg:
		if typed.RunID != m.activeRunID {
			return m, nil
		}
		m.compilingPrompt = false
		if typed.Err != nil {
			m.lastError = fmt.Sprintf(m.t().ErrPromptFmt, typed.Err)
			m.appendLog(m.lastError)
			return m, nil
		}
		m.bundle = typed.Bundle
		m.promptScroll = 0
		m.appendLog(m.t().LogPromptReady)
		return m, nil
	case exportStartedMsg:
		if typed.RunID != m.activeRunID {
			return m, nil
		}
		m.exporting = true
		m.lastError = ""
		m.appendLog(fmt.Sprintf(m.t().LogExportingFmt, typed.OutputPath))
		return m, nil
	case exportFinishedMsg:
		if typed.RunID != m.activeRunID {
			return m, nil
		}
		m.exporting = false
		if typed.Err != nil {
			m.lastError = fmt.Sprintf(m.t().ErrExportFmt, typed.Err)
			m.appendLog(m.lastError)
			return m, nil
		}
		m.lastExportedPath = typed.OutputPath
		m.state = StateDone
		m.appendLog(fmt.Sprintf(m.t().LogExportedFmt, typed.OutputPath))
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "ctrl+c":
		return m, tea.Quit
	case "ctrl+l":
		if m.lang == LangZH {
			m.lang = LangEN
		} else {
			m.lang = LangZH
		}
		return m, nil
	case "q":
		if m.state == StatePathInput || m.state == StateDone {
			return m, tea.Quit
		}
	}

	switch m.state {
	case StatePathInput:
		return m.updatePathInput(msg)
	case StateScanning:
		return m.updateScanning(msg)
	case StateSpecSummary:
		return m.updateSpecSummary(msg)
	case StatePromptPreview:
		return m.updatePromptPreview(msg)
	case StateDone:
		return m.updateDone(msg)
	default:
		return m, nil
	}
}

func (m Model) updatePathInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		path := strings.TrimSpace(m.pathInput)
		if path == "" {
			m.lastError = m.t().ErrPathEmpty
			return m, nil
		}
		m.pathInput = path
		m.resetRunState()
		m.activeRunID++
		m.state = StateScanning
		return m, startScanCmd(m.activeRunID, path, m.services.Scanner)
	case tea.KeyBackspace, tea.KeyDelete:
		runes := []rune(m.pathInput)
		if len(runes) > 0 {
			m.pathInput = string(runes[:len(runes)-1])
		}
		return m, nil
	default:
		if msg.Type == tea.KeyRunes {
			m.pathInput += msg.String()
			m.lastError = ""
		}
		return m, nil
	}
}

func (m Model) updateScanning(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		m.resetForNewRun()
		m.appendLog(m.t().LogCancelled)
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) updateSpecSummary(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "g":
		if m.loading || m.compilingPrompt || m.exporting {
			return m, nil
		}
		m.lastError = ""
		return m, startCompilePromptCmd(m.activeRunID, m.projectSpec, m.services.PromptCompiler)
	case "esc":
		m.resetForNewRun()
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) updatePromptPreview(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "k":
		if m.promptScroll > 0 {
			m.promptScroll--
		}
		return m, nil
	case "down", "j":
		maxScroll := m.maxPromptScroll()
		if m.promptScroll < maxScroll {
			m.promptScroll++
		}
		return m, nil
	case "pgup":
		m.promptScroll -= m.promptPageSize()
		if m.promptScroll < 0 {
			m.promptScroll = 0
		}
		return m, nil
	case "pgdown":
		m.promptScroll += m.promptPageSize()
		if maxScroll := m.maxPromptScroll(); m.promptScroll > maxScroll {
			m.promptScroll = maxScroll
		}
		return m, nil
	case "e":
		if !m.bundle.Ready() || m.compilingPrompt || m.exporting {
			return m, nil
		}
		return m, startExportCmd(m.activeRunID, m.bundle, m.outputPath, m.services.Exporter)
	case "esc":
		if m.compilingPrompt || m.exporting {
			return m, nil
		}
		m.state = StateSpecSummary
		m.lastError = ""
		return m, nil
	default:
		return m, nil
	}
}

func (m Model) updateDone(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "r":
		m.resetForNewRun()
		return m, nil
	case "q", "enter":
		return m, tea.Quit
	default:
		return m, nil
	}
}

func (m *Model) resetRunState() {
	m.facts = spec.RepoFacts{}
	m.projectSpec = spec.ProjectSpec{}
	m.bundle = spec.PromptBundle{}
	m.promptScroll = 0
	m.lastError = ""
	m.lastExportedPath = ""
	m.loading = false
	m.compilingPrompt = false
	m.exporting = false
}

func (m *Model) resetForNewRun() {
	m.activeRunID++
	m.resetRunState()
	m.state = StatePathInput
}

func (m *Model) failToInput(prefix string, err error) {
	m.state = StatePathInput
	m.lastError = fmt.Sprintf("%s: %v", prefix, err)
	m.appendLog(m.lastError)
}

func (m *Model) appendLog(line string) {
	if strings.TrimSpace(line) == "" {
		return
	}
	stamped := fmt.Sprintf("%s  %s", time.Now().Format("15:04:05"), line)
	m.logs = append(m.logs, stamped)
	limit := m.config.MaxLogLines
	if limit <= 0 {
		return
	}
	if len(m.logs) > limit {
		m.logs = m.logs[len(m.logs)-limit:]
	}
}
