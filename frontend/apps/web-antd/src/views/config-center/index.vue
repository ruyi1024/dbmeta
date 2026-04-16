<script lang="ts" setup>
import { Page } from '@vben/common-ui';

import { $t } from '#/locales';
</script>

<template>
  <Page auto-content-height :description="$t('page.configCenter.description')">
    <div class="rounded-2xl bg-card p-6 text-sm text-muted-foreground">
      {{ $t('page.configCenter.placeholder') }}
    </div>
  </Page>
</template>
