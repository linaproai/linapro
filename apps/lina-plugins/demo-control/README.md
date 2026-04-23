# demo-control

`demo-control` is the official LinaPro source plugin for demo-environment read-only protection.

Enable this plugin by adding `demo-control` to the host `plugin.autoEnable` list when the target environment should run in read-only demo mode.

## Scope

This plugin owns:

- environment-level demo request guarding based on `HTTP Method`
- write-operation interception for host and plugin APIs under `/api/v1`
- the minimal session whitelist required for login and logout in demo mode
