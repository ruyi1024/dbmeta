import React, { useState } from 'react';
import { ModalForm } from '@ant-design/pro-components';
import { message, Modal } from 'antd';
import { AIModel, testModelConfig } from '../api';

export type UpdateFormProps = {
  onCancel: () => void;
  modalVisible: boolean;
  values: Partial<AIModel>;
  onSubmit: (values: any, id: number) => Promise<boolean>;
  children?: React.ReactNode;
};

const UpdateForm: React.FC<UpdateFormProps> = (props) => {
  const { onCancel, modalVisible, values, onSubmit, children } = props;
  const [testing, setTesting] = useState(false);

  return (
    <ModalForm
      title="编辑AI模型"
      width="800px"
      open={modalVisible}
      onOpenChange={(open) => {
        if (!open) {
          onCancel();
        }
      }}
      onFinish={async (formValues) => {
        // 保存前进行可用性检测
        setTesting(true);
        try {
          // 使用testModelConfig接口测试配置（因为可能修改了配置）
          const testResult = await testModelConfig({ ...values, ...formValues });
          if (!testResult.success) {
            Modal.confirm({
              title: '可用性检测失败',
              content: testResult.error || testResult.message || '模型配置测试失败，是否仍要保存？',
              okText: '仍要保存',
              cancelText: '取消',
              onOk: async () => {
                const success = await onSubmit(formValues, values.id!);
                if (success) {
                  onCancel();
                }
                setTesting(false);
              },
              onCancel: () => {
                setTesting(false);
              },
            });
            return false;
          }
          message.success('可用性检测通过');
        } catch (error: any) {
          Modal.confirm({
            title: '可用性检测失败',
            content: error?.message || '模型配置测试失败，是否仍要保存？',
            okText: '仍要保存',
            cancelText: '取消',
            onOk: async () => {
              const success = await onSubmit(formValues, values.id!);
              if (success) {
                onCancel();
              }
              setTesting(false);
            },
            onCancel: () => {
              setTesting(false);
            },
          });
          return false;
        } finally {
          setTesting(false);
        }

        // 检测通过后保存
        const success = await onSubmit(formValues, values.id!);
        if (success) {
          onCancel();
        }
        return success;
      }}
      modalProps={{
        destroyOnClose: true,
      }}
      submitter={{
        submitButtonProps: {
          loading: testing,
        },
      }}
      initialValues={{
        ...values,
        enabled: values.enabled === 1,
        stream_enabled: values.stream_enabled === 1,
      }}
    >
      {children}
    </ModalForm>
  );
};

export default UpdateForm;

