<script lang="ts" setup>
import { computed, defineComponent, h, reactive, ref } from 'vue';
import Exceljs from 'exceljs';
import { saveAs } from 'file-saver';
import { CheckCircleOutlined, CloseCircleOutlined, DownloadOutlined } from '@ant-design/icons-vue';

import {
  Alert,
  Button,
  Card,
  Col,
  Drawer,
  Form,
  Input,
  List,
  message,
  Modal,
  Row,
  Select,
  Space,
  Table,
  Tabs,
  Tag,
  TypographyParagraph,
} from 'ant-design-vue';

import { baseRequestClient } from '#/api/request';
import { $t } from '#/locales';

interface OptionItem {
  id?: number;
  name?: string;
  table_name?: string;
  value?: string;
}

const typeList = ref<OptionItem[]>([]);
const favoriteList = ref<any[]>([]);

const openFavorite = ref(false);
const openAiGenerate = ref(false);
const aiQuestion = ref('');
const aiGenerating = ref(false);

const globalFormState = reactive({
  database: '',
  datasource: '',
  type: '',
});

interface QueryTab {
  key: string;
  title: string;
  selectedSql: string;
  sqlContent: string;
  loading: boolean;
  tableDataColumn: any[];
  tableDataList: any[];
  tableDataMsg: string;
  tableDataSuccess: boolean;
  tableDataTotal: number;
  queryTimes: number;
  columnWidths: Record<string, number>;
}

const tabs = ref<QueryTab[]>([
  {
    key: '1',
    title: `${$t('page.dataQuery.tab')} 1`,
    selectedSql: '',
    sqlContent: '',
    loading: false,
    tableDataColumn: [],
    tableDataList: [],
    tableDataMsg: '',
    tableDataSuccess: false,
    tableDataTotal: 0,
    queryTimes: 0,
    columnWidths: {},
  },
]);
const activeTabKey = ref('1');
const tabCounter = ref(2);

const datasourceList = ref<any[]>([]);
const databaseList = ref<any[]>([]);
const tableList = ref<any[]>([]);
const tableKeyword = ref('');
const tablePanelCollapsed = ref(false);
/** 左侧点击的表名；后端 showIndex/showColumn/showCreate 等依赖此字段生成 SQL，不能为空 */
const selectedTable = ref('');

const currentTab = computed(
  () => tabs.value.find((tab) => tab.key === activeTabKey.value) ?? tabs.value[0],
);
const activeSqlContent = computed({
  get: () => currentTab.value?.sqlContent ?? '',
  set: (value: string) => {
    updateTab(activeTabKey.value, { sqlContent: value });
  },
});
const currentDataColumns = computed(() =>
  (currentTab.value?.tableDataColumn || []).map((col: any) => {
    const dataIndex = col.dataIndex || col.key;
    return {
      dataIndex,
      key: dataIndex,
      title: col.title,
      width: currentTab.value?.columnWidths?.[dataIndex] || col.width || 160,
      onHeaderCell: () => ({
        width: currentTab.value?.columnWidths?.[dataIndex] || col.width || 160,
        onResize: (_evt: MouseEvent, payload: { size: { width: number } }) => {
          resizeColumn(dataIndex, payload.size.width);
        },
      }),
    };
  }),
);
const filteredTableList = computed(() => {
  const keyword = tableKeyword.value.trim().toLowerCase();
  if (!keyword) {
    return tableList.value;
  }
  return tableList.value.filter((item: any) =>
    String(item.table_name || '')
      .toLowerCase()
      .includes(keyword),
  );
});

const ResizableHeaderCell = defineComponent({
  name: 'ResizableHeaderCell',
  props: {
    width: { type: Number, default: undefined },
    onResize: { type: Function, default: undefined },
  },
  setup(props, { slots, attrs }) {
    function onMouseDown(event: MouseEvent) {
      if (!props.onResize || !props.width) {
        return;
      }
      event.preventDefault();
      const startX = event.clientX;
      const startWidth = props.width;
      const minWidth = 90;
      const handleMouseMove = (moveEvent: MouseEvent) => {
        const diff = moveEvent.clientX - startX;
        const newWidth = Math.max(minWidth, startWidth + diff);
        (props.onResize as any)(moveEvent, { size: { width: newWidth } });
      };
      const handleMouseUp = () => {
        document.removeEventListener('mousemove', handleMouseMove);
        document.removeEventListener('mouseup', handleMouseUp);
      };
      document.addEventListener('mousemove', handleMouseMove);
      document.addEventListener('mouseup', handleMouseUp);
    }
    return () =>
      h(
        'th',
        { ...attrs, style: { position: 'relative', width: `${props.width}px` } },
        [
          slots.default?.(),
          props.width
            ? h('span', {
                class: 'resize-handle',
                onMousedown: onMouseDown,
              })
            : null,
        ],
      );
  },
});

