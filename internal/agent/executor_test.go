package agent

import (
	"runtime"
	"testing"
)

func TestResolveVars_DollarBrace(t *testing.T) {
	result, err := newTestExecutor().resolveVariables("echo ${name}", map[string]interface{}{"name": "spark"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "echo spark" {
		t.Errorf("expected 'echo spark', got %q", result)
	}
}

func TestResolveVars_DoubleBrace(t *testing.T) {
	result, err := newTestExecutor().resolveVariables("echo {{name}}", map[string]interface{}{"name": "spark"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "echo spark" {
		t.Errorf("expected 'echo spark', got %q", result)
	}
}

func TestResolveVars_BothSyntax(t *testing.T) {
	result, err := newTestExecutor().resolveVariables("${a} {{b}}", map[string]interface{}{"a": "hello", "b": "world"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "hello world" {
		t.Errorf("expected 'hello world', got %q", result)
	}
}

func TestResolveVars_UnknownVar(t *testing.T) {
	result, err := newTestExecutor().resolveVariables("echo ${unknown}", map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "echo ${unknown}" {
		t.Errorf("expected unchanged command, got %q", result)
	}
}

func TestResolveVars_InjectionBlocked(t *testing.T) {
	_, err := newTestExecutor().resolveVariables("echo ${cmd}", map[string]interface{}{"cmd": "safe; rm -rf /"})
	if err == nil {
		t.Error("expected error for injection in variable value")
	}
}

func TestValidateVar_Semicolon(t *testing.T) {
	if err := validateVariableValue("k", "a;b"); err == nil {
		t.Error("expected error for semicolon")
	}
}

func TestValidateVar_DoubleAnd(t *testing.T) {
	if err := validateVariableValue("k", "a&&b"); err == nil {
		t.Error("expected error for &&")
	}
}

func TestValidateVar_DoubleOr(t *testing.T) {
	if err := validateVariableValue("k", "a||b"); err == nil {
		t.Error("expected error for ||")
	}
}

func TestValidateVar_Backtick(t *testing.T) {
	if err := validateVariableValue("k", "a`b"); err == nil {
		t.Error("expected error for backtick")
	}
}

func TestValidateVar_DollarParen(t *testing.T) {
	if err := validateVariableValue("k", "$(cmd)"); err == nil {
		t.Error("expected error for $(")
	}
}

func TestValidateVar_Newline(t *testing.T) {
	if err := validateVariableValue("k", "a\nb"); err == nil {
		t.Error("expected error for newline")
	}
}

func TestValidateVar_CarriageReturn(t *testing.T) {
	if err := validateVariableValue("k", "a\rb"); err == nil {
		t.Error("expected error for carriage return")
	}
}

func TestValidateVar_SafeValue(t *testing.T) {
	if err := validateVariableValue("k", "hello world"); err != nil {
		t.Errorf("expected nil for safe value, got: %v", err)
	}
}

func TestValidateVar_EmptyValue(t *testing.T) {
	if err := validateVariableValue("k", ""); err != nil {
		t.Errorf("expected nil for empty value, got: %v", err)
	}
}

func TestPlatform_EmptyList(t *testing.T) {
	if !newTestExecutor().isPlatformSupported([]string{}) {
		t.Error("expected true for empty platform list")
	}
}

func TestPlatform_CrossPlatform(t *testing.T) {
	if !newTestExecutor().isPlatformSupported([]string{"cross-platform"}) {
		t.Error("expected true for cross-platform")
	}
}

func TestPlatform_EmptyStringInList(t *testing.T) {
	if !newTestExecutor().isPlatformSupported([]string{""}) {
		t.Error("expected true for empty string in list")
	}
}

func TestPlatform_CurrentOS(t *testing.T) {
	platform := runtime.GOOS
	if platform == "darwin" {
		platform = "macos"
	}
	if !newTestExecutor().isPlatformSupported([]string{platform}) {
		t.Errorf("expected current platform %s to be supported", platform)
	}
}

func TestPlatform_OtherOS(t *testing.T) {
	other := "windows"
	if runtime.GOOS == "windows" {
		other = "linux"
	}
	if newTestExecutor().isPlatformSupported([]string{other}) {
		t.Errorf("expected platform %s to not be supported on %s", other, runtime.GOOS)
	}
}

func TestPlatform_MacOSAlias(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin only")
	}
	if !newTestExecutor().isPlatformSupported([]string{"macos"}) {
		t.Error("expected macos to be supported on darwin")
	}
}

func TestPlatform_DarwinAlias(t *testing.T) {
	if runtime.GOOS != "darwin" {
		t.Skip("darwin only")
	}
	if !newTestExecutor().isPlatformSupported([]string{"darwin"}) {
		t.Error("expected darwin to be supported on darwin")
	}
}
