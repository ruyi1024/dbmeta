import React, { useState, useEffect } from 'react';
import { Modal, Form, Input, Select, Button, Space, Card, Divider, message, Row, Col } from 'antd';
import { PlusOutlined, DeleteOutlined, DatabaseOutlined, RobotOutlined } from '@ant-design/icons';
import { getDatasourceTypeList, getDatasourceList, getEnabledAIModels } from '../service';

const { TextArea } = Input;
const { Option } = Select;

interface TaskModalProps {
  open: boolean;
  onCancel: () => void;
  onSuccess: () => void;
  editData?: any; // 编辑时的数据
  mode: 'create' | 'edit'; // 模式：创建或编辑
}

const TaskModal: React.FC<TaskModalProps> = ({
  open,
  onCancel,
  onSuccess,
  editData,
  mode,
}) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [sqlQueries, setSqlQueries] = useState(['']);
  
  // 数据源类型和数据源相关状态
  const [datasourceTypeList, setDatasourceTypeList] = useState<any[]>([]);
  const [datasourceList, setDatasourceList] = useState<any[]>([]);
  const [selectedDatasourceType, setSelectedDatasourceType] = useState<string>('');
  
  // AI模型列表
  const [aiModelList, setAiModelList] = useState<any[]>([]);

  // 当编辑数据变化时，填充表单
  useEffect(() => {
    if (open && editData && mode === 'edit') {
      form.setFieldsValue({
        task_name: editData.task_name,
        task_description: editData.task_description,
        datasource_type: editData.datasource_type,
        datasource_id: editData.datasource_id,
        ai_model_id: editData.ai_model_id,
        report_email: editData.report_email,
        cron_expression: editData.cron_expression,
        prompt: editData.prompt,
      });
      
      // 设置SQL查询
      if (editData.sql_queries && editData.sql_queries.length > 0) {
        setSqlQueries(editData.sql_queries);
      } else {
        setSqlQueries(['']);
      }
      
      // 如果是编辑模式，需要加载对应的数据源列表
      if (editData.datasource_type) {
        setSelectedDatasourceType(editData.datasource_type);
        fetchDatasourceList(editData.datasource_type);
      }
    } else if (open && mode === 'create') {
      // 创建模式时重置表单
      form.resetFields();
      setSqlQueries(['']);
      setSelectedDatasourceType('');
      setDatasourceList([]);
    }
      }, [open, editData, mode, form]);

  // 组件挂载时获取数据源类型列表和AI模型列表
  useEffect(() => {
    if (open) {
      fetchDatasourceTypeList();
      fetchAIModelList();
    }
  }, [open]);

  // 获取数据源类型列表
  const fetchDatasourceTypeList = async () => {
    try {
      const result = await getDatasourceTypeList();
      if (result.success) {
        setDatasourceTypeList(result.data || []);
      } else {
        message.error('获取数据源类型列表失败: ' + result.msg);
      }
    } catch (error) {
      console.error('获取数据源类型列表失败:', error);
      message.error('获取数据源类型列表失败');
    }
  };

  // 获取数据源列表
  const fetchDatasourceList = async (datasourceType: string) => {
    try {
      const result = await getDatasourceList({ type: datasourceType });
      if (result.success) {
        setDatasourceList(result.data || []);
      } else {
        message.error('获取数据源列表失败: ' + result.msg);
      }
    } catch (error) {
      console.error('获取数据源列表失败:', error);
      message.error('获取数据源列表失败');
    }
  };

  // 获取AI模型列表
  const fetchAIModelList = async () => {
    try {
      const result = await getEnabledAIModels();
      if (result.success) {
        setAiModelList(result.data || []);
      } else {
        message.error('获取AI模型列表失败: ' + result.message);
      }
    } catch (error) {
      console.error('获取AI模型列表失败:', error);
      message.error('获取AI模型列表失败');
    }
  };

  // 数据源类型选择变化处理
  const handleDatasourceTypeChange = (value: string) => {
    setSelectedDatasourceType(value);
    setDatasourceList([]);
    // 清空数据源选择
    form.setFieldsValue({ datasource_id: undefined });
    
    if (value) {
      fetchDatasourceList(value);
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);

      const formData = {
        ...values,
        sql_queries: sqlQueries.filter(sql => sql.trim() !== ''),
        status: 1,
      };

      let url = '/api/v1/task/analysis/create';
      let method = 'POST';

      if (mode === 'edit' && editData?.id) {
        url = `/api/v1/task/analysis/update`;
        method = 'PUT';
        formData.id = editData.id;
      }

      const response = await fetch(url, {
        method,
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify(formData),
      });

      const result = await response.json();
      if (result.success) {
        message.success(mode === 'create' ? '创建成功' : '更新成功');
        form.resetFields();
        setSqlQueries(['']);
        onSuccess();
        onCancel();
      } else {
        message.error(result.msg || (mode === 'create' ? '创建失败' : '更新失败'));
      }
    } catch (error) {
      console.error(mode === 'create' ? '创建任务失败:' : '更新任务失败:', error);
      message.error(mode === 'create' ? '创建失败，请重试' : '更新失败，请重试');
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

    const datasourceType = form.getFieldValue('datasource_type');
    const datasourceId = form.getFieldValue('datasource_id');

    if (!datasourceType) {
      message.warning('请先选择数据源类型');
      return;
    }

    if (!datasourceId) {
      message.warning('请先选择数据源');
      return;
    }

    try {
      const response = await fetch('/api/v1/task/analysis/test-sql', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ 
          sql,
          datasource_type: datasourceType,
          datasource_id: datasourceId
        }),
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
    (<Modal
      title={mode === 'create' ? '创建智能任务' : '编辑智能任务'}
      open={open}
      onCancel={onCancel}
      footer={[
        <Button key="cancel" onClick={onCancel}>
          取消
        </Button>,
        <Button key="submit" type="primary" loading={loading} onClick={handleSubmit}>
          {mode === 'create' ? '创建' : '更新'}
        </Button>,
      ]}
      width={1200}
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{ status: 1 }}
      >
        <Row gutter={24}>
          {/* 左侧栏 */}
          <Col span={12}>
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
          </Col>

          {/* 右侧栏 */}
          <Col span={12}>
            {/* 数据源类型选择 */}
            <Form.Item
              name="datasource_type"
              label="数据源类型"
              rules={[{ required: true, message: '请选择数据源类型' }]}
            >
              <Select
                placeholder="请选择数据源类型"
                onChange={handleDatasourceTypeChange}
                showSearch
                allowClear
              >
                {datasourceTypeList.map((item) => (
                  <Option key={item.name} value={item.name}>
                    {item.name}
                  </Option>
                ))}
              </Select>
            </Form.Item>

            {/* 数据源选择 */}
            <Form.Item
              name="datasource_id"
              label="选择数据源"
              rules={[{ required: true, message: '请选择数据源' }]}
            >
              <Select
                placeholder="请选择数据源"
                showSearch
                allowClear
                disabled={!selectedDatasourceType}
              >
                {datasourceList.map((item) => (
                  <Option key={item.id} value={item.id}>
                    {item.name} [{item.host}:{item.port}]
                  </Option>
                ))}
              </Select>
            </Form.Item>

            {/* AI模型选择 */}
            <Form.Item
              name="ai_model_id"
              label={
                <Space>
                  <RobotOutlined />
                  AI模型
                </Space>
              }
              rules={[{ required: true, message: '请选择AI模型' }]}
            >
              <Select
                placeholder="请选择AI模型"
                showSearch
                allowClear
                optionFilterProp="children"
                filterOption={(input, option) =>
                  (option?.children as unknown as string)?.toLowerCase().includes(input.toLowerCase())
                }
              >
                {aiModelList.map((model) => (
                  <Option key={model.id} value={model.id}>
                    {model.name} ({model.provider})
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
                 onClick={testDify}
               >
                 测试Dify连接
               </Button>
             </div>
          </Col>
        </Row>
      </Form>
    </Modal>)
  );
};

export default TaskModal; 