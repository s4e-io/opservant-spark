package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/s4e-io/opservant-spark/internal/config"
)

// Level represents logging levels in strict ascending severity order.
// TRACE < DEBUG < INFO < WARN < ERROR < FATAL
type Level int

const (
	TRACE Level = iota // 0 — ultra-verbose, step-by-step
	DEBUG              // 1 — diagnostic detail
	INFO               // 2 — normal operational events
	WARN               // 3 — unexpected but recoverable
	ERROR              // 4 — operation failed, service continues
	FATAL              // 5 — unrecoverable, process must exit
)

// String returns the canonical string representation of a log level.
func (l Level) String() string {
	switch l {
	case TRACE:
		return "TRACE"
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "INFO"
	}
}

// Log writes structured, levelled output to console and optionally to a file.
type Log struct {
	level         Level
	fileLogger    *log.Logger
	consoleLogger *log.Logger
	execLogger    *log.Logger
	logFile       *os.File
	execFile      *os.File
}

// New creates a new logger instance.
func New(level string, cfg config.LoggingConfig) *Log {
	l := &Log{
		level: parseLogLevel(level),
	}

	l.consoleLogger = log.New(os.Stdout, "", 0)

	if cfg.LogToFile {
		if err := l.setupFileLogger(cfg.LogDir); err != nil {
			fmt.Printf("Failed to setup file logger: %v\n", err)
		}
	}

	return l
}

