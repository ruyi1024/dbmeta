import React, { useState, useEffect } from 'react';
import { Modal, Table, Tag, Space, Form, DatePicker, Select, Button, message } from 'antd';
import { queryDataAlarmLogs } from '../service';
import type { DataAlarmLogItem, DataAlarmLogParams } from '../data.d';
import dayjs from 'dayjs';

const { RangePicker } = DatePicker;
const { Option } = Select;

interface AlarmLogModalProps {
  open: boolean;
  onCancel: () => void;
  alarmId?: number;
  alarmName?: string;
}

const AlarmLogModal: React.FC<AlarmLogModalProps> = ({
  open,
  onCancel,
  alarmId,
  alarmName,
}) => {
  const [logs, setLogs] = useState<DataAlarmLogItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [filters, setFilters] = useState<DataAlarmLogParams>({});
  const [form] = Form.useForm();

  const fetchLogs = async (params?: DataAlarmLogParams) => {
    if (!alarmId) return;

    setLoading(true);
    try {
      const response = await queryDataAlarmLogs({
        alarm_id: alarmId,
        pageSize,
        currentPage,
        ...filters,
        ...params,
      });

      if (response.success) {
        setLogs(response.data || []);
        setTotal(response.total || 0);
      }
    } catch (error) {
      console.error('获取日志失败:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (open && alarmId) {
      fetchLogs();
    }
  }, [open, alarmId, currentPage, pageSize]);

  const handleSearch = (values: any) => {
    const newFilters: DataAlarmLogParams = {};
    
    if (values.status) {
      newFilters.status = values.status;
    }
    
    if (values.dateRange && values.dateRange.length === 2) {
      newFilters.start_date = values.dateRange[0].format('YYYY-MM-DD');
      newFilters.end_date = values.dateRange[1].format('YYYY-MM-DD');
    }

    setFilters(newFilters);
    setCurrentPage(1);
    fetchLogs(newFilters);
  };

  const handleReset = () => {
    form.resetFields();
    setFilters({});
    setCurrentPage(1);
    fetchLogs({});
  };

  const getStatusTag = (status: string) => {
    const statusMap: { [key: string]: { color: string; text: string } } = {
      running: { color: 'processing', text: '执行中' },
      success: { color: 'success', text: '成功' },
      failed: { color: 'error', text: '失败' },
      triggered: { color: 'warning', text: '已触发' },
    };
    const config = statusMap[status] || { color: 'default', text: status };
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: '开始时间',
      dataIndex: 'start_time',
      key: 'start_time',
      width: 180,
    },
    {
      title: '完成时间',
      dataIndex: 'complete_time',
      key: 'complete_time',
      width: 180,
      render: (text: string) => text || '-',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => getStatusTag(status),
    },
    {
      title: '数据量',
      dataIndex: 'data_count',
      key: 'data_count',
      width: 100,
    },
    {
      title: '规则匹配',
      dataIndex: 'rule_matched',
      key: 'rule_matched',
      width: 100,
      render: (matched: boolean) => (
        <Tag color={matched ? 'warning' : 'default'}>
          {matched ? '是' : '否'}
        </Tag>
      ),
    },
    {
      title: '邮件已发送',
      dataIndex: 'email_sent',
      key: 'email_sent',
      width: 120,
      render: (sent: boolean) => (
        <Tag color={sent ? 'success' : 'default'}>
          {sent ? '是' : '否'}
        </Tag>
      ),
    },
    {
      title: '错误信息',
      dataIndex: 'error_message',
      key: 'error_message',
      ellipsis: true,
      render: (text: string) => text || '-',
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
    },
  ];

  return (
    <Modal
      title={`告警日志 - ${alarmName || ''}`}
      open={open}
      onCancel={onCancel}
      footer={null}
      width={1200}
      destroyOnClose
    >
      <Form
        form={form}
        layout="inline"
        onFinish={handleSearch}
        style={{ marginBottom: 16 }}
      >
        <Form.Item name="status" label="状态">
          <Select placeholder="选择状态" allowClear style={{ width: 150 }}>
            <Option value="running">执行中</Option>
            <Option value="success">成功</Option>
            <Option value="failed">失败</Option>
            <Option value="triggered">已触发</Option>
          </Select>
        </Form.Item>
        <Form.Item name="dateRange" label="日期范围">
          <RangePicker />
        </Form.Item>
        <Form.Item>
          <Space>
            <Button type="primary" htmlType="submit">
              查询
            </Button>
            <Button onClick={handleReset}>
              重置
            </Button>
          </Space>
        </Form.Item>
      </Form>

      <Table
        columns={columns}
        dataSource={logs}
        loading={loading}
        rowKey="id"
        pagination={{
          current: currentPage,
          pageSize: pageSize,
          total: total,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 条`,
          onChange: (page, size) => {
            setCurrentPage(page);
            setPageSize(size || 10);
          },
        }}
      />
    </Modal>
  );
};

export default AlarmLogModal;

