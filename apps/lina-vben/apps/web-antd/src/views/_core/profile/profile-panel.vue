<script setup lang="ts">
import type { SysUser } from '#/api/system/user';

import { computed } from 'vue';

import { preferences } from '@vben/preferences';

import { Card, Descriptions, DescriptionsItem, Tooltip } from 'ant-design-vue';

import { userUpdateAvatar } from '#/api/system/user';
import { CropperAvatar } from '#/components/cropper';

const props = defineProps<{ profile?: SysUser }>();

defineEmits<{
  uploadFinish: [];
}>();

const avatar = computed(
  () => props.profile?.avatar || preferences.app.defaultAvatar,
);
</script>

<template>
  <Card :loading="!profile" class="h-full lg:w-1/3">
    <div v-if="profile" class="flex flex-col items-center gap-[24px]">
      <div class="flex flex-col items-center gap-[20px]">
        <Tooltip title="点击上传头像">
          <CropperAvatar
            :show-btn="false"
            :upload-api="userUpdateAvatar"
            :value="avatar"
            width="120"
            @change="$emit('uploadFinish')"
          />
        </Tooltip>
        <div class="flex flex-col items-center gap-[8px]">
          <span class="text-foreground text-xl font-bold">
            {{ profile.nickname || '未知' }}
          </span>
          <span class="text-foreground/60 text-sm">
            {{ profile.username }}
          </span>
        </div>
      </div>
      <div class="w-full px-[24px]">
        <Descriptions :column="1">
          <DescriptionsItem label="账号">
            {{ profile.username }}
          </DescriptionsItem>
          <DescriptionsItem label="手机">
            {{ profile.phone || '未绑定手机' }}
          </DescriptionsItem>
          <DescriptionsItem label="邮箱">
            {{ profile.email || '未绑定邮箱' }}
          </DescriptionsItem>
          <DescriptionsItem label="上次登录">
            {{ profile.loginDate || '暂无登录记录' }}
          </DescriptionsItem>
        </Descriptions>
      </div>
    </div>
  </Card>
</template>
