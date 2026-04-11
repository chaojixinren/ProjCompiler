package parsers

import (
	"context"
	"fmt"
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

	stdlibImports := make(map[string]bool, len(parsed.Imports))
	for _, imp := range parsed.Imports {
		if err := ctx.Err(); err != nil {
			return ParseResult{}, err
		}
		name := strings.Trim(imp.Path.Value, `"`)
		if name == "" {
			continue
		}
		isStdlib := !strings.Contains(name, ".")
		var localName string
		if imp.Name != nil && imp.Name.Name != "_" && imp.Name.Name != "." {
			localName = imp.Name.Name
		} else if imp.Name == nil {
			parts := strings.Split(name, "/")
			localName = parts[len(parts)-1]
		}
		if localName != "" && isStdlib {
			stdlibImports[localName] = true
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
				if skipCall(call, stdlibImports) {
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

		case *ast.GenDecl:
			if decl.Tok != token.TYPE {
				break
			}
			for _, typeSpec := range decl.Specs {
				ts, ok := typeSpec.(*ast.TypeSpec)
				if !ok || ts.Name == nil {
					continue
				}
				p.extractTypeDecl(file, parsed.Name.Name, fset, pkgID, ts, &out)
			}
		}
		return true
	})

	return out, nil
}

var goBuiltins = map[string]bool{
	"append":  true,
	"cap":     true,
	"clear":   true,
	"close":   true,
	"complex": true,
	"copy":    true,
	"delete":  true,
	"imag":    true,
	"len":     true,
	"make":    true,
	"max":     true,
	"min":     true,
	"new":     true,
	"panic":   true,
	"print":   true,
	"println": true,
	"real":    true,
	"recover": true,
	"error":   true,
}

func skipCall(call *ast.CallExpr, stdlibImports map[string]bool) bool {
	switch expr := call.Fun.(type) {
	case *ast.Ident:
		return goBuiltins[expr.Name]
	case *ast.SelectorExpr:
		switch x := expr.X.(type) {
		case *ast.Ident:
			return stdlibImports[x.Name]
		default:
			return true
		}
	}
	return false
}

func (p *GoProvider) extractTypeDecl(file types.File, packageName string, fset *token.FileSet, pkgID string, ts *ast.TypeSpec, out *ParseResult) {
	name := ts.Name.Name
	qualified := packageName + "." + name
	span := nodeSpan(fset, ts)

	var kind types.SymbolKind
	switch ts.Type.(type) {
	case *ast.StructType:
		kind = types.SymbolKindStruct
	case *ast.InterfaceType:
		kind = types.SymbolKindInterface
	default:
		return
	}

	symID := types.NewID(file.ID, string(kind), qualified)
	ev := newEvidence(file, "go-ast.type", span, qualified)
	out.Evidences = append(out.Evidences, ev)
	conf := newConfidence("observed", 1.0, "ast-type-decl")
	out.Confidences = append(out.Confidences, conf)

	out.Symbols = append(out.Symbols, types.Symbol{
		ID:            symID,
		FileID:        file.ID,
		Kind:          kind,
		Name:          name,
		QualifiedName: qualified,
		PackageName:   packageName,
		Span:          span,
		SourceCapture: "@definition.type",
		EvidenceIDs:   []string{ev.ID},
	})
	out.Relations = append(out.Relations, types.Relation{
		ID:           types.NewID(pkgID, "contains", symID),
		Type:         types.RelationContains,
		FromID:       pkgID,
		ToID:         symID,
		Resolver:     "go-ast",
		ConfidenceID: conf.ID,
		EvidenceIDs:  []string{ev.ID},
	})

	switch st := ts.Type.(type) {
	case *ast.StructType:
		if st.Fields == nil {
			break
		}
		for _, field := range st.Fields.List {
			fieldType := formatFieldType(field.Type)
			for _, ident := range field.Names {
				fieldRef := ident.Name + " " + fieldType
				fieldConf := newConfidence("observed", 1.0, "ast-struct-field")
				out.Confidences = append(out.Confidences, fieldConf)
				out.Relations = append(out.Relations, types.Relation{
					ID:           types.NewID(symID, "field", ident.Name),
					Type:         types.RelationDefinesField,
					FromID:       symID,
					ToRef:        fieldRef,
					Resolver:     "go-ast",
					ConfidenceID: fieldConf.ID,
					EvidenceIDs:  []string{ev.ID},
				})
			}
			if len(field.Names) == 0 {
				fieldRef := formatFieldType(field.Type)
				fieldConf := newConfidence("observed", 1.0, "ast-struct-embed")
				out.Confidences = append(out.Confidences, fieldConf)
				out.Relations = append(out.Relations, types.Relation{
					ID:           types.NewID(symID, "embed", fieldRef),
					Type:         types.RelationDefinesField,
					FromID:       symID,
					ToRef:        fieldRef,
					Resolver:     "go-ast",
					ConfidenceID: fieldConf.ID,
					EvidenceIDs:  []string{ev.ID},
				})
			}
		}

	case *ast.InterfaceType:
		if st.Methods == nil {
			break
		}
		for _, method := range st.Methods.List {
			for _, ident := range method.Names {
				methodRef := ident.Name + "()"
				methodConf := newConfidence("observed", 1.0, "ast-iface-method")
				out.Confidences = append(out.Confidences, methodConf)
				out.Relations = append(out.Relations, types.Relation{
					ID:           types.NewID(symID, "method", ident.Name),
					Type:         types.RelationDefinesField,
					FromID:       symID,
					ToRef:        methodRef,
					Resolver:     "go-ast",
					ConfidenceID: methodConf.ID,
					EvidenceIDs:  []string{ev.ID},
				})
			}
		}
	}
}

func formatFieldType(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + formatFieldType(t.X)
	case *ast.SelectorExpr:
		return formatFieldType(t.X) + "." + t.Sel.Name
	case *ast.ArrayType:
		return "[]" + formatFieldType(t.Elt)
	case *ast.MapType:
		return fmt.Sprintf("map[%s]%s", formatFieldType(t.Key), formatFieldType(t.Value))
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.FuncType:
		return "func(...)"
	case *ast.ChanType:
		return "chan " + formatFieldType(t.Value)
	case *ast.StructType:
		return "struct{...}"
	case *ast.Ellipsis:
		return "..." + formatFieldType(t.Elt)
	case *ast.ParenExpr:
		return "(" + formatFieldType(t.X) + ")"
	case *ast.IndexExpr:
		return formatFieldType(t.X) + "[" + formatFieldType(t.Index) + "]"
	default:
		return "any"
	}
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
