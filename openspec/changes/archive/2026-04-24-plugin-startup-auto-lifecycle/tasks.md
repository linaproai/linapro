## 1. Config and startup flow

- [x] 1.1 Extend the host main config model, template examples, and config validation for `plugin.autoEnable` so startup auto-enable is driven by a plugin ID list
- [x] 1.2 Add a plugin startup bootstrap phase to the host startup flow and move it ahead of plugin route registration, plugin cron wiring, and dynamic bundle warm-up

## 2. Lifecycle bootstrap implementation

- [x] 2.1 Implement source-plugin auto-install and auto-enable execution driven by `plugin.autoEnable`, including shared-action protection on the cluster primary node
- [x] 2.2 Implement dynamic-plugin auto-install and auto-enable execution driven by `plugin.autoEnable`, reusing existing authorization snapshots plus the `desired_state/current_state` and targeted reconcile mechanisms
- [x] 2.3 Implement fail-fast behavior, convergence waiting, and enabled-snapshot refresh for auto-enabled plugins so later plugin wiring only happens after bootstrap finishes

## 3. Testing and verification

- [x] 3.1 Add tests for `plugin.autoEnable` parsing and invalid-config rejection
- [x] 3.2 Add source-plugin auto-enable bootstrap tests that cover discovered-state preservation, auto-install, and auto-enable paths
- [x] 3.3 Add dynamic-plugin auto-enable bootstrap tests that cover existing authorization snapshots, missing authorization snapshots, and convergence behavior in both single-node and cluster primary/follower flows

## 4. Documentation and operations guidance

- [x] 4.1 Update plugin technical docs and config guidance with `plugin.autoEnable` examples in the host main config file and the prerequisite for dynamic-plugin authorization snapshots

## Feedback

- [x] **FB-1**: Simplify plugin auto-enable config into the host main config file's `plugin.autoEnable` plugin ID list
- [x] **FB-2**: Change plugin install/enable status values in `catalog/status.go` to strong typed enums and clean up constants in that file that should not stay as loosely defined status strings
- [x] **FB-3**: Clearly label plugins managed by `plugin.autoEnable` in the plugin-management page and add warnings before disable/uninstall actions that say the change takes effect immediately but will be restored after restart unless the config is changed
