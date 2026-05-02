package models

type Playbook struct {
	Slug        string `json:"slug" yaml:"slug"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	RiskScore   int      `json:"risk_score,omitempty" yaml:"risk_score,omitempty"`
	TargetTags  []string `json:"target_tags,omitempty" yaml:"target_tags,omitempty"`
	SupportedOS []string `json:"supported_os,omitempty" yaml:"supported_os,omitempty"`

	RiskLevel           string `json:"risk_level,omitempty" yaml:"risk_level,omitempty"`
	Timeout             int    `json:"timeout_seconds,omitempty" yaml:"timeout_seconds,omitempty"`
	HumanInTheLoop      bool   `json:"human_in_the_loop,omitempty" yaml:"human_in_the_loop,omitempty"`
	AutoRevertOnFailure bool   `json:"auto_revert_on_failure,omitempty" yaml:"auto_revert_on_failure,omitempty"`

	Variables map[string]interface{} `json:"variables,omitempty" yaml:"variables,omitempty"`
	Notes     []string               `json:"notes,omitempty" yaml:"notes,omitempty"`

	Tasks []Task `json:"tasks" yaml:"tasks"`
}

type Task struct {
	Slug        string `json:"slug" yaml:"slug"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`

	DependsOn   []string               `json:"depends_on,omitempty" yaml:"depends_on,omitempty"`
	Variables   map[string]interface{} `json:"variables,omitempty" yaml:"variables,omitempty"`
	SupportedOS []string               `json:"supported_os,omitempty" yaml:"supported_os,omitempty"`
	Actions     []Action               `json:"actions" yaml:"actions"`
	Timeout     int                    `json:"timeout_seconds,omitempty" yaml:"timeout_seconds,omitempty"`
}

type Action struct {
	Slug        string `json:"slug" yaml:"slug"`
	Name        string `json:"name" yaml:"name"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Command     string `json:"command" yaml:"command"`

	SupportedOS []string `json:"supported_os,omitempty" yaml:"supported_os,omitempty"`
	RiskLevel   string   `json:"risk_level,omitempty" yaml:"risk_level,omitempty"`
	DependsOn   []string `json:"depends_on,omitempty" yaml:"depends_on,omitempty"`

	Timeout          int    `json:"timeout_seconds,omitempty" yaml:"timeout_seconds,omitempty"`
	// RollbackCmd must target the same platform as Command — no platform check is performed at rollback time.
	RollbackCmd      string `json:"rollback_command,omitempty" yaml:"rollback_command,omitempty"`
	RollbackTimeout  int    `json:"rollback_timeout_seconds,omitempty" yaml:"rollback_timeout_seconds,omitempty"`
	ApprovalRequired bool   `json:"approval_required,omitempty" yaml:"approval_required,omitempty"`

	Environment   map[string]string `json:"environment,omitempty" yaml:"environment,omitempty"`
	WorkingDir    string            `json:"working_dir,omitempty" yaml:"working_dir,omitempty"`
	RequiresAdmin bool              `json:"requires_admin,omitempty" yaml:"requires_admin,omitempty"`
}
