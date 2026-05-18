# locales

This directory stores locale extensions used only by the `web-antd` application, such as dayjs locale setup, Ant Design Vue locale wiring, and app-scoped translations.

The application discovers app-scoped message bundles from `langs/<locale>/*.json`. The language switcher is hydrated from `GET /api/v1/i18n/runtime/locales`, so adding a built-in language should only require locale JSON resources plus host-side `i18n.locales` config metadata, not frontend TypeScript language-list edits.

## Runtime Metadata

`GET /api/v1/i18n/runtime/locales` is the source of truth for the language switcher, default locale, native names, and whether the switcher is enabled. Runtime text direction is fixed to `ltr`; language changes keep both `<html dir>` and Ant Design Vue `ConfigProvider.direction` on `ltr`.

Third-party locale wiring should be derived from the locale code convention and the generated locale loader keys where possible. Do not add a frontend language registry or per-language fallback map when a new built-in language is added.

## Runtime Bundle Cache

Runtime UI messages are loaded from `GET /api/v1/i18n/runtime/messages?lang=<locale>` and persisted in `localStorage` under `linapro:i18n:runtime:<locale>` for 7 days.

The cache stores `{etag, messages, savedAt}`. A fresh persisted bundle renders immediately, then the app refreshes in the background with `If-None-Match`; a `304 Not Modified` response keeps the current bundle unchanged.

## Request Errors And Visible Copy

Request error rendering must prefer backend `messageKey` plus `messageParams`, then fall back to the backend-rendered `message`, and only then use the request library fallback text. This keeps the active frontend language in control when structured backend errors are available.

User-visible page copy must use `$t` or runtime i18n messages. Do not place literal Chinese, English, or mixed-language labels directly in titles, form schemas, table columns, placeholders, buttons, empty states, toasts, notifications, or confirmation dialogs.

Run these checks after changing frontend runtime copy:

```bash
make i18n.check
```

`i18n.check` is intentionally strict and may still report existing backend or plugin findings while the active runtime-message governance cleanup is in progress.
