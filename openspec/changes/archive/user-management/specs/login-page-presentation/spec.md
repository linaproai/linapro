## ADDED Requirements

### Requirement: Login Page Presents by Resolution Strategy
Login page SHALL present different UI based on current effective resolution strategy:
- subdomain strategy: No tenant input box (auto-resolved from URL).
- header strategy: Show tenant code input box.
- jwt/session/default strategy: No tenant input, after login select-tenant decides.

When 1:N user login returns pre_token + tenant list, frontend switches to tenant dropdown selector; account password form and tenant selector must not display simultaneously.

### Requirement: Prompt Mode Visibility
When ambiguous behavior = `prompt`, selector SHALL provide clear guidance; `reject` mode shows error only.

### Requirement: Multi-Tenant Disabled Degrades
When multi-tenant not enabled, login page SHALL only show traditional username/password, login success directly enters workbench.

### Requirement: Tenant Transition Loading State
After selecting tenant, prioritize showing "entering tenant" transition interface with loading indicator.
