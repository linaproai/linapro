# Regression Feedback Localization UI

## Why

Manual regression found gaps in English localization, default seed display, dynamic plugin permission mounting, table and form layout, and high-risk action confirmation. This change groups those findings into a traceable regression-fix iteration with automated checks.

## What Changes

- Align monitoring defaults, built-in job projection, and fallback intervals to one minute.
- Remove remaining Chinese copy from framework-delivered English pages and seed displays.
- Make role display consistent between user management and role management.
- Mount dynamic plugin route permission buttons under the owning plugin menu.
- Improve menu tree expand interactions.
- Localize generated Unassigned department nodes and built-in config display.
- Improve English layout for post, dictionary, and service-monitoring tables/forms.
- Add confirmation for scheduled-job Run Now.
- Address follow-up feedback for logo depth, shell action rendering, theme preference, quick navigation, project cards, plugin demo protection, and built-in deletion protection.

## I18n Impact

This change updates runtime language resources, plugin resources, packed resources, frontend locale JSON, and related projection services. It also records that locale JSON should not expose markdown-only backticks in ordinary UI text.