function updateTab(key: string, patch: Partial<QueryTab>) {
  tabs.value = tabs.value.map((tab) => (tab.key === key ? { ...tab, ...patch } : tab));
}

function addTab() {
  const key = `${tabCounter.value}`;
  tabs.value.push({
    key,
    title: `${$t('page.dataQuery.tab')} ${tabCounter.value}`,
    selectedSql: '',
    sqlContent: '',
    loading: false,
    tableDataColumn: [],
    tableDataList: [],
    tableDataMsg: '',
    tableDataSuccess: false,
    tableDataTotal: 0,
    queryTimes: 0,
    columnWidths: {},
  });
  activeTabKey.value = key;
  tabCounter.value += 1;
}

function removeTab(key: string) {
  if (tabs.value.length <= 1) {
    message.warning($t('page.dataQuery.message.keepOneTab'));
    return;
  }
  const idx = tabs.value.findIndex((tab) => tab.key === key);
  tabs.value = tabs.value.filter((tab) => tab.key !== key);
  if (activeTabKey.value === key) {
    const fallback = tabs.value[Math.max(0, idx - 1)]?.key || tabs.value[0]?.key || '1';
    activeTabKey.value = fallback;
  }
}

function resizeColumn(dataIndex: string, width: number) {
  const widths = { ...(currentTab.value?.columnWidths || {}) };
  widths[dataIndex] = width;
  updateTab(activeTabKey.value, { columnWidths: widths });
}

function getSqlToExecute() {
  const selected = currentTab.value?.selectedSql?.trim();
  if (selected) {
    return selected;
  }
  return currentTab.value?.sqlContent || '';
}

function handleSqlSelection(event: Event) {
  const target = event.target as HTMLTextAreaElement;
  const selected = target.value.slice(target.selectionStart || 0, target.selectionEnd || 0).trim();
  updateTab(activeTabKey.value, { selectedSql: selected });
}

async function writeLog(doType: string) {
  try {
    await baseRequestClient.post('/v1/query/writeLog', {
      datasource_type: globalFormState.type,
      datasource: globalFormState.datasource,
      database: globalFormState.database,
      sql: getSqlToExecute(),
      query_type: doType,
    });
  } catch {
    // 审计日志异常不阻断主流程
  }
}

async function loadDatasourceTypes() {
  const response = await baseRequestClient.get('/v1/query/datasource_type');
  const payload = (response as any)?.data ?? response;
  typeList.value = payload?.data ?? payload ?? [];
}

async function onTypeChange(val: any) {
  if (!val) {
    return;
  }
  globalFormState.type = val;
  globalFormState.datasource = '';
  globalFormState.database = '';
  selectedTable.value = '';
  tabs.value = tabs.value.map((tab) => ({ ...tab, sqlContent: '', selectedSql: '' }));
  databaseList.value = [];
  tableList.value = [];

  const response = await baseRequestClient.get(`/v1/query/datasource?type=${encodeURIComponent(val)}`);
  const payload = (response as any)?.data ?? response;
  datasourceList.value = payload?.data ?? payload ?? [];
}

async function onDatasourceChange(val: any) {
  if (!val) {
    return;
  }
  globalFormState.datasource = val;
  globalFormState.database = '';
  selectedTable.value = '';
  tableList.value = [];
  tabs.value = tabs.value.map((tab) => ({ ...tab, sqlContent: '', selectedSql: '' }));

  const response = await baseRequestClient.get(
    `/v1/query/database?datasource=${encodeURIComponent(val)}&type=${encodeURIComponent(globalFormState.type)}`,
  );
  const payload = (response as any)?.data ?? response;
  databaseList.value = payload?.data ?? payload ?? [];
}

