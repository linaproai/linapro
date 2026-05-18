<script setup lang="ts">
interface Props {
  testId?: string;
}

defineProps<Props>();
</script>

<template>
  <div
    :data-testid="testId"
    class="rounded-md border-l-4 border-[var(--ant-color-primary)] bg-[var(--ant-color-primary-bg)] px-3 py-2 text-[15px] font-semibold leading-6 text-[var(--ant-color-primary)]"
  >
    <slot />
  </div>
</template>
