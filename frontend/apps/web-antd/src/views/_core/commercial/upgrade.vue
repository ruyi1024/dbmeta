<script lang="ts" setup>
import { Page } from '@vben/common-ui';

import { Button, Result } from 'ant-design-vue';
import { useRouter } from 'vue-router';
import { $t } from '#/locales';

defineOptions({ name: 'CommercialUpgradePage' });

const router = useRouter();

function goHome() {
  void router.push('/');
}
</script>

<template>
  <Page>
    <Result
      status="info"
      :title="$t('page.commercial.title')"
      :sub-title="$t('page.commercial.subtitle')"
    >
      <template #extra>
        <Button type="primary" @click="goHome">
          {{ $t('page.commercial.backHome') }}
        </Button>
      </template>
    </Result>
  </Page>
</template>
