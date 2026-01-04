import { PlusOutlined, FormOutlined, DeleteOutlined } from '@ant-design/icons';
import { Button, message, Input, Form, Card, Row, Col, Alert } from 'antd';
import React, { useState, useRef } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import ProTable, { ProColumns, ActionType } from '@ant-design/pro-table';

import { useAccess } from 'umi';

/* eslint-disable no-template-curly-in-string */
const validateMessages = {
  required: '${label}是必填项!',
  types: {
    email: '${label} is not a valid email!',
    number: '${label} is not a valid number!',
  },
  number: {
    range: '${label} must be between ${min} and ${max}',
  },
};

const sendTest: React.FC<{}> = () => {
  const [loading, setLoading] = useState<boolean>(false);
  const actionRef = useRef<ActionType>();
  const access = useAccess();

  const [form] = Form.useForm();

  const [formValues, setFormValues] = useState({
    email_list: "",
    sms_list: "",
    phone_list: "",
    wechat_list: "",
    weburl: "",
  });


  //表单提交查询执行请求
  const asyncFetch = (values: {}, sendUrl: string) => {
    console.info(values);
    setLoading(true);
    const params = { ...values, };
    const headers = new Headers();
    headers.append('Content-Type', 'application/json');
    fetch(sendUrl, {
      method: 'post',
      headers: headers,
      body: JSON.stringify(params),
    })
      .then((response) => response.json())
      .then((json) => {
        console.info(json.msg);
        setLoading(false);
        if (json.success == true) {
          message.success(json.msg);
        } else {
          message.error(json.msg);
        }

      })
      .catch((error) => {
        setLoading(false);
        console.log('post data failed', error);
      });
  };


  const onFinish = (fieldValue: []) => {
    const values = {
      email_list: fieldValue["email_list"],
    };
    setFormValues(values);
    asyncFetch(values, '/api/v1/alarm/test/send_email');
  };

  const onFinish2 = (fieldValue: []) => {
    const values = {
      sms_list: fieldValue["sms_list"],
    };
    setFormValues(values);
    asyncFetch(values, '/api/v1/alarm/test/send_sms');
  };

  const onFinish3 = (fieldValue: []) => {
    const values = {
      phone_list: fieldValue["phone_list"],
    };
    setFormValues(values);
    asyncFetch(values, '/api/v1/alarm/test/send_phone');
  };

  const onFinish4 = (fieldValue: []) => {
    const values = {
      wechat_list: fieldValue["wechat_list"],
    };
    setFormValues(values);
    asyncFetch(values, '/api/v1/alarm/test/send_wechat');
  };

  const onFinish5 = (fieldValue: []) => {
    const values = {
      weburl: fieldValue["weburl"],
    };
    setFormValues(values);
    asyncFetch(values, '/api/v1/alarm/test/send_weburl');
  };

  return (
    <PageContainer >
      <Row gutter={[16, 24]} style={{ marginTop: '0px' }}>

        <Col span={8}>
          <Card title="发送邮件测试" bordered={false} >
            <Form name="nest-messages" layout="vertical" onFinish={onFinish} >
              <Form.Item name="email_list" label="收件人邮箱" initialValue={formValues.email_list} tooltip={"发送前确保邮件网关配置正确，多个收件人使用英文分号分隔"} rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item wrapperCol={{ offset: 8 }}>
                <Button type="primary" htmlType="submit" loading={loading}>
                  发送邮件
                </Button>
              </Form.Item>
            </Form>
          </Card>
        </Col>

        <Col span={8}>
          <Card title="发送短信测试" bordered={false} >
            <Form name="nest-messages" layout="vertical" onFinish={onFinish2} >
              <Form.Item name="sms_list" label="收件人手机" initialValue={formValues.sms_list} tooltip={"发送前确保短信网关配置正确，多个收件人使用英文分号分隔"} rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item wrapperCol={{ offset: 8 }}>
                <Button type="primary" htmlType="submit" loading={loading}>
                  发送短信
                </Button>
              </Form.Item>
            </Form>
          </Card>
        </Col>

        <Col span={8}>
          <Card title="电话通知测试" bordered={false} >
            <Form name="nest-messages" layout="vertical" onFinish={onFinish3} >
              <Form.Item name="phone_list" label="接听人手机" initialValue={formValues.phone_list} tooltip={"发送前确保电话网关配置正确，多个接听人使用英文分号分隔"} rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item wrapperCol={{ offset: 8 }}>
                <Button type="primary" htmlType="submit" loading={loading}>
                  拨打电话
                </Button>
              </Form.Item>
            </Form>
          </Card>
        </Col>

      </Row>

      <Row gutter={[16, 24]} style={{ marginTop: '10px' }}>
        <Col span={8}>
          <Card title="微信通知测试" bordered={false} >
            <Form name="nest-messages" layout="vertical" onFinish={onFinish4} >
              <Form.Item name="wechat_list" label="微信ID列表" initialValue={formValues.wechat_list} tooltip={"发送前确保微信服务配置正确，多个接收者使用英文分号分隔"} rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item wrapperCol={{ offset: 8 }}>
                <Button type="primary" htmlType="submit" loading={loading}>
                  发送微信
                </Button>
              </Form.Item>
            </Form>
          </Card>
        </Col>

        <Col span={8}>
          <Card title="WebHook接口测试" bordered={false} >
            <Form name="nest-messages" layout="vertical" onFinish={onFinish5} >
              <Form.Item name="weburl" label="WebUrl地址" initialValue={formValues.weburl} tooltip={"发送前确保微信服务配置正确，多个接收者使用英文分号分隔"} rules={[{ required: true }]}>
                <Input />
              </Form.Item>
              <Form.Item wrapperCol={{ offset: 8 }}>
                <Button type="primary" htmlType="submit" loading={loading}>
                  发送微信
                </Button>
              </Form.Item>
            </Form>
          </Card>
        </Col>
      </Row>
    </PageContainer>
  );
};

export default sendTest;
