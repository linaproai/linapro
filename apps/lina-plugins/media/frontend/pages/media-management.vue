<script lang="ts">
export const pluginPageMeta = {
  routePath: "/media",
  title: "媒体管理",
};
</script>

<script setup lang="ts">
import type {
  MediaAlias,
  MediaBindingKind,
  MediaDeviceBinding,
  MediaStrategy,
  MediaTenantBinding,
  MediaTenantDeviceBinding,
  MediaTenantWhite,
} from "./media-client";

import { nextTick, ref } from "vue";

import { useAccess } from "@vben/access";
import { Page, useVbenModal } from "@vben/common-ui";
import { IconifyIcon } from "@vben/icons";

import {
  Descriptions,
  DescriptionsItem,
  Form,
  FormItem,
  Input,
  message,
  Popconfirm,
  Select,
  Space,
  TabPane,
  Tabs,
  Tag,
} from "ant-design-vue";

import { useVbenVxeGrid } from "#/adapter/vxe-table";

import AliasModal from "./components/alias-modal.vue";
import BindingModal from "./components/binding-modal.vue";
import StrategyModal from "./components/strategy-modal.vue";
import TenantWhiteModal from "./components/tenant-white-modal.vue";
import {
  deleteMediaAlias,
  deleteMediaDeviceBinding,
  deleteMediaStrategy,
  deleteMediaTenantBinding,
  deleteMediaTenantDeviceBinding,
  deleteMediaTenantWhite,
  listMediaAliases,
  listMediaDeviceBindings,
  listMediaStrategies,
  listMediaTenantBindings,
  listMediaTenantDeviceBindings,
  listMediaTenantWhites,
  resolveMediaStrategy,
  setGlobalMediaStrategy,
  updateMediaStrategyEnable,
} from "./media-client";

const { hasAccessByCodes } = useAccess();

const accessCodes = {
  add: "media:management:add",
  edit: "media:management:edit",
  remove: "media:management:remove",
} as const;

const switchOptions = [
  { label: "开启", value: 1 },
  { label: "关闭", value: 2 },
];
const globalOptions = [
  { label: "是", value: 1 },
  { label: "否", value: 2 },
];

const mediaTabs = [
  {
    key: "strategies",
    label: "策略管理",
    icon: "lucide:sliders-horizontal",
  },
  {
    key: "deviceBindings",
    label: "设备绑定",
    icon: "lucide:cctv",
  },
  {
    key: "tenantBindings",
    label: "租户绑定",
    icon: "lucide:building-2",
  },
  {
    key: "tenantDeviceBindings",
    label: "租户设备绑定",
    icon: "lucide:network",
  },
  {
    key: "resolve",
    label: "策略解析",
    icon: "lucide:radar",
  },
  {
    key: "aliases",
    label: "流别名",
    icon: "lucide:route",
  },
  {
    key: "tenantWhites",
    label: "租户白名单",
    icon: "lucide:shield-check",
  },
] as const;

const activeTab = ref("strategies");
const resolveForm = ref({
  tenantId: "",
  deviceId: "",
});
const resolveResult = ref<Awaited<
  ReturnType<typeof resolveMediaStrategy>
> | null>(null);
const resolving = ref(false);

const [StrategyModalRef, strategyModalApi] = useVbenModal({
  connectedComponent: StrategyModal,
});
const [BindingModalRef, bindingModalApi] = useVbenModal({
  connectedComponent: BindingModal,
});
const [AliasModalRef, aliasModalApi] = useVbenModal({
  connectedComponent: AliasModal,
});
const [TenantWhiteModalRef, tenantWhiteModalApi] = useVbenModal({
  connectedComponent: TenantWhiteModal,
});

