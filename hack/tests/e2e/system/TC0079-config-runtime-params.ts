import { test, expect } from '../../fixtures/auth';
import { ConfigPage } from '../../pages/ConfigPage';

const runtimeParams = [
  { key: 'sys.jwt.expire', name: '认证管理-JWT Token 有效期' },
  { key: 'sys.session.timeout', name: '在线用户-会话超时时间' },
  { key: 'sys.upload.maxSize', name: '文件管理-上传大小上限' },
  { key: 'sys.login.blackIPList', name: '用户登录-IP 黑名单列表' },
];

const removedParams = [
  { key: 'sys.logger.traceID.enabled', name: '日志-TraceID 输出开关' },
  { key: 'sys.user.initPassword', name: '已下线的初始化密码参数' },
];

test.describe('TC0079 参数设置内置运行时参数', () => {
  for (const [index, runtimeParam] of runtimeParams.entries()) {
    const subCase = String.fromCharCode('a'.charCodeAt(0) + index);

    test(`TC0079${subCase}: 可检索到 ${runtimeParam.key}`, async ({ adminPage }) => {
      const configPage = new ConfigPage(adminPage);
      await configPage.goto();

      await configPage.fillSearchField('参数键名', runtimeParam.key);
      await configPage.clickSearch();

      const hasConfig = await configPage.hasConfig(runtimeParam.key);
      expect(hasConfig).toBeTruthy();
      await expect(adminPage.locator('.vxe-body--row').first()).toContainText(
        runtimeParam.name,
      );
    });
  }

  for (const [index, removedParam] of removedParams.entries()) {
    const subCase = String.fromCharCode(
      'a'.charCodeAt(0) + runtimeParams.length + index,
    );

    test(`TC0079${subCase}: 不再检索到 ${removedParam.key}`, async ({ adminPage }) => {
      const configPage = new ConfigPage(adminPage);
      await configPage.goto();

      await configPage.fillSearchField('参数键名', removedParam.key);
      await configPage.clickSearch();

      const hasConfig = await configPage.hasConfig(removedParam.key);
      expect(hasConfig).toBeFalsy();
    });
  }
});
