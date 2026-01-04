import { message, Popconfirm, Button, Switch } from 'antd';
import React, { useState, useRef } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';
import { ProFormText, ProFormSelect, ProFormTextArea, ProFormDigit, ProFormSwitch } from '@ant-design/pro-components';
import CreateForm from './forms/CreateForm';
import UpdateForm from './forms/UpdateForm';
import {
  AIModel,
  getModels,
  createModel,
  updateModel,
  deleteModel,
  testModel,
  toggleModel,
} from './api';

/**
 * 添加模型
 */
const handleAdd = async (fields: Partial<AIModel>) => {
  const hide = message.loading('正在添加');
  try {
    await createModel(fields);
    hide();
    message.success('添加成功');
    return true;
  } catch (error) {
    hide();
    message.error('添加失败请重试！');
    return false;
  }
};

/**
 * 更新模型
 */
const handleUpdate = async (fields: Partial<AIModel>, id: number) => {
  const hide = message.loading('正在配置');
  try {
    await updateModel(id, fields);
    hide();
    message.success('修改成功');
    return true;
  } catch (error) {
    hide();
    message.error('修改失败请重试！');
    return false;
  }
};

/**
 * 删除模型
 */
const handleRemove = async (selectedRows: AIModel[]) => {
  const hide = message.loading('正在删除');
  if (!selectedRows || selectedRows.length === 0) return true;
  try {
    for (const row of selectedRows) {
      await deleteModel(row.id);
    }
    hide();
    message.success('删除成功，即将刷新');
    return true;
  } catch (error) {
    hide();
    message.error('删除失败，请重试');
    return false;
  }
};

/**
 * 测试模型连接
 */
const handleTest = async (id: number) => {
  const hide = message.loading('正在测试连接');
  try {
    const result = await testModel(id);
    hide();
    if (result.success) {
      message.success('连接测试成功');
    } else {
      message.error(result.message || '连接测试失败');
    }
    return result.success;
  } catch (error) {
    hide();
    message.error('测试失败，请重试');
    return false;
  }
};

