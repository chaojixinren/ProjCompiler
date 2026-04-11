package spec

import "time"

type ProjectSpec struct {
	Goal                string
	TechProfile         TechProfile
	ImplementationShape ImplementationShape
	CriticalConstraints []Constraint
	AcceptanceChecks    []AcceptanceCheck
	OpenQuestions       []OpenQuestion
	EvidenceLog         []EvidenceItem
	Understanding       UnderstandingSpec
}

type TechProfile struct {
	Platform             string
	Languages            []string
	Frameworks           []string
	BuildSystem          string
	PackageManager       string
	ExternalIntegrations []string
}

type ImplementationShape struct {
	CoreModules []ModuleSpec
	KeyFlows    []KeyFlow
	APIRoutes   []APIRoute
	DataFlows   []DataFlow
}

type APIRoute struct {
	Method     string
	Path       string
	Handler    string
	Middleware []string
	Signal     FactSignal
}

type DataFlow struct {
	Name   string
	Steps  []string
	Signal FactSignal
}

type ModuleSpec struct {
	ID             string
	Name           string
	Responsibility string
	SymbolIDs      []string
	EvidenceIDs    []string
	Signal         FactSignal
}

type KeyFlow struct {
	ID            string
	Name          string
	Summary       string
	Trigger       string
	EntrySymbolID string
	StepSymbolIDs []string
	RelationIDs   []string
	EvidenceIDs   []string
	Signal        FactSignal
}

type AcceptanceCheck struct {
	Description string
	Signal      FactSignal
}

type OpenQuestion struct {
	ID               string
	Question         string
	Category         string
	Blocking         bool
	Reason           string
	MissingEntityIDs []string
	EvidenceIDs      []string
	Signal           FactSignal
}

type EvidenceItem struct {
	ID          string
	Label       string
	Source      string
	EvidenceIDs []string
	Signal      FactSignal
}

type UnderstandingSpec struct {
	SchemaVersion string
	Repo          RepoIdentity
	Snapshot      SnapshotMeta
	Files         []UnderstandingFile
	Documents     []UnderstandingDocument
	Symbols       []UnderstandingSymbol
	Relations     []UnderstandingRelation
	Flows         []UnderstandingFlow
	Evidences     []UnderstandingEvidence
	Confidences   []UnderstandingConfidence
	Stats         UnderstandingStats
}

type RepoIdentity struct {
	RootPath string
	RepoName string
	VCSRef   string
}

type SnapshotMeta struct {
	RunID          string
	BuiltAt        time.Time
	ScannerVersion string
	ParserVersion  string
	Mode           string
}

type UnderstandingStats struct {
	FileCount       int
	SymbolCount     int
	RelationCount   int
	FlowCount       int
	QuestionCount   int
	ConstraintCount int
}

type UnderstandingFileRole string

const (
	UnderstandingFileRoleSource    UnderstandingFileRole = "source"
	UnderstandingFileRoleDoc       UnderstandingFileRole = "doc"
	UnderstandingFileRoleManifest  UnderstandingFileRole = "manifest"
	UnderstandingFileRoleConfig    UnderstandingFileRole = "config"
	UnderstandingFileRoleTest      UnderstandingFileRole = "test"
	UnderstandingFileRoleGenerated UnderstandingFileRole = "generated"
	UnderstandingFileRoleOther     UnderstandingFileRole = "other"
)

type UnderstandingParseStatus string

const (
	UnderstandingParseStatusParsed  UnderstandingParseStatus = "parsed"
	UnderstandingParseStatusSkipped UnderstandingParseStatus = "skipped"
	UnderstandingParseStatusFailed  UnderstandingParseStatus = "failed"
)

type UnderstandingFile struct {
	ID          string
	Path        string
	AbsPath     string
	Language    string
	Role        UnderstandingFileRole
	SizeBytes   int64
	ContentHash string
	ParseStatus UnderstandingParseStatus
	SkipReason  string
	EvidenceIDs []string
}

type UnderstandingDocument struct {
	ID            string
	FileID        string
	Title         string
	Summary       string
	ConstraintIDs []string
	EvidenceIDs   []string
}

type UnderstandingSymbolKind string

const (
	UnderstandingSymbolKindPackage   UnderstandingSymbolKind = "package"
	UnderstandingSymbolKindFunction  UnderstandingSymbolKind = "function"
	UnderstandingSymbolKindMethod    UnderstandingSymbolKind = "method"
	UnderstandingSymbolKindImport    UnderstandingSymbolKind = "import"
	UnderstandingSymbolKindModule    UnderstandingSymbolKind = "module"
	UnderstandingSymbolKindStruct    UnderstandingSymbolKind = "struct"
	UnderstandingSymbolKindInterface UnderstandingSymbolKind = "interface"
)

type Span struct {
	StartByte uint32
	EndByte   uint32
	StartLine int
	EndLine   int
	StartCol  int
	EndCol    int
}

type UnderstandingSymbol struct {
	ID            string
	FileID        string
	Kind          UnderstandingSymbolKind
	Name          string
	QualifiedName string
	PackageName   string
	IsEntrypoint  bool
	Span          Span
	SourceCapture string
	EvidenceIDs   []string
}

type UnderstandingRelationType string

const (
	UnderstandingRelationContains     UnderstandingRelationType = "contains"
	UnderstandingRelationImports      UnderstandingRelationType = "imports"
	UnderstandingRelationCalls        UnderstandingRelationType = "calls"
	UnderstandingRelationReadsEnv     UnderstandingRelationType = "reads_env"
	UnderstandingRelationEntryFlow    UnderstandingRelationType = "entry_flow"
	UnderstandingRelationDefinesField UnderstandingRelationType = "defines_field"
)

type UnderstandingRelation struct {
	ID           string
	Type         UnderstandingRelationType
	FromID       string
	ToID         string
	ToRef        string
	Resolver     string
	ConfidenceID string
	EvidenceIDs  []string
}

type UnderstandingFlow struct {
	ID            string
	Name          string
	Trigger       string
	EntrySymbolID string
	StepSymbolIDs []string
	EvidenceIDs   []string
}

type UnderstandingEvidence struct {
	ID          string
	SourceKind  string
	Path        string
	Span        Span
	Extractor   string
	SnippetHash string
	Note        string
}

type UnderstandingConfidenceBand string

const (
	UnderstandingConfidenceBandHigh   UnderstandingConfidenceBand = "high"
	UnderstandingConfidenceBandMedium UnderstandingConfidenceBand = "medium"
	UnderstandingConfidenceBandLow    UnderstandingConfidenceBand = "low"
)

type UnderstandingConfidence struct {
	ID         string
	Kind       string
	Score      float64
	Band       UnderstandingConfidenceBand
	ReasonCode string
}