async function onDatabaseChange(val: any) {
  if (!val) {
    return;
  }
  globalFormState.database = val;
  selectedTable.value = '';
  updateTab(activeTabKey.value, { sqlContent: '', selectedSql: '' });
  const response = await baseRequestClient.get(
    `/v1/query/table?datasource=${encodeURIComponent(globalFormState.datasource)}&database=${encodeURIComponent(val)}&type=${encodeURIComponent(globalFormState.type)}`,
  );
  const payload = (response as any)?.data ?? response;
  tableList.value = payload?.data ?? payload ?? [];
}

function onClickTable(tableName: string) {
  selectedTable.value = tableName;
  let sql = '';
  if (['MySQL', 'TiDB', 'Doris', 'MariaDB', 'GreatSQL', 'OceanBase', 'ClickHouse', 'PostgreSQL'].includes(globalFormState.type)) {
    sql = `select * from ${tableName} limit 100`;
  } else if (globalFormState.type === 'Oracle') {
    sql = `select * from ${globalFormState.database}.${tableName} where rownum<=100`;
  } else if (globalFormState.type === 'SQLServer') {
    sql = `select top 100 * from ${tableName}`;
  } else if (globalFormState.type === 'MongoDB') {
    sql = `select.from('${tableName}').where('_id','!=','').limit(100)`;
  }
  updateTab(activeTabKey.value, { sqlContent: sql, selectedSql: '' });
}

async function executeQuery(queryType = 'execute') {
  const sqlToExecute = getSqlToExecute();
  const needsTableForMeta =
    queryType === 'showIndex' ||
    queryType === 'showColumn' ||
    queryType === 'showCreate' ||
    queryType === 'showTableSize';
  if (needsTableForMeta && !selectedTable.value) {
    message.warning($t('page.dataQuery.message.selectTableFirst'));
    return;
  }
  if (!globalFormState.type || !globalFormState.datasource) {
    message.warning($t('page.dataQuery.message.selectDatasourceFirst'));
    return;
  }
  if ((queryType === 'execute' || queryType === 'doExplain') && !sqlToExecute?.trim()) {
    message.warning($t('page.dataQuery.message.selectDatasourceAndInputSql'));
    return;
  }
  updateTab(activeTabKey.value, { loading: true });
  try {
    const response = await baseRequestClient.post('/v1/query/doQuery', {
      database: globalFormState.database,
      datasource: globalFormState.datasource,
      datasource_type: globalFormState.type,
      query_type: queryType,
      sql: sqlToExecute,
      table: selectedTable.value,
    });
    const payload = (response as any)?.data ?? response;
    const currentWidths = currentTab.value?.columnWidths || {};
    const columns = payload.columns || [];
    const incomingWidths: Record<string, number> = {};
    columns.forEach((col: any) => {
      const idx = col.dataIndex || col.key;
      if (idx && !currentWidths[idx]) {
        incomingWidths[idx] = col.width || 160;
      }
    });
    updateTab(activeTabKey.value, {
      loading: false,
      tableDataSuccess: !!payload.success,
      tableDataMsg: payload.msg || '',
      tableDataList: payload.data || [],
      tableDataColumn: columns,
      tableDataTotal: payload.total || 0,
      queryTimes: payload.times || 0,
      columnWidths: { ...currentWidths, ...incomingWidths },
    });
  } catch (error: any) {
    updateTab(activeTabKey.value, {
      loading: false,
      tableDataSuccess: false,
      tableDataMsg: error?.message || $t('page.dataQuery.message.executeFailed'),
    });
  }
}

async function favoriteSql() {
  if (!globalFormState.type || !globalFormState.datasource || !currentTab.value?.sqlContent) {
    message.warning($t('page.dataQuery.message.favoriteIncomplete'));
    return;
  }
  const response = await baseRequestClient.post('/v1/favorite/list', {
    content: currentTab.value?.sqlContent,
    database_name: globalFormState.database,
    datasource: globalFormState.datasource,
    datasource_type: globalFormState.type,
  });
  const payload = (response as any)?.data ?? response;
  message.success(
    payload?.success
      ? $t('page.dataQuery.message.favoriteSuccess')
      : $t('page.dataQuery.message.favoriteFailed'),
  );
}

