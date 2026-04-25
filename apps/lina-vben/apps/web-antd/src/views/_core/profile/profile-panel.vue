<script setup lang="ts">
import type { SysUser } from '#/api/system/user';

import { computed } from 'vue';

import { preferences } from '@vben/preferences';

import { Card, Descriptions, DescriptionsItem, Tooltip } from 'ant-design-vue';

import { userUpdateAvatar } from '#/api/system/user';
import { CropperAvatar } from '#/components/cropper';
import { $t } from '#/locales';

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
        <Tooltip :title="$t('pages.profile.panel.uploadAvatar')">
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
            {{ profile.nickname || $t('pages.status.unknown') }}
          </span>
          <span class="text-foreground/60 text-sm">
            {{ profile.username }}
          </span>
        </div>
      </div>
      <div class="w-full px-[24px]">
        <Descriptions :column="1">
          <DescriptionsItem :label="$t('pages.profile.panel.account')">
            {{ profile.username }}
          </DescriptionsItem>
          <DescriptionsItem :label="$t('pages.fields.phone')">
            {{ profile.phone || $t('pages.profile.panel.unboundPhone') }}
          </DescriptionsItem>
          <DescriptionsItem :label="$t('pages.fields.email')">
            {{ profile.email || $t('pages.profile.panel.unboundEmail') }}
          </DescriptionsItem>
          <DescriptionsItem :label="$t('pages.profile.panel.lastLogin')">
            {{ profile.loginDate || $t('pages.profile.panel.noLoginRecord') }}
          </DescriptionsItem>
        </Descriptions>
      </div>
    </div>
  </Card>
</template>
