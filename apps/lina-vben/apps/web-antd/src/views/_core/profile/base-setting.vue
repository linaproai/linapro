<script setup lang="ts">
import type { SysUser } from '#/api/system/user';
import type { VbenFormSchema } from '#/adapter/form';

import { onMounted } from 'vue';

import { useVbenForm, z } from '#/adapter/form';
import { updateProfile } from '#/api/system/user';

import { message } from 'ant-design-vue';

const props = defineProps<{ profile: SysUser }>();

const emit = defineEmits<{ updated: [] }>();

const formSchema: VbenFormSchema[] = [
  {
    fieldName: 'nickname',
    component: 'Input',
    label: '昵称',
    rules: 'required',
    componentProps: {
      placeholder: '请输入昵称',
    },
  },
  {
    fieldName: 'email',
    component: 'Input',
    label: '邮箱',
    rules: z.string().email('请输入正确的邮箱').optional().or(z.literal('')),
    componentProps: {
      placeholder: '请输入邮箱',
    },
  },
  {
    fieldName: 'phone',
    component: 'Input',
    label: '手机号码',
    rules: z
      .string()
      .regex(/^1[3-9]\d{9}$/, '请输入正确的手机号')
      .optional()
      .or(z.literal('')),
    componentProps: {
      placeholder: '请输入手机号码',
    },
  },
  {
    fieldName: 'sex',
    component: 'RadioGroup',
    label: '性别',
    defaultValue: 0,
    componentProps: {
      buttonStyle: 'solid',
      optionType: 'button',
      options: [
        { label: '未知', value: 0 },
        { label: '男', value: 1 },
        { label: '女', value: 2 },
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
    labelWidth: 80,
    componentProps: {
      class: 'w-full',
    },
  },
  wrapperClass: 'grid-cols-1',
  resetButtonOptions: { show: false },
  submitButtonOptions: { content: '更新信息' },
  async handleSubmit(values) {
    buttonLoading(true);
    try {
      await updateProfile(values);
      message.success('更新成功');
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
  <div class="mt-[16px] md:w-full lg:w-1/2 2xl:w-2/5">
    <Form />
  </div>
</template>