async function openFavoriteDrawer() {
  if (!globalFormState.type || !globalFormState.datasource) {
    message.warning($t('page.dataQuery.message.selectDatasourceBeforeFavorite'));
    return;
  }
  const response = await baseRequestClient.get('/v1/favorite/list', {
    params: {
      database_name: globalFormState.database || undefined,
      datasource: globalFormState.datasource,
      datasource_type: globalFormState.type,
    },
  });
  const payload = (response as any)?.data ?? response;
  favoriteList.value = payload?.data ?? payload ?? [];
  openFavorite.value = true;
}

async function deleteFavorite(id: number) {
  const response = await baseRequestClient.delete('/v1/favorite/list', { data: { id } } as any);
  const payload = (response as any)?.data ?? response;
  if (payload?.success) {
    message.success($t('page.dataQuery.message.deleteSuccess'));
    openFavoriteDrawer();
  } else {
    message.error($t('page.dataQuery.message.deleteFailed'));
  }
}

async function doAiGenerate() {
  if (!aiQuestion.value.trim()) {
    message.warning($t('page.dataQuery.message.enterAiQuestion'));
    return;
  }
  aiGenerating.value = true;
  try {
    const [host, port] = globalFormState.datasource.split(':');
    const response = await baseRequestClient.post('/v1/ai/dbquery', {
      database_name: globalFormState.database,
      datasource_type: globalFormState.type,
      host,
      page: 1,
      page_size: 1,
      port,
      question: aiQuestion.value.trim(),
    });
    const payload = (response as any)?.data ?? response;
    const sql = payload?.data?.sql_query;
    if (payload?.success && sql) {
      updateTab(activeTabKey.value, { sqlContent: sql, selectedSql: '' });
      message.success($t('page.dataQuery.message.aiGenerateSuccess'));
      openAiGenerate.value = false;
      aiQuestion.value = '';
    } else {
      message.error(payload?.message || $t('page.dataQuery.message.aiGenerateFailed'));
    }
  } catch (error: any) {
    message.error(error?.message || $t('page.dataQuery.message.aiGenerateFailed'));
  } finally {
    aiGenerating.value = false;
  }
}

async function exportExcel() {
  if (!currentTab.value?.tableDataList?.length || !currentTab.value?.tableDataColumn?.length) {
    message.warning($t('page.dataQuery.message.noExportData'));
    return;
  }
  const workbook = new Exceljs.Workbook();
  const worksheet = workbook.addWorksheet($t('page.dataQuery.export.sheetName'));
  worksheet.properties.defaultRowHeight = 20;
  worksheet.columns = currentTab.value.tableDataColumn.map((col: any) => ({
    header: col.title,
    key: col.dataIndex || col.key,
    width: ((currentTab.value?.columnWidths?.[col.dataIndex || col.key] || col.width || 160) as number) / 8,
  }));
  worksheet.addRows(currentTab.value.tableDataList);

  const headerRow = worksheet.getRow(1);
  headerRow.eachCell((cell) => {
    cell.fill = { type: 'pattern', pattern: 'solid', fgColor: { argb: '0099CC' } };
    cell.font = { bold: true, size: 11, name: 'SimSun', color: { argb: 'FFFFFF' } };
    cell.alignment = { vertical: 'middle', horizontal: 'center' };
  });
  const binary = await workbook.xlsx.writeBuffer();
  const filename = `${globalFormState.type || 'query'}-${Date.now()}.xlsx`;
  saveAs(new Blob([binary], { type: 'application/octet-stream' }), filename);
  await writeLog('exportExcel');
}

function onCopyResult() {
  message.warning($t('page.dataQuery.message.copyLogged'));
  void writeLog('copyData');
}

/** 数据源是否可用（后端 QueryAll 可能把 status 扫成字符串 "1"） */
function isDatasourceStatusOk(item: { status?: unknown }): boolean {
  return Number(item.status) === 1;
}

loadDatasourceTypes();
</script>

