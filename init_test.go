package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestNormalizeLanguage(t *testing.T) {
	tests := map[string]string{"CPP": "c++", " c# ": "csharp", "golang": "go", "JS": "javascript", "python": "python"}
	for input, want := range tests {
		if got := normalizeLanguage(input); got != want {
			t.Errorf("normalizeLanguage(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestValidateName(t *testing.T) {
	for _, name := range []string{"", ".", "..", "../escape", "a/b", `a\\b`} {
		if _, err := validateName(name); err == nil {
			t.Errorf("validateName(%q) returned no error", name)
		}
	}
	if got, err := validateName("my-app"); err != nil || got != "my-app" {
		t.Fatalf("validateName(my-app) = %q, %v", got, err)
	}
}

func TestExecuteVersion(t *testing.T) {
	var out bytes.Buffer
	if err := execute([]string{"--version"}, &out); err != nil {
		t.Fatal(err)
	}
	if got, want := out.String(), "code-init 1.0.0\n"; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestExecuteCreatesFallbackFile(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing", "nested")
	var out bytes.Buffer
	if err := execute([]string{"python", dir, "hello"}, &out); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(dir, "hello.py")
	if _, err := os.Stat(file); err != nil {
		t.Fatalf("fallback file not created: %v", err)
	}
	if err := execute([]string{"python", dir, "hello"}, &out); err == nil {
		t.Fatal("expected existing destination error")
	}
}

func TestExecuteAcceptsBackslashSeparatedPath(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "missing", "nested")
	backslashPath := strings.ReplaceAll(dir, "/", `\`)
	if err := execute([]string{"python", backslashPath, "hello"}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(dir, "hello.py")); err != nil {
		t.Fatalf("fallback file not created at normalized path: %v", err)
	}
}

func TestExecuteUsesNativeTool(t *testing.T) {
	oldLookPath, oldRun := lookPath, runCommand
	t.Cleanup(func() { lookPath, runCommand = oldLookPath, oldRun })
	lookPath = func(file string) (string, error) {
		if file == "go" {
			return "/bin/go", nil
		}
		return "", errors.New("not found")
	}
	var got command
	runCommand = func(c command) error { got = c; return nil }

	dir := t.TempDir()
	if err := execute([]string{"go", dir, "example.com/unsafe"}, &bytes.Buffer{}); err == nil {
		t.Fatal("expected invalid name error")
	}
	if err := execute([]string{"go", dir, "api"}, &bytes.Buffer{}); err != nil {
		t.Fatal(err)
	}
	want := command{name: "go", args: []string{"mod", "init", "api"}, dir: filepath.Join(dir, "api")}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("command = %#v, want %#v", got, want)
	}
}

func TestReadOSRelease(t *testing.T) {
	file := filepath.Join(t.TempDir(), "os-release")
	if err := os.WriteFile(file, []byte("ID=ubuntu\nPRETTY_NAME=\"Ubuntu 24.04 LTS\"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	id, pretty, err := readOSRelease(file)
	if err != nil {
		t.Fatal(err)
	}
	if id != "ubuntu" || pretty != "Ubuntu 24.04 LTS" {
		t.Fatalf("got %q, %q", id, pretty)
	}
}

func TestSwiftUsesCompilerPackageNotOpenStackClient(t *testing.T) {
	if got := languages["swift"].executable; got != "swiftc" {
		t.Fatalf("Swift toolchain probe = %q, want swiftc", got)
	}
	if got := packageManagers["dnf"].packages["swift"]; got != "swift-lang" {
		t.Fatalf("DNF Swift package = %q, want swift-lang", got)
	}
	for managerName, manager := range packageManagers {
		if got := manager.packages["swift"]; got == "swift" {
			t.Errorf("%s maps Swift to the OpenStack client package", managerName)
		}
	}
}

func TestPackageMappingsOnlyInstallNativeProjectTools(t *testing.T) {
	for managerName, manager := range packageManagers {
		for language, packageName := range manager.packages {
			spec, ok := languages[language]
			if !ok {
				t.Errorf("%s maps unknown language %q", managerName, language)
				continue
			}
			if spec.create == nil || spec.executable == "" {
				t.Errorf("%s maps %s even though it has no native project tool", managerName, language)
			}
			if packageName == "" {
				t.Errorf("%s has an empty package for %s", managerName, language)
			}
		}
	}
}

func TestInstallCommands(t *testing.T) {
	tests := []struct {
		name      string
		got, want command
	}{
		{"apt", packageManagers["apt-get"].install("golang"), command{name: "sudo", args: []string{"apt-get", "install", "-y", "golang"}}},
		{"brew", packageManagers["brew"].install("go"), command{name: "brew", args: []string{"install", "go"}}},
		{"winget", packageManagers["winget"].install("GoLang.Go"), command{name: "winget", args: []string{"install", "--id=GoLang.Go"}}},
		{"nix", packageManagers["nix-env"].install("go"), command{name: "nix-env", args: []string{"-iA", "nixpkgs.go"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !reflect.DeepEqual(tt.got, tt.want) {
				t.Fatalf("got %#v, want %#v", tt.got, tt.want)
			}
		})
	}
}
