#!/usr/bin/env python3
"""Smoke tests for the LinaPro bootstrap installer scripts."""

from __future__ import annotations

import os
import subprocess
import sys
import tarfile
import tempfile
from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[3]
INSTALL_SH = REPO_ROOT / "hack" / "scripts" / "install" / "install.sh"
INSTALL_PS1 = REPO_ROOT / "hack" / "scripts" / "install" / "install.ps1"

EXPECTED_SHELL_OPTIONS = {
    "--repo",
    "--ref",
    "--dir",
    "--name",
    "--current-dir",
    "--force",
    "--help",
}
EXPECTED_POWERSHELL_OPTIONS = {
    "-Repo",
    "-Ref",
    "-Dir",
    "-Name",
    "-CurrentDir",
    "-Force",
    "-Help",
}


def assert_true(condition: bool, message: str) -> None:
    if not condition:
        raise AssertionError(message)


def create_fixture_archive(base_dir: Path) -> Path:
    source_root = base_dir / "linapro-main"
    source_root.mkdir(parents=True)
    (source_root / "README.md").write_text("# LinaPro test fixture\n", encoding="utf-8")
    (source_root / ".gitignore").write_text("temp/\n", encoding="utf-8")
    nested_dir = source_root / "apps" / "lina-core"
    nested_dir.mkdir(parents=True)
    (nested_dir / "main.go").write_text("package main\n", encoding="utf-8")

    archive_path = base_dir / "fixture.tar.gz"
    with tarfile.open(archive_path, "w:gz") as archive:
        archive.add(source_root, arcname=source_root.name)
    return archive_path


def run_install(
    *args: str,
    cwd: Path,
    archive_path: Path,
    stable_ref: str | None = "v0.0.0",
) -> subprocess.CompletedProcess[str]:
    env = os.environ.copy()
    env["LINAPRO_INSTALL_ARCHIVE_PATH"] = str(archive_path)
    if stable_ref is not None:
        env["LINAPRO_INSTALL_STABLE_REF"] = stable_ref
    return subprocess.run(
        [str(INSTALL_SH), *args],
        cwd=str(cwd),
        env=env,
        text=True,
        capture_output=True,
        check=False,
    )


def test_install_into_named_directory() -> None:
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_path = Path(temp_dir)
        archive_path = create_fixture_archive(temp_path)
        workspace = temp_path / "workspace"
        workspace.mkdir()
        target_dir = workspace / "demo-install"

        result = run_install("--dir", str(target_dir), cwd=workspace, archive_path=archive_path)
        assert_true(result.returncode == 0, f"install.sh failed:\nSTDOUT:\n{result.stdout}\nSTDERR:\n{result.stderr}")
        assert_true((target_dir / "README.md").exists(), "README.md should be installed into the target directory")
        assert_true((target_dir / ".gitignore").exists(), "hidden files should be copied into the target directory")
        assert_true("Project directory:" in result.stdout, "installer output should include the project directory")
        assert_true("Environment check:" in result.stdout, "installer output should include the environment check")


def test_install_into_current_directory() -> None:
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_path = Path(temp_dir)
        archive_path = create_fixture_archive(temp_path)
        workspace = temp_path / "workspace"
        workspace.mkdir()

        result = run_install("--current-dir", cwd=workspace, archive_path=archive_path)
        assert_true(result.returncode == 0, f"install.sh current-dir mode failed:\nSTDOUT:\n{result.stdout}\nSTDERR:\n{result.stderr}")
        assert_true((workspace / "README.md").exists(), "README.md should be copied into the current directory")
        assert_true((workspace / "apps" / "lina-core" / "main.go").exists(), "nested files should be copied into the current directory")


def test_default_invocation_uses_stable_ref_and_default_directory_name() -> None:
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_path = Path(temp_dir)
        archive_path = create_fixture_archive(temp_path)
        workspace = temp_path / "workspace"
        workspace.mkdir()
        target_dir = workspace / "linapro"
        resolved_target_dir = target_dir.resolve()

        result = run_install(cwd=workspace, archive_path=archive_path, stable_ref="v9.9.9")
        assert_true(result.returncode == 0, f"default install failed:\nSTDOUT:\n{result.stdout}\nSTDERR:\n{result.stderr}")
        assert_true(target_dir.exists(), "default invocation should install into ./linapro")
        assert_true((target_dir / "README.md").exists(), "default invocation should copy repository contents")
        assert_true("Resolved ref: v9.9.9" in result.stdout, "default invocation should report the resolved stable ref")
        assert_true(
            f"Target directory: {resolved_target_dir}" in result.stdout,
            "default invocation should report the resolved target directory",
        )


def test_reject_non_empty_target_without_force() -> None:
    with tempfile.TemporaryDirectory() as temp_dir:
        temp_path = Path(temp_dir)
        archive_path = create_fixture_archive(temp_path)
        workspace = temp_path / "workspace"
        workspace.mkdir()
        target_dir = workspace / "existing"
        target_dir.mkdir()
        (target_dir / "custom.txt").write_text("keep me\n", encoding="utf-8")

        result = run_install("--dir", str(target_dir), cwd=workspace, archive_path=archive_path)
        assert_true(result.returncode != 0, "install.sh should reject a non-empty target without --force")
        combined_output = f"{result.stdout}\n{result.stderr}"
        assert_true("not empty" in combined_output, "installer should explain why the target directory was rejected")


def test_option_contracts_match_documented_shape() -> None:
    help_result = subprocess.run(
        [str(INSTALL_SH), "--help"],
        cwd=str(REPO_ROOT),
        text=True,
        capture_output=True,
        check=False,
    )
    assert_true(help_result.returncode == 0, "install.sh --help should succeed")

    for option in EXPECTED_SHELL_OPTIONS:
        assert_true(option in help_result.stdout, f"shell installer help is missing option {option}")

    powershell_text = INSTALL_PS1.read_text(encoding="utf-8")
    for option in EXPECTED_POWERSHELL_OPTIONS:
        assert_true(option in powershell_text, f"PowerShell installer is missing parameter {option}")

    assert_true(
        "LINAPRO_INSTALL_ARCHIVE_PATH" in help_result.stdout and "LINAPRO_INSTALL_ARCHIVE_PATH" in powershell_text,
        "both installers should expose the same archive override contract",
    )
    assert_true(
        "LINAPRO_INSTALL_STABLE_REF" in help_result.stdout and "LINAPRO_INSTALL_STABLE_REF" in powershell_text,
        "both installers should expose the same stable ref override contract",
    )


def main() -> int:
    tests = [
        test_install_into_named_directory,
        test_install_into_current_directory,
        test_default_invocation_uses_stable_ref_and_default_directory_name,
        test_reject_non_empty_target_without_force,
        test_option_contracts_match_documented_shape,
    ]

    failures = []
    for test in tests:
        try:
            test()
            print(f"PASS {test.__name__}")
        except Exception as exc:  # noqa: BLE001
            failures.append((test.__name__, exc))
            print(f"FAIL {test.__name__}: {exc}", file=sys.stderr)

    if failures:
        print("\nTest failures:", file=sys.stderr)
        for name, exc in failures:
            print(f"- {name}: {exc}", file=sys.stderr)
        return 1

    return 0


if __name__ == "__main__":
    sys.exit(main())