<template>
  <div class="query-page p-5">
    <Row :gutter="[12, 12]">
      <Col :span="24">
        <Card>
          <Form class="filter-form mb-0">
            <div class="query-grid">
              <Form.Item :label="$t('page.dataQuery.form.datasourceType')" class="query-item">
                <Select
                  v-model:value="globalFormState.type"
                  class="query-control"
                  :placeholder="$t('page.dataQuery.placeholder.datasourceType')"
                  @change="onTypeChange"
                >
                  <Select.Option v-for="item in typeList" :key="item.name" :value="item.name">
                    {{ item.name }}
                  </Select.Option>
                </Select>
              </Form.Item>
              <Form.Item :label="$t('page.dataQuery.form.datasource')" class="query-item">
                <Select
                  v-model:value="globalFormState.datasource"
                  class="query-control"
                  :placeholder="$t('page.dataQuery.placeholder.datasource')"
                  show-search
                  @change="onDatasourceChange"
                >
                  <Select.Option
                    v-for="item in datasourceList"
                    :key="`${item.host}:${item.port}`"
                    :value="`${item.host}:${item.port}`"
                  >
                    <span class="datasource-option-label">
                      <span class="datasource-option-name">{{ item.name }}</span>
                      <span class="datasource-option-status">
                        [
                        <CheckCircleOutlined
                          v-if="isDatasourceStatusOk(item)"
                          class="datasource-status-icon datasource-status-icon--ok"
                        />
                        <CloseCircleOutlined v-else class="datasource-status-icon datasource-status-icon--bad" />
                        {{ isDatasourceStatusOk(item) ? $t('page.dataQuery.datasource.available') : $t('page.dataQuery.datasource.unavailable') }}
                        ]
                      </span>
                    </span>
                  </Select.Option>
                </Select>
              </Form.Item>
              <Form.Item v-if="globalFormState.type !== 'Redis'" :label="$t('page.dataQuery.form.database')" class="query-item">
                <Select
                  v-model:value="globalFormState.database"
                  class="query-control"
                  :placeholder="$t('page.dataQuery.placeholder.database')"
                  show-search
                  @change="onDatabaseChange"
                >
                  <Select.Option v-for="item in databaseList" :key="item.database_name" :value="item.database_name">
                    {{ item.database_name }}
                  </Select.Option>
                </Select>
              </Form.Item>
            </div>
          </Form>
        </Card>
      </Col>
    </Row>

    <div class="query-workspace mt-3">
      <div
        v-if="globalFormState.type !== 'Redis'"
        :class="['workspace-left', { 'workspace-left-collapsed': tablePanelCollapsed }]"
      >
        <div class="workspace-left-header">
          <span v-if="!tablePanelCollapsed">{{ $t('page.dataQuery.tableListTitle') }}</span>
          <Button
            size="small"
            type="text"
            :title="tablePanelCollapsed ? $t('page.dataQuery.expandTableList') : $t('page.dataQuery.collapseTableList')"
            @click="tablePanelCollapsed = !tablePanelCollapsed"
          >
            {{ tablePanelCollapsed ? '>>' : '<<' }}
          </Button>
        </div>
        <div v-if="!tablePanelCollapsed" class="workspace-left-body">
          <Input
            v-model:value="tableKeyword"
            allow-clear
            size="small"
            class="table-search mb-2"
            :placeholder="$t('page.dataQuery.placeholder.searchTable')"
          />
          <List :data-source="filteredTableList" size="small">
            <template #renderItem="{ item }">
              <List.Item>
                <a
                  :class="{ 'table-link-active': selectedTable === item.table_name }"
                  @click="onClickTable(item.table_name)"
                >
                  {{ item.table_name }}
                </a>
              </List.Item>
            </template>
          </List>
        </div>
      </div>
      <div class="workspace-right">
        <Card>
          <Alert
            v-if="globalFormState.database"
            type="info"
            show-icon
            :message="`${$t('page.dataQuery.currentEngine')}${globalFormState.type}，${$t('page.dataQuery.currentDatabase')}${globalFormState.database}`"
            class="mb-2"
          />
          <Tabs
            v-model:active-key="activeTabKey"
            type="editable-card"
            :hide-add="false"
            @edit="(targetKey, action) => (action === 'add' ? addTab() : removeTab(String(targetKey)))"
          >
            <Tabs.TabPane v-for="tab in tabs" :key="tab.key" :tab="tab.title" :closable="tabs.length > 1">
              <Input.TextArea
                :id="`sql-editor-${tab.key}`"
                v-model:value="activeSqlContent"
                :rows="6"
                :placeholder="$t('page.dataQuery.placeholder.sqlInput')"
                @select="handleSqlSelection"
              />
            </Tabs.TabPane>
          </Tabs>
          <div class="mt-2">
            <Space>
              <Button size="small" @click="openAiGenerate = true">{{ $t('page.dataQuery.action.aiGenerate') }}</Button>
              <Button size="small" @click="favoriteSql">{{ $t('page.dataQuery.action.addFavorite') }}</Button>
              <Button size="small" @click="openFavoriteDrawer">{{ $t('page.dataQuery.action.openFavorite') }}</Button>
              <Tag v-if="currentTab?.selectedSql" color="processing">{{ $t('page.dataQuery.selectedFragmentTip') }}</Tag>
            </Space>
          </div>
          <div class="mt-3">
            <Space wrap>
              <Tag v-if="selectedTable" color="blue">{{ $t('page.dataQuery.currentTable') }}{{ selectedTable }}</Tag>
              <Button type="primary" @click="executeQuery('execute')">{{ $t('page.dataQuery.action.executeSql') }}</Button>
              <Button @click="executeQuery('doExplain')">{{ $t('page.dataQuery.action.explain') }}</Button>
              <Button @click="executeQuery('showIndex')">{{ $t('page.dataQuery.action.showIndex') }}</Button>
              <Button @click="executeQuery('showColumn')">{{ $t('page.dataQuery.action.showColumns') }}</Button>
              <Button @click="executeQuery('showCreate')">{{ $t('page.dataQuery.action.showCreate') }}</Button>
              <Button @click="executeQuery('showTableSize')">{{ $t('page.dataQuery.action.showTableSize') }}</Button>
            </Space>
          </div>
        </Card>

        <Card class="mt-3" @copy="onCopyResult">
          <Alert
            v-if="currentTab?.tableDataMsg"
            :type="currentTab?.tableDataSuccess ? 'success' : 'error'"
            :message="currentTab?.tableDataSuccess ? `${$t('page.dataQuery.message.executeSuccessPrefix')}${currentTab?.queryTimes}${$t('page.dataQuery.message.ms')}${currentTab?.tableDataMsg}` : `${$t('page.dataQuery.message.executeFailedPrefix')}${currentTab?.tableDataMsg}`"
            banner
            class="mb-3"
          />
          <div v-if="currentTab?.tableDataSuccess">
            <div class="mb-2 flex justify-end">
              <Space>
                <span>{{ $t('page.dataQuery.foundDataPrefix') }} {{ currentTab?.tableDataTotal }} {{ $t('page.dataQuery.foundDataSuffix') }}</span>
                <Button class="export-excel-btn" type="primary" @click="exportExcel">
                  <template #icon><DownloadOutlined /></template>
                  {{ $t('page.dataQuery.action.exportExcel') }}
                </Button>
              </Space>
            </div>
            <Table
              bordered
              :loading="currentTab?.loading"
              :data-source="currentTab?.tableDataList"
              :columns="currentDataColumns"
              :scroll="{ x: 'max-content', y: 'calc(100vh - 320px)' }"
              size="small"
              :components="{ header: { cell: ResizableHeaderCell } }"
            />
          </div>
        </Card>
      </div>
    </div>

    <Drawer
      v-model:open="openFavorite"
      :title="$t('page.dataQuery.favorite.title')"
      placement="right"
      :width="1000"
    >
      <Table
        row-key="id"
        :data-source="favoriteList"
        :columns="[
          { title: $t('page.dataQuery.favorite.columns.createdAt'), dataIndex: 'gmt_created', key: 'gmt_created', width: 300 },
          { title: $t('page.dataQuery.favorite.columns.content'), dataIndex: 'content', key: 'content' },
          { title: $t('page.dataQuery.favorite.columns.action'), key: 'option', width: 100 }
        ]"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.key === 'content'">
            <TypographyParagraph
              class="favorite-sql-copy mb-0"
              :copyable="{ text: String(record.content ?? '') }"
              :ellipsis="{ rows: 2, expandable: true }"
            >
              {{ record.content }}
            </TypographyParagraph>
          </template>
          <template v-else-if="column.key === 'option'">
            <Button type="link" danger size="small" @click="deleteFavorite(record.id)">{{ $t('page.dataQuery.action.delete') }}</Button>
          </template>
        </template>
      </Table>
    </Drawer>

    <Modal
      v-model:open="openAiGenerate"
      :title="$t('page.dataQuery.ai.title')"
      :confirm-loading="aiGenerating"
      @ok="doAiGenerate"
      @cancel="aiQuestion = ''"
    >
      <Alert
        type="info"
        show-icon
        :message="$t('page.dataQuery.ai.hint')"
        class="mb-3"
      />
      <Input.TextArea
        v-model:value="aiQuestion"
        :rows="6"
        :placeholder="$t('page.dataQuery.ai.placeholder')"
      />
    </Modal>
  </div>
