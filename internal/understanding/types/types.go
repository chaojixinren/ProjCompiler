package types

import "projcompiler/internal/spec"

// The understanding package now uses spec-native types so ProjectSpec is the
// single public contract for both understanding and planning.
type RepoIdentity = spec.RepoIdentity
type SnapshotMeta = spec.SnapshotMeta
type UnderstandingStats = spec.UnderstandingStats

type FileRole = spec.UnderstandingFileRole

const (
	FileRoleSource    FileRole = spec.UnderstandingFileRoleSource
	FileRoleDoc       FileRole = spec.UnderstandingFileRoleDoc
	FileRoleManifest  FileRole = spec.UnderstandingFileRoleManifest
	FileRoleConfig    FileRole = spec.UnderstandingFileRoleConfig
	FileRoleTest      FileRole = spec.UnderstandingFileRoleTest
	FileRoleGenerated FileRole = spec.UnderstandingFileRoleGenerated
	FileRoleOther     FileRole = spec.UnderstandingFileRoleOther
)

type ParseStatus = spec.UnderstandingParseStatus

const (
	ParseStatusParsed  ParseStatus = spec.UnderstandingParseStatusParsed
	ParseStatusSkipped ParseStatus = spec.UnderstandingParseStatusSkipped
	ParseStatusFailed  ParseStatus = spec.UnderstandingParseStatusFailed
)

type File = spec.UnderstandingFile
type Document = spec.UnderstandingDocument

type SymbolKind = spec.UnderstandingSymbolKind

const (
	SymbolKindPackage  SymbolKind = spec.UnderstandingSymbolKindPackage
	SymbolKindFunction SymbolKind = spec.UnderstandingSymbolKindFunction
	SymbolKindMethod   SymbolKind = spec.UnderstandingSymbolKindMethod
	SymbolKindImport   SymbolKind = spec.UnderstandingSymbolKindImport
	SymbolKindModule   SymbolKind = spec.UnderstandingSymbolKindModule
)

type Span = spec.Span
type Symbol = spec.UnderstandingSymbol

type RelationType = spec.UnderstandingRelationType

const (
	RelationContains  RelationType = spec.UnderstandingRelationContains
	RelationImports   RelationType = spec.UnderstandingRelationImports
	RelationCalls     RelationType = spec.UnderstandingRelationCalls
	RelationReadsEnv  RelationType = spec.UnderstandingRelationReadsEnv
	RelationEntryFlow RelationType = spec.UnderstandingRelationEntryFlow
)

type Relation = spec.UnderstandingRelation
type Flow = spec.UnderstandingFlow
type Constraint = spec.Constraint
type OpenQuestion = spec.OpenQuestion
type Evidence = spec.UnderstandingEvidence

type ConfidenceBand = spec.UnderstandingConfidenceBand

const (
	ConfidenceBandHigh   ConfidenceBand = spec.UnderstandingConfidenceBandHigh
	ConfidenceBandMedium ConfidenceBand = spec.UnderstandingConfidenceBandMedium
	ConfidenceBandLow    ConfidenceBand = spec.UnderstandingConfidenceBandLow
)

type Confidence = spec.UnderstandingConfidence
