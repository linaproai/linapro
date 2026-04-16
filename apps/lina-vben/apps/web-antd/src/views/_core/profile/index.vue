<script setup lang="ts">
import type { SysUser } from '#/api/system/user';

import { onMounted, ref } from 'vue';

import { Page } from '@vben/common-ui';
import { useUserStore } from '@vben/stores';

import { TabPane, Tabs } from 'ant-design-vue';

import { getProfile } from '#/api/system/user';
import { useAuthStore } from '#/store';

import BaseSetting from './base-setting.vue';
import PasswordSetting from './password-setting.vue';
import ProfilePanel from './profile-panel.vue';

const profile = ref<SysUser>();
const authStore = useAuthStore();
const userStore = useUserStore();

async function loadProfile() {
  profile.value = await getProfile();
}

onMounted(loadProfile);

async function handleUploadFinish() {
  await loadProfile();
  const userInfo = await authStore.fetchUserInfo();
  userStore.setUserInfo(userInfo);
}

async function handleProfileUpdated() {
  await loadProfile();
  const userInfo = await authStore.fetchUserInfo();
  userStore.setUserInfo(userInfo);
}
</script>
<template>
  <Page>
    <div class="flex flex-col gap-[16px] lg:flex-row lg:items-stretch">
      <!-- 左侧 -->
      <ProfilePanel :profile="profile" @upload-finish="handleUploadFinish" />
      <!-- 右侧 -->
      <div
        class="bg-background rounded-[var(--radius)] px-[16px] pt-[4px] lg:flex-1"
      >
        <Tabs default-active-key="basic">
          <TabPane key="basic" tab="基本设置">
            <BaseSetting
              v-if="profile"
              :profile="profile"
              @updated="handleProfileUpdated"
            />
          </TabPane>
          <TabPane key="password" tab="安全设置">
            <PasswordSetting @updated="handleProfileUpdated" />
          </TabPane>
        </Tabs>
      </div>
    </div>
  </Page>
</template>
