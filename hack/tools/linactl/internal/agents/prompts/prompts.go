// Package prompts manages repository-local symbolic links from supported AI
// coding agents' project commands/prompts directories to canonical source
// directory.
//
// This subpackage is part of the multi-resource agents framework. Unlike
// the skills resource (where every binding shares a single source path
// .agents/skills), prompts bindings bridge each agent's commands/prompts
// root to .agents/prompts so all prompt catalogs under that source root
// become visible without creating one symlink per catalog.
//
// All bindings are directory-kind symlinks managed via the resource-
// agnostic engine in linactl/internal/agents/common. Real directories and
// files are never automatically removed, even with FORCE=1.
package prompts
