# code-init

`code-init` creates a new project with a language's native tooling when available, or creates an empty source file when that language has no standard project initializer.

## Version

1.0.0

## Install

Requires Go 1.26 or newer.

```sh
go install .
```

Or build a local binary:

```sh
go build -o code-init .
```

## Usage

```text
code-init <language> <path> <name>
code-init --version
```

Examples:

```sh
code-init go ~/Projects my-api
code-init rust ~/Projects parser
code-init python ~/Projects hello
```

Go and Rust use `go mod init` and `cargo new`. Python has no canonical project initializer, so the last example creates `~/Projects/hello.py`.

The command refuses to overwrite an existing project or source file. Project names must be a single path component; this prevents accidental writes outside the requested destination.

## Languages

Ada, Bash, Bun, C, C++, Clojure, COBOL, Crystal, C#, CSS, D, Dart, Elixir, Elm, Erlang, F#, Fortran, Gleam, Go, Groovy, Haskell, HTML, Java, JavaScript, Julia, Kotlin, LaTeX, Lua, Nim, Node.js, Objective-C, Objective-C++, OCaml, Odin, Pascal, Perl, PHP, PowerShell, Python, R, Ruby, Rust, Scala, shell, Solidity, Swift, TypeScript, V, and Zig.

Common aliases such as `cpp`, `c#`, `golang`, `js`, `py`, `rs`, and `ts` are accepted.

## Platforms and package managers

When a native project tool is missing, `code-init` can install known packages using:

- Debian family: APT
- Fedora/RHEL family: DNF or YUM
- Arch family: Pacman
- openSUSE: Zypper
- Alpine: APK
- Void: XBPS
- Gentoo: Portage
- Solus: eopkg
- NixOS: nix-env
- macOS: Homebrew
- Windows: WinGet
- FreeBSD: pkg
- OpenBSD: pkg_add

Unknown Linux distributions are supported when one of these package-manager executables can be detected. If no known package is available, creation falls back to an empty language source file rather than failing. Installation may request administrator privileges and uses the configured operating-system repositories; review repository trust and network/cost policies before use.

## Development

```sh
gofmt -w *.go
go vet ./...
go test ./...
go test -race ./...
go build ./...
```

The implementation uses only the Go standard library.
