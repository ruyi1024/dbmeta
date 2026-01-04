import React, { useState, useEffect } from 'react';
import {
  Modal,
  Form,
  Input,
  Select,
  Button,
  Space,
  Card,
  Divider,
  message,
  Tooltip,
  Switch,
  Row,
  Col,
} from 'antd';
import {
  PlusOutlined,
  DeleteOutlined,
  PlayCircleOutlined,
  QuestionCircleOutlined,
  MailOutlined,
  DatabaseOutlined,
  RobotOutlined,
} from '@ant-design/icons';
import type { AnalysisTaskFormData } from '../data.d';
import { testSqlQuery, testDifyConnection } from '../service';

const { TextArea } = Input;
const { Option } = Select;

interface AnalysisTaskFormProps {
  open: boolean;
  onCancel: () => void;
  onSubmit: (values: AnalysisTaskFormData) => void;
  initialValues?: Partial<AnalysisTaskFormData>;
  loading?: boolean;
}

const AnalysisTaskForm: React.FC<AnalysisTaskFormProps> = ({
  open,
  onCancel,
  onSubmit,
  initialValues,
  loading = false,
}) => {
  const [form] = Form.useForm();
  const [sqlQueries, setSqlQueries] = useState<string[]>(['']);
  const [testingSql, setTestingSql] = useState<number[]>([]);
  const [testingDify, setTestingDify] = useState(false);

  useEffect(() => {
    if (open && initialValues) {
      form.setFieldsValue({
        ...initialValues,
        status: initialValues.status ?? 1,
      });
      if (initialValues.sql_queries && initialValues.sql_queries.length > 0) {
        setSqlQueries([...initialValues.sql_queries]);
      }
    } else if (open) {
      form.resetFields();
      setSqlQueries(['']);
    }
      }, [open, initialValues, form]);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      const formData: AnalysisTaskFormData = {
        ...values,
        sql_queries: sqlQueries.filter(sql => sql.trim() !== ''),
      };
      onSubmit(formData);
    } catch (error) {
      console.error('表单验证失败:', error);
    }
  };

  const addSqlQuery = () => {
    setSqlQueries([...sqlQueries, '']);
  };

  const removeSqlQuery = (index: number) => {
    if (sqlQueries.length > 1) {
      const newQueries = sqlQueries.filter((_, i) => i !== index);
      setSqlQueries(newQueries);
    }
  };

  const updateSqlQuery = (index: number, value: string) => {
    const newQueries = [...sqlQueries];
    newQueries[index] = value;
    setSqlQueries(newQueries);
  };

  const testSql = async (index: number) => {
    const sql = sqlQueries[index];
    if (!sql.trim()) {
      message.warning('请先输入SQL语句');
      return;
    }

    setTestingSql([...testingSql, index]);
    try {
      const response = await testSqlQuery(sql);
      if (response.success) {
        message.success(`SQL测试成功，返回 ${response.data?.length || 0} 条数据`);
      } else {
        message.error(`SQL测试失败: ${response.msg}`);
      }
    } catch (error) {
      message.error('SQL测试失败，请检查SQL语法');
    } finally {
      setTestingSql(testingSql.filter(i => i !== index));
    }
  };

  const testDify = async () => {
    setTestingDify(true);
    try {
      const response = await testDifyConnection();
      if (response.success) {
        message.success('Dify连接测试成功');
      } else {
        message.error(`Dify连接测试失败: ${response.msg}`);
      }
    } catch (error) {
      message.error('Dify连接测试失败');
    } finally {
      setTestingDify(false);
    }
  };

  const cronExamples = [
    { label: '每天凌晨2点', value: '0 2 * * *' },
    { label: '每周一上午9点', value: '0 9 * * 1' },
    { label: '每月1号上午8点', value: '0 8 1 * *' },
    { label: '每小时执行', value: '0 * * * *' },
    { label: '每30分钟执行', value: '*/30 * * * *' },
    { label: '每5分钟执行', value: '*/5 * * * *' },
  ];

  return (
    (<Modal
      title={initialValues ? '编辑分析任务' : '创建分析任务'}
      open={open}
      onCancel={onCancel}
      footer={[
        <Button key="cancel" onClick={onCancel}>
          取消
        </Button>,
        <Button key="submit" type="primary" loading={loading} onClick={handleSubmit}>
          确定
        </Button>,
      ]}
      width={800}
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{ status: 1 }}
      >
        <Row gutter={16}>
          <Col span={12}>
            <Form.Item
              name="task_name"
              label="任务名称"
              rules={[{ required: true, message: '请输入任务名称' }]}
            >
              <Input placeholder="请输入任务名称" />
            </Form.Item>
          </Col>
          <Col span={12}>
            <Form.Item
              name="status"
              label="任务状态"
              valuePropName="checked"
            >
              <Switch
                checkedChildren="启用"
                unCheckedChildren="禁用"
                defaultChecked
              />
            </Form.Item>
          </Col>
        </Row>

        <Form.Item
          name="task_description"
          label="任务描述"
          rules={[{ required: true, message: '请输入任务描述' }]}
        >
          <TextArea
            rows={3}
            placeholder="请描述这个分析任务的用途和目标"
          />
        </Form.Item>

        <Form.Item
          name="report_email"
          label={
            <Space>
              <MailOutlined />
              报告发送邮箱
              <Tooltip title="分析报告将发送到此邮箱">
                <QuestionCircleOutlined />
              </Tooltip>
            </Space>
          }
          rules={[
            { required: true, message: '请输入邮箱地址' },
            { type: 'email', message: '请输入有效的邮箱地址' },
          ]}
        >
          <Input placeholder="example@company.com" />
        </Form.Item>

        <Form.Item
          name="cron_expression"
          label={
            <Space>
              计划任务
              <Tooltip title="Cron表达式格式：分 时 日 月 周">
                <QuestionCircleOutlined />
              </Tooltip>
            </Space>
          }
          rules={[{ required: true, message: '请输入Cron表达式' }]}
        >
          <Select
            placeholder="选择或输入Cron表达式"
            showSearch
            allowClear
            dropdownRender={(menu) => (
              <div>
                {menu}
                <Divider style={{ margin: '8px 0' }} />
                <div style={{ padding: '8px' }}>
                  <div style={{ marginBottom: '8px', fontWeight: 'bold' }}>常用表达式：</div>
                  {cronExamples.map((example) => (
                    <div
                      key={example.value}
                      style={{
                        padding: '4px 8px',
                        cursor: 'pointer',
                        borderRadius: '4px',
                      }}
                      onMouseEnter={(e) => {
                        e.currentTarget.style.backgroundColor = '#f5f5f5';
                      }}
                      onMouseLeave={(e) => {
                        e.currentTarget.style.backgroundColor = 'transparent';
                      }}
                      onClick={() => form.setFieldsValue({ cron_expression: example.value })}
                    >
                      {example.label}: {example.value}
                    </div>
                  ))}
                </div>
              </div>
            )}
          >
            {cronExamples.map((example) => (
              <Option key={example.value} value={example.value}>
                {example.label} ({example.value})
              </Option>
            ))}
          </Select>
        </Form.Item>

        <Card
          title={
            <Space>
              <DatabaseOutlined />
              取数SQL
              <Tooltip title="支持多个SQL查询，数据将合并后发送给AI分析">
                <QuestionCircleOutlined />
              </Tooltip>
            </Space>
          }
          size="small"
          extra={
            <Button
              type="dashed"
              icon={<PlusOutlined />}
              onClick={addSqlQuery}
              size="small"
            >
              添加SQL
            </Button>
          }
        >
          {sqlQueries.map((sql, index) => (
            <div key={index} style={{ marginBottom: '16px' }}>
              <Space.Compact style={{ width: '100%' }}>
                <TextArea
                  rows={4}
                  placeholder={`SQL查询 ${index + 1}...`}
                  value={sql}
                  onChange={(e) => updateSqlQuery(index, e.target.value)}
                  style={{ flex: 1 }}
                />
                <Space direction="vertical">
                  <Button
                    type="primary"
                    icon={<PlayCircleOutlined />}
                    loading={testingSql.includes(index)}
                    onClick={() => testSql(index)}
                    size="small"
                  >
                    测试
                  </Button>
                  {sqlQueries.length > 1 && (
                    <Button
                      type="text"
                      danger
                      icon={<DeleteOutlined />}
                      onClick={() => removeSqlQuery(index)}
                      size="small"
                    />
                  )}
                </Space>
              </Space.Compact>
            </div>
          ))}
        </Card>

        <Form.Item
          name="prompt"
          label={
            <Space>
              <RobotOutlined />
              提示词
              <Tooltip title="指导AI如何分析SQL查询结果并生成报告">
                <QuestionCircleOutlined />
              </Tooltip>
            </Space>
          }
          rules={[{ required: true, message: '请输入提示词' }]}
        >
          <TextArea
            rows={6}
            placeholder={`请分析以下数据并生成一份详细的分析报告：

1. 数据概览：总结数据的基本情况
2. 关键指标：识别重要的业务指标
3. 趋势分析：分析数据变化趋势
4. 异常检测：发现数据异常或问题
5. 建议措施：基于分析结果提出改进建议

请确保报告结构清晰，内容详实，便于决策参考。`}
          />
        </Form.Item>

        <div style={{ textAlign: 'center', marginTop: '16px' }}>
          <Button
            type="dashed"
            icon={<RobotOutlined />}
            loading={testingDify}
            onClick={testDify}
          >
            测试Dify连接
          </Button>
        </div>
      </Form>
    </Modal>)
  );
};

export default AnalysisTaskForm; 