const [StrategyGrid, strategyGridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: "Input",
        fieldName: "keyword",
        label: "策略名称",
      },
      {
        component: "Select",
        componentProps: {
          allowClear: true,
          options: switchOptions,
        },
        fieldName: "enable",
        label: "启用状态",
      },
      {
        component: "Select",
        componentProps: {
          allowClear: true,
          options: globalOptions,
        },
        fieldName: "global",
        label: "全局策略",
      },
    ],
    commonConfig: {
      componentProps: {
        allowClear: true,
      },
      labelWidth: 88,
    },
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3",
  },
  gridOptions: {
    columns: [
      {
        field: "name",
        minWidth: 180,
        title: "策略名称",
      },
      {
        field: "strategy",
        minWidth: 260,
        showOverflow: "tooltip",
        title: "策略内容",
      },
      {
        field: "enable",
        slots: { default: "enable" },
        title: "启用状态",
        width: 110,
      },
      {
        field: "global",
        slots: { default: "global" },
        title: "全局策略",
        width: 110,
      },
      {
        field: "updateTime",
        minWidth: 170,
        title: "更新时间",
      },
      {
        field: "action",
        fixed: "right",
        slots: { default: "action" },
        title: "操作",
        width: 250,
      },
    ],
    height: "100%",
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues: Record<string, any> = {},
        ) => {
          return await listMediaStrategies({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: {
      keyField: "id",
    },
    id: "media-strategy-grid",
  },
});

const [DeviceBindingGrid, deviceBindingGridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: "Input",
        fieldName: "keyword",
        label: "设备国标 ID",
      },
    ],
    commonConfig: {
      componentProps: {
        allowClear: true,
      },
      labelWidth: 96,
    },
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3",
  },
  gridOptions: {
    columns: [
      {
        field: "deviceId",
        minWidth: 220,
        title: "设备国标 ID",
      },
      {
        field: "strategyName",
        minWidth: 180,
        title: "媒体策略",
      },
      {
        field: "action",
        fixed: "right",
        slots: { default: "deviceBindingAction" },
        title: "操作",
        width: 170,
      },
    ],
    height: "100%",
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues: Record<string, any> = {},
        ) => {
          return await listMediaDeviceBindings({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: {
      keyField: "rowKey",
    },
    id: "media-device-binding-grid",
  },
});

const [TenantBindingGrid, tenantBindingGridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: "Input",
        fieldName: "keyword",
        label: "租户 ID",
      },
    ],
    commonConfig: {
      componentProps: {
        allowClear: true,
      },
      labelWidth: 96,
    },
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3",
  },
  gridOptions: {
    columns: [
      {
        field: "tenantId",
        minWidth: 180,
        title: "租户 ID",
      },
      {
        field: "strategyName",
        minWidth: 180,
        title: "媒体策略",
      },
      {
        field: "action",
        fixed: "right",
        slots: { default: "tenantBindingAction" },
        title: "操作",
        width: 170,
      },
    ],
    height: "100%",
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues: Record<string, any> = {},
        ) => {
          return await listMediaTenantBindings({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: {
      keyField: "rowKey",
    },
    id: "media-tenant-binding-grid",
  },
});

const [TenantDeviceBindingGrid, tenantDeviceBindingGridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: "Input",
        fieldName: "keyword",
        label: "关键字",
      },
    ],
    commonConfig: {
      componentProps: {
        allowClear: true,
      },
      labelWidth: 72,
    },
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3",
  },
  gridOptions: {
    columns: [
      {
        field: "tenantId",
        minWidth: 170,
        title: "租户 ID",
      },
      {
        field: "deviceId",
        minWidth: 220,
        title: "设备国标 ID",
      },
      {
        field: "strategyName",
        minWidth: 180,
        title: "媒体策略",
      },
      {
        field: "action",
        fixed: "right",
        slots: { default: "tenantDeviceBindingAction" },
        title: "操作",
        width: 170,
      },
    ],
    height: "100%",
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues: Record<string, any> = {},
        ) => {
          return await listMediaTenantDeviceBindings({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: {
      keyField: "rowKey",
    },
    id: "media-tenant-device-binding-grid",
  },
});

