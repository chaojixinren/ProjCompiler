package spec

type ProjectSpec struct {
	Goal                string
	TechProfile         TechProfile
	ImplementationShape ImplementationShape
	CriticalConstraints []Constraint
	AcceptanceChecks    []AcceptanceCheck
	OpenQuestions       []OpenQuestion
	EvidenceLog         []EvidenceItem
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
}

type ModuleSpec struct {
	Name           string
	Responsibility string
	Signal         FactSignal
}

type KeyFlow struct {
	Name    string
	Summary string
	Signal  FactSignal
}

type AcceptanceCheck struct {
	Description string
	Signal      FactSignal
}

type OpenQuestion struct {
	Question string
	Blocking bool
	Signal   FactSignal
}

type EvidenceItem struct {
	Label  string
	Source string
	Signal FactSignal
}
