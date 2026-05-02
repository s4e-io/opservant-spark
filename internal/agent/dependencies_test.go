package agent

import (
	"testing"

	"github.com/s4e-io/opservant-spark/internal/models"
)

func TestCheckActionDeps_NoDeps(t *testing.T) {
	action := &models.Action{Slug: "a1"}
	ok, reason := newTestExecutor().checkActionDependencies(action, map[string]bool{})
	if !ok {
		t.Errorf("expected true for no deps, got false: %s", reason)
	}
}

func TestCheckActionDeps_NotExecuted(t *testing.T) {
	action := &models.Action{Slug: "a1", DependsOn: []string{"a0"}}
	ok, reason := newTestExecutor().checkActionDependencies(action, map[string]bool{})
	if ok {
		t.Error("expected false for unexecuted dep")
	}
	if reason == "" {
		t.Error("expected non-empty reason")
	}
}

func TestCheckActionDeps_Failed(t *testing.T) {
	action := &models.Action{Slug: "a1", DependsOn: []string{"a0"}}
	ok, reason := newTestExecutor().checkActionDependencies(action, map[string]bool{"a0": false})
	if ok {
		t.Error("expected false for failed dep")
	}
	if reason == "" {
		t.Error("expected non-empty reason")
	}
}

func TestCheckActionDeps_Succeeded(t *testing.T) {
	action := &models.Action{Slug: "a1", DependsOn: []string{"a0"}}
	ok, _ := newTestExecutor().checkActionDependencies(action, map[string]bool{"a0": true})
	if !ok {
		t.Error("expected true for succeeded dep")
	}
}

func TestCheckActionDeps_MultiAllPass(t *testing.T) {
	action := &models.Action{Slug: "a2", DependsOn: []string{"a0", "a1"}}
	ok, _ := newTestExecutor().checkActionDependencies(action, map[string]bool{"a0": true, "a1": true})
	if !ok {
		t.Error("expected true when all deps pass")
	}
}

func TestCheckActionDeps_MultiOneFails(t *testing.T) {
	action := &models.Action{Slug: "a2", DependsOn: []string{"a0", "a1"}}
	ok, _ := newTestExecutor().checkActionDependencies(action, map[string]bool{"a0": true, "a1": false})
	if ok {
		t.Error("expected false when one dep fails")
	}
}

func TestCheckTaskDeps_NoDeps(t *testing.T) {
	task := &models.Task{Slug: "t1"}
	ok, reason := newTestExecutor().checkTaskDependencies(task, map[string]bool{})
	if !ok {
		t.Errorf("expected true for no deps, got false: %s", reason)
	}
}

func TestCheckTaskDeps_NotExecuted(t *testing.T) {
	task := &models.Task{Slug: "t1", DependsOn: []string{"t0"}}
	ok, _ := newTestExecutor().checkTaskDependencies(task, map[string]bool{})
	if ok {
		t.Error("expected false for unexecuted dep")
	}
}

func TestCheckTaskDeps_Failed(t *testing.T) {
	task := &models.Task{Slug: "t1", DependsOn: []string{"t0"}}
	ok, _ := newTestExecutor().checkTaskDependencies(task, map[string]bool{"t0": false})
	if ok {
		t.Error("expected false for failed dep")
	}
}

func TestCheckTaskDeps_Succeeded(t *testing.T) {
	task := &models.Task{Slug: "t1", DependsOn: []string{"t0"}}
	ok, _ := newTestExecutor().checkTaskDependencies(task, map[string]bool{"t0": true})
	if !ok {
		t.Error("expected true for succeeded dep")
	}
}
