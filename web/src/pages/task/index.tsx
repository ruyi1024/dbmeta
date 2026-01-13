import { PlusOutlined, FormOutlined, DeleteOutlined, FileTextOutlined, PlayCircleOutlined, InfoCircleOutlined } from '@ant-design/icons';
import { Button, Divider, message, Popconfirm, Select, Modal, Card, Tooltip } from 'antd';

const { Option } = Select;
import React, { useState, useRef, useEffect } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';
import CreateForm from './components/CreateForm';
import UpdateForm from './components/UpdateForm';
import { TableListItem } from './data.d';
import { query, update, add, remove, queryTaskLogs, executeTask, queryTodayStats } from './service';
import { useAccess } from 'umi';
import { Badge } from 'antd';
import { ChartCard, MiniArea } from '../portal/components/Charts';
import styles from '../portal/style.less';

/**
 * 添加节点
 * @param fields
 */
const handleAdd = async (fields: TableListItem) => {
  const hide = message.loading('正在添加');
  try {
    await add({ ...fields });
    hide();
    message.success('添加成功');
    return true;
  } catch (error) {
    hide();
    message.error('添加失败请重试！');
    return false;
  }
};

/**
 * 更新节点
 * @param fields
 */
const handleUpdate = async (fields: Partial<TableListItem>, task_key: string) => {
  const hide = message.loading('正在配置');
  try {
    await update({
      ...fields,
      "task_key": task_key,
    });
    hide();
    message.success('修改成功');
    return true;
  } catch (error) {
    hide();
    message.error('修改失败请重试！');
    return false;
  }
};

/**
 *  删除节点
 * @param selectedRows
 */
const handleRemove = async (task_key: string) => {
  const hide = message.loading('正在删除');
  try {
    await remove({
      "task_key": task_key,
    });
    hide();
    message.success('删除成功，即将刷新');
    return true;
  } catch (error) {
    hide();
    message.error('删除失败，请重试');
    return false;
  }
};

const formInitValue = { "task_key": "", "task_name": "", "task_description": "", "crontab": "", "enable": "" }

