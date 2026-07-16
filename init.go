package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const version = "1.0.0"

type languageSpec struct {
	executable string
	suffix     string
	create     func(target, name string) command
}

type command struct {
	name string
	args []string
	dir  string
}

type packageManager struct {
	install  func(string) command
	packages map[string]string
}

var languages = map[string]languageSpec{
	"ada":           {"gnat", "adb", nil},
	"bash":          {"bash", "sh", nil},
	"bun":           {"bun", "js", inTarget("bun", "init", "-y")},
	"c":             {"gcc", "c", nil},
	"c++":           {"g++", "cpp", nil},
	"clojure":       {"clojure", "clj", nil},
	"cobol":         {"cobc", "cob", nil},
	"crystal":       {"crystal", "cr", atParent("crystal", "init", "app")},
	"csharp":        {"dotnet", "cs", dotnetProject("console")},
	"css":           {"", "css", nil},
	"d":             {"dmd", "d", nil},
	"dart":          {"dart", "dart", atParent("dart", "create")},
	"elixir":        {"mix", "ex", atParent("mix", "new")},
	"elm":           {"elm", "elm", inTarget("elm", "init")},
	"erlang":        {"erl", "erl", nil},
	"f#":            {"dotnet", "fs", dotnetProject("console", "--language", "F#")},
	"fortran":       {"gfortran", "f90", nil},
	"gleam":         {"gleam", "gleam", atParent("gleam", "new")},
	"go":            {"go", "go", goProject},
	"groovy":        {"groovy", "groovy", nil},
	"haskell":       {"cabal", "hs", cabalProject},
	"html":          {"", "html", nil},
	"java":          {"java", "java", nil},
	"javascript":    {"npm", "js", inTarget("npm", "init", "-y")},
	"julia":         {"julia", "jl", nil},
	"kotlin":        {"kotlinc", "kt", nil},
	"latex":         {"pdflatex", "tex", nil},
	"lua":           {"lua", "lua", nil},
	"nim":           {"nim", "nim", nil},
	"node":          {"npm", "js", inTarget("npm", "init", "-y")},
	"objective-c":   {"clang", "m", nil},
	"objective-c++": {"clang++", "mm", nil},
	"ocaml":         {"dune", "ml", atParent("dune", "init", "project")},
	"odin":          {"odin", "odin", nil},
	"pascal":        {"fpc", "pas", nil},
	"perl":          {"perl", "pl", nil},
	"php":           {"php", "php", nil},
	"powershell":    {"pwsh", "ps1", nil},
	"python":        {"python3", "py", nil},
	"r":             {"R", "R", nil},
	"ruby":          {"ruby", "rb", nil},
	"rust":          {"cargo", "rs", rustProject},
	"scala":         {"scala", "scala", nil},
	"shell":         {"sh", "sh", nil},
	"solidity":      {"forge", "sol", atParent("forge", "init")},
	"swift":         {"swiftc", "swift", swiftProject},
	"typescript":    {"npm", "ts", typescriptProject},
	"v":             {"v", "v", nil},
	"zig":           {"zig", "zig", inTarget("zig", "init")},
}

var aptPackages = map[string]string{
	"crystal": "crystal", "elixir": "elixir", "elm": "elm-compiler", "go": "golang-go",
	"haskell": "cabal-install", "javascript": "npm", "node": "npm", "ocaml": "dune",
	"rust": "cargo", "typescript": "npm",
}

var dnfPackages = map[string]string{
	"elixir": "elixir", "go": "golang", "haskell": "cabal-install", "javascript": "nodejs22-npm",
	"node": "nodejs22-npm", "ocaml": "ocaml-dune", "rust": "cargo", "swift": "swift-lang",
	"typescript": "nodejs22-npm", "zig": "zig",
}

var yumPackages = map[string]string{
	"go": "golang", "javascript": "npm", "node": "npm", "typescript": "npm",
}

var pacmanPackages = map[string]string{
	"bun": "bun", "crystal": "crystal", "csharp": "dotnet-sdk", "dart": "dart", "elixir": "elixir",
	"elm": "elm", "f#": "dotnet-sdk", "gleam": "gleam", "go": "go", "haskell": "cabal-install",
	"javascript": "npm", "node": "npm", "ocaml": "dune", "rust": "rust", "solidity": "foundry",
	"swift": "swift-language", "typescript": "npm", "zig": "zig",
}

