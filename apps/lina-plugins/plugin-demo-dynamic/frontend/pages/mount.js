const pluginID = "plugin-demo-dynamic";
const apiBasePath = `/api/v1/extensions/${pluginID}`;
const pageTitle = "动态插件示例已生效";
const pageDescription =
  "该页面来自 plugin-demo-dynamic 的动态挂载入口，用于验证宿主主内容区展示与独立静态页面跳转。";
const buttonLabel = "打开独立页面";
const gridTitle = "示例记录";
const emptyText =
  "当前只有安装 SQL 初始化的默认记录，你可以继续新增、编辑或删除自定义记录。";
const defaultRecordPageSize = 10;

const hostStyleId = "plugin-demo-dynamic-mount-style";

const featureItems = [
  {
    label: "接入方式",
    value: "宿主内嵌挂载",
    description:
      "通过宿主页壳动态加载页面入口，并沿用宿主登录态访问后端动态路由。",
  },
  {
    label: "数据示例",
    value: "安装 SQL + CRUD",
    description: "安装时创建插件自有业务表，页面可直接完成增删查改和附件下载。",
  },
  {
    label: "卸载治理",
    value: "可选清理数据",
    description:
      "禁用不删数据，卸载时由宿主弹窗决定是否同时清理表数据和存储文件。",
  },
];

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

function createJSONHeaders(accessToken, extraHeaders = {}) {
  const headers = {
    Accept: "application/json",
    ...extraHeaders,
  };
  if (accessToken) {
    headers.Authorization = `Bearer ${accessToken}`;
  }
  return headers;
}

async function parseErrorMessage(response) {
  const fallback = `请求失败 (${response.status})`;
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

function readFileAsBase64(file) {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onload = () => {
      const content = String(reader.result || "");
      const marker = "base64,";
      const markerIndex = content.indexOf(marker);
      if (markerIndex < 0) {
        reject(new Error("附件读取失败"));
        return;
      }
      resolve(content.slice(markerIndex + marker.length));
    };
    reader.onerror = () => reject(new Error("附件读取失败"));
    reader.readAsDataURL(file);
  });
}

