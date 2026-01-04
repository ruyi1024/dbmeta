import React, { useState, useEffect, useRef } from 'react';
import { PageContainer, ProTable } from '@ant-design/pro-components';
import type { ActionType, ProColumns } from '@ant-design/pro-components';
import { Button, message, Modal, Form, Input, InputNumber, Switch, Select, Space, Card, Collapse } from 'antd';
import { PlusOutlined, EditOutlined, DeleteOutlined, MinusCircleOutlined, PlusCircleOutlined } from '@ant-design/icons';
import { getRules, createRule, updateRule, deleteRule, SemanticSqlRule, QuestionFlowItem } from '../services/chatQuery';

const { TextArea } = Input;
const { Panel } = Collapse;

const RulesPage: React.FC = () => {
  const [form] = Form.useForm();
  const actionRef = useRef<ActionType>();
  const [modalVisible, setModalVisible] = useState(false);
  const [editingRule, setEditingRule] = useState<SemanticSqlRule | null>(null);

  const columns: ProColumns<SemanticSqlRule>[] = [
    {
      title: 'ID',
      dataIndex: 'id',
      width: 80,
      search: false,
    },
    {
      title: '规则名称',
      dataIndex: 'rule_name',
      width: 200,
    },
    {
      title: '语义模式',
      dataIndex: 'semantic_pattern',
      width: 250,
      ellipsis: true,
      render: (text) => <span title={text as string}>{text}</span>,
    },
    {
      title: '查询类型',
      dataIndex: 'query_type',
      width: 120,
      valueEnum: {
        status: { text: '状态' },
        performance: { text: '性能' },
        metadata: { text: '元数据' },
        custom: { text: '自定义' },
      },
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      width: 100,
      search: false,
      sorter: true,
    },
    {
      title: '启用',
      dataIndex: 'enabled',
      width: 80,
      valueEnum: {
        0: { text: '禁用', status: 'Default' },
        1: { text: '启用', status: 'Success' },
      },
      render: (_, record) => (record.enabled === 1 ? '是' : '否'),
    },
    {
      title: '使用本地MySQL',
      dataIndex: 'use_local_db',
      width: 120,
      valueEnum: {
        0: { text: '远程数据源', status: 'Default' },
        1: { text: '本地MySQL', status: 'Success' },
      },
      render: (_, record) => (record.use_local_db === 1 ? '是' : '否'),
    },
    {
      title: '描述',
      dataIndex: 'description',
      ellipsis: true,
      search: false,
      render: (text) => <span title={text as string}>{text || '-'}</span>,
    },
    {
      title: '操作',
      valueType: 'option',
      width: 150,
      render: (_, record) => [
        <Button
          key="edit"
          type="link"
          icon={<EditOutlined />}
          onClick={() => handleEdit(record)}
        >
          编辑
        </Button>,
        <Button
          key="delete"
          type="link"
          danger
          icon={<DeleteOutlined />}
          onClick={() => handleDelete(record.id)}
        >
          删除
        </Button>,
      ],
    },
  ];

  const handleAdd = () => {
    setEditingRule(null);
    form.resetFields();
    setModalVisible(true);
  };

  const handleEdit = (record: SemanticSqlRule) => {
    setEditingRule(record);
    // 处理question_flow和parameter_mapping（如果是字符串，需要解析）
    const formValues: any = { ...record };
    if (record.question_flow && typeof record.question_flow === 'string') {
      try {
        formValues.question_flow = JSON.parse(record.question_flow);
      } catch (e) {
        formValues.question_flow = [];
      }
    }
    if (record.parameter_mapping && typeof record.parameter_mapping === 'string') {
      try {
        formValues.parameter_mapping = JSON.stringify(JSON.parse(record.parameter_mapping), null, 2);
      } catch (e) {
        formValues.parameter_mapping = '';
      }
    } else if (record.parameter_mapping) {
      formValues.parameter_mapping = JSON.stringify(record.parameter_mapping, null, 2);
    }
    form.setFieldsValue(formValues);
    setModalVisible(true);
  };

  const handleDelete = async (id: number) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除此规则吗？',
      onOk: async () => {
        try {
          const response = await deleteRule(id);
          if (response.success) {
            message.success('删除成功');
            actionRef.current?.reload();
          }
        } catch (error) {
          message.error('删除失败');
        }
      },
    });
  };

  const handleSubmit = async () => {
    try {
      const values = await form.validateFields();
      
      // 处理question_flow：将选项字符串转换为数组
      if (values.question_flow && Array.isArray(values.question_flow)) {
        values.question_flow = values.question_flow.map((item: any) => {
          if (item.options && typeof item.options === 'string') {
            item.options = item.options.split('\n').filter((v: string) => v.trim());
          }
          return item;
        });
      }
      
      // 处理parameter_mapping：如果是字符串，尝试解析为JSON
      if (values.parameter_mapping && typeof values.parameter_mapping === 'string') {
        try {
          values.parameter_mapping = JSON.parse(values.parameter_mapping);
        } catch (e) {
          message.error('参数映射配置格式错误，应为JSON格式');
          return;
        }
      }
      
      if (editingRule) {
        const response = await updateRule(editingRule.id, values);
        if (response.success) {
          message.success('更新成功');
          setModalVisible(false);
          actionRef.current?.reload();
        }
      } else {
        const response = await createRule(values);
        if (response.success) {
          message.success('创建成功');
          setModalVisible(false);
          actionRef.current?.reload();
        }
      }
    } catch (error) {
      console.error('提交失败:', error);
      message.error('提交失败，请检查表单数据');
    }
  };

  return (
    <PageContainer title="语义-SQL规则管理">
      <ProTable<SemanticSqlRule>
        headerTitle="规则列表"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 'auto',
        }}
        toolBarRender={() => [
          <Button
            type="primary"
            key="primary"
            icon={<PlusOutlined />}
            onClick={handleAdd}
          >
            新建规则
          </Button>,
        ]}
        request={async (params) => {
          try {
            const response = await getRules();
            if (response.success) {
              let data = response.data || [];
              
              // 前端过滤
              if (params.rule_name) {
                data = data.filter(item => 
                  item.rule_name?.includes(params.rule_name as string)
                );
              }
              if (params.query_type) {
                data = data.filter(item => item.query_type === params.query_type);
              }
              if (params.enabled !== undefined) {
                data = data.filter(item => item.enabled === params.enabled);
              }

              return {
                data,
                success: true,
                total: data.length,
              };
            }
            return { data: [], success: false, total: 0 };
          } catch (error) {
            return { data: [], success: false, total: 0 };
          }
        }}
        columns={columns}
      />

      <Modal
        title={editingRule ? '编辑规则' : '新建规则'}
        open={modalVisible}
        onOk={handleSubmit}
        onCancel={() => setModalVisible(false)}
        width={800}
      >
        <Form
          form={form}
          layout="vertical"
          initialValues={{
            enabled: 1,
            use_local_db: 0,
            multi_round_enabled: 0,
            priority: 0,
            query_type: 'custom',
            question_flow: [],
          }}
        >
          <Form.Item
            name="rule_name"
            label="规则名称"
            rules={[{ required: true, message: '请输入规则名称' }]}
          >
            <Input placeholder="例如：查询数据库状态" />
          </Form.Item>

          <Form.Item
            name="semantic_pattern"
            label="语义模式（正则表达式）"
            rules={[{ required: true, message: '请输入语义模式' }]}
            extra="支持正则表达式，例如：查询.*数据库.*状态"
          >
            <TextArea
              rows={3}
              placeholder="例如：查询.*数据库.*状态"
            />
          </Form.Item>

          <Form.Item
            name="sql_template"
            label="SQL模板"
            rules={[{ required: true, message: '请输入SQL模板' }]}
            extra="支持参数占位符 {host}, {port}, {database}, {table} 等"
          >
            <TextArea
              rows={5}
              placeholder="例如：SELECT status, status_text FROM datasource WHERE host = '{host}' AND port = '{port}'"
            />
          </Form.Item>

          <Form.Item
            name="query_type"
            label="查询类型"
            rules={[{ required: true, message: '请选择查询类型' }]}
          >
            <Select>
              <Select.Option value="status">状态</Select.Option>
              <Select.Option value="performance">性能</Select.Option>
              <Select.Option value="metadata">元数据</Select.Option>
              <Select.Option value="custom">自定义</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            name="description"
            label="描述"
          >
            <TextArea rows={2} placeholder="规则描述" />
          </Form.Item>

          <Space>
            <Form.Item
              name="enabled"
              label="启用"
              valuePropName="checked"
              getValueFromEvent={(checked) => (checked ? 1 : 0)}
              getValueProps={(value) => ({ checked: value === 1 })}
            >
              <Switch />
            </Form.Item>

            <Form.Item
              name="use_local_db"
              label="使用本地MySQL"
              valuePropName="checked"
              getValueFromEvent={(checked) => (checked ? 1 : 0)}
              getValueProps={(value) => ({ checked: value === 1 })}
              extra="开启后，规则生成的SQL将在本项目配置的MySQL中执行"
            >
              <Switch />
            </Form.Item>

            <Form.Item
              name="priority"
              label="优先级"
              extra="数字越大优先级越高"
            >
              <InputNumber min={0} max={100} />
            </Form.Item>
          </Space>

          <Form.Item
            name="multi_round_enabled"
            label="启用多轮对话"
            valuePropName="checked"
            getValueFromEvent={(checked: boolean) => (checked ? 1 : 0)}
            getValueProps={(value: number) => ({ checked: value === 1 })}
            extra="开启后，系统会按顺序收集必要信息后再生成SQL"
          >
            <Switch />
          </Form.Item>

          <Form.Item
            noStyle
            shouldUpdate={(prevValues, currentValues) => prevValues.multi_round_enabled !== currentValues.multi_round_enabled}
          >
            {({ getFieldValue }) => {
              const multiRoundEnabled = getFieldValue('multi_round_enabled') === 1;
              return multiRoundEnabled ? (
                <>
                  <Card size="small" title="问题流程配置" style={{ marginBottom: 16 }}>
                    <Form.List name="question_flow">
                      {(fields, { add, remove }) => (
                        <>
                          {fields.map(({ key, name, ...restField }) => (
                            <Card key={key} size="small" style={{ marginBottom: 8 }}>
                              <Space direction="vertical" style={{ width: '100%' }}>
                                <Space>
                                  <Form.Item
                                    {...restField}
                                    name={[name, 'key']}
                                    label="参数键名"
                                    rules={[{ required: true, message: '请输入参数键名' }]}
                                    style={{ width: 200 }}
                                  >
                                    <Input placeholder="如：user_id" />
                                  </Form.Item>
                                  <Form.Item
                                    {...restField}
                                    name={[name, 'type']}
                                    label="输入类型"
                                    initialValue="text"
                                    style={{ width: 150 }}
                                  >
                                    <Select>
                                      <Select.Option value="text">文本</Select.Option>
                                      <Select.Option value="select">选择</Select.Option>
                                      <Select.Option value="number">数字</Select.Option>
                                      <Select.Option value="email">邮箱</Select.Option>
                                    </Select>
                                  </Form.Item>
                                  <Form.Item
                                    {...restField}
                                    name={[name, 'required']}
                                    label="必填"
                                    valuePropName="checked"
                                    initialValue={true}
                                  >
                                    <Switch />
                                  </Form.Item>
                                  <Button
                                    type="link"
                                    danger
                                    icon={<MinusCircleOutlined />}
                                    onClick={() => remove(name)}
                                  >
                                    删除
                                  </Button>
                                </Space>
                                <Form.Item
                                  {...restField}
                                  name={[name, 'question']}
                                  label="提示问题"
                                  rules={[{ required: true, message: '请输入提示问题' }]}
                                >
                                  <Input placeholder="如：请输入查询用户的ID" />
                                </Form.Item>
                                <Form.Item
                                  noStyle
                                  shouldUpdate={(prevValues, currentValues) => {
                                    const prevType = prevValues.question_flow?.[name]?.type;
                                    const currentType = currentValues.question_flow?.[name]?.type;
                                    return prevType !== currentType;
                                  }}
                                >
                                  {({ getFieldValue }) => {
                                    const itemType = getFieldValue(['question_flow', name, 'type']);
                                    const optionsSQL = getFieldValue(['question_flow', name, 'options_sql']);
                                    return itemType === 'select' ? (
                                      <>
                                        <Form.Item
                                          {...restField}
                                          name={[name, 'options_sql']}
                                          label="选项SQL（可选）"
                                          extra="通过SQL动态获取选项列表，SQL应返回单列数据。如果设置了SQL，将优先使用SQL获取选项。"
                                        >
                                          <TextArea
                                            rows={4}
                                            placeholder="SELECT DISTINCT event_type FROM alarm_rule WHERE status = 'active'"
                                          />
                                        </Form.Item>
                                        <Form.Item
                                          {...restField}
                                          name={[name, 'options']}
                                          label="选项列表（静态）"
                                          rules={[
                                            {
                                              validator: (_, value) => {
                                                const sql = getFieldValue(['question_flow', name, 'options_sql']);
                                                if (!sql && (!value || value.length === 0)) {
                                                  return Promise.reject(new Error('请填写选项SQL或选项列表'));
                                                }
                                                return Promise.resolve();
                                              }
                                            }
                                          ]}
                                          extra={optionsSQL ? "已设置选项SQL，静态选项将作为备用" : "每行一个选项，当未设置选项SQL时使用"}
                                        >
                                          <TextArea
                                            rows={3}
                                            placeholder="MySQL&#10;PostgreSQL&#10;Oracle"
                                            onChange={(e) => {
                                              const options = e.target.value.split('\n').filter(v => v.trim());
                                              form.setFieldValue(['question_flow', name, 'options'], options);
                                            }}
                                          />
                                        </Form.Item>
                                      </>
                                    ) : null;
                                  }}
                                </Form.Item>
                                <Form.Item
                                  {...restField}
                                  name={[name, 'description']}
                                  label="参数描述（可选）"
                                >
                                  <Input placeholder="参数说明" />
                                </Form.Item>
                              </Space>
                            </Card>
                          ))}
                          <Button
                            type="dashed"
                            onClick={() => add()}
                            block
                            icon={<PlusCircleOutlined />}
                          >
                            添加问题
                          </Button>
                        </>
                      )}
                    </Form.List>
                  </Card>

                  <Card size="small" title="参数映射配置（可选）" style={{ marginBottom: 16 }}>
                    <Form.Item
                      name="parameter_mapping"
                      extra="将收集的参数映射到SQL模板中的占位符，例如：{'user_id': 'id', 'db_type': 'type'}"
                    >
                      <TextArea
                        rows={4}
                        placeholder='{"user_id": "id", "db_type": "type"}'
                      />
                    </Form.Item>
                  </Card>
                </>
              ) : null;
            }}
          </Form.Item>
        </Form>
      </Modal>
    </PageContainer>
  );
};

export default RulesPage;

