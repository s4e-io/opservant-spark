package processor

import (
	"testing"

	"github.com/s4e-io/opservant-spark/internal/models"
)

func TestEnrich_VariablesPropagated(t *testing.T) {
	pb := &models.Playbook{
		Slug:      "test",
		Variables: map[string]interface{}{"key": "value"},
		Tasks:     []models.Task{{Slug: "t", Name: "T"}},
	}
	result, err := newTestEnricher().Enrich(pb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Tasks[0].Variables["key"] != "value" {
		t.Errorf("expected variable propagated to task, got %v", result.Tasks[0].Variables["key"])
	}
}

func TestEnrich_TaskVarNotOverridden(t *testing.T) {
	pb := &models.Playbook{
		Slug:      "test",
		Variables: map[string]interface{}{"key": "playbook-val"},
		Tasks: []models.Task{
			{Slug: "t", Name: "T", Variables: map[string]interface{}{"key": "task-val"}},
		},
	}
	result, err := newTestEnricher().Enrich(pb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Tasks[0].Variables["key"] != "task-val" {
		t.Errorf("task variable should not be overridden, got %v", result.Tasks[0].Variables["key"])
	}
}

func TestEnrich_EmptyTargetTagsNil(t *testing.T) {
	pb := &models.Playbook{
		Slug:       "test",
		TargetTags: []string{},
		Tasks:      []models.Task{{Slug: "t", Name: "T"}},
	}
	result, err := newTestEnricher().Enrich(pb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TargetTags != nil {
		t.Errorf("expected TargetTags nil, got %v", result.TargetTags)
	}
}

func TestEnrich_EmptyNotesNil(t *testing.T) {
	pb := &models.Playbook{
		Slug:  "test",
		Notes: []string{},
		Tasks: []models.Task{{Slug: "t", Name: "T"}},
	}
	result, err := newTestEnricher().Enrich(pb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Notes != nil {
		t.Errorf("expected Notes nil, got %v", result.Notes)
	}
}

func TestEnrich_EmptyTaskDependsOnNil(t *testing.T) {
	pb := &models.Playbook{
		Slug:  "test",
		Tasks: []models.Task{{Slug: "t", Name: "T", DependsOn: []string{}}},
	}
	result, err := newTestEnricher().Enrich(pb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Tasks[0].DependsOn != nil {
		t.Errorf("expected DependsOn nil, got %v", result.Tasks[0].DependsOn)
	}
}

func TestEnrich_EmptyTaskSupportedOSNil(t *testing.T) {
	pb := &models.Playbook{
		Slug:  "test",
		Tasks: []models.Task{{Slug: "t", Name: "T", SupportedOS: []string{}}},
	}
	result, err := newTestEnricher().Enrich(pb)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Tasks[0].SupportedOS != nil {
		t.Errorf("expected SupportedOS nil, got %v", result.Tasks[0].SupportedOS)
	}
}
