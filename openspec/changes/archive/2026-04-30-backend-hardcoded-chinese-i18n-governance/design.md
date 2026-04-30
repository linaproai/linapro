# Design

## Overview

The remediation separates backend Chinese string findings into actionable categories. Caller-visible failures are converted to structured `bizerr` codes; user-visible projections and deliverables use runtime i18n resources or structured fields; developer diagnostics become stable English messages; generated schema text is fixed at its source.

## Error Governance

Each affected host or plugin module defines reusable error codes in its own `*_code.go` file. Business call sites create or wrap those errors through `bizerr`, preserving `errorCode`, `messageKey`, `messageParams`, and English fallback text. Plugin business errors keep translation resources in the owning plugin's `manifest/i18n/<locale>/` directory.

## Projection and Export Governance

Backend-owned labels, export headers, status fallbacks, and runtime display reasons are resolved by request language. User-entered data remains unchanged. When a UI can render the text more safely, the backend returns structured codes and values instead of localized text.

## Developer Diagnostics

Internal plugin platform diagnostics use stable English text. If such diagnostics cross a public API or plugin boundary, the boundary wraps them as structured errors so callers do not depend on free-form text.

## Generated Schema Governance

Generated DAO, DO, and Entity files are not edited manually. If generated descriptions enter OpenAPI or user-visible docs, SQL comments or generation inputs are updated and code generation is rerun.

## Regression Guard

The scanner covers high-risk patterns such as Chinese `gerror`, `errors.New`, `fmt.Errorf`, message fields, reason fields, label maps, export headers, and plugin diagnostics. Allowlist entries must include file, category, reason, and scope.

## I18n Impact

This change directly affects runtime i18n resources and apidoc/schema text governance. Host-owned text stays in lina-core resources; plugin-owned text stays in plugin resources; English API source text remains in DTOs or generation inputs.