func (l *Log) setupFileLogger(logDir string) error {
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile, err := os.OpenFile(filepath.Join(logDir, "spark.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}
	l.logFile = logFile
	l.fileLogger = log.New(logFile, "", 0)

	execFile, err := os.OpenFile(filepath.Join(logDir, "execution.log"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
	if err != nil {
		logFile.Close()
		return fmt.Errorf("failed to open execution log file: %w", err)
	}
	l.execFile = execFile
	l.execLogger = log.New(execFile, "", 0)

	return nil
}

// Close flushes and closes all open file handles.
func (l *Log) Close() {
	if l.logFile != nil {
		l.logFile.Close()
	}
	if l.execFile != nil {
		l.execFile.Close()
	}
}

// parseLogLevel parses a level string into a Level constant.
func parseLogLevel(level string) Level {
	switch strings.ToLower(level) {
	case "trace":
		return TRACE
	case "debug":
		return DEBUG
	case "info":
		return INFO
	case "warn", "warning":
		return WARN
	case "error":
		return ERROR
	case "fatal":
		return FATAL
	default:
		return INFO
	}
}

// shouldLog reports whether a message at the given level should be emitted.
func (l *Log) shouldLog(level Level) bool {
	return level >= l.level
}

// formatMessage formats log message with timestamp and level.
func (l *Log) formatMessage(label string, format string, args ...interface{}) string {
	timestamp := time.Now().Format("2006-01-02 15:04:05.000000")
	message := fmt.Sprintf(format, args...)
	return fmt.Sprintf("%s [%s] %s", timestamp, label, message)
}

func (l *Log) log(level Level, label string, format string, args ...interface{}) {
	if !l.shouldLog(level) {
		return
	}

	message := l.formatMessage(label, format, args...)

	if l.consoleLogger != nil {
		l.consoleLogger.Println(l.addColor(level, message))
	}

	if l.fileLogger != nil {
		l.fileLogger.Println(message)
	}
}

// LogWithCategory emits a log entry with an execution category.
func (l *Log) LogWithCategory(level Level, category string, format string, args ...interface{}) {
	if !l.shouldLog(level) {
		return
	}

	message := l.formatMessage(level.String(), format, args...)

	if l.consoleLogger != nil {
		l.consoleLogger.Println(l.addColor(level, message))
	}

	if l.fileLogger != nil {
		l.fileLogger.Println(message)
	}
}

// addColor adds ANSI color codes for console output.
func (l *Log) addColor(level Level, message string) string {
	const (
		colorReset    = "\033[0m"
		colorRed      = "\033[31m"
		colorGreen    = "\033[32m"
		colorYellow   = "\033[33m"
		colorMagenta  = "\033[35m"
		colorCyan     = "\033[36m"
		colorDarkGray = "\033[90m"
	)

	switch level {
	case TRACE:
		return colorDarkGray + message + colorReset
	case DEBUG:
		return colorCyan + message + colorReset
	case INFO:
		return colorGreen + message + colorReset
	case WARN:
		return colorYellow + message + colorReset
	case ERROR:
		return colorRed + message + colorReset
	case FATAL:
		return colorMagenta + message + colorReset
	default:
		return message
	}
}

// --- Public logging methods ---

func (l *Log) Trace(format string, args ...interface{}) { l.log(TRACE, "TRACE", format, args...) }
func (l *Log) Debug(format string, args ...interface{}) { l.log(DEBUG, "DEBUG", format, args...) }
func (l *Log) Info(format string, args ...interface{})  { l.log(INFO, "INFO", format, args...) }
func (l *Log) Warn(format string, args ...interface{})  { l.log(WARN, "WARN", format, args...) }
func (l *Log) Error(format string, args ...interface{}) { l.log(ERROR, "ERROR", format, args...) }

func (l *Log) Fatal(format string, args ...interface{}) {
	l.log(FATAL, "FATAL", format, args...)
	os.Exit(1)
}

// --- Execution lifecycle helpers ---

func (l *Log) logToExecLog(format string, args ...interface{}) {
	if l.execLogger != nil {
		timestamp := time.Now().Format("2006-01-02 15:04:05")
		message := fmt.Sprintf(format, args...)
		l.execLogger.Printf("%s %s", timestamp, message)
	}
}

func (l *Log) LogPlaybookStart(playbookSlug string, riskLevel string, timeout time.Duration) {
	l.LogWithCategory(INFO, "execution", "Starting playbook: %s, risk: %s, timeout: %v", playbookSlug, riskLevel, timeout)
	l.logToExecLog("PLAYBOOK STARTED: %s, risk: %s, timeout: %v", playbookSlug, riskLevel, timeout)
}

func (l *Log) LogPlaybookComplete(playbookSlug string, success bool, duration time.Duration) {
	if success {
		l.LogWithCategory(INFO, "execution", "Playbook completed: %s, duration: %v", playbookSlug, duration)
		l.logToExecLog("PLAYBOOK COMPLETED: %s, duration: %v", playbookSlug, duration)
	} else {
		l.LogWithCategory(ERROR, "execution", "Playbook failed: %s, duration: %v", playbookSlug, duration)
		l.logToExecLog("PLAYBOOK FAILED: %s, duration: %v", playbookSlug, duration)
	}
}

func (l *Log) LogTaskStart(taskSlug, taskName string) {
	l.LogWithCategory(INFO, "execution", "Starting task: %s, name: %s", taskSlug, taskName)
	l.logToExecLog("TASK STARTED: %s", taskSlug)
}

func (l *Log) LogTaskComplete(taskSlug string, success bool) {
	if success {
		l.LogWithCategory(INFO, "execution", "Task completed: %s", taskSlug)
		l.logToExecLog("TASK COMPLETED: %s", taskSlug)
	} else {
		l.LogWithCategory(ERROR, "execution", "Task failed: %s", taskSlug)
		l.logToExecLog("TASK FAILED: %s", taskSlug)
	}
}

func (l *Log) LogActionStart(actionSlug, actionName, command string) {
	l.LogWithCategory(DEBUG, "execution", "Starting action: %s, name: %s", actionSlug, actionName)
	l.LogWithCategory(TRACE, "execution", "Action command:\n%s", command)
	l.logToExecLog("ACTION STARTED: %s, command: %s", actionSlug, command)
}

func (l *Log) LogActionComplete(actionSlug string, success bool, output string, duration time.Duration) {
	if success {
		l.LogWithCategory(DEBUG, "execution", "Action completed: %s, duration: %v", actionSlug, duration)
		if output != "" {
			l.LogWithCategory(DEBUG, "execution", "Action output:\n%s", strings.TrimSpace(output))
		}
		l.logToExecLog("ACTION COMPLETED: %s, duration: %v\nOutput: %s", actionSlug, duration, strings.TrimSpace(output))
	} else {
		l.LogWithCategory(ERROR, "execution", "Action failed: %s, duration: %v", actionSlug, duration)
		if output != "" {
			l.LogWithCategory(ERROR, "execution", "Action output:\n%s", strings.TrimSpace(output))
		}
		l.logToExecLog("ACTION FAILED: %s, duration: %v\nOutput: %s", actionSlug, duration, strings.TrimSpace(output))
	}
}

func (l *Log) LogActionSkipped(actionSlug, reason string) {
	l.LogWithCategory(WARN, "execution", "Action skipped: %s, reason: %s", actionSlug, reason)
	l.logToExecLog("ACTION SKIPPED: %s, reason: %s", actionSlug, reason)
}
