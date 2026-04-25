<script setup lang="ts">
import type { SysUser } from '#/api/system/user';
import type { VbenFormSchema } from '#/adapter/form';

import { onMounted } from 'vue';

import { useVbenForm, z } from '#/adapter/form';
import { updateProfile } from '#/api/system/user';
import { $t } from '#/locales';

import { message } from 'ant-design-vue';

const props = defineProps<{ profile: SysUser }>();

const emit = defineEmits<{ updated: [] }>();

const formSchema: VbenFormSchema[] = [
  {
    fieldName: 'nickname',
    component: 'Input',
    label: $t('pages.fields.nickname'),
    rules: 'required',
    componentProps: {
      placeholder: $t('pages.profile.placeholders.nickname'),
    },
  },
  {
    fieldName: 'email',
    component: 'Input',
    label: $t('pages.fields.email'),
    rules: z
      .string()
      .email($t('pages.profile.validation.email'))
      .optional()
      .or(z.literal('')),
    componentProps: {
      placeholder: $t('pages.profile.placeholders.email'),
    },
  },
  {
    fieldName: 'phone',
    component: 'Input',
    label: $t('pages.fields.phone'),
    rules: z
      .string()
      .regex(/^1[3-9]\d{9}$/, $t('pages.profile.validation.phone'))
      .optional()
      .or(z.literal('')),
    componentProps: {
      placeholder: $t('pages.profile.placeholders.phone'),
    },
  },
  {
    fieldName: 'sex',
    component: 'RadioGroup',
    label: $t('pages.fields.sex'),
    defaultValue: 0,
    componentProps: {
      buttonStyle: 'solid',
      optionType: 'button',
      options: [
        { label: $t('pages.status.unknown'), value: 0 },
        { label: $t('pages.status.male'), value: 1 },
        { label: $t('pages.status.female'), value: 2 },
      ],
    },
  },
];

function buttonLoading(loading: boolean) {
  formApi.setState({ submitButtonOptions: { loading } });
}

const [Form, formApi] = useVbenForm({
  schema: formSchema,
  commonConfig: {
    labelWidth: 140,
    componentProps: {
      class: 'w-full',
    },
  },
  wrapperClass: 'grid-cols-1',
  resetButtonOptions: { show: false },
  submitButtonOptions: { content: $t('pages.profile.actions.updateProfile') },
  async handleSubmit(values) {
    buttonLoading(true);
    try {
      await updateProfile(values);
      message.success($t('pages.common.updateSuccess'));
      emit('updated');
    } finally {
      buttonLoading(false);
    }
  },
});

onMounted(() => {
  formApi.setValues({
    nickname: props.profile.nickname,
    email: props.profile.email,
    phone: props.profile.phone,
    sex: props.profile.sex,
  });
});
</script>
<template>
  <div data-testid="profile-base-form" class="mt-[16px] w-full max-w-[30rem]">
    <Form />
  </div>
</template>
