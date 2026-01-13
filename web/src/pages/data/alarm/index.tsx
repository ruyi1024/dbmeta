import React, { useState, useRef } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import ProTable from '@ant-design/pro-table';
import { Button, Divider, message, Popconfirm, Space, Badge } from 'antd';
import { EditOutlined, DeleteOutlined, PlayCircleOutlined, FileTextOutlined } from '@ant-design/icons';
import AlarmModal from './components/AlarmModal';
import AlarmLogModal from './components/AlarmLogModal';
import { useAccess } from 'umi';

const DataAlarmList: React.FC = () => {
  const [alarmModalVisible, setAlarmModalVisible] = useState(false);
  const [alarmModalMode, setAlarmModalMode] = useState<'create' | 'edit'>('create');
  const [currentAlarm, setCurrentAlarm] = useState<any>(null);
  const [logModalVisible, setLogModalVisible] = useState(false);
  const [currentAlarmForLog, setCurrentAlarmForLog] = useState<any>(null);
  const actionRef = useRef<any>(null);
  const access = useAccess();

  const handleAlarmSuccess = () => {
    if (actionRef.current) {
      actionRef.current.reload();
    }
  };

  const handleCreate = () => {
    setAlarmModalMode('create');
    setCurrentAlarm(null);
    setAlarmModalVisible(true);
  };

  const handleEdit = (record: any) => {
    setAlarmModalMode('edit');
    setCurrentAlarm(record);
    setAlarmModalVisible(true);
  };

  const handleDelete = async (record: any) => {
    try {
      const response = await fetch(`/api/v1/data/alarm/delete/${record.id}`, {
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
      console.error('删除告警失败:', error);
      message.error('删除失败，请重试');
    }
  };

  const handleExecute = async (record: any) => {
    try {
      const response = await fetch('/api/v1/data/alarm/execute', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ id: record.id }),
      });
      const result = await response.json();
      if (result.success) {
        message.success('执行成功');
        if (actionRef.current) {
          actionRef.current.reload();
        }
      } else {
        message.error(result.msg || '执行失败');
      }
    } catch (error) {
      console.error('执行告警失败:', error);
      message.error('执行失败，请重试');
    }
  };

  const handleViewLog = (record: any) => {
    setCurrentAlarmForLog(record);
    setLogModalVisible(true);
  };

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '告警名称',
      dataIndex: 'alarm_name',
      key: 'alarm_name',
      width: 200,
    },
    {
      title: '告警描述',
      dataIndex: 'alarm_description',
      key: 'alarm_description',
      width: 250,
      ellipsis: true,
    },
    {
      title: '数据源类型',
      dataIndex: 'datasource_type',
      key: 'datasource_type',
      width: 120,
    },
    {
      title: '规则',
      key: 'rule',
      width: 150,
      render: (_: any, record: any) => {
        const operatorMap: { [key: string]: string } = {
          '>': '大于',
          '<': '小于',
          '=': '等于',
          '>=': '大于等于',
          '<=': '小于等于',
          '!=': '不等于',
        };
        return `数据量 ${operatorMap[record.rule_operator] || record.rule_operator} ${record.rule_value}`;
      },
    },
    {
      title: '接收邮箱',
      dataIndex: 'email_to',
      key: 'email_to',
      width: 200,
      ellipsis: true,
    },
    {
      title: 'Cron表达式',
      dataIndex: 'cron_expression',
      key: 'cron_expression',
      width: 150,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: number) => (
        <Badge status={status === 1 ? 'success' : 'default'} text={status === 1 ? '启用' : '禁用'} />
      ),
    },
    {
      title: '上次运行',
      dataIndex: 'last_run_time',
      key: 'last_run_time',
      width: 180,
      render: (text: string) => text || '-',
    },
    {
      title: '下次运行',
      dataIndex: 'next_run_time',
      key: 'next_run_time',
      width: 180,
      render: (text: string) => text || '-',
    },
    {
      title: '操作',
      key: 'action',
      width: 200,
      fixed: 'right' as const,
      render: (_: any, record: any) => (
        <Space split={<Divider type="vertical" />}>
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
            onClick={() => handleExecute(record)}
          >
            执行
          </Button>
          <Button
            type="link"
            size="small"
            icon={<FileTextOutlined />}
            onClick={() => handleViewLog(record)}
          >
            日志
          </Button>
          <Popconfirm
            title="确定要删除这个告警吗？"
            onConfirm={() => {
              if (!access.canAdmin) { message.error('操作权限受限，请联系平台管理员'); return }
              handleDelete(record)
            }}
            okText="确定"
            cancelText="取消"
          >
            <Button
              type="link"
              danger
              size="small"
              icon={<DeleteOutlined />}
            >
              删除
            </Button>
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <PageContainer>
      <ProTable
        headerTitle="数据告警"
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
            创建告警
          </Button>,
        ]}
        request={async (params) => {
          try {
            const response = await fetch('/api/v1/data/alarm/list?' + new URLSearchParams(params as any));
            const data = await response.json();
            return {
              data: data.data || [],
              success: data.success,
              total: data.total || 0,
            };
          } catch (error) {
            console.error('获取告警列表失败:', error);
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
        scroll={{ x: 1810 }}
      />

      {/* 告警模态框 */}
      <AlarmModal
        open={alarmModalVisible}
        mode={alarmModalMode}
        editData={currentAlarm}
        onCancel={() => setAlarmModalVisible(false)}
        onSuccess={handleAlarmSuccess}
      />

      {/* 告警日志模态框 */}
      <AlarmLogModal
        open={logModalVisible}
        alarmId={currentAlarmForLog?.id}
        alarmName={currentAlarmForLog?.alarm_name}
        onCancel={() => setLogModalVisible(false)}
      />
    </PageContainer>
  );
};

export default DataAlarmList;

