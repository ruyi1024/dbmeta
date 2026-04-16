<script lang="ts" setup>
import { onMounted, reactive, ref } from 'vue';

import { Button, Card, Form, Input, Tabs, message } from 'ant-design-vue';

import { baseRequestClient } from '#/api/request';
import { $t } from '#/locales';

defineOptions({ name: 'SettingNoticePage' });

interface NoticeForm {
  accessKeyId: string;
  accessKeySecret: string;
  mailFrom: string;
  mailHost: string;
  mailPass: string;
  mailPort: string;
  mailUser: string;
  phonePlayTimes: string;
  phoneTemplateCode: string;
  smsSignName: string;
  smsTemplateCode: string;
  wechatAppId: string;
  wechatAppSecret: string;
  wechatSendTemplateId: string;
}

function extractApiBody(response: unknown): Record<string, unknown> {
  if (!response || typeof response !== 'object') return {};
  const r = response as Record<string, unknown>;
  if ('data' in r && r.data !== undefined && typeof r.data === 'object' && 'status' in r) {
    return (r.data ?? {}) as Record<string, unknown>;
  }
  return r;
}

const loading = ref(false);
const saving = ref<'aliyun' | 'mail' | 'wechat' | ''>('');
const activeTab = ref<'aliyun' | 'mail' | 'wechat'>('mail');
const form = reactive<NoticeForm>({
  accessKeyId: '',
  accessKeySecret: '',
  mailFrom: '',
  mailHost: '',
  mailPass: '',
  mailPort: '',
  mailUser: '',
  phonePlayTimes: '',
  phoneTemplateCode: '',
  smsSignName: '',
  smsTemplateCode: '',
  wechatAppId: '',
  wechatAppSecret: '',
  wechatSendTemplateId: '',
});

function assignForm(data: Record<string, unknown>) {
  form.mailHost = String(data.mailHost ?? '');
  form.mailPort = String(data.mailPort ?? '');
  form.mailUser = String(data.mailUser ?? '');
  form.mailPass = String(data.mailPass ?? '');
  form.mailFrom = String(data.mailFrom ?? '');
  form.accessKeyId = String(data.accessKeyId ?? '');
  form.accessKeySecret = String(data.accessKeySecret ?? '');
  form.smsSignName = String(data.smsSignName ?? '');
  form.smsTemplateCode = String(data.smsTemplateCode ?? '');
  form.phoneTemplateCode = String(data.phoneTemplateCode ?? '');
  form.phonePlayTimes = String(data.phonePlayTimes ?? '');
  form.wechatAppId = String(data.wechatAppId ?? '');
  form.wechatAppSecret = String(data.wechatAppSecret ?? '');
  form.wechatSendTemplateId = String(data.wechatSendTemplateId ?? '');
}

async function fetchNotice() {
  loading.value = true;
  try {
    const response = await baseRequestClient.get('/v1/setting/notice');
    const body = extractApiBody(response);
    assignForm((body.data ?? {}) as Record<string, unknown>);
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.notice.message.loadFailed'));
  } finally {
    loading.value = false;
  }
}

async function saveMail() {
  saving.value = 'mail';
  try {
    const response = await baseRequestClient.put('/v1/setting/notice/mail', {
      mailFrom: form.mailFrom,
      mailHost: form.mailHost,
      mailPass: form.mailPass,
      mailPort: form.mailPort,
      mailUser: form.mailUser,
    });
    const body = extractApiBody(response);
    if (body.success !== true) {
      message.error(String(body.msg ?? $t('page.notice.message.saveFailed')));
      return;
    }
    message.success($t('page.notice.message.mailSaved'));
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.notice.message.saveFailed'));
  } finally {
    saving.value = '';
  }
}

async function saveAliyun() {
  saving.value = 'aliyun';
  try {
    const response = await baseRequestClient.put('/v1/setting/notice/aliyun', {
      accessKeyId: form.accessKeyId,
      accessKeySecret: form.accessKeySecret,
      phonePlayTimes: form.phonePlayTimes,
      phoneTemplateCode: form.phoneTemplateCode,
      smsSignName: form.smsSignName,
      smsTemplateCode: form.smsTemplateCode,
    });
    const body = extractApiBody(response);
    if (body.success !== true) {
      message.error(String(body.msg ?? $t('page.notice.message.saveFailed')));
      return;
    }
    message.success($t('page.notice.message.aliyunSaved'));
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.notice.message.saveFailed'));
  } finally {
    saving.value = '';
  }
}

