package processor

import (
	"testing"
)

func TestValidate_OK(t *testing.T) {
	if err := newTestValidator().Validate(validPlaybook()); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidate_NoSlug(t *testing.T) {
	p := validPlaybook()
	p.Slug = ""
	if err := newTestValidator().Validate(p); err == nil {
		t.Error("expected error for missing slug")
	}
}

func TestValidate_NoName(t *testing.T) {
	p := validPlaybook()
	p.Name = ""
	if err := newTestValidator().Validate(p); err == nil {
		t.Error("expected error for missing name")
	}
}

func TestValidate_NoTasks(t *testing.T) {
	p := validPlaybook()
	p.Tasks = nil
	if err := newTestValidator().Validate(p); err == nil {
		t.Error("expected error for empty tasks")
	}
}

func TestValidate_BadRiskLevel(t *testing.T) {
	p := validPlaybook()
	p.RiskLevel = "extreme"
	if err := newTestValidator().Validate(p); err != nil {
		t.Errorf("bad risk_level is warning not error, got: %v", err)
	}
}

func TestValidate_EmptyRiskLevel(t *testing.T) {
	p := validPlaybook()
	p.RiskLevel = ""
	if err := newTestValidator().Validate(p); err != nil {
		t.Errorf("empty risk_level is warning not error, got: %v", err)
	}
}

func TestValidate_ZeroTimeout(t *testing.T) {
	p := validPlaybook()
	p.Timeout = 0
	if err := newTestValidator().Validate(p); err != nil {
		t.Errorf("zero timeout is warning not error, got: %v", err)
	}
}

func TestValidate_TaskNoSlug(t *testing.T) {
	p := validPlaybook()
	p.Tasks[0].Slug = ""
	if err := newTestValidator().Validate(p); err == nil {
		t.Error("expected error for task missing slug")
	}
}

func TestValidate_TaskNoName(t *testing.T) {
	p := validPlaybook()
	p.Tasks[0].Name = ""
	if err := newTestValidator().Validate(p); err == nil {
		t.Error("expected error for task missing name")
	}
}

func TestValidate_TaskNoActions(t *testing.T) {
	p := validPlaybook()
	p.Tasks[0].Actions = nil
	if err := newTestValidator().Validate(p); err == nil {
		t.Error("expected error for task with no actions")
	}
}

func TestValidate_ActionNoSlug(t *testing.T) {
	p := validPlaybook()
	p.Tasks[0].Actions[0].Slug = ""
	if err := newTestValidator().Validate(p); err == nil {
		t.Error("expected error for action missing slug")
	}
}

func TestValidate_ActionNoName(t *testing.T) {
	p := validPlaybook()
	p.Tasks[0].Actions[0].Name = ""
	if err := newTestValidator().Validate(p); err == nil {
		t.Error("expected error for action missing name")
	}
}

func TestValidate_ActionNoCommand(t *testing.T) {
	p := validPlaybook()
	p.Tasks[0].Actions[0].Command = ""
	if err := newTestValidator().Validate(p); err == nil {
		t.Error("expected error for action missing command")
	}
}

func TestValidate_ActionBadOS(t *testing.T) {
	p := validPlaybook()
	p.Tasks[0].Actions[0].SupportedOS = []string{"amiga"}
	if err := newTestValidator().Validate(p); err != nil {
		t.Errorf("bad supported_os is warning not error, got: %v", err)
	}
}

func TestValidate_ActionNoSupportedOS(t *testing.T) {
	p := validPlaybook()
	p.Tasks[0].Actions[0].SupportedOS = nil
	if err := newTestValidator().Validate(p); err != nil {
		t.Errorf("missing supported_os is warning not error, got: %v", err)
	}
}

func TestValidate_ActionZeroTimeout(t *testing.T) {
	p := validPlaybook()
	p.Tasks[0].Actions[0].Timeout = 0
	if err := newTestValidator().Validate(p); err != nil {
		t.Errorf("zero action timeout is warning not error, got: %v", err)
	}
}

func TestValidate_CriticalNoHITL(t *testing.T) {
	p := validPlaybook()
	p.RiskLevel = "critical"
	p.HumanInTheLoop = false
	if err := newTestValidator().Validate(p); err != nil {
		t.Errorf("critical+no HITL is warning not error, got: %v", err)
	}
}

func TestValidate_HighNoRollback(t *testing.T) {
	p := validPlaybook()
	p.RiskLevel = "high"
	if err := newTestValidator().Validate(p); err != nil {
		t.Errorf("high+no rollback is warning not error, got: %v", err)
	}
}

func TestValidate_CriticalNoRollback(t *testing.T) {
	p := validPlaybook()
	p.RiskLevel = "critical"
	p.HumanInTheLoop = true
	if err := newTestValidator().Validate(p); err != nil {
		t.Errorf("critical+no rollback is warning not error, got: %v", err)
	}
}

func TestValidate_AutoRevertNoRollback(t *testing.T) {
	p := validPlaybook()
	p.AutoRevertOnFailure = true
	if err := newTestValidator().Validate(p); err != nil {
		t.Errorf("auto_revert+no rollback is warning not error, got: %v", err)
	}
}

func TestValidate_HighWithRollback(t *testing.T) {
	p := validPlaybook()
	p.RiskLevel = "high"
	p.Tasks[0].Actions[0].RollbackCmd = "echo rollback"
	if err := newTestValidator().Validate(p); err != nil {
		t.Errorf("high+rollback should pass, got: %v", err)
	}
}
