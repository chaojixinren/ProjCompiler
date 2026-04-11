package app

import (
	"testing"

	"projcompiler/internal/config"
)

func TestServiceStatusesIncludeUnifiedSpecServices(t *testing.T) {
	cfg := config.Config{
		Paths: config.Paths{WorkingDir: t.TempDir()},
	}
	application := New(cfg, Services{})

	statusByName := map[string]bool{}
	for _, status := range application.ServiceStatuses() {
		statusByName[status.Name] = status.Implemented
	}

	if _, ok := statusByName["SpecBuilder"]; !ok {
		t.Fatalf("service status for SpecBuilder is missing")
	}
	if _, ok := statusByName["Exporter"]; !ok {
		t.Fatalf("service status for Exporter is missing")
	}
}
