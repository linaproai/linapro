const pluginID = "plugin-demo-dynamic";
const apiBasePath = `/api/v1/extensions/${pluginID}`;
const defaultRecordPageSize = 10;
const standaloneI18nStoragePrefix = `linapro:${pluginID}:standalone-i18n:`;

const standaloneMessageKeys = {
  badge: "plugin.plugin-demo-dynamic.page.standalone.badge",
  card1Body: "plugin.plugin-demo-dynamic.page.standalone.card1Body",
  card1Title: "plugin.plugin-demo-dynamic.page.standalone.card1Title",
  card2Body: "plugin.plugin-demo-dynamic.page.standalone.card2Body",
  card2Title: "plugin.plugin-demo-dynamic.page.standalone.card2Title",
  footer: "plugin.plugin-demo-dynamic.page.standalone.footer",
  heroTitle: "plugin.plugin-demo-dynamic.page.standalone.heroTitle",
  lead: "plugin.plugin-demo-dynamic.page.standalone.lead",
  summary1Body: "plugin.plugin-demo-dynamic.page.standalone.summary1Body",
  summary1Title: "plugin.plugin-demo-dynamic.page.standalone.summary1Title",
  summary2Body: "plugin.plugin-demo-dynamic.page.standalone.summary2Body",
  summary2Title: "plugin.plugin-demo-dynamic.page.standalone.summary2Title",
  summary3Body: "plugin.plugin-demo-dynamic.page.standalone.summary3Body",
  summary3Title: "plugin.plugin-demo-dynamic.page.standalone.summary3Title",
  summaryTitle: "plugin.plugin-demo-dynamic.page.standalone.summaryTitle",
  title: "plugin.plugin-demo-dynamic.page.standalone.title",
};

const hostStyleId = "plugin-demo-dynamic-mount-style";

function translate(context, key, fallback) {
  if (context && typeof context.t === "function") {
    const translated = context.t(key, fallback);
    if (translated) {
      return translated;
    }
  }
  return fallback;
}

function formatTemplate(template, parameters = {}) {
  return template.replace(/\{(\w+)\}/g, (_match, key) => {
    const value = parameters[key];
    return value == null ? "" : String(value);
  });
}

function buildStandaloneI18nStorageKey() {
  if (window.crypto && typeof window.crypto.randomUUID === "function") {
    return `${standaloneI18nStoragePrefix}${window.crypto.randomUUID()}`;
  }
  return `${standaloneI18nStoragePrefix}${Date.now()}-${Math.random()
    .toString(36)
    .slice(2)}`;
}

function collectStandaloneRuntimeMessages(context) {
  const messages = {};
  for (const [targetKey, runtimeKey] of Object.entries(standaloneMessageKeys)) {
    const translated = translate(context, runtimeKey, "");
    if (translated) {
      messages[targetKey] = translated;
    }
  }
  return messages;
}

function persistStandaloneRuntimeMessages(context, locale) {
  const messages = collectStandaloneRuntimeMessages(context);
  if (Object.keys(messages).length === 0) {
    return "";
  }

  try {
    const storage = window.localStorage;
    if (!storage) {
      return "";
    }
    const storageKey = buildStandaloneI18nStorageKey();
    storage.setItem(
      storageKey,
      JSON.stringify({
        locale,
        messages,
      }),
    );
    return storageKey;
  } catch (_error) {
    return "";
  }
}

function resolveSupportedLocale(locale) {
  const normalized = String(locale || "").trim();
  if (!normalized) {
    return "en-US";
  }
  const lowerLocale = normalized.toLowerCase();
  if (
    lowerLocale === "zh" ||
    lowerLocale === "zh-cn" ||
    lowerLocale.startsWith("zh-hans")
  ) {
    return "zh-CN";
  }
  if (lowerLocale.startsWith("en")) {
    return "en-US";
  }
  return normalized;
}

