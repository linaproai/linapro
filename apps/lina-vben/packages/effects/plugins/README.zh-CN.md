# @vben/plugins

该目录用于存放第三方集成以及相关插件封装。

## 规范

所有插件都应通过 subpath export 方式引入，这样应用只会按需启用自己真正使用的集成。

示例：

```json
{
  "exports": {
    "./echarts": {
      "types": "./src/echarts/index.ts",
      "default": "./src/echarts/index.ts"
    }
  }
}
```

```ts
import { useEcharts } from '@vben/plugins/echarts';
```
