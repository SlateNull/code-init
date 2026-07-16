# code-init

`code-init` creates projects with each language's native tooling when available. If a language has no standard project initializer—or a verified toolchain package is unavailable—it safely creates `<name>.<suffix>` instead.

## Install

No Go toolchain is required. Release binaries are available for Linux, macOS, and Windows on AMD64 and ARM64.

### Linux and macOS

Install the latest release with `curl`:

```sh
curl -fsSL https://raw.githubusercontent.com/SlateNull/code-init/main/install.sh | sh
```

The installer downloads the correct binary to `~/.local/bin`, verifies its SHA-256 checksum, and prints a PATH command if `~/.local/bin` is not already available.

Install to another directory:

```sh
curl -fsSL https://raw.githubusercontent.com/SlateNull/code-init/main/install.sh | CODE_INIT_INSTALL_DIR=/usr/local/bin sh
```

Installing to a protected system directory may require elevated permissions. Prefer the default user-local installation unless the machine is intentionally managed system-wide.

Install a specific version:

```sh
curl -fsSL https://raw.githubusercontent.com/SlateNull/code-init/main/install.sh | CODE_INIT_VERSION=1.0.0 sh
```

### Windows

Run in PowerShell. Windows includes both PowerShell and `curl.exe`:

```powershell
irm https://raw.githubusercontent.com/SlateNull/code-init/main/install.ps1 | iex
```

The installer downloads the correct executable to `%LOCALAPPDATA%\Programs\code-init`, verifies its SHA-256 checksum, and adds that directory to the user PATH. Open a new terminal after the first installation.

To install a specific version:

```powershell
$env:CODE_INIT_VERSION = "1.0.0"; irm https://raw.githubusercontent.com/SlateNull/code-init/main/install.ps1 | iex
```

### Manual download

You can also download a binary and `checksums.txt` from the [GitHub Releases page](https://github.com/SlateNull/code-init/releases). Verify the checksum, rename the executable to `code-init` (`code-init.exe` on Windows), and place it on your PATH.

### Build from source

If Go is already installed:

```sh
go install github.com/SlateNull/code-init@latest
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

On Windows, both separator styles are accepted:

```powershell
code-init go C:\Projects my-api
code-init python C:/Projects hello
```

The destination path and missing parent directories are created automatically. Go and Rust use `go mod init` and `cargo new`; Python has no canonical initializer, so the Python example creates `hello.py`. Existing projects and files are never overwritten, and project names cannot contain path separators.

## Supported languages

Ada, Bash, Bun, C, C++, Clojure, COBOL, Crystal, C#, CSS, D, Dart, Elixir, Elm, Erlang, F#, Fortran, Gleam, Go, Groovy, Haskell, HTML, Java, JavaScript, Julia, Kotlin, LaTeX, Lua, Nim, Node.js, Objective-C, Objective-C++, OCaml, Odin, Pascal, Perl, PHP, PowerShell, Python, R, Ruby, Rust, Scala, shell, Solidity, Swift, TypeScript, V, and Zig.

Common aliases such as `cpp`, `c#`, `golang`, `js`, `py`, `rs`, and `ts` are accepted.

## Platforms and package managers

When native project tooling is missing, `code-init` can install conservative, manager-specific packages through:

- Debian and Ubuntu: APT
- Fedora and RHEL family: DNF or YUM
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

Unknown Linux distributions are supported when a known package-manager executable is detected. Package mappings are intentionally omitted when a repository name is ambiguous, version-sensitive, or does not provide the exact executable used by the native initializer. In those cases, `code-init` creates the language source file instead of risking installation of unrelated software.

Package installation can require administrator approval, network access, and acceptance of repository or WinGet terms. On Windows, a newly installed tool may not appear in the current process PATH until a new terminal is opened.

## Releases

Pushing a signed or annotated `v*` tag triggers `.github/workflows/release.yml`. The workflow runs vet and race tests, cross-builds six binaries, creates `checksums.txt`, and publishes a GitHub Release using the repository's built-in `GITHUB_TOKEN`.

Create the first release:

```sh
git tag -a v1.0.0 -m "code-init 1.0.0"
git push origin v1.0.0
```

The one-command installers work only after that release workflow has successfully published the matching assets.

## Development

```sh
gofmt -w *.go
go vet ./...
go test ./...
go test -race ./...
go build ./...
```

The implementation uses only the Go standard library.

## License

[MIT](LICENSE)