function buildPageCopy(context) {
  const t = (key, fallback) => translate(context, key, fallback);

  const paginationSummaryTemplate = t(
    "plugin.plugin-demo-dynamic.page.pagination.summary",
    "Page {page}/{pages}, showing {start}-{end} of {total}",
  );
  const createdAtTemplate = t(
    "plugin.plugin-demo-dynamic.page.table.createdAt",
    "Created at: {value}",
  );
  const pendingAttachmentTemplate = t(
    "plugin.plugin-demo-dynamic.page.modal.pendingAttachment",
    "Pending attachment: {name}",
  );
  const currentAttachmentTemplate = t(
    "plugin.plugin-demo-dynamic.page.modal.currentAttachment",
    "Current attachment: {name}",
  );
  const deleteConfirmTemplate = t(
    "plugin.plugin-demo-dynamic.page.message.deleteRecordConfirm",
    'Delete record "{title}"?',
  );

  return {
    actions: {
      addRecord: t(
        "plugin.plugin-demo-dynamic.page.action.addRecord",
        "Add Record",
      ),
      cancel: t("plugin.plugin-demo-dynamic.page.action.cancel", "Cancel"),
      createRecord: t(
        "plugin.plugin-demo-dynamic.page.action.createRecord",
        "Create Record",
      ),
      deleteRecord: t(
        "plugin.plugin-demo-dynamic.page.action.deleteRecord",
        "Delete",
      ),
      downloadAttachment: t(
        "plugin.plugin-demo-dynamic.page.action.downloadAttachment",
        "Download Attachment",
      ),
      editRecord: t(
        "plugin.plugin-demo-dynamic.page.action.editRecord",
        "Edit",
      ),
      nextPage: t("plugin.plugin-demo-dynamic.page.action.nextPage", "Next"),
      previousPage: t(
        "plugin.plugin-demo-dynamic.page.action.previousPage",
        "Previous",
      ),
      reloadList: t(
        "plugin.plugin-demo-dynamic.page.action.reloadList",
        "Reload List",
      ),
      saveEdit: t(
        "plugin.plugin-demo-dynamic.page.action.saveEdit",
        "Save Changes",
      ),
      savePending: t(
        "plugin.plugin-demo-dynamic.page.action.savePending",
        "Saving...",
      ),
    },
    badge: t("plugin.plugin-demo-dynamic.page.badge", "WASM Plugin Demo"),
    emptyText: t(
      "plugin.plugin-demo-dynamic.page.emptyText",
      "Only the install-SQL seed record exists right now. You can still create, edit, or delete custom records.",
    ),
    emptyTitle: t(
      "plugin.plugin-demo-dynamic.page.emptyTitle",
      "No demo records yet",
    ),
    featureItems: [
      {
        description: t(
          "plugin.plugin-demo-dynamic.page.feature.integration.description",
          "The host shell imports the page entry dynamically and keeps using the host login session for backend dynamic routes.",
        ),
        label: t(
          "plugin.plugin-demo-dynamic.page.feature.integration.label",
          "Integration",
        ),
        value: t(
          "plugin.plugin-demo-dynamic.page.feature.integration.value",
          "Host Embedded Mount",
        ),
      },
      {
        description: t(
          "plugin.plugin-demo-dynamic.page.feature.data.description",
          "Install-time SQL creates the plugin-owned business table so the page can handle CRUD and attachment downloads directly.",
        ),
        label: t(
          "plugin.plugin-demo-dynamic.page.feature.data.label",
          "Data Sample",
        ),
        value: t(
          "plugin.plugin-demo-dynamic.page.feature.data.value",
          "Install SQL + CRUD",
        ),
      },
      {
        description: t(
          "plugin.plugin-demo-dynamic.page.feature.lifecycle.description",
          "Disable keeps data intact, while uninstall lets the host decide whether data and files are cleaned together.",
        ),
        label: t(
          "plugin.plugin-demo-dynamic.page.feature.lifecycle.label",
          "Lifecycle",
        ),
        value: t(
          "plugin.plugin-demo-dynamic.page.feature.lifecycle.value",
          "Optional Cleanup",
        ),
      },
    ],
    gridTitle: t("plugin.plugin-demo-dynamic.page.gridTitle", "Demo Records"),
    metrics: [
      {
        description: t(
          "plugin.plugin-demo-dynamic.page.metric.loader.description",
          "Loaded and mounted by the host shell at runtime",
        ),
        label: t(
          "plugin.plugin-demo-dynamic.page.metric.loader.label",
          "Dynamic Loader",
        ),
        value: t(
          "plugin.plugin-demo-dynamic.page.metric.loader.value",
          "Mounted by the host shell at runtime",
        ),
      },
      {
        description: t(
          "plugin.plugin-demo-dynamic.page.metric.schema.description",
          "Install-time SQL creates plugin-owned sample data",
        ),
        label: t(
          "plugin.plugin-demo-dynamic.page.metric.schema.label",
          "SQL Bootstrap",
        ),
        value: t(
          "plugin.plugin-demo-dynamic.page.metric.schema.value",
          "Install-time SQL bootstraps plugin-owned sample data",
        ),
      },
      {
        description: t(
          "plugin.plugin-demo-dynamic.page.metric.storage.description",
          "Demo records can attach files stored in the plugin namespace",
        ),
        label: t(
          "plugin.plugin-demo-dynamic.page.metric.storage.label",
          "File Storage",
        ),
        value: t(
          "plugin.plugin-demo-dynamic.page.metric.storage.value",
          "Records can attach files in the plugin namespace",
        ),
      },
      {
        description: t(
          "plugin.plugin-demo-dynamic.page.metric.lifecycle.description",
          "Uninstall can optionally keep or remove data and files together",
        ),
        label: t(
          "plugin.plugin-demo-dynamic.page.metric.lifecycle.label",
          "Lifecycle Cleanup",
        ),
        value: t(
          "plugin.plugin-demo-dynamic.page.metric.lifecycle.value",
          "Optional data and file cleanup",
        ),
      },
    ],
    messages: {
      deleteConfirm: (title) =>
        formatTemplate(deleteConfirmTemplate, {
          title,
        }),
      downloadFailed: t(
        "plugin.plugin-demo-dynamic.page.message.downloadFailed",
        "Failed to download the attachment",
      ),
      downloadStarted: t(
        "plugin.plugin-demo-dynamic.page.message.downloadStarted",
        "Attachment download started",
      ),
      fetchRecordsFailed: t(
        "plugin.plugin-demo-dynamic.page.message.fetchRecordsFailed",
        "Failed to load demo records",
      ),
      fileReadFailed: t(
        "plugin.plugin-demo-dynamic.page.message.fileReadFailed",
        "Failed to read the attachment",
      ),
      recordCreated: t(
        "plugin.plugin-demo-dynamic.page.message.recordCreated",
        "Demo record created",
      ),
      recordDeleted: t(
        "plugin.plugin-demo-dynamic.page.message.recordDeleted",
        "Demo record deleted",
      ),
      recordUpdated: t(
        "plugin.plugin-demo-dynamic.page.message.recordUpdated",
        "Demo record updated",
      ),
      requestFailed: t(
        "plugin.plugin-demo-dynamic.page.message.requestFailed",
        "Request failed ({status})",
      ),
      saveRecordFailed: t(
        "plugin.plugin-demo-dynamic.page.message.saveRecordFailed",
        "Failed to save the demo record",
      ),
      titleRequired: t(
        "plugin.plugin-demo-dynamic.page.message.titleRequired",
        "Record title is required",
      ),
      deleteRecordFailed: t(
        "plugin.plugin-demo-dynamic.page.message.deleteRecordFailed",
        "Failed to delete the demo record",
      ),
    },
    modal: {
      attachmentFieldHint: t(
        "plugin.plugin-demo-dynamic.page.modal.attachmentFieldHint",
        "Upload one sample attachment. If uninstall also clears plugin storage, the uploaded file is removed together.",
      ),
      attachmentFieldLabel: t(
        "plugin.plugin-demo-dynamic.page.modal.attachmentFieldLabel",
        "Attachment",
      ),
      contentFieldLabel: t(
        "plugin.plugin-demo-dynamic.page.modal.contentFieldLabel",
        "Content",
      ),
      createTitle: t(
        "plugin.plugin-demo-dynamic.page.modal.createTitle",
        "Create Demo Record",
      ),
      currentAttachment: (name) =>
        formatTemplate(currentAttachmentTemplate, {
          name,
        }),
      editTitle: t(
        "plugin.plugin-demo-dynamic.page.modal.editTitle",
        "Edit Demo Record",
      ),
      pendingAttachment: (name) =>
        formatTemplate(pendingAttachmentTemplate, {
          name,
        }),
      removeAttachment: t(
        "plugin.plugin-demo-dynamic.page.modal.removeAttachment",
        "Remove the current attachment on submit",
      ),
      summary: t(
        "plugin.plugin-demo-dynamic.page.modal.summary",
        "Record content is stored in the table created by the plugin-demo-dynamic install SQL. Uploaded files are stored in the plugin-authorized storage path.",
      ),
      titleFieldLabel: t(
        "plugin.plugin-demo-dynamic.page.modal.titleFieldLabel",
        "Record Title",
      ),
    },
    pageDescription: t(
      "plugin.plugin-demo-dynamic.page.description",
      "This page is mounted from the plugin-demo-dynamic embedded entry so the host can verify content rendering and the hosted standalone jump flow.",
    ),
    pageTitle: t(
      "plugin.plugin-demo-dynamic.page.title",
      "Dynamic Plugin Demo Is Live",
    ),
    panelTitle: t(
      "plugin.plugin-demo-dynamic.page.panelTitle",
      "Current Validation Scope",
    ),
    standaloneButton: t(
      "plugin.plugin-demo-dynamic.page.standaloneButton",
      "Open Standalone Page",
    ),
    standaloneHint: t(
      "plugin.plugin-demo-dynamic.page.standaloneHint",
      "Open the hosted static page in a new window to verify plugin asset publishing.",
    ),
    table: {
      createdAt: (value) =>
        formatTemplate(createdAtTemplate, {
          value,
        }),
      headers: {
        actions: t(
          "plugin.plugin-demo-dynamic.page.table.header.actions",
          "Actions",
        ),
        attachment: t(
          "plugin.plugin-demo-dynamic.page.table.header.attachment",
          "Attachment",
        ),
        content: t(
          "plugin.plugin-demo-dynamic.page.table.header.content",
          "Content",
        ),
        title: t("plugin.plugin-demo-dynamic.page.table.header.title", "Title"),
        updatedAt: t(
          "plugin.plugin-demo-dynamic.page.table.header.updatedAt",
          "Updated At",
        ),
      },
      loading: t(
        "plugin.plugin-demo-dynamic.page.table.loading",
        "Loading demo records...",
      ),
      noAttachment: t(
        "plugin.plugin-demo-dynamic.page.table.noAttachment",
        "No attachment",
      ),
      paginationSummary: (page, pages, start, end, total) =>
        formatTemplate(paginationSummaryTemplate, {
          end,
          page,
          pages,
          start,
          total,
        }),
    },
    workspaceSummary: t(
      "plugin.plugin-demo-dynamic.page.workspaceSummary",
      "This area reads the table created by the plugin-demo-dynamic install SQL and uses the plugin backend routes for create, update, delete, and attachment downloads. Disabling the plugin does not clear the data.",
    ),
  };
}

