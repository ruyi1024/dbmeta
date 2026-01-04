import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Form, message, DatePicker, Select, Space, Button, Menu } from 'antd';
import LineChart from '@/components/Chart/LineChart';
import AreaChart from '@/components/Chart/AreaChart';
import moment from 'moment';
import { PageContainer } from "@ant-design/pro-layout";
import { AppstoreOutlined, MailOutlined } from "@ant-design/icons";

const { RangePicker } = DatePicker;
const { Option } = Select;


function onBlur() {
  console.log('blur');
}

function onFocus() {
  console.log('focus');
}

function onSearch(val) {
  console.log('search:', val);
}


const DemoLine: React.FC = () => {

  const [form] = Form.useForm();
  const [data, setData] = useState({
    connectionsChartList: [],
    memChartList: [],
    networkBytesChartList: [],
    networkRequestsChartList: [],
    opcountersChartList: [],
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
    fetch('/api/v1/meta/node/list_search?module=mongodb')
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


  const asyncFetch = (values: []) => {
    const params = { ...formValues, ...values };
    let headers = new Headers();
    headers.append('Content-Type', 'application/json');
    fetch('/api/v1/performance/mongodb/chart', {
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

  const onFinishFailed = (errorInfo) => {
    message.error('查询失败');
  };

  return (
    <>
      <Menu mode="horizontal" selectedKeys={[current]} >
        <Menu.Item key="dashboard" icon={<MailOutlined />}>
          <a href="/performance/mongodb/health" rel="noopener noreferrer">
            Mongodb健康大盘
          </a>
        </Menu.Item>
        <Menu.Item key="chart" icon={<AppstoreOutlined />}>
          <a href="/performance/mongodb/chart" rel="noopener noreferrer">
            Mongodb性能图表
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
              label={'选择MongoDB实例'}
              rules={[{ required: true, message: '请选择' }]}
            >
              <Select showSearch style={{ width: 260 }} placeholder="">
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
                onChange={onchange}
                ranges={{}}
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
                  <LineChart data={data.opcountersChartList} unit="" />
                </div>
              </Card>
            </Col>
          </Row>
          <Row gutter={[16, 24]}>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.networkBytesChartList} unit="" />
                </div>
              </Card>
            </Col>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.networkRequestsChartList} unit="kb" />
                </div>
              </Card>
            </Col>
          </Row>

          <Row gutter={[16, 24]}>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.memChartList} unit="" />
                </div>
              </Card>
            </Col>
            <Col span={12}>

            </Col>
          </Row>
        </>
      }
    </>
  );
};

export default DemoLine;
