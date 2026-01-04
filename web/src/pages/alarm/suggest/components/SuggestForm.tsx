import React, { useEffect } from 'react';
import {Form, Input, Space, Button } from 'antd';
import {Drawer} from "antd";
import type {SuggestListItem } from "../data";
import '../style.css'
import SuggestFormInput from "@/pages/alarm/suggest/components/SuggestFormInput";

type SuggestFormProps = {
  updateVisible: boolean;
  onSubmit: (values: SuggestListItem) => void;
  onClose: () => void;
  values: SuggestListItem;
}


const layout = {
  labelCol: { span: 5 },
  wrapperCol: { span: 16 },
};

const SuggestForm: React.FC<SuggestFormProps> = ({
  updateVisible,
  onSubmit,
  onClose,
  values
  }) => {
  // const intl = useIntl();
  const [form] = Form.useForm();

  useEffect(() => {
    if(values !== null && values.modify){
      form.setFieldsValue({...values})
    } else {
      form.resetFields();
    }
  }, [values]);


  return (
    <Drawer
      // destroyOnClose
      mask
      width={1000}
      height={680}
      title={values.modify ? `修改 ${values.event_key}` : '新增'}
      open={updateVisible}
      onClose={onClose}
      placement={'top'}
      forceRender={false}
      footer={
        <Space>
          <Button onClick={onClose}>关闭</Button>
          <Button onClick={()=>{
            form
              .validateFields()
              .then(vals => {
                const data = {...vals, modify: values.modify};
                if (values.modify) {
                  // @ts-ignore
                  data.id = values.id || 0
                }
                onSubmit({...data});
                form.resetFields();
              })
              .catch(info => {
                console.log('Validate Failed:', info);
              });

          }} type="primary">提交</Button>
        </Space>
      }
      >

        <Form
          {...layout}
          form={form}
        >
          <Form.Item name="event_type" label="事件类型" rules={[{ required: true }]}>
            <Input style={{width: 240}} />
          </Form.Item>
          <Form.Item name="event_key" label="事件指标" rules={[{ required: true }]}>
           <Input style={{width: 240}} />
          </Form.Item>
          <Form.Item name="content" label={"建议内容"}>
            <SuggestFormInput />
          </Form.Item>
        </Form>
    </Drawer>
  );
}

export default SuggestForm;
