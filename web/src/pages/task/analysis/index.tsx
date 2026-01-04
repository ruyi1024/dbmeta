import React, { useState, useRef } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import ProTable from '@ant-design/pro-table';
import { Button, Divider, message, Popconfirm, Space } from 'antd';
import { EditOutlined, DeleteOutlined, PlayCircleOutlined, FileTextOutlined } from '@ant-design/icons';
import TaskModal from './components/TaskModal';
import AnalysisTaskLogModal from './components/AnalysisTaskLogModal';
import { useAccess } from 'umi';

const AnalysisTaskList: React.FC = () => {
  const [taskModalVisible, setTaskModalVisible] = useState(false);
  const [taskModalMode, setTaskModalMode] = useState<'create' | 'edit'>('create');
  const [currentTask, setCurrentTask] = useState<any>(null);
  const [logModalVisible, setLogModalVisible] = useState(false);
  const [currentTaskForLog, setCurrentTaskForLog] = useState<any>(null);
  const actionRef = useRef<any>(null);
  const access = useAccess();

  const handleTaskSuccess = () => {
    if (actionRef.current) {
      actionRef.current.reload();
    }
  };

  const handleCreate = () => {
    setTaskModalMode('create');
    setCurrentTask(null);
    setTaskModalVisible(true);
  };

  const handleEdit = (record: any) => {
    setTaskModalMode('edit');
    setCurrentTask(record);
    setTaskModalVisible(true);
  };

  const handleDelete = async (record: any) => {
    try {
      const response = await fetch(`/api/v1/task/analysis/delete/${record.id}`, {
        method: 'DELETE',
      });
      const result = await response.json();
      if (result.success) {
        message.success('删除成功');
        if (actionRef.current) {
          actionRef.current.reload();
        }
      } else {
        message.error(result.msg || '删除失败');
      }
    } catch (error) {
      console.error('删除任务失败:', error);
      message.error('删除失败，请重试');
    }
  };

  const handleExecute = async (record: any) => {
    try {
      const response = await fetch('/api/v1/task/analysis/execute', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ id: record.id }),
      });
      const result = await response.json();
      if (result.success) {
        message.success('任务执行已启动');
      } else {
        message.error(result.msg || '任务执行失败');
      }
    } catch (error) {
      console.error('执行任务失败:', error);
      message.error('执行失败，请重试');
    }
  };

  const handleViewLogs = (record: any) => {
    setCurrentTaskForLog(record);
    setLogModalVisible(true);
  };

  const columns = [
    {
      title: '任务名称',
      dataIndex: 'task_name',
      width: 200,
    },
    {
      title: '任务描述',
      dataIndex: 'task_description',
      width: 300,
      ellipsis: true,
    },
    {
      title: '计划任务',
      dataIndex: 'cron_expression',
      width: 150,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      valueEnum: {
        0: { text: '禁用', status: 'default' },
        1: { text: '启用', status: 'success' },
      },
    },
    {
      title: '最后执行',
      dataIndex: 'last_run_time',
      width: 160,
    },
    {
      title: '下次执行',
      dataIndex: 'next_run_time',
      width: 160,
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      width: 160,
      valueType: 'dateTime',
    },
    {
      title: '操作',
      key: 'action',
      width: 280,
      fixed: 'right' as const,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EditOutlined />}
            onClick={() => {
              if (!access.canAdmin) { message.error('操作权限受限，请联系平台管理员'); return }
              handleEdit(record)
            }}
          >
            编辑
          </Button>
          <Button
            type="link"
            size="small"
            icon={<PlayCircleOutlined />}
            onClick={() => {
              if (!access.canAdmin) { message.error('操作权限受限，请联系平台管理员'); return }
              handleExecute(record)
            }}
          >
            执行
          </Button>
          <Button
            type="link"
            size="small"
            icon={<FileTextOutlined />}
            onClick={() => handleViewLogs(record)}
          >
            日志
          </Button>
        
          <Popconfirm
            title={`确认要删除数据【${record.task_name}】,删除后不可恢复，是否继续？`}
            placement={"left"}
            onConfirm={ () => {
              if (!access.canAdmin) { message.error('操作权限受限，请联系平台管理员'); return }
              handleDelete(record)
            }}
          >
            <a style={{ color: 'red' }}><DeleteOutlined /> 删除</a>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <PageContainer>
      <ProTable
        headerTitle="智能任务"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={() => [
          <Button
            type="primary"
            key="create"
            onClick={() => {
              if (!access.canAdmin) { message.error('操作权限受限，请联系平台管理员'); return }
              handleCreate()
            }}
          >
            创建任务
          </Button>,
        ]}
        request={async (params) => {
          try {
            const response = await fetch('/api/v1/task/analysis/list?' + new URLSearchParams(params as any));
            const data = await response.json();
            return {
              data: data.data || [],
              success: data.success,
              total: data.total || 0,
            };
          } catch (error) {
            console.error('获取任务列表失败:', error);
            return {
              data: [],
              success: false,
              total: 0,
            };
          }
        }}
        columns={columns}
        pagination={{
          pageSize: 10,
        }}
      />

      {/* 任务模态框 */}
      <TaskModal
        open={taskModalVisible}
        mode={taskModalMode}
        editData={currentTask}
        onCancel={() => setTaskModalVisible(false)}
        onSuccess={handleTaskSuccess}
      />

      {/* 任务日志模态框 */}
      <AnalysisTaskLogModal
        open={logModalVisible}
        taskId={currentTaskForLog?.id}
        taskName={currentTaskForLog?.task_name}
        onCancel={() => setLogModalVisible(false)}
      />
    </PageContainer>
  );
};

export default AnalysisTaskList; 