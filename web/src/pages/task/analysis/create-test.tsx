import React, { useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { Card, Button, message, Form, Input, Select, Space, Divider } from 'antd';
import { PlusOutlined, DatabaseOutlined, RobotOutlined } from '@ant-design/icons';

const { TextArea } = Input;
const { Option } = Select;

const CreateTaskTest: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [sqlQueries, setSqlQueries] = useState(['']);

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);

      const formData = {
        ...values,
        sql_queries: sqlQueries.filter(sql => sql.trim() !== ''),
        status: 1,
      };

      console.log('提交的数据:', formData);

      const response = await fetch('/api/v1/task/analysis/create', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(formData),
      });

      const result = await response.json();
      console.log('API响应:', result);

      if (result.success) {
        message.success('创建成功');
        form.resetFields();
        setSqlQueries(['']);
      } else {
        message.error(result.msg || '创建失败');
      }
    } catch (error) {
      console.error('创建任务失败:', error);
      message.error('创建失败，请重试');
    } finally {
      setLoading(false);
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

    try {
      const response = await fetch('/api/v1/task/analysis/test-sql', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ sql }),
      });

      const result = await response.json();
      if (result.success) {
        message.success(`SQL测试成功，返回 ${result.data?.length || 0} 条数据`);
      } else {
        message.error(`SQL测试失败: ${result.msg}`);
      }
    } catch (error) {
      message.error('SQL测试失败，请检查SQL语法');
    }
  };

  const testDify = async () => {
    try {
      const response = await fetch('/api/v1/task/analysis/test-dify', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
      });

      const result = await response.json();
      if (result.success) {
        message.success('Dify连接测试成功');
      } else {
        message.error(`Dify连接测试失败: ${result.msg}`);
      }
    } catch (error) {
      message.error('Dify连接测试失败');
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
    <PageContainer>
      <Card title="创建分析任务测试" style={{ maxWidth: 800, margin: '0 auto' }}>
        <Form
          form={form}
          layout="vertical"
          initialValues={{ status: 1 }}
        >
          <Form.Item
            name="task_name"
            label="任务名称"
            rules={[{ required: true, message: '请输入任务名称' }]}
          >
            <Input placeholder="请输入任务名称" />
          </Form.Item>

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
            label="报告发送邮箱"
            rules={[
              { required: true, message: '请输入邮箱地址' },
              { type: 'email', message: '请输入有效的邮箱地址' },
            ]}
          >
            <Input placeholder="example@company.com" />
          </Form.Item>

          <Form.Item
            name="cron_expression"
            label="计划任务"
            rules={[{ required: true, message: '请输入Cron表达式' }]}
          >
            <Select
              placeholder="选择或输入Cron表达式"
              showSearch
              allowClear
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
                      size="small"
                      onClick={() => testSql(index)}
                    >
                      测试
                    </Button>
                    {sqlQueries.length > 1 && (
                      <Button
                        type="text"
                        danger
                        size="small"
                        onClick={() => removeSqlQuery(index)}
                      >
                        删除
                      </Button>
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

          <Divider />

          <Space>
            <Button
              type="dashed"
              icon={<RobotOutlined />}
              onClick={testDify}
            >
              测试Dify连接
            </Button>
            <Button
              type="primary"
              loading={loading}
              onClick={handleSubmit}
            >
              创建任务
            </Button>
          </Space>
        </Form>
      </Card>
    </PageContainer>
  );
};

export default CreateTaskTest; 