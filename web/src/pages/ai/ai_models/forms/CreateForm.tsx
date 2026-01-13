import React, { useState } from 'react';
import { ModalForm } from '@ant-design/pro-components';
import { message, Modal } from 'antd';
import { testModelConfig } from '../api';

export type CreateFormProps = {
  onCancel: () => void;
  modalVisible: boolean;
  onSubmit: (values: any) => Promise<boolean>;
  children?: React.ReactNode;
};

const CreateForm: React.FC<CreateFormProps> = (props) => {
  const { onCancel, modalVisible, onSubmit, children } = props;
  const [testing, setTesting] = useState(false);

  return (
    <ModalForm
      title="新建AI模型"
      width="800px"
      open={modalVisible}
      onOpenChange={(open) => {
        if (!open) {
          onCancel();
        }
      }}
      onFinish={async (values) => {
        // 保存前进行可用性检测
        setTesting(true);
        try {
          const testResult = await testModelConfig(values);
          if (!testResult.success) {
            Modal.confirm({
              title: '可用性检测失败',
              content: testResult.error || testResult.message || '模型配置测试失败，是否仍要保存？',
              okText: '仍要保存',
              cancelText: '取消',
              onOk: async () => {
                const success = await onSubmit(values);
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
              const success = await onSubmit(values);
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
        const success = await onSubmit(values);
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
        priority: 0,
        enabled: 0,
        timeout: 30,
        max_tokens: 2000,
        temperature: 0.7,
        stream_enabled: 0,
      }}
    >
      {children}
    </ModalForm>
  );
};

export default CreateForm;

