# @vben/plugins

This directory stores third-party integrations and related plugin wrappers.

## Guideline

Import every plugin through a subpath export so applications can opt in only to the integrations they need.

Example:

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
