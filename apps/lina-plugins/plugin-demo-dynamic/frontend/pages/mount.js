const pageTitle = "动态插件示例已生效";
const pageDescription =
  "该页面来自 plugin-demo-dynamic 的动态挂载入口，用于验证宿主主内容区展示与独立静态页面跳转。";
const buttonLabel = "打开独立页面";

const hostStyleId = "plugin-demo-dynamic-mount-style";

const featureItems = [
  {
    label: "接入方式",
    value: "宿主内嵌挂载",
    description: "通过宿主页壳在主内容区动态加载页面入口。",
  },
  {
    label: "资源形态",
    value: "WASM 静态资源",
    description: "页面资源由插件产物释放后统一托管访问。",
  },
  {
    label: "验证目标",
    value: "主窗口 + 独立页",
    description: "同时验证内嵌展示与独立页面跳转体验。",
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

    .plugin-demo-dynamic-page__button.ant-btn {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      gap: 8px;
      height: 42px;
      border: 1px solid var(--dynamic-shell-accent);
      background: var(--dynamic-shell-accent);
      color: #ffffff;
      padding-inline: 20px;
      border-radius: 10px;
      font-size: 14px;
      line-height: 1.5715;
      font-weight: 600;
      appearance: none;
      cursor: pointer;
      transition:
        background-color 0.2s ease,
        border-color 0.2s ease,
        box-shadow 0.2s ease,
        transform 0.2s ease;
      box-shadow: 0 12px 30px rgba(22, 119, 255, 0.22);
    }

    .plugin-demo-dynamic-page__button.ant-btn:hover,
    .plugin-demo-dynamic-page__button.ant-btn:focus-visible {
      border-color: #4096ff;
      background: #4096ff;
      color: #ffffff;
      cursor: pointer;
      transform: translateY(-1px);
      box-shadow: 0 16px 34px rgba(22, 119, 255, 0.28);
      outline: none;
    }

    .plugin-demo-dynamic-page__button.ant-btn:active {
      transform: translateY(0);
      box-shadow: 0 10px 24px rgba(22, 119, 255, 0.24);
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

    @media (max-width: 960px) {
      .plugin-demo-dynamic-page__hero,
      .plugin-demo-dynamic-page__grid {
        grid-template-columns: 1fr;
      }
    }

    @media (max-width: 640px) {
      .plugin-demo-dynamic-page {
        padding: 0;
      }

      .plugin-demo-dynamic-page__hero,
      .plugin-demo-dynamic-page__grid {
        padding-inline: 18px;
      }

      .plugin-demo-dynamic-page__hero {
        padding-top: 22px;
        padding-bottom: 18px;
      }

      .plugin-demo-dynamic-page__grid {
        padding-bottom: 20px;
      }

      .plugin-demo-dynamic-page__title {
        font-size: 26px;
      }

      .plugin-demo-dynamic-page__panel-metrics {
        grid-template-columns: 1fr;
      }
    }
  `;
  documentRef.head.append(styleElement);
}

function buildMetric(title, label) {
  const wrapper = document.createElement("div");
  wrapper.className = "plugin-demo-dynamic-page__metric";

  const value = document.createElement("strong");
  value.className = "plugin-demo-dynamic-page__metric-value";
  value.textContent = title;

  const text = document.createElement("span");
  text.className = "plugin-demo-dynamic-page__metric-label";
  text.textContent = label;

  wrapper.append(value, text);
  return wrapper;
}

function buildFeatureCard(item) {
  const card = document.createElement("article");
  card.className = "plugin-demo-dynamic-page__card";

  const label = document.createElement("span");
  label.className = "plugin-demo-dynamic-page__card-label";
  label.textContent = item.label;

  const value = document.createElement("h2");
  value.className = "plugin-demo-dynamic-page__card-value";
  value.textContent = item.value;

  const description = document.createElement("p");
  description.className = "plugin-demo-dynamic-page__card-description";
  description.textContent = item.description;

  card.append(label, value, description);
  return card;
}

export function mount(context) {
  const documentRef = context.container.ownerDocument;
  ensureMountStyles(documentRef);

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
  actionButton.className =
    "ant-btn ant-btn-primary plugin-demo-dynamic-page__button";
  actionButton.setAttribute("data-testid", "plugin-demo-dynamic-open-standalone");

  const buttonText = documentRef.createElement("span");
  buttonText.textContent = buttonLabel;
  actionButton.append(buttonText);

  actionButton.addEventListener("click", () => {
    const standaloneURL = new URL("./standalone.html", context.baseURL).toString();
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
    buildMetric("动态加载", "由宿主页壳动态导入并挂载页面入口"),
    buildMetric("静态托管", "独立页面资源由宿主统一托管访问"),
    buildMetric("宿主管理", "启用禁用卸载仍走统一治理链路"),
    buildMetric("双端体验", "主窗口与新窗口两种访问方式"),
  );

  sidePanel.append(panelTitle, metrics);
  hero.append(intro, sidePanel);

  const featureGrid = documentRef.createElement("div");
  featureGrid.className = "plugin-demo-dynamic-page__grid";
  for (const item of featureItems) {
    featureGrid.append(buildFeatureCard(item));
  }

  shell.append(hero, featureGrid);
  root.append(shell);
  context.container.replaceChildren(root);

  return {
    unmount() {
      root.remove();
    },
  };
}