const [AliasGrid, aliasGridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: "Input",
        fieldName: "keyword",
        label: "关键字",
      },
    ],
    commonConfig: {
      componentProps: {
        allowClear: true,
      },
      labelWidth: 72,
    },
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3",
  },
  gridOptions: {
    columns: [
      {
        field: "alias",
        minWidth: 180,
        title: "流别名",
      },
      {
        field: "streamPath",
        minWidth: 240,
        title: "真实流路径",
      },
      {
        field: "autoRemove",
        slots: { default: "autoRemove" },
        title: "自动移除",
        width: 110,
      },
      {
        field: "createTime",
        minWidth: 170,
        title: "创建时间",
      },
      {
        field: "action",
        fixed: "right",
        slots: { default: "aliasAction" },
        title: "操作",
        width: 170,
      },
    ],
    height: "100%",
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues: Record<string, any> = {},
        ) => {
          return await listMediaAliases({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: {
      keyField: "id",
    },
    id: "media-alias-grid",
  },
});

const [TenantWhiteGrid, tenantWhiteGridApi] = useVbenVxeGrid({
  formOptions: {
    schema: [
      {
        component: "Input",
        fieldName: "keyword",
        label: "关键字",
      },
      {
        component: "Select",
        componentProps: {
          allowClear: true,
          options: [
            { label: "开启", value: 1 },
            { label: "关闭", value: 0 },
          ],
        },
        fieldName: "enable",
        label: "启用状态",
      },
    ],
    commonConfig: {
      componentProps: {
        allowClear: true,
      },
      labelWidth: 72,
    },
    wrapperClass: "grid-cols-1 md:grid-cols-2 lg:grid-cols-3",
  },
  gridOptions: {
    columns: [
      {
        field: "tenantId",
        minWidth: 170,
        title: "租户 ID",
      },
      {
        field: "ip",
        minWidth: 150,
        title: "白名单地址",
      },
      {
        field: "description",
        minWidth: 160,
        title: "白名单描述",
      },
      {
        field: "enable",
        slots: { default: "tenantWhiteEnable" },
        title: "启用状态",
        width: 110,
      },
      {
        field: "updateTime",
        minWidth: 170,
        title: "更新时间",
      },
      {
        field: "action",
        fixed: "right",
        slots: { default: "tenantWhiteAction" },
        title: "操作",
        width: 170,
      },
    ],
    height: "100%",
    keepSource: true,
    pagerConfig: {},
    proxyConfig: {
      ajax: {
        query: async (
          { page }: { page: { currentPage: number; pageSize: number } },
          formValues: Record<string, any> = {},
        ) => {
          return await listMediaTenantWhites({
            pageNum: page.currentPage,
            pageSize: page.pageSize,
            ...formValues,
          });
        },
      },
    },
    rowConfig: {
      keyField: "rowKey",
    },
    id: "media-tenant-white-grid",
  },
});

function canAdd() {
  return hasAccessByCodes([accessCodes.add]);
}

function canEdit() {
  return hasAccessByCodes([accessCodes.edit]);
}

function canRemove() {
  return hasAccessByCodes([accessCodes.remove]);
}

function switchLabel(value: number) {
  return value === 1 ? "开启" : "关闭";
}

function activeGridApi() {
  if (activeTab.value === "deviceBindings") return deviceBindingGridApi;
  if (activeTab.value === "tenantBindings") return tenantBindingGridApi;
  if (activeTab.value === "tenantDeviceBindings")
    return tenantDeviceBindingGridApi;
  if (activeTab.value === "aliases") return aliasGridApi;
  if (activeTab.value === "tenantWhites") return tenantWhiteGridApi;
  if (activeTab.value === "strategies") return strategyGridApi;
  return null;
}

function refreshActiveGridLayout() {
  const gridApi = activeGridApi();
  if (!gridApi) return;
  void nextTick(() => {
    gridApi.grid?.recalculate?.();
  });
}

function handleTabChange(key: string | number) {
  activeTab.value = String(key);
  refreshActiveGridLayout();
}

function handleAddStrategy() {
  strategyModalApi.setData({ id: undefined });
  strategyModalApi.open();
}

function handleEditStrategy(row: MediaStrategy) {
  strategyModalApi.setData({ id: row.id });
  strategyModalApi.open();
}

