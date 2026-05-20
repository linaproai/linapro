// Package prompts manages repository-local symbolic links from supported AI
// coding agents' project commands/prompts directories to canonical source
// directories under .agents/prompts/.
//
// This subpackage is part of the multi-resource agents framework. Unlike
// the skills resource (where every binding shares a single source path
// .agents/skills), each prompts agent declares its own SourcePath
// explicitly because different agents might surface different prompt
// catalogs (the initial registry only links the OpenSpec /opsx slash
// commands, but additional sources can be added per agent later).
//
// All bindings are directory-kind symlinks managed via the resource-
// agnostic engine in linactl/internal/agents/common. Real directories and
// files are never automatically removed, even with FORCE=1.
package prompts