async function saveWechat() {
  saving.value = 'wechat';
  try {
    const response = await baseRequestClient.put('/v1/setting/notice/wechat', {
      wechatAppId: form.wechatAppId,
      wechatAppSecret: form.wechatAppSecret,
      wechatSendTemplateId: form.wechatSendTemplateId,
    });
    const body = extractApiBody(response);
    if (body.success !== true) {
      message.error(String(body.msg ?? $t('page.notice.message.saveFailed')));
      return;
    }
    message.success($t('page.notice.message.wechatSaved'));
  } catch (e: unknown) {
    message.error((e as Error)?.message || $t('page.notice.message.saveFailed'));
  } finally {
    saving.value = '';
  }
}

onMounted(() => {
  void fetchNotice();
});
</script>

<template>
  <div class="p-5">
    <Card :loading="loading" :title="$t('page.notice.title')">
      <Tabs v-model:active-key="activeTab" tab-position="left">
        <Tabs.TabPane key="mail" :tab="$t('page.notice.tab.mail')">
          <Form layout="vertical">
            <div class="form-grid">
              <Form.Item :label="$t('page.notice.form.smtpHost')">
                <Input
                  v-model:value="form.mailHost"
                  :placeholder="$t('page.notice.placeholder.smtpHost')"
                />
              </Form.Item>
              <Form.Item :label="$t('page.notice.form.smtpPort')">
                <Input
                  v-model:value="form.mailPort"
                  :placeholder="$t('page.notice.placeholder.smtpPort')"
                />
              </Form.Item>
              <Form.Item :label="$t('page.notice.form.mailUser')">
                <Input v-model:value="form.mailUser" />
              </Form.Item>
              <Form.Item :label="$t('page.notice.form.mailPass')">
                <Input.Password v-model:value="form.mailPass" />
              </Form.Item>
              <Form.Item :label="$t('page.notice.form.mailFrom')">
                <Input v-model:value="form.mailFrom" />
              </Form.Item>
            </div>
            <div class="mt-2 flex justify-end">
              <Button type="primary" :loading="saving === 'mail'" @click="saveMail">
                {{ $t('page.notice.action.saveMail') }}
              </Button>
            </div>
          </Form>
        </Tabs.TabPane>
        <Tabs.TabPane key="aliyun" :tab="$t('page.notice.tab.aliyun')">
          <Form layout="vertical">
            <div class="form-grid">
              <Form.Item :label="$t('page.notice.form.accessKeyId')">
                <Input v-model:value="form.accessKeyId" />
              </Form.Item>
              <Form.Item :label="$t('page.notice.form.accessKeySecret')">
                <Input.Password v-model:value="form.accessKeySecret" />
              </Form.Item>
              <Form.Item :label="$t('page.notice.form.smsSignName')">
                <Input v-model:value="form.smsSignName" />
              </Form.Item>
              <Form.Item :label="$t('page.notice.form.smsTemplateCode')">
                <Input v-model:value="form.smsTemplateCode" />
              </Form.Item>
              <Form.Item :label="$t('page.notice.form.phoneTemplateCode')">
                <Input v-model:value="form.phoneTemplateCode" />
              </Form.Item>
              <Form.Item :label="$t('page.notice.form.phonePlayTimes')">
                <Input v-model:value="form.phonePlayTimes" />
              </Form.Item>
            </div>
            <div class="mt-2 flex justify-end">
              <Button type="primary" :loading="saving === 'aliyun'" @click="saveAliyun">
                {{ $t('page.notice.action.saveAliyun') }}
              </Button>
            </div>
          </Form>
        </Tabs.TabPane>
        <Tabs.TabPane key="wechat" :tab="$t('page.notice.tab.wechat')">
          <Form layout="vertical">
            <div class="form-grid">
              <Form.Item :label="$t('page.notice.form.wechatAppId')">
                <Input v-model:value="form.wechatAppId" />
              </Form.Item>
              <Form.Item :label="$t('page.notice.form.wechatAppSecret')">
                <Input.Password v-model:value="form.wechatAppSecret" />
              </Form.Item>
              <Form.Item :label="$t('page.notice.form.wechatSendTemplateId')">
                <Input v-model:value="form.wechatSendTemplateId" />
              </Form.Item>
            </div>
            <div class="mt-2 flex justify-end">
              <Button type="primary" :loading="saving === 'wechat'" @click="saveWechat">
                {{ $t('page.notice.action.saveWechat') }}
              </Button>
            </div>
          </Form>
        </Tabs.TabPane>
      </Tabs>
    </Card>
  </div>
</template>

<style scoped>
.form-grid {
  column-gap: 12px;
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
}

@media (max-width: 900px) {
  .form-grid {
    grid-template-columns: 1fr;
  }
}
</style>

