<script lang="ts" setup>
import { Page } from '@vben/common-ui';

import { Card, Empty } from 'ant-design-vue';

defineOptions({ name: 'Categorization' });
</script>

<template>
  <Page auto-content-height description="按业务域与合规要求对数据资产进行类目与标签管理。">
    <Card>
      <Empty description="功能建设中" />
    </Card>
  </Page>
</template>