async function handleSetGlobalStrategy(row: MediaStrategy) {
  await setGlobalMediaStrategy(row.id);
  message.success("全局策略已更新");
  await strategyGridApi.query();
}

async function handleToggleStrategy(row: MediaStrategy) {
  await updateMediaStrategyEnable(row.id, row.enable === 1 ? 2 : 1);
  message.success("启用状态已更新");
  await strategyGridApi.query();
}

async function handleDeleteStrategy(row: MediaStrategy) {
  await deleteMediaStrategy(row.id);
  message.success("媒体策略已删除");
  await strategyGridApi.query();
}

function handleAddBinding(kind: MediaBindingKind) {
  bindingModalApi.setData({
    deviceId: undefined,
    kind,
    strategyId: undefined,
    tenantId: undefined,
  });
  bindingModalApi.open();
}

function handleEditDeviceBinding(row: MediaDeviceBinding) {
  bindingModalApi.setData({ ...row, kind: "device" });
  bindingModalApi.open();
}

function handleEditTenantBinding(row: MediaTenantBinding) {
  bindingModalApi.setData({ ...row, kind: "tenant" });
  bindingModalApi.open();
}

function handleEditTenantDeviceBinding(row: MediaTenantDeviceBinding) {
  bindingModalApi.setData({ ...row, kind: "tenantDevice" });
  bindingModalApi.open();
}

async function handleDeleteDeviceBinding(row: MediaDeviceBinding) {
  await deleteMediaDeviceBinding(row.deviceId);
  message.success("设备策略绑定已删除");
  await deviceBindingGridApi.query();
}

async function handleDeleteTenantBinding(row: MediaTenantBinding) {
  await deleteMediaTenantBinding(row.tenantId);
  message.success("租户策略绑定已删除");
  await tenantBindingGridApi.query();
}

async function handleDeleteTenantDeviceBinding(row: MediaTenantDeviceBinding) {
  await deleteMediaTenantDeviceBinding(row.tenantId, row.deviceId);
  message.success("租户设备策略绑定已删除");
  await tenantDeviceBindingGridApi.query();
}

async function handleResolveStrategy() {
  resolving.value = true;
  try {
    resolveResult.value = await resolveMediaStrategy(resolveForm.value);
  } finally {
    resolving.value = false;
  }
}

function handleAddAlias() {
  aliasModalApi.setData({ id: undefined });
  aliasModalApi.open();
}

function handleEditAlias(row: MediaAlias) {
  aliasModalApi.setData({ id: row.id });
  aliasModalApi.open();
}

async function handleDeleteAlias(row: MediaAlias) {
  await deleteMediaAlias(row.id);
  message.success("流别名已删除");
  await aliasGridApi.query();
}

function tenantWhiteRowKey(row: Pick<MediaTenantWhite, "ip" | "tenantId">) {
  return `${row.tenantId}:${row.ip}`;
}

function handleAddTenantWhite() {
  tenantWhiteModalApi.setData({ ip: undefined, tenantId: undefined });
  tenantWhiteModalApi.open();
}

function handleEditTenantWhite(row: MediaTenantWhite) {
  tenantWhiteModalApi.setData({ ip: row.ip, tenantId: row.tenantId });
  tenantWhiteModalApi.open();
}

async function handleDeleteTenantWhite(row: MediaTenantWhite) {
  await deleteMediaTenantWhite(row.tenantId, row.ip);
  message.success("租户白名单已删除");
  await tenantWhiteGridApi.query();
}

function reloadStrategies() {
  strategyGridApi.query();
}

function reloadBindings(kind?: MediaBindingKind) {
  if (kind === "device") {
    deviceBindingGridApi.query();
  } else if (kind === "tenant") {
    tenantBindingGridApi.query();
  } else if (kind === "tenantDevice") {
    tenantDeviceBindingGridApi.query();
  }
}

function reloadAliases() {
  aliasGridApi.query();
}

function reloadTenantWhites() {
  tenantWhiteGridApi.query();
}
</script>

