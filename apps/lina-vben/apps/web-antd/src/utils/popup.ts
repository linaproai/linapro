import { $t } from '@vben/locales';

import { Modal } from 'ant-design-vue';

/**
 * 获取表单默认值获取器
 */
export function defaultFormValueGetter(formApi: {
  form: { values: Record<string, any> };
}) {
  return () => formApi.form?.values || {};
}

function confirmDiscardChanges() {
  return new Promise<boolean>((resolve) => {
    Modal.confirm({
      title: $t('pages.common.confirmTitle'),
      content: $t('pages.common.discardChangesConfirm'),
      okText: $t('pages.common.confirm'),
      cancelText: $t('pages.common.cancel'),
      zIndex: 2100,
      onCancel: () => resolve(false),
      onOk: () => resolve(true),
    });
  });
}

/**
 * 用于抽屉/弹窗关闭前检测表单是否有变化
 */
export function useBeforeCloseDiff(options: {
  initializedGetter: () => Promise<any> | any;
  currentGetter: () => Promise<any> | any;
}) {
  let initialized = false;
  let initializedValue: any;

  const markInitialized = async () => {
    initializedValue = await options.initializedGetter();
    initialized = true;
  };

  const resetInitialized = () => {
    initialized = false;
    initializedValue = undefined;
  };

  const onBeforeClose = async () => {
    if (!initialized) {
      return true;
    }
    const currentValue = await options.currentGetter();
    const hasChanged =
      JSON.stringify(currentValue) !== JSON.stringify(initializedValue);
    if (hasChanged) {
      return await confirmDiscardChanges();
    }
    return true;
  };

  return {
    onBeforeClose,
    markInitialized,
    resetInitialized,
  };
}
