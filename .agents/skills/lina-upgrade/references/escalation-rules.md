# Escalation Rules

The `lina-upgrade` skill must stop automatic execution and ask for human intervention when any of these conditions occur.

1. A conflict appears in any Tier 1 path.
2. A Tier 2 three-way merge has low AI confidence or semantic ambiguity.
3. A SQL migration may destroy user data, such as `DROP COLUMN` against a column that has existing values.
4. An e2e smoke test fails and automatic rollback or repair does not restore a verified state.
5. The user states that they modified a core file and upstream changed the same file.

The escalation report must include the path or command, the detected tier when applicable, the target version, and the exact reason automatic execution stopped.