<template>
  <Page
    :auto-content-height="true"
    content-class="flex min-h-0 flex-col"
    data-testid="media-management-page"
  >
    <Tabs
      v-model:active-key="activeTab"
      :animated="false"
      class="media-tabs flex min-h-0 flex-1 flex-col overflow-hidden"
      @change="handleTabChange"
    >
      <template #renderTabBar="{ DefaultTabBar, ...props }">
        <div class="media-tabs-bar">
          <component :is="DefaultTabBar" v-bind="props" />
        </div>
      </template>

      <TabPane key="strategies">
        <template #tab>
          <span class="media-tab-label">
            <IconifyIcon :icon="mediaTabs[0].icon" />
            <span>{{ mediaTabs[0].label }}</span>
          </span>
        </template>
        <div class="media-tab-pane">
          <StrategyGrid
            class="min-h-0 flex-1 overflow-hidden"
            table-title="媒体策略"
          >
            <template #toolbar-tools>
              <a-button
                v-if="canAdd()"
                data-testid="media-strategy-add"
                type="primary"
                @click="handleAddStrategy"
              >
                <template #icon>
                  <IconifyIcon icon="lucide:plus" />
                </template>
                新增策略
              </a-button>
            </template>

            <template #enable="{ row }">
              <Tag :color="row.enable === 1 ? 'green' : 'default'">
                {{ switchLabel(row.enable) }}
              </Tag>
            </template>

            <template #global="{ row }">
              <Tag :color="row.global === 1 ? 'blue' : 'default'">
                {{ row.global === 1 ? "是" : "否" }}
              </Tag>
            </template>

            <template #action="{ row }">
              <Space>
                <ghost-button
                  v-if="canEdit()"
                  :data-testid="`media-strategy-edit-${row.id}`"
                  @click.stop="handleEditStrategy(row)"
                >
                  编辑
                </ghost-button>
                <ghost-button
                  v-if="canEdit()"
                  :data-testid="`media-strategy-toggle-${row.id}`"
                  @click.stop="handleToggleStrategy(row)"
                >
                  {{ row.enable === 1 ? "关闭" : "开启" }}
                </ghost-button>
                <ghost-button
                  v-if="canEdit() && row.global !== 1"
                  :data-testid="`media-strategy-global-${row.id}`"
                  @click.stop="handleSetGlobalStrategy(row)"
                >
                  设为全局
                </ghost-button>
                <Popconfirm
                  v-if="canRemove()"
                  title="确认删除该媒体策略？已被绑定引用的策略不能删除。"
                  @confirm="handleDeleteStrategy(row)"
                >
                  <ghost-button
                    danger
                    :data-testid="`media-strategy-delete-${row.id}`"
                    @click.stop=""
                  >
                    删除
                  </ghost-button>
                </Popconfirm>
              </Space>
            </template>
          </StrategyGrid>
        </div>
      </TabPane>

      <TabPane key="deviceBindings">
        <template #tab>
          <span class="media-tab-label">
            <IconifyIcon :icon="mediaTabs[1].icon" />
            <span>{{ mediaTabs[1].label }}</span>
          </span>
        </template>
        <div class="media-tab-pane">
          <DeviceBindingGrid
            class="min-h-0 flex-1 overflow-hidden"
            table-title="设备策略绑定"
          >
            <template #toolbar-tools>
              <a-button
                v-if="canEdit()"
                data-testid="media-device-binding-add"
                type="primary"
                @click="handleAddBinding('device')"
              >
                <template #icon>
                  <IconifyIcon icon="lucide:link" />
                </template>
                新增设备绑定
              </a-button>
            </template>

            <template #deviceBindingAction="{ row }">
              <Space>
                <ghost-button
                  v-if="canEdit()"
                  :data-testid="`media-device-binding-edit-${row.rowKey}`"
                  @click.stop="handleEditDeviceBinding(row)"
                >
                  编辑
                </ghost-button>
                <Popconfirm
                  v-if="canRemove()"
                  title="确认删除该设备策略绑定？"
                  @confirm="handleDeleteDeviceBinding(row)"
                >
                  <ghost-button
                    danger
                    :data-testid="`media-device-binding-delete-${row.rowKey}`"
                    @click.stop=""
                  >
                    删除
                  </ghost-button>
                </Popconfirm>
              </Space>
            </template>
          </DeviceBindingGrid>
        </div>
      </TabPane>

      <TabPane key="tenantBindings">
        <template #tab>
          <span class="media-tab-label">
            <IconifyIcon :icon="mediaTabs[2].icon" />
            <span>{{ mediaTabs[2].label }}</span>
          </span>
        </template>
        <div class="media-tab-pane">
          <TenantBindingGrid
            class="min-h-0 flex-1 overflow-hidden"
            table-title="租户策略绑定"
          >
            <template #toolbar-tools>
              <a-button
                v-if="canEdit()"
                data-testid="media-tenant-binding-add"
                type="primary"
                @click="handleAddBinding('tenant')"
              >
                <template #icon>
                  <IconifyIcon icon="lucide:link" />
                </template>
                新增租户绑定
              </a-button>
            </template>

            <template #tenantBindingAction="{ row }">
              <Space>
                <ghost-button
                  v-if="canEdit()"
                  :data-testid="`media-tenant-binding-edit-${row.rowKey}`"
                  @click.stop="handleEditTenantBinding(row)"
                >
                  编辑
                </ghost-button>
                <Popconfirm
                  v-if="canRemove()"
                  title="确认删除该租户策略绑定？"
                  @confirm="handleDeleteTenantBinding(row)"
                >
                  <ghost-button
                    danger
                    :data-testid="`media-tenant-binding-delete-${row.rowKey}`"
                    @click.stop=""
                  >
                    删除
                  </ghost-button>
                </Popconfirm>
              </Space>
            </template>
          </TenantBindingGrid>
        </div>
      </TabPane>

      <TabPane key="tenantDeviceBindings">
        <template #tab>
          <span class="media-tab-label">
            <IconifyIcon :icon="mediaTabs[3].icon" />
            <span>{{ mediaTabs[3].label }}</span>
          </span>
        </template>
        <div class="media-tab-pane">
          <TenantDeviceBindingGrid
            class="min-h-0 flex-1 overflow-hidden"
            table-title="租户设备策略绑定"
          >
            <template #toolbar-tools>
              <a-button
                v-if="canEdit()"
                data-testid="media-tenant-device-binding-add"
                type="primary"
                @click="handleAddBinding('tenantDevice')"
              >
                <template #icon>
                  <IconifyIcon icon="lucide:link" />
                </template>
                新增租户设备绑定
              </a-button>
            </template>

            <template #tenantDeviceBindingAction="{ row }">
              <Space>
                <ghost-button
                  v-if="canEdit()"
                  :data-testid="`media-tenant-device-binding-edit-${row.rowKey}`"
                  @click.stop="handleEditTenantDeviceBinding(row)"
                >
                  编辑
                </ghost-button>
                <Popconfirm
                  v-if="canRemove()"
                  title="确认删除该租户设备策略绑定？"
                  @confirm="handleDeleteTenantDeviceBinding(row)"
                >
                  <ghost-button
                    danger
                    :data-testid="`media-tenant-device-binding-delete-${row.rowKey}`"
                    @click.stop=""
                  >
                    删除
                  </ghost-button>
                </Popconfirm>
              </Space>
            </template>
          </TenantDeviceBindingGrid>
        </div>
      </TabPane>

      <TabPane key="resolve">
        <template #tab>
          <span class="media-tab-label">
            <IconifyIcon :icon="mediaTabs[4].icon" />
            <span>{{ mediaTabs[4].label }}</span>
          </span>
        </template>
        <div class="media-tab-pane">
          <div
            class="flex-shrink-0 border border-solid border-gray-200 bg-white p-4 dark:border-gray-700 dark:bg-gray-900"
          >
            <Form layout="inline">
              <FormItem label="租户 ID">
                <Input
                  v-model:value="resolveForm.tenantId"
                  allow-clear
                  placeholder="tenant-a"
                />
              </FormItem>
              <FormItem label="设备国标 ID">
                <Input
                  v-model:value="resolveForm.deviceId"
                  allow-clear
                  placeholder="34020000001320000001"
                />
              </FormItem>
              <FormItem>
                <a-button
                  :loading="resolving"
                  type="primary"
                  @click="handleResolveStrategy"
                >
                  解析生效策略
                </a-button>
              </FormItem>
            </Form>
            <Descriptions
              v-if="resolveResult"
              :column="4"
              bordered
              class="mt-4"
              size="small"
            >
              <DescriptionsItem label="匹配结果">
                <Tag :color="resolveResult.matched ? 'green' : 'default'">
                  {{ resolveResult.matched ? "已匹配" : "未匹配" }}
                </Tag>
              </DescriptionsItem>
              <DescriptionsItem label="来源">
                {{ resolveResult.sourceLabel }}
              </DescriptionsItem>
              <DescriptionsItem label="策略 ID">
                {{ resolveResult.strategyId || "-" }}
              </DescriptionsItem>
              <DescriptionsItem label="策略名称">
                {{ resolveResult.strategyName || "-" }}
              </DescriptionsItem>
            </Descriptions>
          </div>
        </div>
      </TabPane>

      <TabPane key="aliases">
        <template #tab>
          <span class="media-tab-label">
            <IconifyIcon :icon="mediaTabs[5].icon" />
            <span>{{ mediaTabs[5].label }}</span>
          </span>
        </template>
        <div class="media-tab-pane">
          <AliasGrid
            class="min-h-0 flex-1 overflow-hidden"
            table-title="流别名"
          >
            <template #toolbar-tools>
              <a-button
                v-if="canAdd()"
                data-testid="media-alias-add"
                type="primary"
                @click="handleAddAlias"
              >
                <template #icon>
                  <IconifyIcon icon="lucide:plus" />
                </template>
                新增别名
              </a-button>
            </template>

            <template #autoRemove="{ row }">
              <Tag :color="row.autoRemove === 1 ? 'orange' : 'default'">
                {{ row.autoRemove === 1 ? "是" : "否" }}
              </Tag>
            </template>

            <template #aliasAction="{ row }">
              <Space>
                <ghost-button
                  v-if="canEdit()"
                  :data-testid="`media-alias-edit-${row.id}`"
                  @click.stop="handleEditAlias(row)"
                >
                  编辑
                </ghost-button>
                <Popconfirm
                  v-if="canRemove()"
                  title="确认删除该流别名？"
                  @confirm="handleDeleteAlias(row)"
                >
                  <ghost-button
                    danger
                    :data-testid="`media-alias-delete-${row.id}`"
                    @click.stop=""
                  >
                    删除
                  </ghost-button>
                </Popconfirm>
              </Space>
            </template>
          </AliasGrid>
        </div>
      </TabPane>

      <TabPane key="tenantWhites">
        <template #tab>
          <span class="media-tab-label">
            <IconifyIcon :icon="mediaTabs[6].icon" />
            <span>{{ mediaTabs[6].label }}</span>
          </span>
        </template>
        <div class="media-tab-pane">
          <TenantWhiteGrid
            class="min-h-0 flex-1 overflow-hidden"
            table-title="租户白名单"
          >
            <template #toolbar-tools>
              <a-button
                v-if="canAdd()"
                data-testid="media-tenant-white-add"
                type="primary"
                @click="handleAddTenantWhite"
              >
                <template #icon>
                  <IconifyIcon icon="lucide:shield-plus" />
                </template>
                新增白名单
              </a-button>
            </template>

            <template #tenantWhiteEnable="{ row }">
              <Tag :color="row.enable === 1 ? 'green' : 'default'">
                {{ row.enable === 1 ? "开启" : "关闭" }}
              </Tag>
            </template>

            <template #tenantWhiteAction="{ row }">
              <Space>
                <ghost-button
                  v-if="canEdit()"
                  :data-testid="`media-tenant-white-edit-${tenantWhiteRowKey(row)}`"
                  @click.stop="handleEditTenantWhite(row)"
                >
                  编辑
                </ghost-button>
                <Popconfirm
                  v-if="canRemove()"
                  title="确认删除该租户白名单？"
                  @confirm="handleDeleteTenantWhite(row)"
                >
                  <ghost-button
                    danger
                    :data-testid="`media-tenant-white-delete-${tenantWhiteRowKey(row)}`"
                    @click.stop=""
                  >
                    删除
                  </ghost-button>
                </Popconfirm>
              </Space>
            </template>
          </TenantWhiteGrid>
        </div>
      </TabPane>
    </Tabs>

    <StrategyModalRef @reload="reloadStrategies" />
    <BindingModalRef @reload="reloadBindings" />
    <AliasModalRef @reload="reloadAliases" />
    <TenantWhiteModalRef @reload="reloadTenantWhites" />
  </Page>
