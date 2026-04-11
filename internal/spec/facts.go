package spec

type RepoFacts struct {
	RootPath    string
	ScanMeta    ScanMeta
	Tree        FileTreeSummary
	Docs        []DocumentFact
	Build       BuildSummary
	EntryPoints []EntryPoint
	Snippets    []CodeSnippet
	Constraints []Constraint
	Signals     []FactSignal
}

type ScanMeta struct {
	IgnoreSources []string
	IgnoredPaths  []IgnoredPath
	MaxDepth      int
	MaxFiles      int
	Truncated     bool
}

type IgnoredPath struct {
	Path   string
	Reason string
}

type FileTreeSummary struct {
	Nodes              []FileNode
	CandidateDocs      []string
	CandidateManifests []string
	CandidateSources   []string
}

type FileNode struct {
	Path string
	Kind FileNodeKind
	Size int64
}

type FileNodeKind string

const (
	FileNodeFile      FileNodeKind = "file"
	FileNodeDirectory FileNodeKind = "directory"
)

type DocumentFact struct {
	Path    string
	Title   string
	Excerpt string
	Signal  FactSignal
}

type BuildSummary struct {
	ModulePath    string
	ManifestFiles []string
	Dependencies  []Dependency
	Hints         []Hint
}

type Dependency struct {
	Name    string
	Version string
	Signal  FactSignal
}

type Hint struct {
	Key    string
	Value  string
	Signal FactSignal
}

type EntryPoint struct {
	Path   string
	Symbol string
	Kind   string
	Score  int
	Signal FactSignal
}

type CodeSnippet struct {
	Path      string
	StartLine int
	EndLine   int
	Purpose   string
	Content   string
	Signal    FactSignal
}

type Constraint struct {
	ID           string
	Text         string
	Category     string
	Scope        string
	ConfidenceID string
	EvidenceIDs  []string
	Signal       FactSignal
}
