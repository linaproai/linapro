<script setup lang="ts">
import type { VbenFormSchema } from '#/adapter/form';

import { computed, ref } from 'vue';

import { useVbenModal } from '@vben/common-ui';
import { $t } from '@vben/locales';
import { useUserStore } from '@vben/stores';

import { message } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { roleOptions } from '#/api/system/role';
import { userBatchUpdate } from '#/api/system/user';
import { useDictStore } from '#/store/dict';
import { useTenantStore } from '#/store/tenant';

import { loadUserTenantOptions } from './tenant-options';

const emit = defineEmits<{ success: [] }>();

const dictStore = useDictStore();
const tenantStore = useTenantStore();
const userStore = useUserStore();

const selectedRows = ref<any[]>([]);
const tenantEnabled = ref(false);
const selectedCount = computed(() => selectedRows.value.length);
const batchSwitchProps = {
  class: 'w-auto',
};

const [Form, formApi] = useVbenForm({
  commonConfig: {
    formItemClass: 'col-span-2',
    labelWidth: 96,
  },
  schema: [],
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2',
});

function buildSchema() {
  const schema: VbenFormSchema[] = [
    {
      component: 'Switch',
      componentProps: batchSwitchProps,
      fieldName: 'updateStatus',
      label: $t('pages.system.user.batchEdit.fields.updateStatus'),
    },
    {
      component: 'RadioGroup',
      dependencies: {
        show: (values: any) => values.updateStatus === true,
        triggerFields: ['updateStatus'],
      },
      fieldName: 'status',
      label: $t('pages.common.status'),
      componentProps: {
        buttonStyle: 'solid',
        optionType: 'button',
      },
    },
    {
      component: 'Switch',
      componentProps: batchSwitchProps,
      fieldName: 'updateRoles',
      label: $t('pages.system.user.batchEdit.fields.updateRoles'),
    },
    {
      component: 'Select',
      dependencies: {
        show: (values: any) => values.updateRoles === true,
        triggerFields: ['updateRoles'],
      },
      fieldName: 'roleIds',
      label: $t('pages.fields.roles'),
      help: $t('pages.system.user.batchEdit.help.roles'),
      componentProps: {
        'data-testid': 'user-batch-role-select',
        allowClear: true,
        class: 'w-full',
        mode: 'multiple',
        optionFilterProp: 'label',
        placeholder: $t('pages.system.user.placeholders.selectRole'),
        showSearch: true,
      },
    },
  ];

  if (tenantEnabled.value) {
    schema.push(
      {
        component: 'Switch',
        componentProps: batchSwitchProps,
        fieldName: 'updateTenant',
        label: $t('pages.system.user.batchEdit.fields.updateTenant'),
      },
      {
        component: 'Select',
        dependencies: {
          show: (values: any) => values.updateTenant === true,
          triggerFields: ['updateTenant'],
        },
        fieldName: 'tenantIds',
        label: $t('pages.system.user.labels.tenantMemberships'),
        help: $t('pages.system.user.batchEdit.help.tenants'),
        componentProps: {
          'data-testid': 'user-batch-tenant-select',
          allowClear: true,
          class: 'w-full',
          disabled: !tenantStore.isPlatform,
          mode: 'multiple',
          optionFilterProp: 'label',
          placeholder: $t('pages.multiTenant.placeholders.selectTenant'),
          showSearch: true,
        },
      },
    );
  }

  return schema;
}

async function setupStatusOptions() {
  const statusOptions =
    await dictStore.getDictOptionsAsync('sys_normal_disable');
  formApi.updateSchema([
    {
      fieldName: 'status',
      componentProps: {
        options: statusOptions.map((d) => ({
          label: d.label,
          value: Number(d.value),
        })),
      },
    },
  ]);
}

async function setupRoleOptions() {
  const roles = await roleOptions();
  formApi.updateSchema([
    {
      fieldName: 'roleIds',
      componentProps: {
        options: roles.map((item) => ({
          label: item.name,
          value: item.id,
        })),
      },
    },
  ]);
}

async function setupTenantOptions() {
  if (!tenantEnabled.value) {
    return;
  }
  const options = await loadUserTenantOptions({
    currentTenant: tenantStore.currentTenant,
    isPlatform: tenantStore.isPlatform,
    tenants: tenantStore.tenants,
    userId: Number(userStore.userInfo?.userId || 0),
  });
  formApi.updateSchema([
    {
      fieldName: 'tenantIds',
      componentProps: {
        options,
      },
    },
  ]);
}

const [Modal, modalApi] = useVbenModal({
  async onOpenChange(open) {
    if (!open) {
      return;
    }

    modalApi.setState({ loading: true });
    try {
      const data = modalApi.getData<{
        rows?: any[];
        tenantEnabled?: boolean;
      }>();
      selectedRows.value = data?.rows ?? [];
      tenantEnabled.value = data?.tenantEnabled ?? false;

      formApi.setState({ schema: buildSchema() });
      await formApi.resetForm();
      await formApi.setValues({
        roleIds: [],
        status: 1,
        tenantIds: tenantStore.isPlatform
          ? []
          : tenantStore.currentTenant
            ? [tenantStore.currentTenant.id]
            : [],
        updateRoles: false,
        updateStatus: false,
        updateTenant: false,
      });

      await Promise.all([
        setupStatusOptions(),
        setupRoleOptions(),
        setupTenantOptions(),
      ]);
    } finally {
      modalApi.setState({ loading: false });
    }
  },
  async onConfirm() {
    const values = await formApi.getValues();
    const updateStatus = values.updateStatus === true;
    const updateRoles = values.updateRoles === true;
    const updateTenant = tenantEnabled.value && values.updateTenant === true;
    if (!updateStatus && !updateRoles && !updateTenant) {
      message.warning($t('pages.system.user.batchEdit.messages.noFields'));
      return false;
    }
    if (updateStatus && values.status === undefined) {
      message.warning($t('pages.system.user.batchEdit.messages.statusRequired'));
      return false;
    }
    if (updateRoles && updateTenant) {
      message.warning(
        $t('pages.system.user.batchEdit.messages.roleTenantConflict'),
      );
      return false;
    }

    const ids = selectedRows.value.map((row) => Number(row.id));
    const tenantIds =
      updateTenant && !tenantStore.isPlatform
        ? tenantStore.currentTenant
          ? [tenantStore.currentTenant.id]
          : []
        : ((values.tenantIds ?? []) as number[]);

    await userBatchUpdate({
      ids,
      roleIds: updateRoles ? ((values.roleIds ?? []) as number[]) : undefined,
      status: updateStatus ? Number(values.status) : undefined,
      tenantIds: updateTenant ? tenantIds : undefined,
      updateRoles,
      updateStatus,
      updateTenant,
    });
    message.success($t('pages.system.user.batchEdit.messages.success'));
    emit('success');
    modalApi.close();
  },
});
</script>

<template>
  <Modal
    :title="$t('pages.system.user.batchEdit.title')"
    class="w-[560px]"
  >
    <div class="mb-4 text-sm text-muted-foreground">
      {{
        $t('pages.system.user.batchEdit.selectedCount', {
          count: selectedCount,
        })
      }}
    </div>
    <Form />
  </Modal>
</template>