</template>

<style scoped>
.media-tabs-bar {
  flex: 0 0 auto;
  padding: 6px;
  margin-bottom: 12px;
  overflow-x: auto;
  background: #f8fafc;
  border: 1px solid #e2e8f0;
  border-radius: 8px;
  box-shadow: 0 1px 2px rgb(15 23 42 / 6%);
}

.media-tabs :deep(.ant-tabs-nav) {
  flex: 0 0 auto;
  min-width: max-content;
  margin: 0;
}

.media-tabs :deep(.ant-tabs-nav::before) {
  display: none;
}

.media-tabs :deep(.ant-tabs-nav-wrap) {
  min-height: 34px;
}

.media-tabs :deep(.ant-tabs-nav-list) {
  gap: 4px;
}

.media-tabs :deep(.ant-tabs-tab) {
  padding: 0;
  margin: 0;
  color: #475569;
  border-radius: 6px;
  transition:
    background-color 0.16s ease,
    color 0.16s ease,
    box-shadow 0.16s ease;
}

.media-tabs :deep(.ant-tabs-tab:hover) {
  color: #0f172a;
  background: #eef2f7;
}

.media-tabs :deep(.ant-tabs-tab-btn) {
  color: inherit;
}

.media-tabs :deep(.ant-tabs-tab-active) {
  color: #0f172a;
  background: #ffffff;
  box-shadow:
    0 1px 2px rgb(15 23 42 / 10%),
    inset 0 0 0 1px rgb(37 99 235 / 16%);
}

