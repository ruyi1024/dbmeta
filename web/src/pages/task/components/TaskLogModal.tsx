import React, { useState, useRef } from 'react';
import { Modal, Badge, Space, Input, Button } from 'antd';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';
import { TaskLogItem, TaskLogParams } from '../data.d';
import { queryTaskLogs } from '../service';

const { Search } = Input;

interface TaskLogModalProps {
  open: boolean;
  onCancel: () => void;
  taskKey: string;
  taskName: string;
}

const TaskLogModal: React.FC<TaskLogModalProps> = ({ open, onCancel, taskKey, taskName }) => {
  const actionRef = useRef<ActionType>();
  const [statusFilter, setStatusFilter] = useState<string>('');

  const columns: ProColumns<TaskLogItem>[] = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
      hideInSearch: true,
    },
    {
      title: '开始时间',
      dataIndex: 'start_time',
      valueType: 'dateTime',
      hideInSearch: true,
      width: 180,
    },
    {
      title: '完成时间',
      dataIndex: 'complete_time',
      valueType: 'dateTime',
      hideInSearch: true,
      width: 180,
      render: (text) => {
        if (!text) {
          return <span style={{ color: '#999' }}>未完成</span>;
        }
        return text;
      },
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      valueEnum: {
        running: { text: '执行中', status: 'Processing' },
        success: { text: '成功', status: 'Success' },
        failed: { text: '失败', status: 'Error' },
      },
      render: (text, record) => {
        if (record.status === 'running') {
          return <Badge status="processing" text="执行中" />;
        } else if (record.status === 'success') {
          return <Badge status="success" text="成功" />;
        } else if (record.status === 'failed') {
          return <Badge status="error" text="失败" />;
        }
        return <Badge status="default" text={text} />;
      },
    },
    {
      title: '执行结果',
      dataIndex: 'result',
      ellipsis: true,
      hideInSearch: true,
      width: 300,
    },
    {
      title: '创建时间',
      dataIndex: 'gmt_created',
      valueType: 'dateTime',
      hideInSearch: true,
      width: 180,
    },
  ];

  const handleRequest = async (params: any, sorter: any, filter: any) => {
    const requestParams: TaskLogParams = {
      task_key: taskKey,
      pageSize: params.pageSize || 10,
      currentPage: params.current || 1,
      sorter: sorter,
    };

    // 添加状态过滤
    if (statusFilter) {
      requestParams.status = statusFilter;
    }

    try {
      const response = await queryTaskLogs(requestParams);
      return {
        data: response.data || [],
        success: response.success,
        total: response.total || 0,
      };
    } catch (error) {
      return {
        data: [],
        success: false,
        total: 0,
      };
    }
  };

  const handleStatusChange = (value: string) => {
    setStatusFilter(value);
    if (actionRef.current) {
      actionRef.current.reload();
    }
  };

  const handleReset = () => {
    setStatusFilter('');
    if (actionRef.current) {
      actionRef.current.reload();
    }
  };

  return (
    (<Modal
      title={`${taskName} - 运行日志`}
      open={open}
      onCancel={onCancel}
      width={1200}
      footer={null}
      destroyOnClose
    >
      <div style={{ marginBottom: 16 }}>
        <Space>
          <span>状态过滤：</span>
          <Input
            placeholder="输入状态 (running/success/failed)"
            value={statusFilter}
            onChange={(e) => handleStatusChange(e.target.value)}
            style={{ width: 200 }}
            allowClear
          />
          <Button onClick={handleReset}>重置</Button>
        </Space>
      </div>
      <ProTable<TaskLogItem>
        actionRef={actionRef}
        rowKey="id"
        search={false}
        options={false}
        pagination={{
          pageSize: 10,
          showSizeChanger: true,
          showQuickJumper: true,
        }}
        request={handleRequest}
        columns={columns}
        size="small"
        scroll={{ y: 400 }}
      />
    </Modal>)
  );
};

export default TaskLogModal; 