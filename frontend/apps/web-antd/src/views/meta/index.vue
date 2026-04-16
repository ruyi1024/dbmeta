<script lang="ts" setup>
import { $t } from '#/locales';

defineOptions({ name: 'MetaPage' });
</script>

<template>
  <div class="p-4">
    <h2 class="text-lg font-medium">{{ $t('page.meta.title') }}</h2>
    <p class="mt-2 text-sm text-gray-500">
      {{ $t('page.meta.description') }}
    </p>
  </div>
</template>
