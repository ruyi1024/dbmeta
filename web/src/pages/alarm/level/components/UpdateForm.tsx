import React from 'react';
import { Modal } from 'antd';

interface UpdateFormProps {
  updateModalVisible: boolean;
  onCancel: () => void;
}

const UpdateForm: React.FC<UpdateFormProps> = (props) => {
  const { updateModalVisible, onCancel } = props;

  return (
    (<Modal
      destroyOnClose
      title="修改告警级别"
      open={updateModalVisible}
      onCancel={() => onCancel()}
      footer={null}
    >
      {props.children}
    </Modal>)
  );
};

export default UpdateForm;
