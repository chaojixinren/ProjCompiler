package scan

type ReadOnlyTool interface {
	Name() string
	IsReadOnly() bool
	IsConcurrencySafe() bool
}

type toolInfo struct {
	name string
}

func (t toolInfo) Name() string {
	return t.name
}

func (toolInfo) IsReadOnly() bool {
	return true
}

func (toolInfo) IsConcurrencySafe() bool {
	return true
}
