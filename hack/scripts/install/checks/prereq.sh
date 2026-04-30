#!/usr/bin/env bash
# Verifies development prerequisites before a LinaPro post-clone installation.

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
INSTALL_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
# shellcheck source=hack/scripts/install/lib/_common.sh
source "$INSTALL_DIR/lib/_common.sh"

critical_failures=0
warnings=0
detected_os="$(detect_os)"

# add_critical records a missing tool that prevents the installer from running.
add_critical() {
  critical_failures=$((critical_failures + 1))
  log_error "$1"
}

# add_warning records a non-blocking environment issue.
add_warning() {
  warnings=$((warnings + 1))
  log_warn "$1"
}

# command_version returns the version string produced by a command.
command_version() {
  local command_name="$1"
  case "$command_name" in
    go)
      go version | awk '{print $3}' | sed 's/^go//'
      ;;
    node)
      node --version | sed 's/^v//'
      ;;
    pnpm)
      pnpm --version | awk 'NR == 1 {print $1}'
      ;;
    *)
      "$command_name" --version | awk 'NR == 1 {print $NF}'
      ;;
  esac
}

# check_versioned_command verifies a command exists and meets the minimum version.
check_versioned_command() {
  local command_name="$1"
  local minimum="$2"
  local hint="$3"
  local actual

  if ! command -v "$command_name" >/dev/null 2>&1; then
    add_critical "Missing $command_name. Install hint: $hint"
    return
  fi

  actual="$(command_version "$command_name" || true)"
  if [ -z "$actual" ]; then
    add_critical "Could not read $command_name version. Install hint: $hint"
    return
  fi
  if ! version_ge "$actual" "$minimum"; then
    add_critical "$command_name $actual is too old. Required >= $minimum. Install hint: $hint"
    return
  fi
  log_info "$command_name $actual detected."
}

# check_plain_command verifies a command exists.
check_plain_command() {
  local command_name="$1"
  local hint="$2"
  local mode="${3:-critical}"

  if command -v "$command_name" >/dev/null 2>&1; then
    log_info "$command_name detected."
    return
  fi

  if [ "$mode" = "warning" ]; then
    add_warning "Missing $command_name. Install hint: $hint"
    return
  fi
  add_critical "Missing $command_name. Install hint: $hint"
}

case "$detected_os" in
  macos)
    go_hint="brew install go"
    node_hint="brew install node"
    pnpm_hint="npm i -g pnpm"
    git_hint="xcode-select --install or brew install git"
    make_hint="xcode-select --install or brew install make"
    mysql_hint="brew install mysql-client"
    make_mode="critical"
    ;;
  linux)
    go_hint="install Go from https://go.dev/doc/install or your distribution packages"
    node_hint="install Node.js 20+ from https://nodejs.org or NodeSource packages"
    pnpm_hint="npm i -g pnpm"
    git_hint="apt-get install -y git or yum install -y git"
    make_hint="apt-get install -y make or yum install -y make"
    mysql_hint="apt-get install -y default-mysql-client or yum install -y mysql"
    make_mode="critical"
    ;;
  windows)
    go_hint="winget install GoLang.Go or scoop install go"
    node_hint="winget install OpenJS.NodeJS or scoop install nodejs"
    pnpm_hint="npm i -g pnpm or scoop install pnpm"
    git_hint="install Git for Windows and run this script from Git Bash"
    make_hint="scoop install make, choco install make, or install mingw32-make"
    mysql_hint="winget install Oracle.MySQL or choco install mysql-cli"
    make_mode="warning"
    ;;
  *)
    go_hint="install Go from https://go.dev/doc/install"
    node_hint="install Node.js 20+ from https://nodejs.org"
    pnpm_hint="npm i -g pnpm"
    git_hint="install git"
    make_hint="install make"
    mysql_hint="install mysql client"
    make_mode="critical"
    ;;
esac

check_versioned_command go 1.22 "$go_hint"
check_versioned_command node 20.0.0 "$node_hint"
check_versioned_command pnpm 8.0.0 "$pnpm_hint"
check_plain_command git "$git_hint"
check_plain_command make "$make_hint" "$make_mode"

if command -v mysql >/dev/null 2>&1; then
  if mysql --version >/dev/null 2>&1; then
    log_info "mysql client detected."
  else
    add_warning "mysql client exists but did not return a version. Install hint: $mysql_hint"
  fi
else
  add_warning "Missing mysql client. Install hint: $mysql_hint"
fi

for port in 5666 8080; do
  if is_port_in_use "$port"; then
    add_warning "Port $port is already in use."
  else
    log_info "Port $port is available."
  fi
done

if [ "$critical_failures" -gt 0 ]; then
  exit 1
fi
if [ "$warnings" -gt 0 ]; then
  exit 2
fi
exit 0