function ensureMountStyles(documentRef) {
  if (documentRef.getElementById(hostStyleId)) {
    return;
  }

  const styleElement = documentRef.createElement("style");
  styleElement.id = hostStyleId;
  styleElement.textContent = `
    .plugin-demo-dynamic-page {
      --dynamic-shell-border: rgba(15, 23, 42, 0.08);
      --dynamic-shell-shadow: 0 20px 48px rgba(15, 23, 42, 0.08);
      --dynamic-shell-accent: #1677ff;
      --dynamic-shell-accent-soft: rgba(22, 119, 255, 0.12);
      --dynamic-shell-text: #0f172a;
      --dynamic-shell-muted: #475569;
      --dynamic-shell-success: #16794f;
      --dynamic-shell-danger: #c62828;
      min-height: 100%;
      padding: 8px;
      color: var(--dynamic-shell-text);
      font-family:
        "PingFang SC",
        "Hiragino Sans GB",
        "Microsoft YaHei",
        "Noto Sans SC",
        sans-serif;
      box-sizing: border-box;
    }

    .plugin-demo-dynamic-page * {
      box-sizing: border-box;
    }

    .plugin-demo-dynamic-page__shell {
      position: relative;
      overflow: hidden;
      border-radius: 20px;
      border: 1px solid var(--dynamic-shell-border);
      background:
        radial-gradient(circle at top right, rgba(22, 119, 255, 0.12), transparent 26%),
        linear-gradient(180deg, #ffffff 0%, #f8fbff 100%);
      box-shadow: var(--dynamic-shell-shadow);
    }

    .plugin-demo-dynamic-page__hero {
      display: grid;
      grid-template-columns: minmax(0, 1.4fr) minmax(280px, 0.9fr);
      gap: 20px;
      padding: 28px;
    }

    .plugin-demo-dynamic-page__badge {
      display: inline-flex;
      align-items: center;
      gap: 8px;
      margin-bottom: 14px;
      padding: 6px 12px;
      border-radius: 999px;
      background: var(--dynamic-shell-accent-soft);
      color: var(--dynamic-shell-accent);
      font-size: 12px;
      font-weight: 700;
      letter-spacing: 0.08em;
    }

    .plugin-demo-dynamic-page__badge::before {
      content: "";
      width: 8px;
      height: 8px;
      border-radius: 999px;
      background: currentColor;
      box-shadow: 0 0 0 5px rgba(22, 119, 255, 0.12);
    }

    .plugin-demo-dynamic-page__title {
      margin: 0;
      font-size: 32px;
      line-height: 1.2;
      font-weight: 700;
      letter-spacing: -0.02em;
    }

    .plugin-demo-dynamic-page__description {
      margin: 14px 0 0;
      max-width: 720px;
      color: var(--dynamic-shell-muted);
      font-size: 15px;
      line-height: 1.8;
    }

    .plugin-demo-dynamic-page__cta {
      display: flex;
      align-items: center;
      gap: 12px;
      margin-top: 24px;
      flex-wrap: wrap;
    }

    .plugin-demo-dynamic-page__button,
    .plugin-demo-dynamic-page__ghost-button,
    .plugin-demo-dynamic-page__danger-button {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
      min-height: 40px;
      padding: 0 18px;
      border-radius: 10px;
      border: 1px solid transparent;
      font-size: 14px;
      font-weight: 600;
      line-height: 1.4;
      cursor: pointer;
      transition:
        background-color 0.2s ease,
        border-color 0.2s ease,
        color 0.2s ease,
        transform 0.2s ease,
        box-shadow 0.2s ease;
    }

    .plugin-demo-dynamic-page__button {
      border-color: var(--dynamic-shell-accent);
      background: var(--dynamic-shell-accent);
      color: #ffffff;
      box-shadow: 0 12px 30px rgba(22, 119, 255, 0.22);
    }

    .plugin-demo-dynamic-page__button:hover,
    .plugin-demo-dynamic-page__button:focus-visible {
      border-color: #4096ff;
      background: #4096ff;
      transform: translateY(-1px);
      box-shadow: 0 16px 34px rgba(22, 119, 255, 0.28);
      outline: none;
    }

    .plugin-demo-dynamic-page__ghost-button {
      border-color: rgba(22, 119, 255, 0.22);
      background: rgba(22, 119, 255, 0.08);
      color: var(--dynamic-shell-accent);
    }

    .plugin-demo-dynamic-page__ghost-button:hover,
    .plugin-demo-dynamic-page__ghost-button:focus-visible {
      border-color: rgba(22, 119, 255, 0.34);
      background: rgba(22, 119, 255, 0.14);
      outline: none;
    }

    .plugin-demo-dynamic-page__danger-button {
      border-color: rgba(198, 40, 40, 0.16);
      background: rgba(198, 40, 40, 0.08);
      color: var(--dynamic-shell-danger);
    }

    .plugin-demo-dynamic-page__danger-button:hover,
    .plugin-demo-dynamic-page__danger-button:focus-visible {
      border-color: rgba(198, 40, 40, 0.28);
      background: rgba(198, 40, 40, 0.14);
      outline: none;
    }

    .plugin-demo-dynamic-page__button:disabled,
    .plugin-demo-dynamic-page__ghost-button:disabled,
    .plugin-demo-dynamic-page__danger-button:disabled {
      opacity: 0.56;
      cursor: not-allowed;
      transform: none;
      box-shadow: none;
    }

    .plugin-demo-dynamic-page__hint {
      color: #64748b;
      font-size: 13px;
      line-height: 1.6;
    }

    .plugin-demo-dynamic-page__panel {
      display: flex;
      flex-direction: column;
      gap: 14px;
      padding: 22px;
      border-radius: 18px;
      border: 1px solid rgba(148, 163, 184, 0.18);
      background: rgba(255, 255, 255, 0.84);
      backdrop-filter: blur(10px);
    }

    .plugin-demo-dynamic-page__panel-title {
      margin: 0;
      font-size: 14px;
      font-weight: 700;
      color: #334155;
      letter-spacing: 0.04em;
    }

    .plugin-demo-dynamic-page__panel-metrics {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 12px;
    }

    .plugin-demo-dynamic-page__metric {
      padding: 14px;
      border-radius: 14px;
      background: #f8fafc;
      border: 1px solid rgba(148, 163, 184, 0.14);
    }

    .plugin-demo-dynamic-page__metric-value {
      display: block;
      margin-bottom: 4px;
      font-size: 18px;
      font-weight: 700;
      color: #0f172a;
    }

    .plugin-demo-dynamic-page__metric-label {
      display: -webkit-box;
      font-size: 12px;
      color: #64748b;
      line-height: 1.6;
      min-height: calc(1.6em * 2);
      overflow: hidden;
      -webkit-box-orient: vertical;
      -webkit-line-clamp: 2;
    }

    .plugin-demo-dynamic-page__grid {
      display: grid;
      grid-template-columns: repeat(3, minmax(0, 1fr));
      gap: 16px;
      padding: 0 28px 28px;
    }

    .plugin-demo-dynamic-page__card {
      padding: 20px;
      border-radius: 18px;
      border: 1px solid rgba(148, 163, 184, 0.14);
      background: #ffffff;
      box-shadow: 0 10px 28px rgba(15, 23, 42, 0.04);
    }

    .plugin-demo-dynamic-page__card-label {
      display: inline-flex;
      margin-bottom: 12px;
      color: #64748b;
      font-size: 12px;
      font-weight: 700;
      letter-spacing: 0.08em;
    }

    .plugin-demo-dynamic-page__card-value {
      margin: 0 0 8px;
      font-size: 18px;
      font-weight: 700;
      color: #0f172a;
    }

    .plugin-demo-dynamic-page__card-description {
      display: -webkit-box;
      margin: 0;
      color: #475569;
      font-size: 14px;
      line-height: 1.75;
      min-height: calc(1.75em * 2);
      overflow: hidden;
      -webkit-box-orient: vertical;
      -webkit-line-clamp: 2;
    }

    .plugin-demo-dynamic-page__workspace {
      padding: 0 28px 28px;
    }

    .plugin-demo-dynamic-page__workspace-card {
      border-radius: 18px;
      border: 1px solid rgba(148, 163, 184, 0.16);
      background: rgba(255, 255, 255, 0.96);
      box-shadow: 0 10px 28px rgba(15, 23, 42, 0.04);
      overflow: hidden;
    }

    .plugin-demo-dynamic-page__workspace-header {
      display: flex;
      align-items: flex-start;
      justify-content: space-between;
      gap: 16px;
      padding: 22px 22px 18px;
      border-bottom: 1px solid rgba(148, 163, 184, 0.14);
    }

    .plugin-demo-dynamic-page__workspace-title {
      margin: 0;
      font-size: 20px;
      font-weight: 700;
      color: #0f172a;
    }

    .plugin-demo-dynamic-page__workspace-summary {
      margin: 8px 0 0;
      color: var(--dynamic-shell-muted);
      font-size: 14px;
      line-height: 1.8;
    }

    .plugin-demo-dynamic-page__toolbar {
      display: flex;
      gap: 12px;
      flex-wrap: wrap;
    }

    .plugin-demo-dynamic-page__feedback {
      padding: 0 22px;
    }

    .plugin-demo-dynamic-page__feedback-item {
      margin: 12px 0 0;
      padding: 12px 14px;
      border-radius: 12px;
      font-size: 14px;
      line-height: 1.6;
      border: 1px solid transparent;
    }

    .plugin-demo-dynamic-page__feedback-item[data-kind="success"] {
      color: var(--dynamic-shell-success);
      background: rgba(22, 121, 79, 0.08);
      border-color: rgba(22, 121, 79, 0.12);
    }

    .plugin-demo-dynamic-page__feedback-item[data-kind="error"] {
      color: var(--dynamic-shell-danger);
      background: rgba(198, 40, 40, 0.08);
      border-color: rgba(198, 40, 40, 0.12);
    }

    .plugin-demo-dynamic-page__table-wrap {
      padding: 18px 22px 22px;
    }

    .plugin-demo-dynamic-page__table {
      width: 100%;
      border-collapse: collapse;
      table-layout: fixed;
    }

    .plugin-demo-dynamic-page__table th,
    .plugin-demo-dynamic-page__table td {
      padding: 14px 12px;
      border-bottom: 1px solid rgba(148, 163, 184, 0.12);
      text-align: left;
      vertical-align: top;
      font-size: 14px;
      line-height: 1.7;
    }

    .plugin-demo-dynamic-page__table th {
      color: #64748b;
      font-size: 12px;
      font-weight: 700;
      letter-spacing: 0.08em;
      text-transform: uppercase;
    }

    .plugin-demo-dynamic-page__table tbody tr:hover {
      background: rgba(22, 119, 255, 0.04);
    }

    .plugin-demo-dynamic-page__cell-title {
      font-weight: 700;
      color: #0f172a;
      word-break: break-word;
    }

    .plugin-demo-dynamic-page__cell-content {
      color: #475569;
      white-space: pre-wrap;
      word-break: break-word;
    }

    .plugin-demo-dynamic-page__cell-meta {
      color: #64748b;
      font-size: 13px;
      line-height: 1.6;
    }

    .plugin-demo-dynamic-page__attachment-link {
      display: inline-flex;
      align-items: center;
      gap: 8px;
      color: var(--dynamic-shell-accent);
      font-weight: 600;
      cursor: pointer;
      border: none;
      background: transparent;
      padding: 0;
    }

    .plugin-demo-dynamic-page__attachment-link:hover,
    .plugin-demo-dynamic-page__attachment-link:focus-visible {
      color: #4096ff;
      outline: none;
    }

    .plugin-demo-dynamic-page__row-actions {
      display: flex;
      gap: 10px;
      flex-wrap: wrap;
    }

    .plugin-demo-dynamic-page__inline-button {
      color: var(--dynamic-shell-accent);
      border: none;
      background: transparent;
      padding: 0;
      font-size: 14px;
      font-weight: 600;
      cursor: pointer;
    }

    .plugin-demo-dynamic-page__inline-button[data-variant="danger"] {
      color: var(--dynamic-shell-danger);
    }

    .plugin-demo-dynamic-page__empty {
      padding: 48px 20px;
      text-align: center;
      color: #64748b;
      font-size: 14px;
      line-height: 1.8;
    }

    .plugin-demo-dynamic-page__empty strong {
      display: block;
      margin-bottom: 8px;
      color: #334155;
      font-size: 16px;
    }

    .plugin-demo-dynamic-page__pagination {
      display: flex;
      align-items: center;
      justify-content: space-between;
      gap: 14px;
      margin-top: 18px;
      flex-wrap: wrap;
    }

    .plugin-demo-dynamic-page__pagination-summary {
      color: #64748b;
      font-size: 13px;
      line-height: 1.7;
    }

    .plugin-demo-dynamic-page__pagination-controls {
      display: inline-flex;
      align-items: center;
      gap: 8px;
      flex-wrap: wrap;
    }

    .plugin-demo-dynamic-page__pagination-button,
    .plugin-demo-dynamic-page__pagination-ellipsis {
      min-width: 36px;
      min-height: 36px;
      border-radius: 10px;
      font-size: 13px;
      line-height: 1;
    }

    .plugin-demo-dynamic-page__pagination-button {
      border: 1px solid rgba(148, 163, 184, 0.24);
      background: #ffffff;
      color: #334155;
      font-weight: 600;
      cursor: pointer;
      transition:
        border-color 0.2s ease,
        background-color 0.2s ease,
        color 0.2s ease,
        transform 0.2s ease;
    }

    .plugin-demo-dynamic-page__pagination-button:hover,
    .plugin-demo-dynamic-page__pagination-button:focus-visible {
      border-color: rgba(22, 119, 255, 0.42);
      color: var(--dynamic-shell-accent);
      outline: none;
      transform: translateY(-1px);
    }

    .plugin-demo-dynamic-page__pagination-button[data-active="true"] {
      border-color: var(--dynamic-shell-accent);
      background: var(--dynamic-shell-accent);
      color: #ffffff;
      box-shadow: 0 10px 22px rgba(22, 119, 255, 0.18);
    }

    .plugin-demo-dynamic-page__pagination-button:disabled {
      opacity: 0.5;
      cursor: not-allowed;
      transform: none;
      box-shadow: none;
    }

    .plugin-demo-dynamic-page__pagination-ellipsis {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      color: #94a3b8;
      font-weight: 700;
    }

    .plugin-demo-dynamic-page__modal-mask {
      position: fixed;
      inset: 0;
      display: none;
      align-items: center;
      justify-content: center;
      padding: 24px;
      background: rgba(15, 23, 42, 0.42);
      z-index: 999;
    }

    .plugin-demo-dynamic-page__modal-mask[data-open="true"] {
      display: flex;
    }

    .plugin-demo-dynamic-page__modal {
      width: min(680px, calc(100vw - 32px));
      max-height: calc(100vh - 48px);
      overflow: auto;
      border-radius: 18px;
      border: 1px solid rgba(148, 163, 184, 0.14);
      background: #ffffff;
      box-shadow: 0 24px 60px rgba(15, 23, 42, 0.18);
    }

    .plugin-demo-dynamic-page__modal-header,
    .plugin-demo-dynamic-page__modal-footer {
      padding: 20px 22px;
    }

    .plugin-demo-dynamic-page__modal-header {
      border-bottom: 1px solid rgba(148, 163, 184, 0.14);
    }

    .plugin-demo-dynamic-page__modal-title {
      margin: 0;
      font-size: 20px;
      font-weight: 700;
      color: #0f172a;
    }

    .plugin-demo-dynamic-page__modal-summary {
      margin: 10px 0 0;
      color: #64748b;
      font-size: 14px;
      line-height: 1.7;
    }

    .plugin-demo-dynamic-page__modal-body {
      padding: 20px 22px;
      display: grid;
      gap: 18px;
    }

    .plugin-demo-dynamic-page__field {
      display: grid;
      gap: 8px;
    }

    .plugin-demo-dynamic-page__field-label {
      color: #334155;
      font-size: 13px;
      font-weight: 700;
      letter-spacing: 0.04em;
    }

    .plugin-demo-dynamic-page__input,
    .plugin-demo-dynamic-page__textarea {
      width: 100%;
      border-radius: 10px;
      border: 1px solid rgba(148, 163, 184, 0.28);
      background: #ffffff;
      color: #0f172a;
      font-size: 14px;
      line-height: 1.6;
      padding: 11px 12px;
    }

    .plugin-demo-dynamic-page__input:focus,
    .plugin-demo-dynamic-page__textarea:focus {
      outline: none;
      border-color: rgba(22, 119, 255, 0.44);
      box-shadow: 0 0 0 3px rgba(22, 119, 255, 0.12);
    }

    .plugin-demo-dynamic-page__textarea {
      min-height: 132px;
      resize: vertical;
    }

    .plugin-demo-dynamic-page__field-hint {
      color: #64748b;
      font-size: 13px;
      line-height: 1.7;
    }

    .plugin-demo-dynamic-page__field-hint[data-kind="warn"] {
      color: #8a5200;
      background: rgba(250, 173, 20, 0.1);
      border: 1px solid rgba(250, 173, 20, 0.14);
      padding: 10px 12px;
      border-radius: 10px;
    }

    .plugin-demo-dynamic-page__file-meta {
      display: flex;
      gap: 10px;
      flex-wrap: wrap;
      color: #334155;
      font-size: 13px;
      line-height: 1.7;
    }

    .plugin-demo-dynamic-page__checkbox {
      display: inline-flex;
      align-items: center;
      gap: 10px;
      color: #334155;
      font-size: 14px;
      line-height: 1.6;
    }

    .plugin-demo-dynamic-page__modal-footer {
      display: flex;
      justify-content: flex-end;
      gap: 12px;
      border-top: 1px solid rgba(148, 163, 184, 0.14);
    }

    @media (max-width: 960px) {
      .plugin-demo-dynamic-page__hero,
      .plugin-demo-dynamic-page__grid {
        grid-template-columns: 1fr;
      }
    }

    @media (max-width: 768px) {
      .plugin-demo-dynamic-page {
        padding: 0;
      }

      .plugin-demo-dynamic-page__hero,
      .plugin-demo-dynamic-page__grid,
      .plugin-demo-dynamic-page__workspace {
        padding-inline: 18px;
      }

      .plugin-demo-dynamic-page__hero {
        padding-top: 22px;
        padding-bottom: 18px;
      }

      .plugin-demo-dynamic-page__grid,
      .plugin-demo-dynamic-page__workspace {
        padding-bottom: 20px;
      }

      .plugin-demo-dynamic-page__title {
        font-size: 26px;
      }

      .plugin-demo-dynamic-page__panel-metrics,
      .plugin-demo-dynamic-page__grid {
        grid-template-columns: 1fr;
      }

      .plugin-demo-dynamic-page__workspace-header {
        flex-direction: column;
      }

      .plugin-demo-dynamic-page__table-wrap {
        overflow-x: auto;
      }

      .plugin-demo-dynamic-page__table {
        min-width: 760px;
      }

      .plugin-demo-dynamic-page__pagination {
        flex-direction: column;
        align-items: flex-start;
      }
    }
  `;
  documentRef.head.append(styleElement);
}

