import React from 'react';
import { ModalForm } from '@ant-design/pro-components';
import { AIModel } from '../api';

export type UpdateFormProps = {
  onCancel: () => void;
  modalVisible: boolean;
  values: Partial<AIModel>;
  onSubmit: (values: any, id: number) => Promise<boolean>;
  children?: React.ReactNode;
};

const UpdateForm: React.FC<UpdateFormProps> = (props) => {
  const { onCancel, modalVisible, values, onSubmit, children } = props;

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
        const success = await onSubmit(formValues, values.id!);
        if (success) {
          onCancel();
        }
        return success;
      }}
      modalProps={{
        destroyOnClose: true,
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