const AIModelList: React.FC = () => {
  const [createModalVisible, handleModalVisible] = useState<boolean>(false);
  const [updateModalVisible, handleUpdateModalVisible] = useState<boolean>(false);
  const [stepFormValues, setStepFormValues] = useState<Partial<AIModel>>({});
  const actionRef = useRef<ActionType>();

  const columns: ProColumns<AIModel>[] = [
    {
      title: 'ID',
      dataIndex: 'id',
      tip: '模型ID是唯一的 key',
      width: 60,
      hideInSearch: true,
    },
    {
      title: '模型名称',
      dataIndex: 'name',
      width: 150,
      ellipsis: true,
    },
    {
      title: '提供商',
      dataIndex: 'provider',
      width: 120,
      valueEnum: {
        ollama: { text: 'Ollama', status: 'Default' },
        lm_studio: { text: 'LM Studio', status: 'Processing' },
        vllm: { text: 'vLLM', status: 'Success' },
        dify_local: { text: 'Dify本地', status: 'Warning' },
        openai: { text: 'OpenAI', status: 'Success' },
        deepseek: { text: 'DeepSeek', status: 'Success' },
        qwen: { text: 'Qwen', status: 'Success' },
      },
    },
    {
      title: 'API地址',
      dataIndex: 'api_url',
      width: 200,
      ellipsis: true,
      hideInSearch: true,
    },
    {
      title: '模型标识',
      dataIndex: 'model_name',
      width: 150,
      ellipsis: true,
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      width: 80,
      sorter: true,
      hideInSearch: true,
    },
    {
      title: '启用',
      dataIndex: 'enabled',
      width: 80,
      valueEnum: {
        0: { text: '禁用', status: 'Default' },
        1: { text: '启用', status: 'Success' },
      },
      render: (_, record) => (
        <Switch
          checked={record.enabled === 1}
          onChange={async (checked: boolean) => {
            const enabled = checked ? 1 : 0;
            try {
              await toggleModel(record.id, enabled);
              message.success(checked ? '已启用' : '已禁用');
              actionRef.current?.reload();
            } catch (error) {
              message.error('操作失败');
            }
          }}
        />
      ),
    },
    {
      title: '超时(秒)',
      dataIndex: 'timeout',
      width: 80,
      hideInSearch: true,
    },
    {
      title: '最大Token',
      dataIndex: 'max_tokens',
      width: 100,
      hideInSearch: true,
    },
    {
      title: '温度',
      dataIndex: 'temperature',
      width: 80,
      hideInSearch: true,
    },
    {
      title: '流式响应',
      dataIndex: 'stream_enabled',
      width: 100,
      valueEnum: {
        0: { text: '否', status: 'Default' },
        1: { text: '是', status: 'Success' },
      },
      hideInSearch: true,
    },
    {
      title: '描述',
      dataIndex: 'description',
      ellipsis: true,
      hideInSearch: true,
    },
    {
      title: '操作',
      dataIndex: 'option',
      valueType: 'option',
      width: 200,
      render: (_, record) => [
        <a
          key="test"
          onClick={async () => {
            await handleTest(record.id);
          }}
        >
          测试
        </a>,
        <a
          key="config"
          onClick={() => {
            handleUpdateModalVisible(true);
            setStepFormValues(record);
          }}
        >
          修改
        </a>,
        <Popconfirm
          key="delete"
          title="确定删除此模型吗？"
          onConfirm={async () => {
            await handleRemove([record]);
            actionRef.current?.reloadAndRest?.();
          }}
          okText="是"
          cancelText="否"
        >
          <a>删除</a>
        </Popconfirm>,
      ],
    },
  ];

  return (
    <PageContainer>
      <ProTable<AIModel>
        headerTitle="AI模型列表"
        actionRef={actionRef}
        rowKey="id"
        search={{
          labelWidth: 120,
        }}
        toolBarRender={() => [
          <Button
            key="add"
            type="primary"
            onClick={() => handleModalVisible(true)}
          >
            新建模型
          </Button>,
        ]}
        request={async (params, sorter, filter) => {
          const result = await getModels();
          return {
            data: result.data || [],
            success: result.success,
            total: result.data?.length || 0,
          };
        }}
        columns={columns}
        rowSelection={{}}
      />
      <CreateForm
        onCancel={() => handleModalVisible(false)}
        modalVisible={createModalVisible}
        onSubmit={handleAdd}
      >
        <ProFormText
          name="name"
          label="模型名称"
          rules={[{ required: true, message: '请输入模型名称!' }]}
          placeholder="如：GPT-4, Qwen-7B"
        />
        <ProFormSelect
          name="provider"
          label="提供商"
          options={[
            { label: 'Ollama', value: 'ollama' },
            { label: 'LM Studio', value: 'lm_studio' },
            { label: 'vLLM', value: 'vllm' },
            { label: 'Dify本地', value: 'dify_local' },
            { label: 'OpenAI', value: 'openai' },
            { label: 'DeepSeek', value: 'deepseek' },
            { label: 'Qwen', value: 'qwen' },
          ]}
          rules={[{ required: true, message: '请选择提供商!' }]}
        />
        <ProFormText
          name="api_url"
          label="API地址"
          rules={[{ required: true, message: '请输入API地址!' }]}
          placeholder="如：https://api.openai.com/v1/chat/completions"
        />
        <ProFormText
          name="api_key"
          label="API密钥"
          placeholder="请输入API密钥（可选，部分提供商不需要）"
        />
        <ProFormText
          name="model_name"
          label="模型标识"
          rules={[{ required: true, message: '请输入模型标识!' }]}
          placeholder="如：gpt-4, qwen-7b-chat"
        />
        <ProFormDigit
          name="priority"
          label="优先级"
          initialValue={0}
          min={0}
          max={100}
          tooltip="数字越大优先级越高"
        />
        <ProFormSwitch
          name="enabled"
          label="启用"
          valuePropName="checked"
          getValueFromEvent={(checked: boolean) => checked ? 1 : 0}
        />
        <ProFormDigit
          name="timeout"
          label="超时时间(秒)"
          initialValue={30}
          min={1}
          max={300}
        />
        <ProFormDigit
          name="max_tokens"
          label="最大Token数"
          initialValue={2000}
          min={1}
          max={100000}
        />
        <ProFormDigit
          name="temperature"
          label="温度参数"
          initialValue={0.7}
          min={0}
          max={2}
          step={0.1}
        />
        <ProFormSwitch
          name="stream_enabled"
          label="支持流式响应"
          valuePropName="checked"
          getValueFromEvent={(checked: boolean) => checked ? 1 : 0}
        />
        <ProFormTextArea
          name="description"
          label="描述"
          placeholder="请输入模型描述"
        />
      </CreateForm>
      {stepFormValues && Object.keys(stepFormValues).length ? (
        <UpdateForm
          onCancel={() => {
            handleUpdateModalVisible(false);
            setStepFormValues({});
          }}
          modalVisible={updateModalVisible}
          values={stepFormValues}
          onSubmit={handleUpdate}
        >
          <ProFormText
            name="id"
            label="ID"
            disabled
          />
          <ProFormText
            name="name"
            label="模型名称"
            rules={[{ required: true, message: '请输入模型名称!' }]}
          />
          <ProFormSelect
            name="provider"
            label="提供商"
            options={[
              { label: 'Ollama', value: 'ollama' },
              { label: 'LM Studio', value: 'lm_studio' },
              { label: 'vLLM', value: 'vllm' },
              { label: 'Dify本地', value: 'dify_local' },
              { label: 'OpenAI', value: 'openai' },
              { label: 'DeepSeek', value: 'deepseek' },
              { label: 'Qwen', value: 'qwen' },
            ]}
            rules={[{ required: true, message: '请选择提供商!' }]}
          />
          <ProFormText
            name="api_url"
            label="API地址"
            rules={[{ required: true, message: '请输入API地址!' }]}
          />
          <ProFormText
            name="api_key"
            label="API密钥"
            placeholder="留空则不更新密钥（输入新密钥将替换旧密钥）"
          />
          <ProFormText
            name="model_name"
            label="模型标识"
            rules={[{ required: true, message: '请输入模型标识!' }]}
          />
          <ProFormDigit
            name="priority"
            label="优先级"
            min={0}
            max={100}
          />
          <ProFormSwitch
            name="enabled"
            label="启用"
            valuePropName="checked"
            getValueFromEvent={(checked: boolean) => checked ? 1 : 0}
            getValueProps={(value: number) => ({ checked: value === 1 })}
          />
          <ProFormDigit
            name="timeout"
            label="超时时间(秒)"
            min={1}
            max={300}
          />
          <ProFormDigit
            name="max_tokens"
            label="最大Token数"
            min={1}
            max={100000}
          />
          <ProFormDigit
            name="temperature"
            label="温度参数"
            min={0}
            max={2}
            step={0.1}
          />
          <ProFormSwitch
            name="stream_enabled"
            label="支持流式响应"
            valuePropName="checked"
            getValueFromEvent={(checked: boolean) => checked ? 1 : 0}
            getValueProps={(value: number) => ({ checked: value === 1 })}
          />
          <ProFormTextArea
            name="description"
            label="描述"
          />
        </UpdateForm>
      ) : null}
    </PageContainer>
  );
};

export default AIModelList;

