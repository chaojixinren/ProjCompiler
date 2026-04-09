package spec

type SignalKind string

const (
	SignalObserved SignalKind = "observed"
	SignalInferred SignalKind = "inferred"
	SignalMissing  SignalKind = "missing"
)

type EvidenceRef struct {
	Path      string
	StartLine int
	EndLine   int
	Note      string
}

type FactSignal struct {
	Kind       SignalKind
	Confidence float64
	Reason     string
	Evidence   []EvidenceRef
}

func ObservedSignal(reason string, evidence ...EvidenceRef) FactSignal {
	return FactSignal{
		Kind:       SignalObserved,
		Confidence: 1.0,
		Reason:     reason,
		Evidence:   evidence,
	}
}

func InferredSignal(confidence float64, reason string, evidence ...EvidenceRef) FactSignal {
	return FactSignal{
		Kind:       SignalInferred,
		Confidence: confidence,
		Reason:     reason,
		Evidence:   evidence,
	}
}

func MissingSignal(reason string, evidence ...EvidenceRef) FactSignal {
	return FactSignal{
		Kind:       SignalMissing,
		Confidence: 0,
		Reason:     reason,
		Evidence:   evidence,
	}
}