export function mount(context) {
  const documentRef = context.container.ownerDocument;
  ensureMountStyles(documentRef);

  const accessToken = context.accessToken || "";
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
  badge.textContent = "WASM 插件示例";

  const heading = documentRef.createElement("h1");
  heading.className = "plugin-demo-dynamic-page__title";
  heading.textContent = pageTitle;

  const description = documentRef.createElement("p");
  description.className = "plugin-demo-dynamic-page__description";
  description.textContent = pageDescription;

  const actions = documentRef.createElement("div");
  actions.className = "plugin-demo-dynamic-page__cta";

  const actionButton = documentRef.createElement("button");
  actionButton.type = "button";
  actionButton.className = "plugin-demo-dynamic-page__button";
  actionButton.setAttribute(
    "data-testid",
    "plugin-demo-dynamic-open-standalone",
  );
  actionButton.textContent = buttonLabel;
  actionButton.addEventListener("click", () => {
    const standaloneURL = new URL(
      "./standalone.html",
      context.baseURL,
    ).toString();
    window.open(standaloneURL, "_blank", "noopener,noreferrer");
  });

  const hint = documentRef.createElement("span");
  hint.className = "plugin-demo-dynamic-page__hint";
  hint.textContent =
    "点击后将在新窗口打开托管的纯静态页面，用于验证插件资源公开访问能力。";

  actions.append(actionButton, hint);
  intro.append(badge, heading, description, actions);

  const sidePanel = documentRef.createElement("aside");
  sidePanel.className = "plugin-demo-dynamic-page__panel";

  const panelTitle = documentRef.createElement("h2");
  panelTitle.className = "plugin-demo-dynamic-page__panel-title";
  panelTitle.textContent = "当前验证范围";

  const metrics = documentRef.createElement("div");
  metrics.className = "plugin-demo-dynamic-page__panel-metrics";
  metrics.append(
    buildMetric("动态加载", "由宿主页壳动态导入并挂载页面入口", documentRef),
    buildMetric("SQL 创建", "安装时生成插件自有业务表示例数据", documentRef),
    buildMetric("附件存储", "示例记录可绑定插件自有存储文件", documentRef),
    buildMetric("治理卸载", "卸载时可选保留或清理数据与文件", documentRef),
  );
  sidePanel.append(panelTitle, metrics);

  hero.append(intro, sidePanel);

  const featureGrid = documentRef.createElement("div");
  featureGrid.className = "plugin-demo-dynamic-page__grid";
  for (const item of featureItems) {
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
  workspaceTitle.textContent = gridTitle;

  const workspaceSummary = documentRef.createElement("p");
  workspaceSummary.className = "plugin-demo-dynamic-page__workspace-summary";
  workspaceSummary.textContent =
    "该区域读取 plugin-demo-dynamic 安装 SQL 创建的数据表，并通过动态插件后端路由完成新增、编辑、删除与附件下载。禁用插件不会清空这些数据。";

  workspaceHeadingBlock.append(workspaceTitle, workspaceSummary);

  const toolbar = documentRef.createElement("div");
  toolbar.className = "plugin-demo-dynamic-page__toolbar";

  const addButton = documentRef.createElement("button");
  addButton.type = "button";
  addButton.className = "plugin-demo-dynamic-page__button";
  addButton.setAttribute("data-testid", "plugin-demo-dynamic-record-add");
  addButton.textContent = "新增记录";

  const reloadButton = documentRef.createElement("button");
  reloadButton.type = "button";
  reloadButton.className = "plugin-demo-dynamic-page__ghost-button";
  reloadButton.textContent = "刷新列表";

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
  modalSummary.textContent =
    "记录内容会写入 plugin-demo-dynamic 安装 SQL 创建的数据表；若上传附件，文件会存入该插件授权的 storage path。";
  modalHeader.append(modalTitle, modalSummary);

  const modalBody = documentRef.createElement("div");
  modalBody.className = "plugin-demo-dynamic-page__modal-body";

  const titleField = documentRef.createElement("label");
  titleField.className = "plugin-demo-dynamic-page__field";
  const titleLabel = documentRef.createElement("span");
  titleLabel.className = "plugin-demo-dynamic-page__field-label";
  titleLabel.textContent = "记录标题";
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
  contentLabel.textContent = "记录内容";
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
  attachmentLabel.textContent = "附件";
  const attachmentHint = documentRef.createElement("div");
  attachmentHint.className = "plugin-demo-dynamic-page__field-hint";
  attachmentHint.textContent =
    "支持上传一个示例附件。卸载插件时若勾选清理存储数据，附件文件也会一并删除。";
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
  removeAttachmentText.textContent = "提交时移除当前附件";
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
  cancelButton.textContent = "取消";

  const submitButton = documentRef.createElement("button");
  submitButton.type = "button";
  submitButton.className = "plugin-demo-dynamic-page__button";
  submitButton.setAttribute("data-testid", "plugin-demo-dynamic-record-submit");
  submitButton.textContent = "保存";

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
    summary.textContent = `第 ${state.pageNum} / ${totalPages} 页，显示第 ${rangeStart}-${rangeEnd} 条，共 ${state.total} 条`;
    pagination.append(summary);

    if (totalPages <= 1) {
      return pagination;
    }

    const controls = documentRef.createElement("div");
    controls.className = "plugin-demo-dynamic-page__pagination-controls";

    const previousButton = buildPaginationButton("上一页", state.pageNum - 1, {
      disabled: state.loading || state.pageNum <= 1,
    });
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

    const nextButton = buildPaginationButton("下一页", state.pageNum + 1, {
      disabled: state.loading || state.pageNum >= totalPages,
    });
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
      loading.textContent = "正在加载示例记录...";
      tableWrap.append(loading);
      return;
    }

    if (state.records.length === 0) {
      const empty = documentRef.createElement("div");
      empty.className = "plugin-demo-dynamic-page__empty";
      empty.setAttribute("data-testid", "plugin-demo-dynamic-record-empty");
      empty.innerHTML = `<strong>暂无示例记录</strong>${emptyText}`;
      tableWrap.append(empty);
      return;
    }

    const table = documentRef.createElement("table");
    table.className = "plugin-demo-dynamic-page__table";

    const thead = documentRef.createElement("thead");
    thead.innerHTML = `
      <tr>
        <th style="width: 24%">标题</th>
        <th style="width: 30%">内容</th>
        <th style="width: 18%">附件</th>
        <th style="width: 16%">更新时间</th>
        <th style="width: 12%">操作</th>
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
      titleBlock.textContent = record.title;
      const createdMeta = documentRef.createElement("div");
      createdMeta.className = "plugin-demo-dynamic-page__cell-meta";
      createdMeta.textContent = `创建时间: ${record.createdAt || "-"}`;
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
        downloadButton.textContent = record.attachmentName || "下载附件";
        downloadButton.addEventListener("click", () => {
          void downloadAttachment(record);
        });
        attachmentCell.append(downloadButton);
      } else {
        attachmentCell.textContent = "无附件";
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
      editButton.textContent = "编辑";
      editButton.disabled = state.submitting;
      editButton.addEventListener("click", () => openModal(record));

      const deleteButton = documentRef.createElement("button");
      deleteButton.type = "button";
      deleteButton.className = "plugin-demo-dynamic-page__inline-button";
      deleteButton.setAttribute("data-variant", "danger");
      deleteButton.textContent = "删除";
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
    modalTitle.textContent = isEditing ? "编辑示例记录" : "新增示例记录";
    submitButton.textContent = state.submitting
      ? "保存中..."
      : isEditing
        ? "保存修改"
        : "创建记录";

    if (state.selectedFile) {
      selectedAttachment.textContent = `待上传附件: ${state.selectedFile.name}`;
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
        existingAttachment.textContent = `当前附件: ${record.attachmentName}`;
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
        headers: createJSONHeaders(accessToken, options.headers || {}),
      },
    );
    if (!response.ok) {
      throw new Error(await parseErrorMessage(response));
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
        error instanceof Error ? error.message : "示例记录加载失败",
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
      throw new Error("记录标题不能为空");
    }
    if (state.selectedFile) {
      payload.attachmentName = state.selectedFile.name;
      payload.attachmentContentBase64 = await readFileAsBase64(
        state.selectedFile,
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
      setFeedback("success", isEditing ? "示例记录已更新" : "示例记录已创建");
      await fetchRecords({ pageNum: 1, resetFeedback: false });
    } catch (error) {
      modalFeedback.hidden = false;
      modalFeedback.setAttribute("data-kind", "warn");
      modalFeedback.textContent =
        error instanceof Error ? error.message : "示例记录保存失败";
    } finally {
      state.submitting = false;
      updateActionState();
      renderTable();
      renderModal();
    }
  }

  async function deleteRecord(record) {
    const confirmed = window.confirm(`确认删除记录“${record.title}”吗？`);
    if (!confirmed) {
      return;
    }
    clearFeedback();
    try {
      await requestJSON(`${apiBasePath}/demo-records/${record.id}`, {
        method: "DELETE",
      });
      setFeedback("success", "示例记录已删除");
      await fetchRecords({ resetFeedback: false });
    } catch (error) {
      setFeedback(
        "error",
        error instanceof Error ? error.message : "示例记录删除失败",
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
          headers: createJSONHeaders(accessToken),
        },
      );
      if (!response.ok) {
        throw new Error(await parseErrorMessage(response));
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
      setFeedback("success", "附件下载已开始");
    } catch (error) {
      setFeedback(
        "error",
        error instanceof Error ? error.message : "附件下载失败",
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
