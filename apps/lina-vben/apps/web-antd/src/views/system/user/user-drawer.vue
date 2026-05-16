<script setup lang="ts">
import { computed, ref } from 'vue';

import { useVbenDrawer } from '@vben/common-ui';
import { $t } from '@vben/locales';
import { useUserStore } from '@vben/stores';

import { message } from 'ant-design-vue';

import { useVbenForm } from '#/adapter/form';
import { roleOptions } from '#/api/system/role';
import {
  getDeptTree,
  getUserPostOptions,
  userAdd,
  userInfo,
  userUpdate,
} from '#/api/system/user';
import { useDictStore } from '#/store/dict';
import { useTenantStore } from '#/store/tenant';

import { drawerSchema } from './data';
import { loadUserTenantOptions } from './tenant-options';

const emit = defineEmits<{ success: [] }>();

const isEdit = ref(false);
const orgEnabled = ref(false);
const tenantEnabled = ref(false);
const userId = ref<number>(0);
const title = computed(() =>
  isEdit.value
    ? $t('pages.system.user.drawer.editTitle')
    : $t('pages.system.user.drawer.createTitle'),
);

const dictStore = useDictStore();
const tenantStore = useTenantStore();
const userStore = useUserStore();

function addFullName(
  tree: any[],
  labelField = 'label',
  separator = ' / ',
  parentName = '',
) {
  for (const node of tree) {
    node.fullName = parentName
      ? `${parentName}${separator}${node[labelField]}`
      : node[labelField];
    if (node.children?.length) {
      addFullName(node.children, labelField, separator, node.fullName);
    }
  }
  return tree;
}

const [Form, formApi] = useVbenForm({
  commonConfig: {
    formItemClass: 'col-span-2',
    componentProps: {
      class: 'w-full',
    },
    labelWidth: 80,
  },
  schema: drawerSchema(false, false),
  showDefaultActions: false,
  wrapperClass: 'grid-cols-2',
});

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
      componentProps: {
        'data-testid': 'user-drawer-tenant-select',
        allowClear: true,
        disabled: !tenantStore.isPlatform,
        mode: 'multiple',
        optionFilterProp: 'label',
        options,
        placeholder: $t('pages.multiTenant.placeholders.selectTenant'),
        showSearch: true,
      },
      fieldName: 'tenantIds',
    },
  ]);
}

async function setupPostOptions(deptId: number | string) {
  const postList = await getUserPostOptions(Number(deptId));
  const options = postList.map((item) => ({
    label: item.postName,
    value: item.postId,
  }));
  const placeholder =
    options.length > 0
      ? $t('pages.system.user.placeholders.selectDept')
      : $t('pages.system.user.messages.noPosts');
  formApi.updateSchema([
    {
      componentProps: { options, placeholder },
      fieldName: 'postIds',
    },
  ]);
}

async function setupRoleOptions() {
  const roleList = await roleOptions();
  const options = roleList.map((item) => ({
    label: item.name,
    value: item.id,
  }));
  formApi.updateSchema([
    {
      componentProps: {
        options,
        placeholder: $t('pages.system.user.placeholders.selectRole'),
      },
      fieldName: 'roleIds',
    },
  ]);
}

async function setupDeptSelect() {
  const deptTree = await getDeptTree();
  addFullName(deptTree, 'label', ' / ');
  formApi.updateSchema([
    {
      componentProps: (formModel: any) => ({
        class: 'w-full',
        fieldNames: {
          key: 'id',
          value: 'id',
          children: 'children',
        },
        async onSelect(deptId: number | string) {
          await setupPostOptions(deptId);
          formModel.postIds = [];
        },
        placeholder: $t('pages.system.user.placeholders.selectDept'),
        showSearch: true,
        treeData: deptTree,
        treeDefaultExpandAll: true,
        treeLine: { showLeafIcon: false },
        treeNodeFilterProp: 'label',
        treeNodeLabelProp: 'fullName',
      }),
      fieldName: 'deptId',
    },
  ]);

  return deptTree;
}

const [Drawer, drawerApi] = useVbenDrawer({
  async onOpenChange(open) {
    if (!open) {
      if (orgEnabled.value) {
        formApi.updateSchema([
          {
            componentProps: {
              options: [],
              placeholder: $t('pages.system.user.placeholders.selectDeptFirst'),
            },
            fieldName: 'postIds',
          },
        ]);
      }
      return;
    }

    drawerApi.setState({ loading: true });

    try {
      const data = drawerApi.getData<{
        isEdit: boolean;
        orgEnabled?: boolean;
        tenantEnabled?: boolean;
        row?: any;
      }>();
      isEdit.value = data?.isEdit ?? false;
      orgEnabled.value = data?.orgEnabled ?? false;
      tenantEnabled.value = data?.tenantEnabled ?? false;

      formApi.setState({
        schema: drawerSchema(
          isEdit.value,
          orgEnabled.value,
          tenantEnabled.value,
          !tenantStore.isPlatform,
        ),
      });

      const setupTasks: Promise<unknown>[] = [
        setupRoleOptions(),
        setupTenantOptions(),
        dictStore.getDictOptionsAsync('sys_normal_disable'),
      ];
      if (orgEnabled.value) {
        setupTasks.push(setupDeptSelect());
      }

      const setupResults = await Promise.all(setupTasks);
      const statusOptions = setupResults[2] as Awaited<
        ReturnType<typeof dictStore.getDictOptionsAsync>
      >;
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

      if (isEdit.value && data?.row) {
        userId.value = data.row.id;
        const user = await userInfo(data.row.id);
        const values: Record<string, any> = {
          username: user.username,
          nickname: user.nickname,
          email: user.email,
          phone: user.phone,
          sex: user.sex,
          status: user.status,
          remark: user.remark,
          roleIds: user.roleIds,
        };
        if (tenantEnabled.value) {
          values.tenantIds =
            tenantStore.isPlatform || user.tenantIds?.length
              ? (user.tenantIds ?? [])
              : tenantStore.currentTenant
                ? [tenantStore.currentTenant.id]
                : [];
        }

        if (orgEnabled.value) {
          values.deptId = user.deptId;
          values.postIds = user.postIds;
        }

        await formApi.setValues(values);
        if (orgEnabled.value && user.deptId) {
          await setupPostOptions(user.deptId);
        }
      } else {
        userId.value = 0;
        await formApi.resetForm();
        if (tenantEnabled.value && !tenantStore.isPlatform) {
          await formApi.setValues({
            tenantIds: tenantStore.currentTenant
              ? [tenantStore.currentTenant.id]
              : [],
          });
        }
      }
    } finally {
      drawerApi.setState({ loading: false });
    }
  },
  async onConfirm() {
    const values = await formApi.getValues();
    if (!orgEnabled.value) {
      Reflect.deleteProperty(values, 'deptId');
      Reflect.deleteProperty(values, 'postIds');
    }
    if (!tenantEnabled.value) {
      Reflect.deleteProperty(values, 'tenantIds');
    } else if (!tenantStore.isPlatform) {
      values.tenantIds = tenantStore.currentTenant
        ? [tenantStore.currentTenant.id]
        : [];
    }

    if (isEdit.value) {
      await userUpdate({
        id: userId.value,
        ...values,
      });
      message.success($t('pages.common.updateSuccess'));
    } else {
      await userAdd(values as any);
      message.success($t('pages.common.createSuccess'));
    }

    emit('success');
    drawerApi.close();
  },
});
</script>

<template>
  <Drawer :title="title" class="w-[600px]">
    <Form />
  </Drawer>
</template>
