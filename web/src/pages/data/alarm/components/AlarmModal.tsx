import React, { useState, useEffect } from 'react';
import { Modal, Form, Input, Select, Button, Space, Card, Divider, message, Row, Col, InputNumber } from 'antd';
import { DatabaseOutlined, MailOutlined } from '@ant-design/icons';
import { getDatasourceTypeList, getDatasourceList, getDatabaseList, testSqlQuery } from '../service';

const { TextArea } = Input;
const { Option } = Select;

interface AlarmModalProps {
  open: boolean;
  onCancel: () => void;
  onSuccess: () => void;
  editData?: any;
  mode: 'create' | 'edit';
}

const AlarmModal: React.FC<AlarmModalProps> = ({
  open,
  onCancel,
  onSuccess,
  editData,
  mode,
}) => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [testingSql, setTestingSql] = useState(false);
  
  // 数据源类型和数据源相关状态
  const [datasourceTypeList, setDatasourceTypeList] = useState<any[]>([]);
  const [datasourceList, setDatasourceList] = useState<any[]>([]);
  const [selectedDatasourceType, setSelectedDatasourceType] = useState<string>('');
  const [databaseList, setDatabaseList] = useState<string[]>([]);
  const [selectedDatasourceId, setSelectedDatasourceId] = useState<number | undefined>(undefined);

  // 当编辑数据变化时，填充表单
  useEffect(() => {
    if (open && editData && mode === 'edit') {
      form.setFieldsValue({
        alarm_name: editData.alarm_name,
        alarm_description: editData.alarm_description,
        datasource_type: editData.datasource_type,
        datasource_id: editData.datasource_id,
        database_name: editData.database_name,
        sql_query: editData.sql_query,
        rule_operator: editData.rule_operator,
        rule_value: editData.rule_value,
        email_content: editData.email_content,
        email_to: editData.email_to,
        cron_expression: editData.cron_expression,
        status: editData.status,
      });
      
      if (editData.datasource_type) {
        setSelectedDatasourceType(editData.datasource_type);
        fetchDatasourceList(editData.datasource_type);
      }
      if (editData.datasource_id) {
        setSelectedDatasourceId(editData.datasource_id);
        fetchDatabaseList(editData.datasource_id);
      }
    } else if (open && mode === 'create') {
      form.resetFields();
      setSelectedDatasourceType('');
      setDatasourceList([]);
      setDatabaseList([]);
      setSelectedDatasourceId(undefined);
    }
  }, [open, editData, mode, form]);

  // 组件挂载时获取数据源类型列表
  useEffect(() => {
    if (open) {
      fetchDatasourceTypeList();
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

  // 数据源类型选择变化处理
  const handleDatasourceTypeChange = (value: string) => {
    setSelectedDatasourceType(value);
    setDatasourceList([]);
    setDatabaseList([]);
    setSelectedDatasourceId(undefined);
    form.setFieldsValue({ datasource_id: undefined, database_name: undefined });
    
    if (value) {
      fetchDatasourceList(value);
    }
  };

  // 数据源选择变化处理
  const handleDatasourceChange = (value: number | undefined) => {
    console.log('数据源选择变化:', value);
    if (value) {
      setSelectedDatasourceId(value);
      setDatabaseList([]);
      form.setFieldsValue({ database_name: undefined });
      fetchDatabaseList(value);
    } else {
      setSelectedDatasourceId(undefined);
      setDatabaseList([]);
      form.setFieldsValue({ database_name: undefined });
    }
  };

  // 获取数据库列表
  const fetchDatabaseList = async (datasourceId: number) => {
    try {
      console.log('开始获取数据库列表, datasourceId:', datasourceId);
      const result = await getDatabaseList(datasourceId);
      console.log('获取数据库列表结果:', result);
      if (result.success) {
        setDatabaseList(result.data || []);
        console.log('数据库列表已设置:', result.data);
      } else {
        message.error('获取数据库列表失败: ' + result.msg);
      }
    } catch (error) {
      console.error('获取数据库列表失败:', error);
      message.error('获取数据库列表失败');
    }
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      setLoading(true);

      const formData = {
        ...values,
        status: values.status ?? 1,
      };

      let url = '/api/v1/data/alarm/create';
      let method = 'POST';

      if (mode === 'edit' && editData?.id) {
        url = `/api/v1/data/alarm/update`;
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
        onSuccess();
        onCancel();
      } else {
        message.error(result.msg || (mode === 'create' ? '创建失败' : '更新失败'));
      }
    } catch (error) {
      console.error(mode === 'create' ? '创建告警失败:' : '更新告警失败:', error);
      message.error(mode === 'create' ? '创建失败，请重试' : '更新失败，请重试');
    } finally {
      setLoading(false);
    }
  };

  const testSql = async () => {
    const sql = form.getFieldValue('sql_query');
    if (!sql || !sql.trim()) {
      message.warning('请先输入SQL语句');
      return;
    }

    const datasourceType = form.getFieldValue('datasource_type');
    const datasourceId = form.getFieldValue('datasource_id');
    const databaseName = form.getFieldValue('database_name');

    if (!datasourceType) {
      message.warning('请先选择数据源类型');
      return;
    }

    if (!datasourceId) {
      message.warning('请先选择数据源');
      return;
    }

    setTestingSql(true);
    try {
      const result = await testSqlQuery(sql, datasourceType, datasourceId, databaseName);
      if (result.success) {
        message.success(`SQL测试成功，返回 ${result.data?.length || 0} 条数据`);
      } else {
        message.error(`SQL测试失败: ${result.msg}`);
      }
    } catch (error) {
      message.error('SQL测试失败，请检查SQL语法');
    } finally {
      setTestingSql(false);
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

  const ruleOperators = [
    { label: '大于', value: '>' },
    { label: '小于', value: '<' },
    { label: '等于', value: '=' },
    { label: '大于等于', value: '>=' },
    { label: '小于等于', value: '<=' },
    { label: '不等于', value: '!=' },
  ];

  return (
    <Modal
      title={mode === 'create' ? '创建数据告警' : '编辑数据告警'}
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
      width={1000}
      destroyOnClose
    >
      <Form
        form={form}
        layout="vertical"
        initialValues={{ status: 1, rule_operator: '>' }}
      >
        <Row gutter={24}>
          {/* 左侧栏 */}
          <Col span={12}>
            <Form.Item
              name="alarm_name"
              label="告警名称"
              rules={[{ required: true, message: '请输入告警名称' }]}
            >
              <Input placeholder="请输入告警名称" />
            </Form.Item>

            <Form.Item
              name="alarm_description"
              label="告警描述"
              rules={[{ required: true, message: '请输入告警描述' }]}
            >
              <TextArea
                rows={3}
                placeholder="请描述这个告警的用途和目标"
              />
            </Form.Item>

            <Form.Item
              name="email_to"
              label={
                <Space>
                  <MailOutlined />
                  接收邮箱
                </Space>
              }
              rules={[
                { required: true, message: '请输入邮箱地址' },
                {
                  validator: (_: any, value: string) => {
                    if (!value) {
                      return Promise.resolve();
                    }
                    // 以英文分号分隔多个邮箱
                    const emails = value.split(';').map((email: string) => email.trim()).filter((email: string) => email);
                    if (emails.length === 0) {
                      return Promise.reject(new Error('请输入至少一个邮箱地址'));
                    }
                    // 验证每个邮箱格式
                    const emailRegex = /^[^\s@]+@[^\s@]+\.[^\s@]+$/;
                    for (const email of emails) {
                      if (!emailRegex.test(email)) {
                        return Promise.reject(new Error(`邮箱格式不正确: ${email}`));
                      }
                    }
                    return Promise.resolve();
                  },
                },
              ]}
            >
              <Input placeholder="example@company.com;example2@company.com（多个邮箱用英文分号分隔）" />
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
                onChange={handleDatasourceChange}
              >
                {datasourceList.map((item) => (
                  <Option key={item.id} value={item.id}>
                    {item.name} [{item.host}:{item.port}]
                  </Option>
                ))}
              </Select>
            </Form.Item>

            {/* 数据库选择 - 只在选择了数据源后显示 */}
            {selectedDatasourceId && (
              <Form.Item
                name="database_name"
                label="选择数据库"
                tooltip="可选，如果SQL中已包含数据库名（如 database.table），可以不选"
              >
                <Select
                  placeholder="请选择数据库（可选）"
                  showSearch
                  allowClear
                  loading={databaseList.length === 0 && selectedDatasourceId !== undefined}
                >
                  {databaseList.map((dbName) => (
                    <Option key={dbName} value={dbName}>
                      {dbName}
                    </Option>
                  ))}
                </Select>
              </Form.Item>
            )}

            <Card
              title={
                <Space>
                  <DatabaseOutlined />
                  SQL查询
                </Space>
              }
              size="small"
            >
              <Form.Item
                name="sql_query"
                rules={[{ required: true, message: '请输入SQL查询语句' }]}
              >
                <TextArea
                  rows={6}
                  placeholder="请输入SQL查询语句（只支持一条SQL，查询结果的数据量将用于规则判断）"
                />
              </Form.Item>
              <Button
                type="primary"
                size="small"
                loading={testingSql}
                onClick={testSql}
                style={{ width: '100%' }}
              >
                测试SQL
              </Button>
            </Card>

            <Form.Item
              label="告警规则"
              required
            >
              <Space>
                <Form.Item
                  name="rule_operator"
                  noStyle
                  rules={[{ required: true, message: '请选择规则操作符' }]}
                >
                  <Select style={{ width: 120 }}>
                    {ruleOperators.map((op) => (
                      <Option key={op.value} value={op.value}>
                        {op.label}
                      </Option>
                    ))}
                  </Select>
                </Form.Item>
                <Form.Item
                  name="rule_value"
                  noStyle
                  rules={[{ required: true, message: '请输入规则值' }]}
                >
                  <InputNumber
                    min={0}
                    placeholder="数据量"
                    style={{ width: 150 }}
                  />
                </Form.Item>
                <span>时触发告警</span>
              </Space>
            </Form.Item>

            <Form.Item
              name="email_content"
              label="邮件内容描述"
              tooltip="自定义邮件内容描述，将显示在邮件表格上方"
            >
              <TextArea
                rows={4}
                placeholder="请输入自定义邮件内容描述（可选，将显示在邮件表格上方）"
              />
            </Form.Item>
          </Col>
        </Row>
      </Form>
    </Modal>
  );
};

export default AlarmModal;