var zypperPackages = map[string]string{
	"elixir": "elixir", "go": "go", "haskell": "cabal-install", "javascript": "npm",
	"node": "npm", "ocaml": "ocaml-dune", "rust": "cargo", "typescript": "npm",
}

var apkPackages = map[string]string{
	"crystal": "crystal", "elixir": "elixir", "elm": "elm", "go": "go", "haskell": "cabal",
	"javascript": "npm", "node": "npm", "ocaml": "dune", "rust": "cargo", "typescript": "npm", "zig": "zig",
}

var xbpsPackages = map[string]string{
	"crystal": "crystal", "elixir": "elixir", "elm": "elm", "go": "go", "haskell": "cabal-install",
	"javascript": "nodejs", "node": "nodejs", "ocaml": "dune", "rust": "rust", "typescript": "nodejs", "zig": "zig",
}

var emergePackages = map[string]string{
	"crystal": "dev-lang/crystal", "dart": "dev-lang/dart-sdk", "elixir": "dev-lang/elixir",
	"go": "dev-lang/go", "haskell": "dev-haskell/cabal-install", "javascript": "net-libs/nodejs",
	"node": "net-libs/nodejs", "ocaml": "dev-ml/dune", "rust": "dev-lang/rust",
	"typescript": "net-libs/nodejs", "zig": "dev-lang/zig",
}

var eopkgPackages = map[string]string{
	"go": "golang", "javascript": "nodejs", "node": "nodejs", "rust": "rust", "typescript": "nodejs",
}

var nixPackages = map[string]string{
	"bun": "bun", "crystal": "crystal", "csharp": "dotnet-sdk", "dart": "dart", "elixir": "elixir",
	"elm": "elmPackages.elm", "f#": "dotnet-sdk", "gleam": "gleam", "go": "go",
	"haskell": "haskellPackages.cabal-install", "javascript": "nodejs", "node": "nodejs",
	"ocaml": "dune_3", "rust": "cargo", "typescript": "nodejs", "zig": "zig",
}

var brewPackages = map[string]string{
	"bun": "bun", "crystal": "crystal", "dart": "dart-lang/dart/dart", "elixir": "elixir",
	"elm": "elm", "gleam": "gleam", "go": "go", "haskell": "cabal-install", "javascript": "node",
	"node": "node", "ocaml": "dune", "rust": "rust", "typescript": "node", "zig": "zig",
}

var wingetPackages = map[string]string{
	"bun": "Oven-sh.Bun", "csharp": "Microsoft.DotNet.SDK.8", "dart": "Google.DartSDK",
	"f#": "Microsoft.DotNet.SDK.8", "go": "GoLang.Go", "javascript": "OpenJS.NodeJS.LTS",
	"node": "OpenJS.NodeJS.LTS", "rust": "Rustlang.Rustup", "typescript": "OpenJS.NodeJS.LTS", "zig": "zig.zig",
}

var freeBSDPackages = map[string]string{
	"bun": "bun", "crystal": "crystal", "elixir": "elixir", "elm": "elm", "gleam": "gleam",
	"go": "go", "haskell": "hs-cabal-install", "ocaml": "ocaml-dune", "rust": "rust", "zig": "zig",
}

var openBSDPackages = map[string]string{
	"elixir": "elixir", "go": "go", "haskell": "cabal-install", "javascript": "npm",
	"node": "npm", "ocaml": "dune", "rust": "rust", "typescript": "npm",
}

var packageManagers = map[string]packageManager{
	"apt-get":      {sudoInstall("apt-get", "install", "-y"), aptPackages},
	"apt":          {sudoInstall("apt", "install", "-y"), aptPackages},
	"dnf":          {sudoInstall("dnf", "install", "-y"), dnfPackages},
	"yum":          {sudoInstall("yum", "install", "-y"), yumPackages},
	"pacman":       {sudoInstall("pacman", "-S", "--needed", "--noconfirm"), pacmanPackages},
	"zypper":       {sudoInstall("zypper", "--non-interactive", "install"), zypperPackages},
	"apk":          {sudoInstall("apk", "add"), apkPackages},
	"xbps-install": {sudoInstall("xbps-install", "-Sy"), xbpsPackages},
	"emerge":       {sudoInstall("emerge", "--ask=n"), emergePackages},
	"eopkg":        {sudoInstall("eopkg", "install", "-y"), eopkgPackages},
	"nix-env":      {directInstall("nix-env", "-iA", "nixpkgs."), nixPackages},
	"brew":         {directInstall("brew", "install"), brewPackages},
	"winget":       {directInstall("winget", "install", "--id="), wingetPackages},
	"pkg":          {sudoInstall("pkg", "install", "-y"), freeBSDPackages},
	"pkg_add":      {sudoInstall("pkg_add"), openBSDPackages},
}

