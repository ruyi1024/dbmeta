import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Form, message, DatePicker, Select, Space, Button, Menu } from 'antd';
import LineChart from '@/components/Chart/LineChart';
import moment from 'moment';
import { AppstoreOutlined, AreaChartOutlined, DashboardOutlined, MailOutlined } from "@ant-design/icons";

const { RangePicker } = DatePicker;
const { Option } = Select;

const DemoLine: React.FC = () => {

  const [form] = Form.useForm();

  const [data, setData] = useState({
    connectionsChartList: [],
    queryChartList: [],
    longQueryChartList: [],
    lockChartList: [],
    tupChartList: [],
    xactsChartList: [],
    CheckpointPctList: [],
    CheckpointWriteList: [],
  });

  const [formValues, setFormValues] = useState({
    host: "",
    port: "",
    start_time: moment().subtract(1, 'hours').format('YYYY-MM-DD HH:mm:ss'),
    end_time: moment().format('YYYY-MM-DD HH:mm:ss'),
  });
  const [instanceList, setInstanceList] = useState([]);
  const [current, setCurrent] = useState<string>("");
  //console.info(this.props.match.params);

  useEffect(() => {
    setCurrent("chart");
    fetch('/api/v1/meta/node/list_search?module=postgresql')
      .then((response) => response.json())
      .then((json) => setInstanceList(json.data))
      .catch((error) => {
        console.log('fetch instances failed', error);
      });
    // setFormValues({"ip":"106.13.177.17","port":"3307","start_time":moment().subtract(1, 'hours').format('YYYY-MM-DD HH:mm:ss'),"end_time":moment().format('YYYY-MM-DD HH:mm:ss')})
    // console.info(formValues);
    // console.info(instanceList)
    // asyncFetch();
  }, []);


  const asyncFetch = (values: {}) => {
    const params = { ...formValues, ...values };
    const headers = new Headers();
    headers.append('Content-Type', 'application/json');
    fetch('/api/v1/performance/postgresql/chart', {
      method: 'post',
      headers: headers,
      body: JSON.stringify(params),
    })
      .then((response) => response.json())
      .then((json) => setData(json.data))
      .catch((error) => {
        console.log('fetch data failed', error);
      });
  };

  const onFinish = (fieldValue: []) => {
    const values = {
      host: fieldValue["instance"].split(":")[0],
      port: fieldValue["instance"].split(":")[1],
      start_time: fieldValue['time_range'][0].format('YYYY-MM-DD HH:mm:ss'),
      end_time: fieldValue['time_range'][1].format('YYYY-MM-DD HH:mm:ss'),
    };
    setFormValues(values);
    asyncFetch(values);
  };

  const onFinishFailed = (errorInfo: any) => {
    console.info(errorInfo);
    message.error('查询失败');
  };

  const disabledDate = (current) => {
    // Can not select days before today and today
    return (current && current > moment().endOf('day')) || (current && current < moment().subtract(7, 'days').endOf('day'));
  }

  // @ts-ignore
  return (
    <>
      <Menu mode="horizontal" selectedKeys={[current]} >
        <Menu.Item key="dashboard" icon={<DashboardOutlined />}>
          <a href="/performance/postgresql/health" rel="noopener noreferrer">
            PostgreSQL健康大盘
          </a>
        </Menu.Item>
        <Menu.Item key="chart" icon={<AreaChartOutlined />}>
          <a href="/performance/postgresql/chart" rel="noopener noreferrer">
            PostgreSQL性能图表
          </a>
        </Menu.Item>
      </Menu>

      <Card bordered={false}>
        <Form
          hideRequiredMark
          style={{ marginTop: 8 }}
          form={form}
          name={'basic'}
          onFinish={onFinish}
          onFinishFailed={onFinishFailed}
          initialValues={{ time_range: [moment().subtract(1, 'hours'), moment()] }}
        >
          <Space>
            <Form.Item
              name={'instance'}
              label={'选择PostgreSQL实例'}
              rules={[{ required: true, message: '选择PostgreSQL实例' }]}
            >
              <Select showSearch style={{ width: 260 }} placeholder="请选择">
                {instanceList && instanceList.map(item => <Option key={item.ip + ":" + item.port} value={item.ip + ":" + item.port}>{item.ip + ":" + item.port}</Option>)}

              </Select>
            </Form.Item>
            <Form.Item
              name={'time_range'}
              label={'选择时间范围'}
              rules={[{ required: true, message: '请选择时间范围' }]}
            >
              <RangePicker
                showTime
                format={'YYYY/MM/DD HH:mm:ss'}
                ranges={{
                  '30分钟': [moment().subtract(30, 'minutes'), moment()],
                  '1小时': [moment().subtract(1, 'hours'), moment()],
                  今天: [moment('00:00:00', 'HH:mm:ss'), moment()],
                  最近1天: [moment().subtract(3, 'days'), moment()],
                }}
                placeholder={['开始时间', '结束时间']}
                disabledDate={disabledDate}
              />
            </Form.Item>
            <Form.Item>
              <Button type={'primary'} htmlType={'submit'}>
                绘制图表
              </Button>
            </Form.Item>
          </Space>
        </Form>
      </Card>

      {
        data.connectionsChartList && data.connectionsChartList.length > 0 &&
        <>
          <Row gutter={[16, 24]}>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.connectionsChartList} unit="" />
                </div>
              </Card>
            </Col>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.queryChartList} unit="" />
                </div>
              </Card>
            </Col>
          </Row>
          <Row gutter={[16, 24]}>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.longQueryChartList} unit="" />
                </div>
              </Card>
            </Col>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.lockChartList} unit="" />
                </div>
              </Card>
            </Col>
          </Row>

          <Row gutter={[16, 24]}>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.tupChartList} unit="" />
                </div>
              </Card>
            </Col>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.xactsChartList} unit="" />
                </div>
              </Card>
            </Col>
          </Row>

          <Row gutter={[16, 24]}>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.CheckpointPctList} unit="" />
                </div>
              </Card>
            </Col>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.CheckpointWriteList} unit="kb" />
                </div>
              </Card>
            </Col>
          </Row>

        </>
      }
    </>
  );
};

export default DemoLine;
