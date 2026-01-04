import React, { useState, useEffect } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { Card, Button, message, Table, Space } from 'antd';

const AnalysisTaskTest: React.FC = () => {
  const [tasks, setTasks] = useState([]);
  const [loading, setLoading] = useState(false);

  const fetchTasks = async () => {
    setLoading(true);
    try {
      const response = await fetch('/api/v1/task/analysis/list');
      const data = await response.json();
      if (data.success) {
        setTasks(data.data || []);
        message.success(`获取到 ${data.data?.length || 0} 个任务`);
      } else {
        message.error(data.msg || '获取任务列表失败');
      }
    } catch (error) {
      console.error('获取任务列表失败:', error);
      message.error('网络请求失败');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchTasks();
  }, []);

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '任务名称',
      dataIndex: 'task_name',
      key: 'task_name',
      width: 200,
    },
    {
      title: '任务描述',
      dataIndex: 'task_description',
      key: 'task_description',
      width: 300,
    },
    {
      title: 'Cron表达式',
      dataIndex: 'cron_expression',
      key: 'cron_expression',
      width: 150,
    },
    {
      title: '报告邮箱',
      dataIndex: 'report_email',
      key: 'report_email',
      width: 200,
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: number) => (
        <span style={{ color: status === 1 ? 'green' : 'red' }}>
          {status === 1 ? '启用' : '禁用'}
        </span>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
    },
  ];

  return (
    <PageContainer>
      <Card title="智能任务测试页面">
        <Space style={{ marginBottom: 16 }}>
          <Button type="primary" onClick={fetchTasks} loading={loading}>
            刷新数据
          </Button>
          <Button onClick={() => {
            message.info('创建任务功能正在开发中...');
          }}>
            创建任务
          </Button>
        </Space>
        
        <Table
          columns={columns}
          dataSource={tasks}
          rowKey="id"
          loading={loading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 条记录`,
          }}
        />
      </Card>
    </PageContainer>
  );
};

export default AnalysisTaskTest; 