<script setup lang="ts">
import type { SupportedLanguagesType } from '@vben/locales';

import { Languages } from '@vben/icons';
import {
  loadLocaleMessages,
  runtimeLocaleOptions,
  runtimeLocaleSwitchEnabled,
} from '@vben/locales';
import { preferences, updatePreferences } from '@vben/preferences';

import { VbenDropdownRadioMenu, VbenIconButton } from '@vben-core/shadcn-ui';

defineOptions({
  name: 'LanguageToggle',
});

async function handleUpdate(value: string | undefined) {
  if (!value) return;
  const locale = value as SupportedLanguagesType;
  updatePreferences({
    app: {
      locale,
    },
  });
  await loadLocaleMessages(locale);
  updatePreferences({
    app: {
      locale,
    },
  });
}
</script>

<template>
  <div v-if="runtimeLocaleSwitchEnabled" data-testid="language-toggle">
    <VbenDropdownRadioMenu
      :menus="runtimeLocaleOptions"
      :model-value="preferences.app.locale"
      @update:model-value="handleUpdate"
    >
      <VbenIconButton
        class="hover:animate-[shrink_0.3s_ease-in-out]"
        data-testid="language-toggle-trigger"
      >
        <Languages class="size-4 text-foreground" />
      </VbenIconButton>
    </VbenDropdownRadioMenu>
  </div>
</template>