</template>

<style scoped>
.datasource-option-label {
  align-items: center;
  display: inline-flex;
  flex-wrap: wrap;
  gap: 4px;
  max-width: 100%;
}

.datasource-option-name {
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.datasource-option-status {
  align-items: center;
  color: hsl(var(--muted-foreground));
  display: inline-flex;
  flex-shrink: 0;
  font-size: 12px;
  gap: 4px;
}

.datasource-status-icon {
  font-size: 12px;
  vertical-align: -0.15em;
}

.datasource-status-icon--ok {
  color: #52c41a;
}

.datasource-status-icon--bad {
  color: #ff4d4f;
}

.query-grid {
  column-gap: 16px;
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  row-gap: 12px;
}

:deep(.filter-form .query-item) {
  margin-bottom: 0;
  min-width: 0;
}

:deep(.filter-form .query-item .ant-form-item-row) {
  align-items: center;
  display: flex;
}

:deep(.filter-form .query-item .ant-form-item-label) {
  flex: 0 0 5.5rem;
  max-width: 7rem;
  padding-right: 8px;
  text-align: right;
}

:deep(.filter-form .query-item .ant-form-item-control) {
  flex: 1;
  min-width: 0;
}

:deep(.filter-form .query-control) {
  max-width: 100%;
  min-width: 0;
  width: 100%;
}

@media (max-width: 1400px) {
  .query-grid {
    grid-template-columns: repeat(3, minmax(0, 1fr));
  }
}

@media (max-width: 1100px) {
  .query-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }
}