var distroManagers = map[string]string{
	"alpine": "apk", "almalinux": "dnf", "amazon": "dnf", "arch": "pacman", "artix": "pacman",
	"cachyos": "pacman", "centos": "dnf", "chimera": "apk", "debian": "apt-get", "devuan": "apt-get",
	"elementary": "apt-get", "endeavouros": "pacman", "fedora": "dnf", "garuda": "pacman", "gentoo": "emerge",
	"kali": "apt-get", "linuxmint": "apt-get", "mageia": "dnf", "manjaro": "pacman", "mx": "apt-get",
	"nixos": "nix-env", "opensuse": "zypper", "opensuse-leap": "zypper", "opensuse-tumbleweed": "zypper",
	"oracle": "dnf", "pop": "apt-get", "raspbian": "apt-get", "rhel": "dnf", "rocky": "dnf",
	"solus": "eopkg", "ubuntu": "apt-get", "void": "xbps-install", "zorin": "apt-get",
}

var lookPath = exec.LookPath
var runCommand = run

func main() {
	if err := execute(os.Args[1:], os.Stdout); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

func execute(args []string, out io.Writer) error {
	if len(args) == 1 && (args[0] == "--version" || args[0] == "-v") {
		fmt.Fprintln(out, "code-init", version)
		return nil
	}
	if len(args) != 3 {
		return errors.New("usage: code-init <language> <path> <name>")
	}

	language := normalizeLanguage(args[0])
	spec, ok := languages[language]
	if !ok {
		return fmt.Errorf("unsupported language %q (supported: %s)", args[0], strings.Join(languageNames(), ", "))
	}
	name, err := validateName(args[2])
	if err != nil {
		return err
	}
	base, err := normalizePath(args[1])
	if err != nil {
		return err
	}
	if err := os.MkdirAll(base, 0o755); err != nil {
		return fmt.Errorf("create destination: %w", err)
	}

	if spec.create != nil {
		if _, err := lookPath(spec.executable); err != nil {
			manager := detectPackageManager()
			if manager != "" {
				if err := installLanguage(manager, language, out); err != nil {
					return err
				}
			}
		}
		if _, err := lookPath(spec.executable); err == nil {
			target := filepath.Join(base, name)
			if _, err := os.Stat(target); err == nil {
				return fmt.Errorf("destination already exists: %s", target)
			} else if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("inspect destination: %w", err)
			}
			cmd := spec.create(target, name)
			if cmd.dir != "" {
				if err := os.MkdirAll(cmd.dir, 0o755); err != nil {
					return fmt.Errorf("create project directory: %w", err)
				}
			}
			if err := runCommand(cmd); err != nil {
				return fmt.Errorf("create %s project: %w", language, err)
			}
			fmt.Fprintf(out, "Created %s project at %s\n", language, target)
			return nil
		}
	}

	file := filepath.Join(base, name+"."+spec.suffix)
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o644)
	if err != nil {
		if errors.Is(err, os.ErrExist) {
			return fmt.Errorf("destination already exists: %s", file)
		}
		return fmt.Errorf("create source file: %w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("close source file: %w", err)
	}
	fmt.Fprintf(out, "Created %s\n", file)
	return nil
}

