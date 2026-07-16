#!/bin/sh
set -eu

REPOSITORY="SlateNull/code-init"
INSTALL_DIR="${CODE_INIT_INSTALL_DIR:-${HOME}/.local/bin}"
VERSION="${CODE_INIT_VERSION:-latest}"

fail() {
	printf 'code-init installer: %s\n' "$1" >&2
	exit 1
}

command -v curl >/dev/null 2>&1 || fail "curl is required"

case "$(uname -s)" in
	Linux) os="linux" ;;
	Darwin) os="darwin" ;;
	*) fail "unsupported operating system; Windows users should run install.ps1" ;;
esac

case "$(uname -m)" in
	x86_64|amd64) arch="amd64" ;;
	aarch64|arm64) arch="arm64" ;;
	*) fail "unsupported architecture: $(uname -m)" ;;
esac

asset="code-init_${os}_${arch}"
if [ "$VERSION" = "latest" ]; then
	base_url="https://github.com/${REPOSITORY}/releases/latest/download"
else
	case "$VERSION" in
		v*) tag="$VERSION" ;;
		*) tag="v$VERSION" ;;
	esac
	base_url="https://github.com/${REPOSITORY}/releases/download/${tag}"
fi

tmp_dir="$(mktemp -d 2>/dev/null)" || fail "could not create a temporary directory"
trap 'rm -rf "$tmp_dir"' EXIT HUP INT TERM

printf 'Downloading %s for %s/%s...\n' "$VERSION" "$os" "$arch"
curl --fail --location --silent --show-error --retry 3 \
	"${base_url}/${asset}" --output "${tmp_dir}/${asset}" || fail "binary download failed; verify that the release and platform asset exist"
curl --fail --location --silent --show-error --retry 3 \
	"${base_url}/checksums.txt" --output "${tmp_dir}/checksums.txt" || fail "checksum download failed"

expected="$(awk -v name="$asset" '$2 == name { print $1 }' "${tmp_dir}/checksums.txt")"
[ -n "$expected" ] || fail "release checksum does not include ${asset}"

if command -v sha256sum >/dev/null 2>&1; then
	actual="$(sha256sum "${tmp_dir}/${asset}" | awk '{ print $1 }')"
elif command -v shasum >/dev/null 2>&1; then
	actual="$(shasum -a 256 "${tmp_dir}/${asset}" | awk '{ print $1 }')"
else
	fail "SHA-256 verification requires sha256sum or shasum"
fi
[ "$actual" = "$expected" ] || fail "checksum verification failed"

mkdir -p "$INSTALL_DIR" || fail "could not create ${INSTALL_DIR}"
chmod 755 "${tmp_dir}/${asset}"
mv "${tmp_dir}/${asset}" "${INSTALL_DIR}/code-init" || fail "could not install to ${INSTALL_DIR}"

printf 'Installed code-init to %s/code-init\n' "$INSTALL_DIR"
case ":${PATH}:" in
	*":${INSTALL_DIR}:"*) ;;
	*) printf 'Add %s to PATH, then open a new terminal:\n  export PATH="%s:$PATH"\n' "$INSTALL_DIR" "$INSTALL_DIR" ;;
esac
