package models

// ActionOutput holds the result of a single executed action, including output and timing.
type ActionOutput struct {
	TaskSlug   string `json:"task_slug"`
	ActionSlug string `json:"action_slug"`
	Command    string `json:"command"`
	Output     string `json:"output"`
	Success    bool   `json:"success"`
	Duration   int64  `json:"duration_ms"`
	Timestamp  string `json:"timestamp"`
}
