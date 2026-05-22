// Package skills manages repository-local symbolic links from supported AI
// coding agents' project skill paths to the canonical .agents/skills directory.
//
// This subpackage is part of the multi-resource agents framework. It owns the
// skills resource agent registry derived from vercel-labs/skills, the link
// planning and apply logic, and the unlink logic used by the agents.skills.link
// and agents.skills.unlink commands. It only operates inside the LinaPro
// repository root and never modifies HOME directories or system-global paths.
//
// Implementation uses Go standard library symlink primitives (os.Symlink,
// os.Readlink, os.Lstat, os.Remove, os.MkdirAll) combined with filepath.Rel
// to keep generated symlinks portable across Windows, Linux and macOS. Real
// directories and files are never automatically removed, even with FORCE=1.
package skills