function buildMetric(title, label, documentRef) {
  const wrapper = documentRef.createElement("div");
  wrapper.className = "plugin-demo-dynamic-page__metric";

  const value = documentRef.createElement("strong");
  value.className = "plugin-demo-dynamic-page__metric-value";
  value.textContent = title;

  const text = documentRef.createElement("span");
  text.className = "plugin-demo-dynamic-page__metric-label";
  text.textContent = label;

  wrapper.append(value, text);
  return wrapper;
}

function buildFeatureCard(item, documentRef) {
  const card = documentRef.createElement("article");
  card.className = "plugin-demo-dynamic-page__card";

  const label = documentRef.createElement("span");
  label.className = "plugin-demo-dynamic-page__card-label";
  label.textContent = item.label;

  const value = documentRef.createElement("h2");
  value.className = "plugin-demo-dynamic-page__card-value";
  value.textContent = item.value;

  const description = documentRef.createElement("p");
  description.className = "plugin-demo-dynamic-page__card-description";
  description.textContent = item.description;

  card.append(label, value, description);
  return card;
}

function createJSONHeaders(accessToken, locale, extraHeaders = {}) {
  const headers = {
    Accept: "application/json",
    ...extraHeaders,
  };
  if (locale) {
    headers["Accept-Language"] = locale;
  }
  if (accessToken) {
    headers.Authorization = `Bearer ${accessToken}`;
  }
  return headers;
}

