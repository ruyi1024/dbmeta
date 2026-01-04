import React from 'react';
import { Modal } from 'antd';
import ProForm, { ProFormText, ProFormSelect, ProFormDigit, ProFormTextArea } from '@ant-design/pro-form';
import type { RuleListItem } from '../data.d';

export interface CreateFormProps {
  modalVisible: boolean;
  onCancel: () => void;
  onSubmit: (values: RuleListItem) => Promise<void>;
}

const CreateForm: React.FC<CreateFormProps> = (props) => {
  const { modalVisible, onCancel, onSubmit } = props;

  return (
    <Modal
      destroyOnClose
      title="新建质量规则"
      open={modalVisible}
      onCancel={() => onCancel()}
      footer={null}
      width={600}
    >
      <ProForm
        onFinish={async (value) => {
          await onSubmit(value as RuleListItem);
        }}
        initialValues={{
          ruleType: '完整性',
          severity: 'medium',
          enabled: 1,
          threshold: 0,
        }}
      >
        <ProFormText
          name="ruleName"
          label="规则名称"
          rules={[{ required: true, message: '请输入规则名称' }]}
          placeholder="请输入规则名称"
        />
        <ProFormSelect
          name="ruleType"
          label="规则类型"
          rules={[{ required: true, message: '请选择规则类型' }]}
          options={[
            { label: '完整性', value: '完整性' },
            { label: '准确性', value: '准确性' },
            { label: '唯一性', value: '唯一性' },
            { label: '一致性', value: '一致性' },
            { label: '及时性', value: '及时性' },
          ]}
        />
        <ProFormTextArea
          name="ruleDesc"
          label="规则描述"
          placeholder="请输入规则描述"
        />
        <ProFormTextArea
          name="ruleConfig"
          label="规则配置(JSON)"
          placeholder='例如: {"max_null_rate": 0.2}'
        />
        <ProFormDigit
          name="threshold"
          label="阈值(%)"
          min={0}
          max={100}
          fieldProps={{ precision: 2 }}
        />
        <ProFormSelect
          name="severity"
          label="严重程度"
          rules={[{ required: true, message: '请选择严重程度' }]}
          options={[
            { label: '高', value: 'high' },
            { label: '中', value: 'medium' },
            { label: '低', value: 'low' },
          ]}
        />
        <ProFormSelect
          name="enabled"
          label="是否启用"
          rules={[{ required: true, message: '请选择是否启用' }]}
          options={[
            { label: '启用', value: 1 },
            { label: '禁用', value: 0 },
          ]}
        />
      </ProForm>
    </Modal>
  );
};

export default CreateForm;

