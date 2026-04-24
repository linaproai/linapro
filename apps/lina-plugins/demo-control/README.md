# demo-control

`demo-control` is the official LinaPro source plugin for demo-environment read-only protection.

Read-only demo mode becomes active whenever this plugin is installed and enabled. If you want the host to auto-enable it at startup, add `demo-control` to the host `plugin.autoEnable` list.

## Scope

This plugin owns:

- environment-level demo request guarding based on `HTTP Method`
- whole-system request interception on the host `/*` scope
- write-operation interception for host and plugin write endpoints across the system
- the minimal session whitelist required for login and logout in demo mode
