import React, { useRef, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';
import { Button, Tag, Space, message, Popconfirm } from 'antd';
import { PlusOutlined, DeleteOutlined, PlayCircleOutlined } from '@ant-design/icons';
import { queryTasks, createTask, updateTaskStatus, deleteTask } from './service';
import type { TaskListItem } from './data.d';
import CreateForm from './components/CreateForm';

const TaskList: React.FC<{}> = () => {
  const actionRef = useRef<ActionType>();
  const [createModalVisible, handleCreateModalVisible] = useState<boolean>(false);

  const handleAdd = async (fields: TaskListItem) => {
    try {
      const response = await createTask(fields);
      if (response.code === 200) {
        message.success('创建成功');
        handleCreateModalVisible(false);
        if (actionRef.current) {
          actionRef.current.reload();
        }
        return true;
      } else {
        message.error(response.msg || '创建失败');
        return false;
      }
    } catch (error) {
      message.error('创建失败');
      return false;
    }
  };

  const handleDelete = async (id: number) => {
    try {
      const response = await deleteTask(id);
      if (response.code === 200) {
        message.success('删除成功');
        if (actionRef.current) {
          actionRef.current.reload();
        }
      } else {
        message.error(response.msg || '删除失败');
      }
    } catch (error) {
      message.error('删除失败');
    }
  };

  const handleStart = async (record: TaskListItem) => {
    try {
      const response = await updateTaskStatus({
        id: record.id,
        status: 'running',
      });
      if (response.code === 200) {
        message.success('任务已启动');
        if (actionRef.current) {
          actionRef.current.reload();
        }
      } else {
        message.error(response.msg || '启动失败');
      }
    } catch (error) {
      message.error('启动失败');
    }
  };

  const columns: ProColumns<TaskListItem>[] = [
    {
      title: '任务名称',
      dataIndex: 'taskName',
      width: 200,
      sorter: true,
    },
    {
      title: '任务类型',
      dataIndex: 'taskType',
      width: 120,
      valueEnum: {
        '全量': { text: '全量评估' },
        '增量': { text: '增量评估' },
        '定时': { text: '定时评估' },
      },
    },
    {
      title: '数据源',
      dataIndex: 'datasourceId',
      width: 100,
      hideInSearch: true,
    },
    {
      title: '数据库',
      dataIndex: 'databaseName',
      width: 150,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      valueEnum: {
        'pending': { text: '待执行', status: 'Default' },
        'running': { text: '执行中', status: 'Processing' },
        'success': { text: '成功', status: 'Success' },
        'failed': { text: '失败', status: 'Error' },
      },
      render: (_, record) => {
        const colorMap: any = {
          'pending': 'default',
          'running': 'processing',
          'success': 'success',
          'failed': 'error',
        };
        const textMap: any = {
          'pending': '待执行',
          'running': '执行中',
          'success': '成功',
          'failed': '失败',
        };
        return <Tag color={colorMap[record.status]}>{textMap[record.status]}</Tag>;
      },
    },
    {
      title: '开始时间',
      dataIndex: 'startTime',
      width: 180,
      valueType: 'dateTime',
      hideInSearch: true,
      sorter: true,
    },
    {
      title: '结束时间',
      dataIndex: 'endTime',
      width: 180,
      valueType: 'dateTime',
      hideInSearch: true,
    },
    {
      title: '执行时长',
      dataIndex: 'duration',
      width: 100,
      hideInSearch: true,
      render: (text: number) => text ? `${text}秒` : '-',
    },
    {
      title: '创建人',
      dataIndex: 'createdBy',
      width: 100,
      hideInSearch: true,
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      width: 180,
      valueType: 'dateTime',
      hideInSearch: true,
      sorter: true,
    },
    {
      title: '操作',
      valueType: 'option',
      width: 200,
      render: (_, record) => [
        record.status === 'pending' && (
          <Button
            key="start"
            type="link"
            size="small"
            icon={<PlayCircleOutlined />}
            onClick={() => handleStart(record)}
          >
            启动
          </Button>
        ),
        <Popconfirm
          key="delete"
          title="确定要删除这个任务吗？"
          onConfirm={() => handleDelete(record.id)}
        >
          <Button
            type="link"
            danger
            size="small"
            icon={<DeleteOutlined />}
          >
            删除
          </Button>
        </Popconfirm>,
      ].filter(Boolean),
    },
  ];

  return (
    <PageContainer>
      <ProTable<TaskListItem>
        headerTitle="评估任务管理"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        request={async (params) => {
          const response = await queryTasks(params);
          return {
            data: response.data?.list || [],
            success: response.code === 200,
            total: response.data?.total || 0,
          };
        }}
        columns={columns}
        pagination={{
          pageSize: 10,
          showSizeChanger: true,
        }}
        toolBarRender={() => [
          <Button
            type="primary"
            key="primary"
            onClick={() => {
              handleCreateModalVisible(true);
            }}
          >
            <PlusOutlined /> 新建任务
          </Button>,
        ]}
      />

      <CreateForm
        onCancel={() => handleCreateModalVisible(false)}
        modalVisible={createModalVisible}
        onSubmit={async (value) => {
          const success = await handleAdd(value as TaskListItem);
          if (success) {
            handleCreateModalVisible(false);
            if (actionRef.current) {
              actionRef.current.reload();
            }
          }
        }}
      />
    </PageContainer>
  );
};

export default TaskList;