.media-tabs :deep(.ant-tabs-tab-active .ant-tabs-tab-btn) {
  color: inherit;
}

.media-tabs :deep(.ant-tabs-ink-bar) {
  display: none;
}

.media-tab-label {
  display: inline-flex;
  align-items: center;
  gap: 6px;
  min-height: 32px;
  padding: 0 12px;
  font-size: 13px;
  line-height: 32px;
  white-space: nowrap;
}

.media-tab-label :deep(.iconify) {
  width: 15px;
  height: 15px;
}

.media-tabs :deep(.ant-tabs-content-holder) {
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
}

:global(.dark) .media-tabs-bar {
  background: #111827;
  border-color: #334155;
  box-shadow: 0 1px 2px rgb(0 0 0 / 20%);
}

:global(.dark) .media-tabs :deep(.ant-tabs-tab) {
  color: #cbd5e1;
}

:global(.dark) .media-tabs :deep(.ant-tabs-tab:hover) {
  color: #f8fafc;
  background: #1e293b;
}

:global(.dark) .media-tabs :deep(.ant-tabs-tab-active) {
  color: #f8fafc;
  background: #0f172a;
  box-shadow:
    0 1px 2px rgb(0 0 0 / 24%),
    inset 0 0 0 1px rgb(96 165 250 / 26%);
}

.media-tabs :deep(.ant-tabs-content) {
  height: 100%;
  min-height: 0;
}

.media-tabs :deep(.ant-tabs-tabpane) {
  height: 100%;
  min-height: 0;
  overflow: hidden;
}

.media-tab-pane {
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: 100%;
  min-height: 0;
  overflow: hidden;
}
</style>
