package parsers

import (
	"context"
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"

	"projcompiler/internal/understanding/types"
)

type GoProvider struct{}

func NewGoProvider() *GoProvider {
	return &GoProvider{}
}

func (p *GoProvider) Name() string {
	return "go-ast"
}

func (p *GoProvider) Supports(file types.File) bool {
	return file.Language == "go"
}

func (p *GoProvider) Parse(ctx context.Context, file types.File, src []byte) (ParseResult, error) {
	if err := ctx.Err(); err != nil {
		return ParseResult{}, err
	}

	fset := token.NewFileSet()
	parsed, err := parser.ParseFile(fset, file.AbsPath, src, parser.ParseComments)
	if err != nil {
		return ParseResult{}, err
	}

	out := ParseResult{
		Symbols:     make([]types.Symbol, 0, 16),
		Relations:   make([]types.Relation, 0, 32),
		Evidences:   make([]types.Evidence, 0, 32),
		Confidences: make([]types.Confidence, 0, 32),
	}

	pkgID := types.NewID(file.ID, "pkg", parsed.Name.Name)
	pkgSpan := nodeSpan(fset, parsed.Name)
	pkgEvidence := newEvidence(file, "go-ast.package", pkgSpan, parsed.Name.Name)
	out.Evidences = append(out.Evidences, pkgEvidence)
	out.Symbols = append(out.Symbols, types.Symbol{
		ID:            pkgID,
		FileID:        file.ID,
		Kind:          types.SymbolKindPackage,
		Name:          parsed.Name.Name,
		QualifiedName: parsed.Name.Name,
		PackageName:   parsed.Name.Name,
		Span:          pkgSpan,
		SourceCapture: "@definition.package",
		EvidenceIDs:   []string{pkgEvidence.ID},
	})

	for _, imp := range parsed.Imports {
		if err := ctx.Err(); err != nil {
			return ParseResult{}, err
		}
		name := strings.Trim(imp.Path.Value, `"`)
		if name == "" {
			continue
		}
		span := nodeSpan(fset, imp)
		ev := newEvidence(file, "go-ast.import", span, name)
		out.Evidences = append(out.Evidences, ev)
		conf := newConfidence("observed", 1.0, "ast-import")
		out.Confidences = append(out.Confidences, conf)
		rel := types.Relation{
			ID:           types.NewID(file.ID, "import", name, strconv.Itoa(span.StartLine)),
			Type:         types.RelationImports,
			FromID:       pkgID,
			ToRef:        name,
			Resolver:     "go-ast",
			ConfidenceID: conf.ID,
			EvidenceIDs:  []string{ev.ID},
		}
		out.Relations = append(out.Relations, rel)
	}

	ast.Inspect(parsed, func(n ast.Node) bool {
		if n == nil {
			return true
		}
		if err := ctx.Err(); err != nil {
			return false
		}
		switch decl := n.(type) {
		case *ast.FuncDecl:
			symbol := p.functionSymbol(file, parsed.Name.Name, fset, decl)
			out.Symbols = append(out.Symbols, symbol)
			ev := newEvidence(file, "go-ast.func", symbol.Span, symbol.QualifiedName)
			out.Evidences = append(out.Evidences, ev)
			out.Symbols[len(out.Symbols)-1].EvidenceIDs = []string{ev.ID}
			containsConf := newConfidence("observed", 1.0, "ast-contains")
			out.Confidences = append(out.Confidences, containsConf)
			out.Relations = append(out.Relations, types.Relation{
				ID:           types.NewID(pkgID, "contains", symbol.ID),
				Type:         types.RelationContains,
				FromID:       pkgID,
				ToID:         symbol.ID,
				Resolver:     "go-ast",
				ConfidenceID: containsConf.ID,
				EvidenceIDs:  []string{ev.ID},
			})

			ast.Inspect(decl.Body, func(child ast.Node) bool {
				call, ok := child.(*ast.CallExpr)
				if !ok {
					return true
				}
				toRef := callTarget(call)
				if toRef == "" {
					return true
				}
				span := nodeSpan(fset, call)
				callEv := newEvidence(file, "go-ast.call", span, toRef)
				out.Evidences = append(out.Evidences, callEv)
				callConf := newConfidence("inferred", 0.7, "ast-call-unresolved")
				out.Confidences = append(out.Confidences, callConf)
				out.Relations = append(out.Relations, types.Relation{
					ID:           types.NewID(symbol.ID, "call", toRef, strconv.Itoa(span.StartLine)),
					Type:         types.RelationCalls,
					FromID:       symbol.ID,
					ToRef:        toRef,
					Resolver:     "go-ast",
					ConfidenceID: callConf.ID,
					EvidenceIDs:  []string{callEv.ID},
				})
				return true
			})
		}
		return true
	})

	return out, nil
}

func (p *GoProvider) functionSymbol(file types.File, packageName string, fset *token.FileSet, decl *ast.FuncDecl) types.Symbol {
	name := decl.Name.Name
	kind := types.SymbolKindFunction
	qualified := packageName + "." + name
	if decl.Recv != nil && len(decl.Recv.List) > 0 {
		kind = types.SymbolKindMethod
		recvType := typeName(decl.Recv.List[0].Type)
		if recvType != "" {
			qualified = packageName + "." + recvType + "." + name
		}
	}
	return types.Symbol{
		ID:            types.NewID(file.ID, string(kind), qualified),
		FileID:        file.ID,
		Kind:          kind,
		Name:          name,
		QualifiedName: qualified,
		PackageName:   packageName,
		IsEntrypoint:  name == "main",
		Span:          nodeSpan(fset, decl),
		SourceCapture: "@definition.function",
	}
}

func typeName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return typeName(t.X)
	case *ast.SelectorExpr:
		return typeName(t.X) + "." + t.Sel.Name
	default:
		return ""
	}
}

func callTarget(call *ast.CallExpr) string {
	switch expr := call.Fun.(type) {
	case *ast.Ident:
		return expr.Name
	case *ast.SelectorExpr:
		left := typeName(expr.X)
		if left == "" {
			return expr.Sel.Name
		}
		return left + "." + expr.Sel.Name
	default:
		return ""
	}
}

func nodeSpan(fset *token.FileSet, node ast.Node) types.Span {
	start := fset.Position(node.Pos())
	end := fset.Position(node.End())
	return types.Span{
		StartByte: uint32(start.Offset),
		EndByte:   uint32(end.Offset),
		StartLine: start.Line,
		EndLine:   end.Line,
		StartCol:  start.Column,
		EndCol:    end.Column,
	}
}

func newEvidence(file types.File, sourceKind string, span types.Span, note string) types.Evidence {
	return types.Evidence{
		ID:          types.NewID("ev", file.ID, sourceKind, note, strconv.Itoa(span.StartLine)),
		SourceKind:  sourceKind,
		Path:        file.Path,
		Span:        span,
		Extractor:   "go-provider/v1",
		SnippetHash: file.ContentHash,
		Note:        note,
	}
}

func newConfidence(kind string, score float64, reasonCode string) types.Confidence {
	return types.Confidence{
		ID:         types.NewID("conf", kind, reasonCode, strconv.FormatFloat(score, 'f', 4, 64)),
		Kind:       kind,
		Score:      score,
		Band:       types.ConfidenceBandFor(score),
		ReasonCode: reasonCode,
	}
}
