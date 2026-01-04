import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Form, message, DatePicker, Select, Space, Button, Menu } from 'antd';
import { DashboardOutlined, AreaChartOutlined } from '@ant-design/icons';
import LineChart from '@/components/Chart/LineChart';
import AreaChart from '@/components/Chart/AreaChart';
import moment from 'moment';
import { PageContainer } from '@ant-design/pro-layout';

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
    clientsChartList: [],
    memoryChartList: [],
    opsChartList: [],
    commandChartList: [],
    keysChartList: [],
    keyspaceChartList: [],
  });
  const [formValues, setFormValues] = useState({
    host: '',
    port: '',
    start_time: moment().subtract(1, 'hours').format('YYYY-MM-DD HH:mm:ss'),
    end_time: moment().format('YYYY-MM-DD HH:mm:ss'),
  });
  const [instanceList, setInstanceList] = useState([]);
  const [current, setCurrent] = useState<string>("");

  //console.info(this.props.match.params);

  useEffect(() => {
    setCurrent("chart");
    fetch('/api/v1/meta/node/list_search?module=redis')
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
    const headers = new Headers();
    headers.append('Content-Type', 'application/json');
    fetch('/api/v1/performance/redis/chart', {
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
      host: fieldValue['instance'].split(':')[0],
      port: fieldValue['instance'].split(':')[1],
      start_time: fieldValue['time_range'][0].format('YYYY-MM-DD HH:mm:ss'),
      end_time: fieldValue['time_range'][1].format('YYYY-MM-DD HH:mm:ss'),
    };
    setFormValues(values);
    asyncFetch(values);
  };

  const onFinishFailed = (errorInfo) => {
    message.error('查询失败');
  };

  const disabledDate = (current) => {
    // Can not select days before today and today
    return (current && current > moment().endOf('day')) || (current && current < moment().subtract(7, 'days').endOf('day'));
  }

  return (
    <>
      <Menu mode="horizontal" selectedKeys={[current]} >
        <Menu.Item key="dashboard" icon={<DashboardOutlined />}>
          <a href="/performance/redis/health" rel="noopener noreferrer">
            Redis健康大盘
          </a>
        </Menu.Item>
        <Menu.Item key="chart" icon={<AreaChartOutlined />}>
          <a href="/performance/redis/chart" rel="noopener noreferrer">
            Redis性能图表
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
              label={'选择Redis实例'}
              rules={[{ required: true, message: '选择Redis实例' }]}
            >
              <Select showSearch style={{ width: 260 }} placeholder="选择Redis实例">
                {instanceList &&
                  instanceList.map((item) => (
                    <Option key={item.ip + ':' + item.port} value={item.ip + ':' + item.port}>
                      {item.ip + ':' + item.port}
                    </Option>
                  ))}
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
        data.clientsChartList && data.clientsChartList.length > 0 &&
        <>
          <Row gutter={[16, 24]}>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.clientsChartList} unit="" />
                </div>
              </Card>
            </Col>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.memoryChartList} unit="MB" />
                </div>
              </Card>
            </Col>
          </Row>
          <Row gutter={[16, 24]}>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.opsChartList} unit="" />
                </div>
              </Card>
            </Col>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.commandChartList} unit="" />
                </div>
              </Card>
            </Col>
          </Row>
          <Row gutter={[16, 24]}>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.keysChartList} unit="" />
                </div>
              </Card>
            </Col>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.keyspaceChartList} unit="" />
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
