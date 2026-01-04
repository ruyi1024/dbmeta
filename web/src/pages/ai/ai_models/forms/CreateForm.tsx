import React from 'react';
import { ModalForm } from '@ant-design/pro-components';

export type CreateFormProps = {
  onCancel: () => void;
  modalVisible: boolean;
  onSubmit: (values: any) => Promise<boolean>;
  children?: React.ReactNode;
};

const CreateForm: React.FC<CreateFormProps> = (props) => {
  const { onCancel, modalVisible, onSubmit, children } = props;

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
        const success = await onSubmit(values);
        if (success) {
          onCancel();
        }
        return success;
      }}
      modalProps={{
        destroyOnClose: true,
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