const TableList: React.FC<{}> = () => {
  const [createModalVisible, handleModalVisible] = useState<boolean>(false);
  const [updateModalVisible, handleUpdateModalVisible] = useState<boolean>(false);
  const [logModalVisible, setLogModalVisible] = useState<boolean>(false);
  const [currentTask, setCurrentTask] = useState<{ taskKey: string; taskName: string }>({ taskKey: '', taskName: '' });
  const [formValues, setFormValues] = useState(formInitValue);
  const actionRef = useRef<ActionType>();
  const access = useAccess();

  // 查看任务日志
  const handleViewLogs = (taskKey: string, taskName: string) => {
    setCurrentTask({ taskKey, taskName });
    setLogModalVisible(true);
  };

  // 手工运行任务
  const handleExecuteTask = async (taskKey: string, taskName: string) => {
    const hide = message.loading('正在执行任务...');
    try {
      const response = await executeTask(taskKey);
      hide();
      if (response.success) {
        message.success(`任务【${taskName}】已开始执行`);
      } else {
        message.error(response.msg || '执行失败');
      }
    } catch (error) {
      hide();
      message.error('执行失败，请重试');
    }
  };


  const columns: ProColumns<TableListItem>[] = [
    {
      title: '任务标识',
      dataIndex: 'task_key',
      initialValue: formValues.task_key,
      sorter: true,
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
    },
    {
      title: '任务名',
      dataIndex: 'task_name',
      initialValue: formValues.task_name,
      sorter: true,
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
    },
    {
      title: '任务描述',
      dataIndex: 'task_description',
      initialValue: formValues.task_description,
      sorter: false,
      search: false,
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
    },
    {
      title: '计划任务',
      dataIndex: 'crontab',
      initialValue: formValues.crontab,
      sorter: false,
      search: false,
    },
    {
      title: '启用',
      dataIndex: 'enable',
      filters: true,
      onFilter: true,
      valueEnum: {
        0: { text: '禁用', status: 'Default' },
        1: { text: '启用', status: 'Success' },
      },
      sorter: false,
      initialValue: formValues.enable,
      formItemProps: {
        rules: [
          {
            required: true,
            message: '此项为必填项',
          },
        ],
      },
      renderFormItem: (_, { type, defaultRender, ...rest }, form) => {
        return (
          <Select>
            <Option key={0} value={0}>
              否
            </Option>
            <Option key={1} value={1}>
              是
            </Option>
          </Select>
        );
      },
    },
    {
      title: '上次运行时间',
      dataIndex: 'last_run_time',
      sorter: true,
      valueType: 'dateTime',
      hideInForm: true,
      search: false,
      render: (text: any) => text || '-',
    },
    {
      title: '下次运行时间',
      dataIndex: 'next_run_time',
      sorter: true,
      valueType: 'dateTime',
      hideInForm: true,
      search: false,
      render: (text: any) => text || '-',
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      render: (_, record) => (
        <>
          <a
            onClick={() => {
              handleUpdateModalVisible(true);
              setFormValues(record);
            }}
          >
            <FormOutlined />修改
          </a>
          <Divider type="vertical" />
          <Popconfirm
            title={`确认要删除数据【${record.task_name}】,删除后不可恢复，是否继续？`}
            placement={"left"}
            onConfirm={async () => {
              if (!access.canAdmin) { message.error('操作权限受限，请联系平台管理员'); return }
              const success = await handleRemove(record.task_key);
              if (success) {
                if (actionRef.current) {
                  actionRef.current.reload();
                }
              }
            }}
          >
            <a><DeleteOutlined />删除</a>
          </Popconfirm>
          <Divider type="vertical" />
          <a 
            onClick={() => {
              const title = record.enable === 1 ? '确认执行任务' : '任务未启用';
              const content = record.enable === 1 
                ? `确认要手工执行任务【${record.task_name}】吗？`
                : `任务【${record.task_name}】当前未启用，是否仍要手工执行？`;
              Modal.confirm({
                title: title,
                content: content,
                onOk: () => handleExecuteTask(record.task_key, record.task_name),
              });
            }}
          >
            <PlayCircleOutlined />手工运行
          </a>
          <Divider type="vertical" />
          <a onClick={() => handleViewLogs(record.task_key, record.task_name)}>
            <FileTextOutlined />查看运行日志
          </a>
        </>
      ),
    },
  ];

  // 任务运行统计数据（模拟数据，后续由后端提供）
  const [taskStats, setTaskStats] = useState({
    todayExecuteCount: 0,
    todayFailedCount: 0,
    hour24ExecuteCount: 0,
    successRate: 0,
    successRateStr: '0%',
    todayExecuteTrendData: [] as Array<{ x: string; y: number }>,
    todayFailedTrendData: [] as Array<{ x: string; y: number }>,
    hour24ExecuteTrendData: [] as Array<{ x: string; y: number }>,
    successRateTrendData: [] as Array<{ x: string; y: number }>,
    lastExecuteTime: '',
  });

  // 获取今日任务执行统计
  useEffect(() => {
    const fetchTodayStats = async () => {
      try {
        const response = await queryTodayStats();
        if (response.success) {
          setTaskStats(prev => ({
            ...prev,
            todayExecuteCount: response.todayTotal || 0,
            todayFailedCount: response.todayFailedTotal || 0,
            hour24ExecuteCount: response.hour24Total || 0,
            todayExecuteTrendData: response.todayTrend || [],
            todayFailedTrendData: response.todayFailedTrend || [],
            hour24ExecuteTrendData: response.hour24Trend || [],
            successRate: response.successRate || 0,
            successRateStr: response.successRateStr || '0%',
            successRateTrendData: response.successRateTrend || [],
            lastExecuteTime: response.lastExecuteTime || '',
          }));
        }
      } catch (error) {
        console.error('获取今日任务统计失败:', error);
        // 如果API调用失败，使用模拟数据
        const mockData = {
          todayExecuteCount: 0,
          todayFailedCount: 0,
          hour24ExecuteCount: 0,
          todayExecuteTrendData: Array.from({ length: 24 }, (_, i) => ({
            x: `${String(i).padStart(2, '0')}:00`,
            y: 0,
          })),
          todayFailedTrendData: Array.from({ length: 24 }, (_, i) => ({
            x: `${String(i).padStart(2, '0')}:00`,
            y: 0,
          })),
          hour24ExecuteTrendData: Array.from({ length: 24 }, (_, i) => ({
            x: `${String(i).padStart(2, '0')}:00`,
            y: 0,
          })),
          successRate: 0,
          successRateStr: '0%',
          successRateTrendData: Array.from({ length: 24 }, (_, i) => ({
            x: `${String(i).padStart(2, '0')}:00`,
            y: 0,
          })),
          lastExecuteTime: '',
        };
        setTaskStats(prev => ({
          ...prev,
          ...mockData,
        }));
      }
    };

    fetchTodayStats();
    // 每5分钟刷新一次数据
    const interval = setInterval(fetchTodayStats, 5 * 60 * 1000);
    return () => clearInterval(interval);
  }, []);

  return (
    (<PageContainer title="计划任务管理平台">
      {/* 任务运行图表分析 */}
      <Card
        className={styles.projectList}
        style={{ marginBottom: 16 }}
        title="任务运行分析概览"
        bordered={false}
        loading={false}
        bodyStyle={{ padding: 0, display: 'flex', flexWrap: 'wrap', width: '100%' }}
      >
        <Card.Grid className={styles.projectGrid} style={{ width: '25%', flex: '1 1 25%', minWidth: '200px', padding: 0 }} key="1">
          <ChartCard
            bordered={false}
            loading={false}
            bodyStyle={{ padding: '20px 16px 8px 16px' }}
            title="今日任务执行次数"
            action={
              <Tooltip title="今日任务执行总次数和每小时趋势">
                <InfoCircleOutlined />
              </Tooltip>
            }
            total={taskStats.todayExecuteCount}
            footer={taskStats.lastExecuteTime ? `最新执行时间：${taskStats.lastExecuteTime}` : ''}
            contentHeight={46}
          >
            <MiniArea color="#1979C9" data={taskStats.todayExecuteTrendData} animate={true} />
          </ChartCard>
        </Card.Grid>
        <Card.Grid className={styles.projectGrid} style={{ width: '25%', flex: '1 1 25%', minWidth: '200px', padding: 0,flexWrap:''}} key="2">
          <ChartCard
            bordered={false}
            loading={false}
            bodyStyle={{ padding: '20px 16px 8px 16px' }}
            title="今日失败次数"
            action={
              <Tooltip title="今日任务失败次数和每小时趋势">
                <InfoCircleOutlined />
              </Tooltip>
            }
            total={taskStats.todayFailedCount}
            footer={taskStats.lastExecuteTime ? `最新执行时间：${taskStats.lastExecuteTime}` : ''}
            contentHeight={46}
          >
            <MiniArea color="#F5222D" data={taskStats.todayFailedTrendData} animate={true} />
          </ChartCard>
        </Card.Grid>
        <Card.Grid className={styles.projectGrid} style={{ width: '25%', flex: '1 1 25%', minWidth: '200px', padding: 0 }} key="3">
          <ChartCard
            bordered={false}
            loading={false}
            bodyStyle={{ padding: '20px 16px 8px 16px' }}
            title="24小时执行次数"
            action={
              <Tooltip title="近24小时任务执行总次数和每小时趋势">
                <InfoCircleOutlined />
              </Tooltip>
            }
            total={taskStats.hour24ExecuteCount}
            footer={taskStats.lastExecuteTime ? `最新执行时间：${taskStats.lastExecuteTime}` : ''}
            contentHeight={46}
          >
            <MiniArea color="#52C41A" data={taskStats.hour24ExecuteTrendData} animate={true} />
          </ChartCard>
        </Card.Grid>
        <Card.Grid className={styles.projectGrid} style={{ width: '25%', flex: '1 1 25%', minWidth: '200px', padding: 0 }} key="4">
          <ChartCard
            bordered={false}
            loading={false}
            bodyStyle={{ padding: '20px 16px 8px 16px' }}
            title="任务执行成功率"
            action={
              <Tooltip title="今日任务执行成功率趋势">
                <InfoCircleOutlined />
              </Tooltip>
            }
            total={taskStats.successRateStr}
            footer={taskStats.lastExecuteTime ? `最新执行时间：${taskStats.lastExecuteTime}` : ''}
            contentHeight={46}
          >
            <MiniArea color="#FAAD14" data={taskStats.successRateTrendData} animate={true} />
          </ChartCard>
        </Card.Grid>
      </Card>

      <ProTable<TableListItem>
        headerTitle="数据列表"
        actionRef={actionRef}
        rowKey="id"
        search={true}
        toolBarRender={() => [
          <Button type="primary"
            onClick={() => {
              handleModalVisible(true);
              setFormValues(formInitValue);
            }}
          >
            <PlusOutlined /> 新建
          </Button>,
        ]}
        request={(params, sorter, filter) => query({ ...params, sorter, filter })}
        columns={columns}
      />
      <CreateForm onCancel={() => handleModalVisible(false)} modalVisible={createModalVisible}>
        <ProTable<TableListItem, TableListItem>
          onSubmit={async (value) => {
            if (!access.canAdmin) { message.error('操作权限受限，请联系平台管理员'); return }
            const success = await handleAdd(value);
            if (success) {
              handleModalVisible(false);
              if (actionRef.current) {
                actionRef.current.reload();
              }
            }
          }}
          rowKey="task_key"
          type="form"
          columns={columns}
        />
      </CreateForm>
      <UpdateForm onCancel={() => handleUpdateModalVisible(false)} updateModalVisible={updateModalVisible}>
        <ProTable<TableListItem, TableListItem>
          onSubmit={async (value) => {
            if (!access.canAdmin) { message.error('操作权限受限，请联系平台管理员'); return }
            const success = await handleUpdate(value, formValues.task_key);
            if (success) {
              handleUpdateModalVisible(false);
              if (actionRef.current) {
                actionRef.current.reload();
              }
            }
          }}
          rowKey="task_key"
          type="form"
          columns={columns}
        />
      </UpdateForm>
      {/* 任务日志弹窗 */}
      <Modal
        title={`${currentTask.taskName} - 运行日志`}
        open={logModalVisible}
        onCancel={() => setLogModalVisible(false)}
        width={1200}
        footer={null}
        destroyOnClose
      >
        <ProTable
          rowKey="id"
          search={false}
          options={false}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
          }}
          request={async (params) => {
            const response = await queryTaskLogs({
              task_key: currentTask.taskKey,
              pageSize: params.pageSize || 10,
              currentPage: params.current || 1,
            });
            return {
              data: response.data || [],
              success: response.success,
              total: response.total || 0,
            };
          }}
          columns={[
            {
              title: 'ID',
              dataIndex: 'id',
              width: 80,
            },
            {
              title: '开始时间',
              dataIndex: 'start_time',
              valueType: 'dateTime',
              width: 180,
            },
            {
              title: '完成时间',
              dataIndex: 'complete_time',
              valueType: 'dateTime',
              width: 180,
              render: (text) => text || '未完成',
            },
            {
              title: '状态',
              dataIndex: 'status',
              width: 100,
              render: (text) => {
                if (text === 'running') return <Badge status="processing" text="执行中" />;
                if (text === 'success') return <Badge status="success" text="成功" />;
                if (text === 'failed') return <Badge status="error" text="失败" />;
                return <Badge status="default" text={text} />;
              },
            },
            {
              title: '执行结果',
              dataIndex: 'result',
              ellipsis: true,
              width: 300,
            },
            {
              title: '创建时间',
              dataIndex: 'gmt_created',
              valueType: 'dateTime',
              width: 180,
            },
          ]}
          size="small"
          scroll={{ y: 400 }}
        />
      </Modal>
    </PageContainer>)
  );
};

export default TableList;