func normalizePath(value string) (string, error) {
	value = filepath.FromSlash(strings.ReplaceAll(strings.TrimSpace(value), `\`, "/"))
	path, err := filepath.Abs(value)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	return filepath.Clean(path), nil
}

func normalizeLanguage(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	aliases := map[string]string{"cpp": "c++", "cxx": "c++", "cs": "csharp", "c#": "csharp", "fs": "f#", "golang": "go", "js": "javascript", "objc": "objective-c", "py": "python", "rs": "rust", "sh": "shell", "ts": "typescript"}
	if canonical, ok := aliases[value]; ok {
		return canonical
	}
	return value
}

func validateName(name string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" || name == "." || name == ".." || filepath.Base(name) != name || strings.ContainsAny(name, `/\\`) {
		return "", fmt.Errorf("invalid project name %q: use a single file name without path separators", name)
	}
	return name, nil
}

func languageNames() []string {
	names := make([]string, 0, len(languages))
	for name := range languages {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func detectPackageManager() string {
	if runtime.GOOS == "windows" {
		if _, err := lookPath("winget"); err == nil {
			return "winget"
		}
		return ""
	}
	if runtime.GOOS == "darwin" {
		if _, err := lookPath("brew"); err == nil {
			return "brew"
		}
		return ""
	}
	if runtime.GOOS == "freebsd" {
		return "pkg"
	}
	if runtime.GOOS == "openbsd" {
		return "pkg_add"
	}

	if id, _, err := readOSRelease("/etc/os-release"); err == nil {
		if manager := distroManagers[id]; manager != "" {
			if _, err := lookPath(manager); err == nil {
				return manager
			}
		}
	}
	for _, manager := range []string{"apt-get", "apt", "dnf", "yum", "pacman", "zypper", "apk", "xbps-install", "emerge", "eopkg", "nix-env"} {
		if _, err := lookPath(manager); err == nil {
			return manager
		}
	}
	return ""
}

func readOSRelease(path string) (string, string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	defer file.Close()
	var id, pretty string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		key, value, ok := strings.Cut(scanner.Text(), "=")
		if !ok {
			continue
		}
		value = strings.Trim(strings.TrimSpace(value), `"'`)
		switch key {
		case "ID":
			id = strings.ToLower(value)
		case "PRETTY_NAME":
			pretty = value
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", err
	}
	return id, pretty, nil
}

func installLanguage(managerName, language string, out io.Writer) error {
	manager, ok := packageManagers[managerName]
	if !ok {
		return nil
	}
	pkg := manager.packages[language]
	if pkg == "" {
		return nil
	}
	fmt.Fprintf(out, "Installing %s with %s...\n", language, managerName)
	if err := runCommand(manager.install(pkg)); err != nil {
		return fmt.Errorf("install %s with %s: %w", language, managerName, err)
	}
	return nil
}

func run(c command) error {
	cmd := exec.Command(c.name, c.args...)
	cmd.Dir, cmd.Stdin, cmd.Stdout, cmd.Stderr = c.dir, os.Stdin, os.Stdout, os.Stderr
	return cmd.Run()
}

func atParent(tool string, prefix ...string) func(string, string) command {
	return func(target, name string) command {
		args := append(append([]string{}, prefix...), name)
		return command{name: tool, args: args, dir: filepath.Dir(target)}
	}
}

func inTarget(tool string, args ...string) func(string, string) command {
	return func(target, _ string) command { return command{name: tool, args: args, dir: target} }
}

func dotnetProject(template string, extra ...string) func(string, string) command {
	return func(target, name string) command {
		args := []string{"new", template, "--name", name, "--output", target}
		args = append(args, extra...)
		return command{name: "dotnet", args: args}
	}
}

func goProject(target, name string) command {
	return command{name: "go", args: []string{"mod", "init", name}, dir: target}
}

func cabalProject(target, name string) command {
	return command{name: "cabal", args: []string{"init", "--non-interactive", "--package-name", name}, dir: target}
}

func swiftProject(target, name string) command {
	return command{name: "swift", args: []string{"package", "init", "--type", "executable", "--name", name}, dir: target}
}

func rustProject(target, name string) command {
	return command{name: "cargo", args: []string{"new", "--name", name, target}}
}

func typescriptProject(target, _ string) command {
	return command{name: "npx", args: []string{"tsc", "--init"}, dir: target}
}

func sudoInstall(program string, args ...string) func(string) command {
	return func(pkg string) command {
		all := append([]string{program}, args...)
		all = append(all, pkg)
		if runtime.GOOS == "windows" {
			return command{name: program, args: append(args, pkg)}
		}
		return command{name: "sudo", args: all}
	}
}

func directInstall(program string, args ...string) func(string) command {
	return func(pkg string) command {
		copied := append([]string{}, args...)
		if len(copied) > 0 && (strings.HasSuffix(copied[len(copied)-1], ".") || strings.HasSuffix(copied[len(copied)-1], "=")) {
			copied[len(copied)-1] += pkg
		} else {
			copied = append(copied, pkg)
		}
		return command{name: program, args: copied}
	}
}
