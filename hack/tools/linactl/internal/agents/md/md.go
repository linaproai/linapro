// Package md manages repository-local symbolic links from supported AI
// coding agents' project guide files (e.g. CLAUDE.md, GEMINI.md,
// .junie/guidelines.md) to the repo-root AGENTS.md project specification.
//
// This subpackage is part of the multi-resource agents framework. Unlike
// the skills and prompts resources, md bindings manage single-file
// symlinks (KindFile in common terms). The state machine, Status enum,
// conflict guarding and selector resolution are otherwise identical to
// the directory-kind resources, so most behavior comes from the common
// engine and only the kind discriminator distinguishes md from skills.
//
// Initial coverage:
//   - link agents create a private guide file symlinked to AGENTS.md
//     (e.g. CLAUDE.md -> AGENTS.md, GEMINI.md -> AGENTS.md).
//   - native agents already read AGENTS.md natively at the repo root and
//     are reported in status output without any filesystem mutation.
//   - rootCollision is not used: AGENTS.md itself lives at the repo root
//     and there is no symmetric collision case to guard against.
//
// Real files are never automatically removed, even with FORCE=1.
package md
