import { ModalForm, ProFormText, ProFormSelect, ProFormTextArea, ProFormDigit } from '@ant-design/pro-components';
import { Form, message } from 'antd';
import { useEffect } from 'react';

export type FormValueType = {
  id: number;
  api_name: string;
  api_url: string;
  api_description?: string;
  protocol: string;
  method: string;
  headers?: string;
  params?: string;
  body?: string;
  token?: string;
  auth_type: string;
  expected_codes: string;
  timeout: number;
  retry_count: number;
  enable: number;
} & Partial<API.TableListItem>;

export type UpdateFormProps = {
  onCancel: (flag?: boolean, formVals?: FormValueType) => void;
  onSubmit: (values: FormValueType) => Promise<void>;
  updateModalVisible: boolean;
  values: Partial<API.TableListItem>;
};

const UpdateForm: React.FC<UpdateFormProps> = (props) => {
  const [form] = Form.useForm();
  const { onSubmit, onCancel, updateModalVisible, values } = props;

  useEffect(() => {
    if (updateModalVisible && values && Object.keys(values).length > 0) {
      // 安全地处理JSON字段
      const formatJsonField = (field: string | undefined) => {
        if (!field) return '';
        try {
          return JSON.stringify(JSON.parse(field), null, 2);
        } catch {
          return field;
        }
      };
      
      form.setFieldsValue({
        ...values,
        headers: formatJsonField(values.headers),
        params: formatJsonField(values.params),
        body: formatJsonField(values.body),
      });
    }
  }, [updateModalVisible, values, form]);

  const handleSubmit = async () => {
    try {
      const fieldsValue = await form.validateFields();
      
      // 处理JSON字段
      if (fieldsValue.headers) {
        try {
          JSON.parse(fieldsValue.headers);
        } catch {
          message.error('请求头格式不正确，请输入有效的JSON格式');
          return;
        }
      }
      
      if (fieldsValue.params) {
        try {
          JSON.parse(fieldsValue.params);
        } catch {
          message.error('请求参数格式不正确，请输入有效的JSON格式');
          return;
        }
      }
      
      if (fieldsValue.body) {
        try {
          JSON.parse(fieldsValue.body);
        } catch {
          message.error('请求体格式不正确，请输入有效的JSON格式');
          return;
        }
      }

      await onSubmit({
        ...fieldsValue,
        enable: fieldsValue.enable ? 1 : 0,
      });
    } catch (error) {
      console.error('表单验证失败:', error);
    }
  };

  return (
    <ModalForm
      title="编辑API配置"
      width="800px"
      form={form}
      open={updateModalVisible}
      onOpenChange={(open) => {
        if (!open) {
          form.resetFields();
          onCancel();
        }
      }}
      onFinish={handleSubmit}
    >
      <ProFormText
        name="id"
        label="ID"
        width="md"
        disabled
      />
      
      <ProFormText
        name="api_name"
        label="API名称"
        width="md"
        rules={[
          {
            required: true,
            message: '请输入API名称!',
          },
        ]}
        placeholder="请输入API名称"
      />
      
      <ProFormText
        name="api_url"
        label="API URL"
        width="md"
        rules={[
          {
            required: true,
            message: '请输入API URL!',
          },
          {
            type: 'url',
            message: '请输入有效的URL!',
          },
        ]}
        placeholder="请输入API URL，如：https://api.example.com/status"
      />
      
      <ProFormTextArea
        name="api_description"
        label="API描述"
        width="md"
        placeholder="请输入API描述"
      />
      
      <ProFormSelect
        name="protocol"
        label="协议类型"
        width="md"
        options={[
          { label: 'HTTP', value: 'HTTP' },
          { label: 'HTTPS', value: 'HTTPS' },
        ]}
        rules={[
          {
            required: true,
            message: '请选择协议类型!',
          },
        ]}
      />
      
      <ProFormSelect
        name="method"
        label="请求方法"
        width="md"
        options={[
          { label: 'GET', value: 'GET' },
          { label: 'POST', value: 'POST' },
          { label: 'PUT', value: 'PUT' },
          { label: 'DELETE', value: 'DELETE' },
        ]}
        rules={[
          {
            required: true,
            message: '请选择请求方法!',
          },
        ]}
      />
      
      <ProFormTextArea
        name="headers"
        label="请求头"
        width="md"
        placeholder="请输入JSON格式的请求头，如：{&quot;Content-Type&quot;: &quot;application/json&quot;}"
        fieldProps={{
          rows: 3,
        }}
      />
      
      <ProFormTextArea
        name="params"
        label="请求参数"
        width="md"
        placeholder="请输入JSON格式的请求参数，如：{&quot;page&quot;: 1, &quot;size&quot;: 10}"
        fieldProps={{
          rows: 3,
        }}
      />
      
      <ProFormTextArea
        name="body"
        label="请求体"
        width="md"
        placeholder="请输入JSON格式的请求体，如：{&quot;username&quot;: &quot;test&quot;, &quot;password&quot;: &quot;test123&quot;}"
        fieldProps={{
          rows: 3,
        }}
      />
      
      <ProFormText
        name="token"
        label="Token"
        width="md"
        placeholder="请输入Token认证信息"
      />
      
      <ProFormSelect
        name="auth_type"
        label="认证类型"
        width="md"
        options={[
          { label: '无认证', value: 'NONE' },
          { label: 'Basic认证', value: 'BASIC' },
          { label: 'Bearer认证', value: 'BEARER' },
          { label: 'API Key', value: 'API_KEY' },
        ]}
        rules={[
          {
            required: true,
            message: '请选择认证类型!',
          },
        ]}
      />
      
      <ProFormText
        name="expected_codes"
        label="期望返回码"
        width="md"
        placeholder="请输入期望的返回码，多个用逗号分隔，如：200,201"
        rules={[
          {
            required: true,
            message: '请输入期望返回码!',
          },
        ]}
      />
      
      <ProFormDigit
        name="timeout"
        label="超时时间(秒)"
        width="md"
        min={1}
        max={300}
        rules={[
          {
            required: true,
            message: '请输入超时时间!',
          },
        ]}
      />
      
      <ProFormDigit
        name="retry_count"
        label="重试次数"
        width="md"
        min={0}
        max={10}
      />
      
      <ProFormSelect
        name="enable"
        label="是否启用"
        width="md"
        options={[
          { label: '启用', value: 1 },
          { label: '禁用', value: 0 },
        ]}
        rules={[
          {
            required: true,
            message: '请选择是否启用!',
          },
        ]}
      />
    </ModalForm>
  );
};

export default UpdateForm;