@media (max-width: 768px) {
  .query-grid {
    grid-template-columns: 1fr;
  }
}

.resize-handle {
  position: absolute;
  right: 0;
  top: 0;
  width: 6px;
  height: 100%;
  cursor: col-resize;
  user-select: none;
}

.favorite-sql-copy :deep(.ant-typography-copy) {
  margin-inline-start: 8px;
  vertical-align: middle;
}

.export-excel-btn {
  border-radius: 10px;
  font-size: 12px;
  font-weight: 600;
  height: 24px;
  padding: 0 12px;
}

.export-excel-btn:hover {
  box-shadow: 0 6px 14px rgb(24 144 255 / 28%);
}

.query-workspace {
  border: 1px solid hsl(var(--border));
  border-radius: 12px;
  display: flex;
  overflow: hidden;
}

.workspace-left {
  background: hsl(var(--card));
  border-right: 1px solid hsl(var(--border));
  display: flex;
  flex: 0 0 240px;
  flex-direction: column;
}

.workspace-left-header {
  align-items: center;
  border-bottom: 1px solid hsl(var(--border));
  display: flex;
  font-weight: 600;
  justify-content: space-between;
  padding: 12px 14px;
}

.workspace-left-body {
  height: 760px;
  overflow: auto;
  padding: 4px 8px;
}

.workspace-left-collapsed {
  flex: 0 0 56px;
}

.table-search {
  width: 100%;
}

.table-link-active {
  color: hsl(var(--primary));
  font-weight: 600;
}

.workspace-right {
  flex: 1;
  min-width: 0;
  padding: 0;
}

@media (max-width: 1100px) {
  .query-workspace {
    flex-direction: column;
  }

  .workspace-left {
    border-bottom: 1px solid hsl(var(--border));
    border-right: 0;
    flex-basis: auto;
  }

  .workspace-left-body {
    height: 260px;
  }
}
</style>
