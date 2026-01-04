import React, { useState, useEffect } from 'react';
import {
  Modal,
  Table,
  Tag,
  Space,
  Button,
  DatePicker,
  Select,
  Card,
  Row,
  Col,
  Statistic,
  Descriptions,
  Divider,
} from 'antd';
import {
  ReloadOutlined,
  DownloadOutlined,
  EyeOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  ExclamationCircleOutlined,
} from '@ant-design/icons';
import ReactMarkdown from 'react-markdown';
import remarkGfm from 'remark-gfm';
import type { AnalysisTaskLogItem, AnalysisTaskLogParams } from '../data.d';
import { queryAnalysisTaskLogs } from '../service';
import moment from 'moment';

const { RangePicker } = DatePicker;
const { Option } = Select;

interface AnalysisTaskLogModalProps {
  open: boolean;
  onCancel: () => void;
  taskId?: number;
  taskName?: string;
}

const AnalysisTaskLogModal: React.FC<AnalysisTaskLogModalProps> = ({
  open,
  onCancel,
  taskId,
  taskName,
}) => {
  const [logs, setLogs] = useState<AnalysisTaskLogItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [total, setTotal] = useState(0);
  const [currentPage, setCurrentPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [filters, setFilters] = useState<AnalysisTaskLogParams>({});

  const fetchLogs = async (params?: AnalysisTaskLogParams) => {
    if (!taskId) return;

    setLoading(true);
    try {
      const response = await queryAnalysisTaskLogs({
        task_id: taskId,
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
    if (open && taskId) {
      fetchLogs();
    }
  }, [open, taskId, currentPage, pageSize]);

  const handleSearch = (values: any) => {
    const newFilters: AnalysisTaskLogParams = {};
    
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
    setFilters({});
    setCurrentPage(1);
    fetchLogs({});
  };

  const getStatusTag = (status: string) => {
    switch (status) {
      case 'success':
        return <Tag color="success" icon={<CheckCircleOutlined />}>成功</Tag>;
      case 'failed':
        return <Tag color="error" icon={<CloseCircleOutlined />}>失败</Tag>;
      case 'running':
        return <Tag color="processing" icon={<ClockCircleOutlined />}>执行中</Tag>;
      default:
        return <Tag color="default" icon={<ExclamationCircleOutlined />}>未知</Tag>;
    }
  };

  const columns = [
    {
      title: '执行时间',
      dataIndex: 'start_time',
      key: 'start_time',
      width: 180,
      render: (text: string) => moment(text).format('YYYY-MM-DD HH:mm:ss'),
    },
    {
      title: '完成时间',
      dataIndex: 'complete_time',
      key: 'complete_time',
      width: 180,
      render: (text: string) => text ? moment(text).format('YYYY-MM-DD HH:mm:ss') : '-',
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
      render: (count: number) => (
        <Tag color="blue">{count} 条</Tag>
      ),
    },
    {
      title: '执行结果',
      dataIndex: 'result',
      key: 'result',
      ellipsis: true,
      render: (text: string) => (
        <div style={{ maxWidth: 300 }}>
          {text}
        </div>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 120,
      render: (_, record: AnalysisTaskLogItem) => (
        <Space>
          <Button
            type="link"
            size="small"
            icon={<EyeOutlined />}
            onClick={() => {
              Modal.info({
                title: '执行详情',
                width: 800,
                content: (
                  <div>
                    <Descriptions column={2} bordered>
                      <Descriptions.Item label="任务名称">{record.task_name}</Descriptions.Item>
                      <Descriptions.Item label="执行状态">{getStatusTag(record.status)}</Descriptions.Item>
                      <Descriptions.Item label="开始时间">{moment(record.start_time).format('YYYY-MM-DD HH:mm:ss')}</Descriptions.Item>
                      <Descriptions.Item label="完成时间">
                        {record.complete_time ? moment(record.complete_time).format('YYYY-MM-DD HH:mm:ss') : '-'}
                      </Descriptions.Item>
                      <Descriptions.Item label="数据量">{record.data_count} 条</Descriptions.Item>
                      <Descriptions.Item label="创建时间">{moment(record.created_at).format('YYYY-MM-DD HH:mm:ss')}</Descriptions.Item>
                    </Descriptions>
                    
                    <Divider />
                    
                    <h4>执行结果</h4>
                    <div style={{ 
                      background: 'linear-gradient(135deg, #f8f9fa 0%, #e9ecef 100%)', 
                      padding: '16px', 
                      borderRadius: '8px',
                      border: '1px solid #dee2e6',
                      maxHeight: '300px',
                      overflow: 'auto',
                      fontFamily: 'Monaco, Menlo, "Ubuntu Mono", monospace',
                      fontSize: '13px',
                      lineHeight: '1.5',
                      color: '#495057'
                    }}>
                      <pre style={{ margin: 0, whiteSpace: 'pre-wrap', wordBreak: 'break-word' }}>
                        {record.result}
                      </pre>
                    </div>
                    
                                         {record.report_content && (
                       <>
                         <Divider />
                         <h4>分析报告</h4>
                         <div 
                           style={{ 
                             background: 'linear-gradient(135deg, #f0f8ff 0%, #e6f3ff 100%)', 
                             padding: '16px', 
                             borderRadius: '8px',
                             border: '1px solid #d1ecf1',
                             maxHeight: '500px',
                             overflow: 'auto',
                             boxShadow: '0 2px 8px rgba(0,0,0,0.1)'
                           }}
                         >
                           <ReactMarkdown 
                             remarkPlugins={[remarkGfm]}
                             components={{
                               h1: ({node, ...props}) => <h1 style={{color: '#1890ff', fontSize: '20px', fontWeight: 600, margin: '16px 0 8px 0'}} {...props} />,
                               h2: ({node, ...props}) => <h2 style={{color: '#1890ff', fontSize: '18px', fontWeight: 600, margin: '14px 0 8px 0'}} {...props} />,
                               h3: ({node, ...props}) => <h3 style={{color: '#1890ff', fontSize: '16px', fontWeight: 600, margin: '12px 0 8px 0'}} {...props} />,
                               h4: ({node, ...props}) => <h4 style={{color: '#52c41a', fontSize: '14px', fontWeight: 500, margin: '10px 0 6px 0'}} {...props} />,
                               h5: ({node, ...props}) => <h5 style={{color: '#52c41a', fontSize: '13px', fontWeight: 500, margin: '8px 0 6px 0'}} {...props} />,
                               h6: ({node, ...props}) => <h6 style={{color: '#52c41a', fontSize: '12px', fontWeight: 500, margin: '6px 0 4px 0'}} {...props} />,
                               p: ({node, ...props}) => <p style={{margin: '8px 0', lineHeight: '1.6', color: '#333'}} {...props} />,
                               ul: ({node, ...props}) => <ul style={{margin: '8px 0', paddingLeft: '20px', lineHeight: '1.6'}} {...props} />,
                               ol: ({node, ...props}) => <ol style={{margin: '8px 0', paddingLeft: '20px', lineHeight: '1.6'}} {...props} />,
                               li: ({node, ...props}) => <li style={{margin: '4px 0', lineHeight: '1.6'}} {...props} />,
                               strong: ({node, ...props}) => <strong style={{fontWeight: 600, color: '#2c3e50'}} {...props} />,
                               em: ({node, ...props}) => <em style={{fontStyle: 'italic', color: '#7f8c8d'}} {...props} />,
                               code: ({node, inline, ...props}) => 
                                 inline ? 
                                   <code style={{background: '#f1f2f6', padding: '2px 4px', borderRadius: '3px', fontSize: '0.9em', color: '#e74c3c'}} {...props} /> :
                                   <code style={{background: '#f8f9fa', padding: '8px 12px', borderRadius: '4px', display: 'block', fontSize: '13px', color: '#495057', border: '1px solid #e9ecef'}} {...props} />,
                               pre: ({node, ...props}) => <pre style={{background: '#f8f9fa', padding: '12px', borderRadius: '4px', overflow: 'auto', fontSize: '13px', lineHeight: '1.5', border: '1px solid #e9ecef'}} {...props} />,
                               blockquote: ({node, ...props}) => <blockquote style={{borderLeft: '4px solid #3498db', paddingLeft: '12px', margin: '12px 0', color: '#7f8c8d', fontStyle: 'italic'}} {...props} />,
                               table: ({node, ...props}) => <table style={{width: '100%', borderCollapse: 'collapse', margin: '12px 0', fontSize: '14px'}} {...props} />,
                               thead: ({node, ...props}) => <thead style={{background: '#f8f9fa'}} {...props} />,
                               tbody: ({node, ...props}) => <tbody {...props} />,
                               tr: ({node, ...props}) => <tr style={{borderBottom: '1px solid #e9ecef'}} {...props} />,
                               th: ({node, ...props}) => <th style={{padding: '8px 12px', textAlign: 'left', fontWeight: 600, color: '#2c3e50', borderBottom: '2px solid #dee2e6'}} {...props} />,
                               td: ({node, ...props}) => <td style={{padding: '8px 12px', color: '#495057'}} {...props} />,
                             }}
                           >
                             {record.report_content}
                           </ReactMarkdown>
                         </div>
                       </>
                     )}
                    
                    {record.error_message && (
                      <>
                        <Divider />
                        <h4>错误信息</h4>
                        <div style={{ 
                          background: '#fff2f0', 
                          padding: '12px', 
                          borderRadius: '4px',
                          border: '1px solid #ffccc7'
                        }}>
                          <pre style={{ whiteSpace: 'pre-wrap', margin: 0, color: '#cf1322' }}>
                            {record.error_message}
                          </pre>
                        </div>
                      </>
                    )}
                  </div>
                ),
              });
            }}
          >
            详情
          </Button>
        </Space>
      ),
    },
  ];

  // 统计信息
  const successCount = logs.filter(log => log.status === 'success').length;
  const failedCount = logs.filter(log => log.status === 'failed').length;
  const runningCount = logs.filter(log => log.status === 'running').length;

  return (
    (<Modal
      title={`执行日志 - ${taskName}`}
      open={open}
      onCancel={onCancel}
      footer={null}
      width={1200}
      destroyOnClose
    >
      <Card size="small" style={{ marginBottom: 16 }}>
        <Row gutter={16}>
          <Col span={6}>
            <Statistic
              title="总执行次数"
              value={total}
              prefix={<ClockCircleOutlined />}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="成功次数"
              value={successCount}
              valueStyle={{ color: '#3f8600' }}
              prefix={<CheckCircleOutlined />}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="失败次数"
              value={failedCount}
              valueStyle={{ color: '#cf1322' }}
              prefix={<CloseCircleOutlined />}
            />
          </Col>
          <Col span={6}>
            <Statistic
              title="执行中"
              value={runningCount}
              valueStyle={{ color: '#1890ff' }}
              prefix={<ClockCircleOutlined />}
            />
          </Col>
        </Row>
      </Card>
      <div style={{ marginBottom: 16 }}>
        <Space>
          <RangePicker
            placeholder={['开始日期', '结束日期']}
            onChange={(dates) => {
              if (dates && dates.length === 2) {
                handleSearch({
                  dateRange: dates,
                  status: filters.status,
                });
              }
            }}
          />
          <Select
            placeholder="选择状态"
            style={{ width: 120 }}
            allowClear
            onChange={(value) => {
              handleSearch({
                status: value,
                dateRange: filters.start_date && filters.end_date ? [
                  moment(filters.start_date),
                  moment(filters.end_date)
                ] : undefined,
              });
            }}
          >
            <Option value="success">成功</Option>
            <Option value="failed">失败</Option>
            <Option value="running">执行中</Option>
          </Select>
          <Button icon={<ReloadOutlined />} onClick={handleReset}>
            重置
          </Button>
          <Button icon={<ReloadOutlined />} onClick={() => fetchLogs()}>
            刷新
          </Button>
        </Space>
      </div>
      <Table
        columns={columns}
        dataSource={logs}
        rowKey="id"
        loading={loading}
        pagination={{
          current: currentPage,
          pageSize: pageSize,
          total: total,
          showSizeChanger: true,
          showQuickJumper: true,
          showTotal: (total, range) => `第 ${range[0]}-${range[1]} 条/共 ${total} 条`,
          onChange: (page, size) => {
            setCurrentPage(page);
            setPageSize(size || 10);
          },
        }}
        scroll={{ x: 1000 }}
      />
    </Modal>)
  );
};

export default AnalysisTaskLogModal; 