async function parseErrorMessage(response, fallbackTemplate) {
  const fallback = formatTemplate(fallbackTemplate, {
    status: response.status,
  });
  const contentType = response.headers.get("content-type") || "";

  try {
    if (contentType.includes("application/json")) {
      const payload = await response.json();
      return (
        payload?.failure?.message ||
        payload?.message ||
        payload?.error?.message ||
        payload?.error ||
        fallback
      );
    }
    const text = (await response.text()).trim();
    return text || fallback;
  } catch (_error) {
    return fallback;
  }
}

function readFileAsBase64(file, errorMessage) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      const content = String(reader.result || "");
      const marker = "base64,";
      const markerIndex = content.indexOf(marker);
      if (markerIndex < 0) {
        reject(new Error(errorMessage));
        return;
      }
      resolve(content.slice(markerIndex + marker.length));
    };
    reader.onerror = () => reject(new Error(errorMessage));
    reader.readAsDataURL(file);
  });
}

export function mount(context) {
  const documentRef = context.container.ownerDocument;
  ensureMountStyles(documentRef);
  const pageCopy = buildPageCopy(context);

  const accessToken = context.accessToken || "";
  const currentLocale = resolveSupportedLocale(
    context.locale ||
      documentRef.documentElement.lang ||
      documentRef.defaultView?.navigator?.language ||
      "",
  );
  let recordFetchToken = 0;
  const state = {
    destroyed: false,
    loading: false,
    submitting: false,
    records: [],
    pageNum: 1,
    pageSize: defaultRecordPageSize,
    total: 0,
    successMessage: "",
    errorMessage: "",
    modalOpen: false,
    editingRecord: null,
    selectedFile: null,
  };

  const root = documentRef.createElement("section");
  root.className = "plugin-demo-dynamic-page";
  root.setAttribute("data-testid", "plugin-demo-dynamic-root");

  const shell = documentRef.createElement("div");
  shell.className = "plugin-demo-dynamic-page__shell";

  const hero = documentRef.createElement("div");
  hero.className = "plugin-demo-dynamic-page__hero";

  const intro = documentRef.createElement("div");

  const badge = documentRef.createElement("span");
  badge.className = "plugin-demo-dynamic-page__badge";
  badge.textContent = pageCopy.badge;

  const heading = documentRef.createElement("h1");
  heading.className = "plugin-demo-dynamic-page__title";
  heading.setAttribute("data-testid", "plugin-demo-dynamic-title");
  heading.textContent = pageCopy.pageTitle;

  const description = documentRef.createElement("p");
  description.className = "plugin-demo-dynamic-page__description";
  description.setAttribute("data-testid", "plugin-demo-dynamic-description");
  description.textContent = pageCopy.pageDescription;

  const actions = documentRef.createElement("div");
  actions.className = "plugin-demo-dynamic-page__cta";

  const actionButton = documentRef.createElement("button");
  actionButton.type = "button";
  actionButton.className = "plugin-demo-dynamic-page__button";
  actionButton.setAttribute(
    "data-testid",
    "plugin-demo-dynamic-open-standalone",
  );
  actionButton.textContent = pageCopy.standaloneButton;
  actionButton.addEventListener("click", () => {
    const standaloneURL = new URL("./standalone.html", context.baseURL);
    if (currentLocale) {
      standaloneURL.searchParams.set("lang", currentLocale);
    }
    const standaloneI18nKey = persistStandaloneRuntimeMessages(
      context,
      currentLocale,
    );
    if (standaloneI18nKey) {
      standaloneURL.searchParams.set("i18nKey", standaloneI18nKey);
    }
    window.open(standaloneURL.toString(), "_blank", "noopener,noreferrer");
  });

  const hint = documentRef.createElement("span");
  hint.className = "plugin-demo-dynamic-page__hint";
  hint.textContent = pageCopy.standaloneHint;

  actions.append(actionButton, hint);
  intro.append(badge, heading, description, actions);

  const sidePanel = documentRef.createElement("aside");
  sidePanel.className = "plugin-demo-dynamic-page__panel";

  const panelTitle = documentRef.createElement("h2");
  panelTitle.className = "plugin-demo-dynamic-page__panel-title";
  panelTitle.textContent = pageCopy.panelTitle;

  const metrics = documentRef.createElement("div");
  metrics.className = "plugin-demo-dynamic-page__panel-metrics";
  metrics.append(
    ...pageCopy.metrics.map((item) =>
      buildMetric(item.label, item.description, documentRef),
    ),
  );
  sidePanel.append(panelTitle, metrics);

  hero.append(intro, sidePanel);

  const featureGrid = documentRef.createElement("div");
  featureGrid.className = "plugin-demo-dynamic-page__grid";
  for (const item of pageCopy.featureItems) {
    featureGrid.append(buildFeatureCard(item, documentRef));
  }

  const workspace = documentRef.createElement("div");
  workspace.className = "plugin-demo-dynamic-page__workspace";

  const workspaceCard = documentRef.createElement("section");
  workspaceCard.className = "plugin-demo-dynamic-page__workspace-card";

  const workspaceHeader = documentRef.createElement("div");
  workspaceHeader.className = "plugin-demo-dynamic-page__workspace-header";

  const workspaceHeadingBlock = documentRef.createElement("div");
  const workspaceTitle = documentRef.createElement("h2");
  workspaceTitle.className = "plugin-demo-dynamic-page__workspace-title";
  workspaceTitle.textContent = pageCopy.gridTitle;

  const workspaceSummary = documentRef.createElement("p");
  workspaceSummary.className = "plugin-demo-dynamic-page__workspace-summary";
  workspaceSummary.textContent = pageCopy.workspaceSummary;

  workspaceHeadingBlock.append(workspaceTitle, workspaceSummary);

  const toolbar = documentRef.createElement("div");
  toolbar.className = "plugin-demo-dynamic-page__toolbar";

  const addButton = documentRef.createElement("button");
  addButton.type = "button";
  addButton.className = "plugin-demo-dynamic-page__button";
  addButton.setAttribute("data-testid", "plugin-demo-dynamic-record-add");
  addButton.textContent = pageCopy.actions.addRecord;

  const reloadButton = documentRef.createElement("button");
  reloadButton.type = "button";
  reloadButton.className = "plugin-demo-dynamic-page__ghost-button";
  reloadButton.textContent = pageCopy.actions.reloadList;

  toolbar.append(addButton, reloadButton);
  workspaceHeader.append(workspaceHeadingBlock, toolbar);

  const feedback = documentRef.createElement("div");
  feedback.className = "plugin-demo-dynamic-page__feedback";

  const tableWrap = documentRef.createElement("div");
  tableWrap.className = "plugin-demo-dynamic-page__table-wrap";
  tableWrap.setAttribute("data-testid", "plugin-demo-dynamic-record-grid");

  const modalMask = documentRef.createElement("div");
  modalMask.className = "plugin-demo-dynamic-page__modal-mask";
  modalMask.setAttribute("data-testid", "plugin-demo-dynamic-record-modal");
  modalMask.setAttribute("data-open", "false");

  const modal = documentRef.createElement("div");
  modal.className = "plugin-demo-dynamic-page__modal";
  modal.addEventListener("click", (event) => event.stopPropagation());

  const modalHeader = documentRef.createElement("div");
  modalHeader.className = "plugin-demo-dynamic-page__modal-header";

  const modalTitle = documentRef.createElement("h3");
  modalTitle.className = "plugin-demo-dynamic-page__modal-title";

  const modalSummary = documentRef.createElement("p");
  modalSummary.className = "plugin-demo-dynamic-page__modal-summary";
  modalSummary.textContent = pageCopy.modal.summary;
  modalHeader.append(modalTitle, modalSummary);

  const modalBody = documentRef.createElement("div");
  modalBody.className = "plugin-demo-dynamic-page__modal-body";

  const titleField = documentRef.createElement("label");
  titleField.className = "plugin-demo-dynamic-page__field";
  const titleLabel = documentRef.createElement("span");
  titleLabel.className = "plugin-demo-dynamic-page__field-label";
  titleLabel.textContent = pageCopy.modal.titleFieldLabel;
  const titleInput = documentRef.createElement("input");
  titleInput.className = "plugin-demo-dynamic-page__input";
  titleInput.setAttribute(
    "data-testid",
    "plugin-demo-dynamic-record-title-input",
  );
  titleInput.maxLength = 128;
  titleField.append(titleLabel, titleInput);

  const contentField = documentRef.createElement("label");
  contentField.className = "plugin-demo-dynamic-page__field";
  const contentLabel = documentRef.createElement("span");
  contentLabel.className = "plugin-demo-dynamic-page__field-label";
  contentLabel.textContent = pageCopy.modal.contentFieldLabel;
  const contentInput = documentRef.createElement("textarea");
  contentInput.className = "plugin-demo-dynamic-page__textarea";
  contentInput.setAttribute(
    "data-testid",
    "plugin-demo-dynamic-record-content-input",
  );
  contentInput.maxLength = 1000;
  contentField.append(contentLabel, contentInput);

  const attachmentField = documentRef.createElement("div");
  attachmentField.className = "plugin-demo-dynamic-page__field";
  const attachmentLabel = documentRef.createElement("span");
  attachmentLabel.className = "plugin-demo-dynamic-page__field-label";
  attachmentLabel.textContent = pageCopy.modal.attachmentFieldLabel;
  const attachmentHint = documentRef.createElement("div");
  attachmentHint.className = "plugin-demo-dynamic-page__field-hint";
  attachmentHint.textContent = pageCopy.modal.attachmentFieldHint;
  const fileInput = documentRef.createElement("input");
  fileInput.type = "file";
  fileInput.className = "plugin-demo-dynamic-page__input";
  fileInput.setAttribute(
    "data-testid",
    "plugin-demo-dynamic-record-file-input",
  );
  const fileMeta = documentRef.createElement("div");
  fileMeta.className = "plugin-demo-dynamic-page__file-meta";
  const existingAttachment = documentRef.createElement("div");
  const selectedAttachment = documentRef.createElement("div");
  fileMeta.append(existingAttachment, selectedAttachment);
  const removeAttachmentLabel = documentRef.createElement("label");
  removeAttachmentLabel.className = "plugin-demo-dynamic-page__checkbox";
  removeAttachmentLabel.hidden = true;
  removeAttachmentLabel.setAttribute(
    "data-testid",
    "plugin-demo-dynamic-record-remove-attachment",
  );
  const removeAttachmentInput = documentRef.createElement("input");
  removeAttachmentInput.type = "checkbox";
  const removeAttachmentText = documentRef.createElement("span");
  removeAttachmentText.textContent = pageCopy.modal.removeAttachment;
  removeAttachmentLabel.append(removeAttachmentInput, removeAttachmentText);
  attachmentField.append(
    attachmentLabel,
    attachmentHint,
    fileInput,
    fileMeta,
    removeAttachmentLabel,
  );

  const modalFeedback = documentRef.createElement("div");
  modalFeedback.className = "plugin-demo-dynamic-page__field-hint";
  modalFeedback.hidden = true;

  modalBody.append(titleField, contentField, attachmentField, modalFeedback);

  const modalFooter = documentRef.createElement("div");
  modalFooter.className = "plugin-demo-dynamic-page__modal-footer";

  const cancelButton = documentRef.createElement("button");
  cancelButton.type = "button";
  cancelButton.className = "plugin-demo-dynamic-page__ghost-button";
  cancelButton.setAttribute("data-testid", "plugin-demo-dynamic-record-cancel");
  cancelButton.textContent = pageCopy.actions.cancel;

  const submitButton = documentRef.createElement("button");
  submitButton.type = "button";
  submitButton.className = "plugin-demo-dynamic-page__button";
  submitButton.setAttribute("data-testid", "plugin-demo-dynamic-record-submit");
  submitButton.textContent = pageCopy.actions.createRecord;

  modalFooter.append(cancelButton, submitButton);
  modal.append(modalHeader, modalBody, modalFooter);
  modalMask.append(modal);
  modalMask.addEventListener("click", () => closeModal());

  workspaceCard.append(workspaceHeader, feedback, tableWrap);
  workspace.append(workspaceCard);
  shell.append(hero, featureGrid, workspace);
  root.append(shell, modalMask);
  context.container.replaceChildren(root);

  function setFeedback(type, message) {
    state.successMessage = type === "success" ? message : "";
    state.errorMessage = type === "error" ? message : "";
    renderFeedback();
  }

  function clearFeedback() {
    state.successMessage = "";
    state.errorMessage = "";
    renderFeedback();
  }

  function renderFeedback() {
    feedback.replaceChildren();
    modalFeedback.hidden = true;
    modalFeedback.textContent = "";
    modalFeedback.removeAttribute("data-kind");

    if (state.errorMessage) {
      const item = documentRef.createElement("div");
      item.className = "plugin-demo-dynamic-page__feedback-item";
      item.setAttribute("data-kind", "error");
      item.textContent = state.errorMessage;
      feedback.append(item);
    }
    if (state.successMessage) {
      const item = documentRef.createElement("div");
      item.className = "plugin-demo-dynamic-page__feedback-item";
      item.setAttribute("data-kind", "success");
      item.textContent = state.successMessage;
      feedback.append(item);
    }
  }

  function updateActionState() {
    addButton.disabled = state.loading || state.submitting;
    reloadButton.disabled = state.loading || state.submitting;
    submitButton.disabled = state.submitting;
    cancelButton.disabled = state.submitting;
  }

  // getTotalPages derives the visible page count from the current total and
  // guarantees the summary logic always has a minimum first page to render.
  function getTotalPages(total = state.total) {
    return Math.max(1, Math.ceil(total / state.pageSize));
  }

  // buildPaginationItems keeps the pagination control compact while still
  // exposing the current page neighborhood and the first/last page anchors.
  function buildPaginationItems(currentPage, totalPages) {
    if (totalPages <= 7) {
      return Array.from({ length: totalPages }, (_value, index) => index + 1);
    }
    if (currentPage <= 4) {
      return [1, 2, 3, 4, 5, "...", totalPages];
    }
    if (currentPage >= totalPages - 3) {
      return [
        1,
        "...",
        totalPages - 4,
        totalPages - 3,
        totalPages - 2,
        totalPages - 1,
        totalPages,
      ];
    }
    return [
      1,
      "...",
      currentPage - 1,
      currentPage,
      currentPage + 1,
      "...",
      totalPages,
    ];
  }

  // buildPaginationButton creates one interactive pagination control and binds
  // it to the shared page-change handler when the item maps to a real page.
  function buildPaginationButton(label, pageNumber, options = {}) {
    const button = documentRef.createElement("button");
    button.type = "button";
    button.className = "plugin-demo-dynamic-page__pagination-button";
    button.textContent = label;
    button.disabled = !!options.disabled;
    if (options.active) {
      button.setAttribute("data-active", "true");
      button.setAttribute("aria-current", "page");
    }
    if (typeof pageNumber === "number") {
      button.setAttribute(
        "data-testid",
        `plugin-demo-dynamic-pagination-page-${pageNumber}`,
      );
      button.addEventListener("click", () => {
        void changePage(pageNumber);
      });
    }
    return button;
  }

  // buildPagination renders both the current-range summary and the page
  // controls so the demo record list can be browsed page by page.
  function buildPagination() {
    if (state.total <= 0) {
      return null;
    }

    const pagination = documentRef.createElement("div");
    pagination.className = "plugin-demo-dynamic-page__pagination";
    pagination.setAttribute(
      "data-testid",
      "plugin-demo-dynamic-record-pagination",
    );

    const totalPages = getTotalPages();
    const rangeStart = (state.pageNum - 1) * state.pageSize + 1;
    const rangeEnd = Math.min(
      state.total,
      rangeStart + Math.max(state.records.length - 1, 0),
    );

    const summary = documentRef.createElement("div");
    summary.className = "plugin-demo-dynamic-page__pagination-summary";
    summary.setAttribute(
      "data-testid",
      "plugin-demo-dynamic-pagination-summary",
    );
    summary.textContent = pageCopy.table.paginationSummary(
      state.pageNum,
      totalPages,
      rangeStart,
      rangeEnd,
      state.total,
    );
    pagination.append(summary);

    if (totalPages <= 1) {
      return pagination;
    }

    const controls = documentRef.createElement("div");
    controls.className = "plugin-demo-dynamic-page__pagination-controls";

    const previousButton = buildPaginationButton(
      pageCopy.actions.previousPage,
      state.pageNum - 1,
      {
        disabled: state.loading || state.pageNum <= 1,
      },
    );
    previousButton.setAttribute(
      "data-testid",
      "plugin-demo-dynamic-pagination-prev",
    );
    controls.append(previousButton);

    for (const item of buildPaginationItems(state.pageNum, totalPages)) {
      if (item === "...") {
        const ellipsis = documentRef.createElement("span");
        ellipsis.className = "plugin-demo-dynamic-page__pagination-ellipsis";
        ellipsis.textContent = "...";
        controls.append(ellipsis);
        continue;
      }
      controls.append(
        buildPaginationButton(String(item), item, {
          active: item === state.pageNum,
          disabled: state.loading || item === state.pageNum,
        }),
      );
    }

    const nextButton = buildPaginationButton(
      pageCopy.actions.nextPage,
      state.pageNum + 1,
      {
        disabled: state.loading || state.pageNum >= totalPages,
      },
    );
    nextButton.setAttribute(
      "data-testid",
      "plugin-demo-dynamic-pagination-next",
    );
    controls.append(nextButton);

    pagination.append(controls);
    return pagination;
  }

  // changePage guards duplicate or invalid page transitions before requesting
  // the next record slice from the backend.
  async function changePage(pageNumber) {
    const targetPage = Math.max(1, Math.min(pageNumber, getTotalPages()));
    if (targetPage === state.pageNum || state.loading || state.submitting) {
      return;
    }
    await fetchRecords({ pageNum: targetPage });
  }

  function renderTable() {
    tableWrap.replaceChildren();

    if (state.loading) {
      const loading = documentRef.createElement("div");
      loading.className = "plugin-demo-dynamic-page__empty";
      loading.textContent = pageCopy.table.loading;
      tableWrap.append(loading);
      return;
    }

    if (state.records.length === 0) {
      const empty = documentRef.createElement("div");
      empty.className = "plugin-demo-dynamic-page__empty";
      empty.setAttribute("data-testid", "plugin-demo-dynamic-record-empty");
      empty.innerHTML = `<strong>${pageCopy.emptyTitle}</strong>${pageCopy.emptyText}`;
      tableWrap.append(empty);
      return;
    }

    const table = documentRef.createElement("table");
    table.className = "plugin-demo-dynamic-page__table";

    const thead = documentRef.createElement("thead");
    thead.innerHTML = `
      <tr>
        <th style="width: 24%">${pageCopy.table.headers.title}</th>
        <th style="width: 30%">${pageCopy.table.headers.content}</th>
        <th style="width: 18%">${pageCopy.table.headers.attachment}</th>
        <th style="width: 16%">${pageCopy.table.headers.updatedAt}</th>
        <th style="width: 12%">${pageCopy.table.headers.actions}</th>
      </tr>
    `;

    const tbody = documentRef.createElement("tbody");
    for (const record of state.records) {
      const row = documentRef.createElement("tr");
      row.setAttribute(
        "data-testid",
        `plugin-demo-dynamic-record-row-${record.id}`,
      );

      const titleCell = documentRef.createElement("td");
      const titleBlock = documentRef.createElement("div");
      titleBlock.className = "plugin-demo-dynamic-page__cell-title";
      titleBlock.textContent = record.title || "";
      const createdMeta = documentRef.createElement("div");
      createdMeta.className = "plugin-demo-dynamic-page__cell-meta";
      createdMeta.textContent = pageCopy.table.createdAt(
        record.createdAt || "-",
      );
      titleCell.append(titleBlock, createdMeta);

      const contentCell = documentRef.createElement("td");
      const contentBlock = documentRef.createElement("div");
      contentBlock.className = "plugin-demo-dynamic-page__cell-content";
      contentBlock.textContent = record.content || "-";
      contentCell.append(contentBlock);

      const attachmentCell = documentRef.createElement("td");
      if (record.hasAttachment) {
        const downloadButton = documentRef.createElement("button");
        downloadButton.type = "button";
        downloadButton.className = "plugin-demo-dynamic-page__attachment-link";
        downloadButton.textContent =
          record.attachmentName || pageCopy.actions.downloadAttachment;
        downloadButton.addEventListener("click", () => {
          void downloadAttachment(record);
        });
        attachmentCell.append(downloadButton);
      } else {
        attachmentCell.textContent = pageCopy.table.noAttachment;
      }

      const updatedCell = documentRef.createElement("td");
      const updatedText = documentRef.createElement("div");
      updatedText.className = "plugin-demo-dynamic-page__cell-meta";
      updatedText.textContent = record.updatedAt || "-";
      updatedCell.append(updatedText);

      const actionCell = documentRef.createElement("td");
      const actionWrap = documentRef.createElement("div");
      actionWrap.className = "plugin-demo-dynamic-page__row-actions";

      const editButton = documentRef.createElement("button");
      editButton.type = "button";
      editButton.className = "plugin-demo-dynamic-page__inline-button";
      editButton.textContent = pageCopy.actions.editRecord;
      editButton.disabled = state.submitting;
      editButton.addEventListener("click", () => openModal(record));

      const deleteButton = documentRef.createElement("button");
      deleteButton.type = "button";
      deleteButton.className = "plugin-demo-dynamic-page__inline-button";
      deleteButton.setAttribute("data-variant", "danger");
      deleteButton.textContent = pageCopy.actions.deleteRecord;
      deleteButton.disabled = state.submitting;
      deleteButton.addEventListener("click", () => {
        void deleteRecord(record);
      });

      actionWrap.append(editButton, deleteButton);
      actionCell.append(actionWrap);

      row.append(
        titleCell,
        contentCell,
        attachmentCell,
        updatedCell,
        actionCell,
      );
      tbody.append(row);
    }

    table.append(thead, tbody);
    tableWrap.append(table);

    const pagination = buildPagination();
    if (pagination) {
      tableWrap.append(pagination);
    }
  }

  function resetModalState() {
    state.selectedFile = null;
    state.editingRecord = null;
    titleInput.value = "";
    contentInput.value = "";
    fileInput.value = "";
    removeAttachmentInput.checked = false;
    removeAttachmentLabel.hidden = true;
    existingAttachment.textContent = "";
    selectedAttachment.textContent = "";
    modalFeedback.hidden = true;
    modalFeedback.textContent = "";
    modalFeedback.removeAttribute("data-kind");
  }

  function renderModal() {
    modalMask.setAttribute("data-open", state.modalOpen ? "true" : "false");
    if (!state.modalOpen) {
      updateActionState();
      return;
    }

    const isEditing = !!state.editingRecord;
    modalTitle.textContent = isEditing
      ? pageCopy.modal.editTitle
      : pageCopy.modal.createTitle;
    submitButton.textContent = state.submitting
      ? pageCopy.actions.savePending
      : isEditing
        ? pageCopy.actions.saveEdit
        : pageCopy.actions.createRecord;

    if (state.selectedFile) {
      selectedAttachment.textContent = pageCopy.modal.pendingAttachment(
        state.selectedFile.name,
      );
    } else {
      selectedAttachment.textContent = "";
    }

    updateActionState();
  }

  function openModal(record = null) {
    clearFeedback();
    resetModalState();
    state.modalOpen = true;
    state.editingRecord = record;
    if (record) {
      titleInput.value = record.title || "";
      contentInput.value = record.content || "";
      if (record.hasAttachment) {
        existingAttachment.textContent = pageCopy.modal.currentAttachment(
          record.attachmentName,
        );
        removeAttachmentLabel.hidden = false;
      }
    }
    renderModal();
  }

  function closeModal(force = false) {
    if (state.submitting && !force) {
      return;
    }
    state.modalOpen = false;
    resetModalState();
    renderModal();
  }

  async function requestJSON(path, options = {}) {
    const response = await fetch(
      new URL(path, window.location.origin).toString(),
      {
        ...options,
        headers: createJSONHeaders(
          accessToken,
          currentLocale,
          options.headers || {},
        ),
      },
    );
    if (!response.ok) {
      throw new Error(
        await parseErrorMessage(response, pageCopy.messages.requestFailed),
      );
    }
    return response.json();
  }

  async function fetchRecords(options = {}) {
    if (state.destroyed) {
      return;
    }
    const resetFeedback = options.resetFeedback !== false;
    let nextPageNum = Number.isInteger(options.pageNum)
      ? options.pageNum
      : state.pageNum;
    nextPageNum = Math.max(1, nextPageNum);
    recordFetchToken += 1;
    const currentFetchToken = recordFetchToken;
    state.loading = true;
    if (resetFeedback) {
      clearFeedback();
    }
    updateActionState();
    renderTable();

    try {
      while (true) {
        const query = new URLSearchParams({
          pageNum: String(nextPageNum),
          pageSize: String(state.pageSize),
        });
        const payload = await requestJSON(
          `${apiBasePath}/demo-records?${query.toString()}`,
        );
        if (state.destroyed || currentFetchToken !== recordFetchToken) {
          return;
        }

        const records = Array.isArray(payload?.list) ? payload.list : [];
        const total = Number.isFinite(Number(payload?.total))
          ? Number(payload.total)
          : records.length;
        const totalPages = Math.max(1, Math.ceil(total / state.pageSize));
        if (total > 0 && nextPageNum > totalPages) {
          nextPageNum = totalPages;
          continue;
        }

        state.pageNum = nextPageNum;
        state.total = total;
        state.records = records;
        break;
      }
    } catch (error) {
      if (state.destroyed || currentFetchToken !== recordFetchToken) {
        return;
      }
      state.records = [];
      state.total = 0;
      setFeedback(
        "error",
        error instanceof Error
          ? error.message
          : pageCopy.messages.fetchRecordsFailed,
      );
    } finally {
      if (state.destroyed || currentFetchToken !== recordFetchToken) {
        return;
      }
      state.loading = false;
      updateActionState();
      renderTable();
    }
  }

  async function buildMutationPayload() {
    const payload = {
      title: titleInput.value.trim(),
      content: contentInput.value.trim(),
      attachmentName: "",
      attachmentContentBase64: "",
      attachmentContentType: "",
      removeAttachment: removeAttachmentInput.checked,
    };

    if (!payload.title) {
      throw new Error(pageCopy.messages.titleRequired);
    }
    if (state.selectedFile) {
      payload.attachmentName = state.selectedFile.name;
      payload.attachmentContentBase64 = await readFileAsBase64(
        state.selectedFile,
        pageCopy.messages.fileReadFailed,
      );
      payload.attachmentContentType =
        state.selectedFile.type || "application/octet-stream";
    }
    return payload;
  }

  async function submitRecord() {
    state.submitting = true;
    updateActionState();
    modalFeedback.hidden = true;

    try {
      const payload = await buildMutationPayload();
      const isEditing = !!state.editingRecord;
      const path = isEditing
        ? `${apiBasePath}/demo-records/${state.editingRecord.id}`
        : `${apiBasePath}/demo-records`;
      const method = isEditing ? "PUT" : "POST";
      await requestJSON(path, {
        method,
        headers: {
          "Content-Type": "application/json",
        },
        body: JSON.stringify(payload),
      });
      closeModal(true);
      setFeedback(
        "success",
        isEditing
          ? pageCopy.messages.recordUpdated
          : pageCopy.messages.recordCreated,
      );
      await fetchRecords({ pageNum: 1, resetFeedback: false });
    } catch (error) {
      modalFeedback.hidden = false;
      modalFeedback.setAttribute("data-kind", "warn");
      modalFeedback.textContent =
        error instanceof Error
          ? error.message
          : pageCopy.messages.saveRecordFailed;
    } finally {
      state.submitting = false;
      updateActionState();
      renderTable();
      renderModal();
    }
  }

  async function deleteRecord(record) {
    const confirmed = window.confirm(
      pageCopy.messages.deleteConfirm(record.title || ""),
    );
    if (!confirmed) {
      return;
    }
    clearFeedback();
    try {
      await requestJSON(`${apiBasePath}/demo-records/${record.id}`, {
        method: "DELETE",
      });
      setFeedback("success", pageCopy.messages.recordDeleted);
      await fetchRecords({ resetFeedback: false });
    } catch (error) {
      setFeedback(
        "error",
        error instanceof Error
          ? error.message
          : pageCopy.messages.deleteRecordFailed,
      );
    }
  }

  async function downloadAttachment(record) {
    clearFeedback();
    try {
      const response = await fetch(
        new URL(
          `${apiBasePath}/demo-records/${record.id}/attachment`,
          window.location.origin,
        ).toString(),
        {
          headers: createJSONHeaders(accessToken, currentLocale),
        },
      );
      if (!response.ok) {
        throw new Error(
          await parseErrorMessage(response, pageCopy.messages.requestFailed),
        );
      }
      const blob = await response.blob();
      const objectURL = URL.createObjectURL(blob);
      const link = documentRef.createElement("a");
      link.href = objectURL;
      link.download = record.attachmentName || "attachment";
      documentRef.body.append(link);
      link.click();
      link.remove();
      URL.revokeObjectURL(objectURL);
      setFeedback("success", pageCopy.messages.downloadStarted);
    } catch (error) {
      setFeedback(
        "error",
        error instanceof Error
          ? error.message
          : pageCopy.messages.downloadFailed,
      );
    }
  }

  addButton.addEventListener("click", () => openModal());
  reloadButton.addEventListener("click", () => {
    void fetchRecords();
  });
  cancelButton.addEventListener("click", () => closeModal());
  submitButton.addEventListener("click", () => {
    void submitRecord();
  });
  fileInput.addEventListener("change", () => {
    state.selectedFile = fileInput.files?.[0] || null;
    renderModal();
  });

  renderFeedback();
  renderTable();
  renderModal();
  updateActionState();
  void fetchRecords();

  return {
    unmount() {
      state.destroyed = true;
      root.remove();
    },
  };
}
