<script lang="ts" setup>
import { Page } from '@vben/common-ui';

import { Card, Empty } from 'ant-design-vue';
import { $t } from '#/locales';

defineOptions({ name: 'Categorization' });
</script>

<template>
  <Page
    auto-content-height
    :description="$t('page.categorization.description')"
  >
    <Card>
      <Empty :description="$t('page.common.underConstruction')" />
    </Card>
  </Page>
</template>
