import React, { useState, useEffect, useRef } from 'react';
import { Card, Row, Col, Form, message, DatePicker, Select, Space, Button, Table, Alert } from 'antd';
import LineChart from '@/components/Chart/LineChart';
import { Gauge } from '@ant-design/plots';
import { PageContainer } from '@ant-design/pro-components';
import moment from 'moment';


const { RangePicker } = DatePicker;
const { Option } = Select;

const ServerChart: React.FC = () => {

  const [form] = Form.useForm();

  const [data, setData] = useState({
    loadChartList: [],
    cpuPercentChartList: [],
    memoryUsedPercentChartList: [],
    diskUsedPercentChartList: [],
  });

  const [formValues, setFormValues] = useState({
    host: "",
    port: "",
    start_time: moment().subtract(1, 'hours').format('YYYY-MM-DD HH:mm:ss'),
    end_time: moment().format('YYYY-MM-DD HH:mm:ss'),
  });
  const [serverList, setServerList] = useState([]);
  const [current, setCurrent] = useState<string>("");
  //console.info(this.props.match.params);

  useEffect(() => {
    fetch('/api/v1/performance/server/list')
      .then((response) => response.json())
      .then((json) => setServerList(json.data))
      .catch((error) => {
        console.log('fetch server list failed', error);
      });
  }, []);


  const asyncFetch = (values: {}) => {
    const params = { ...formValues, ...values };
    const headers = new Headers();
    headers.append('Content-Type', 'application/json');
    fetch('/api/v1/performance/server/chart', {
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
      ip: fieldValue["ip"],
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

  const ticks = [0, 1 / 3, 2 / 3, 1];
  const color = ['#30BF78', '#FAAD14', '#F4664A'];
  const graphRef = useRef(null);
  useEffect(() => {
    if (graphRef.current) {
      let data = 0.7;
      const interval = setInterval(() => {
        if (data >= 1.5) {
          clearInterval(interval);
        }

        data += 0.005;
        graphRef.current.changeData(data > 1 ? data - 1 : data);
      }, 1000);
    }
  }, [graphRef]);
  const config = {
    percent: 0,
    range: {
      ticks: [0, 1],
      color: ['l(0) 0:#30BF78 0.5:#FAAD14 1:#F4664A'],
    },
    indicator: {
      pointer: {
        style: {
          stroke: '#D0D0D0',
        },
      },
      pin: {
        style: {
          stroke: '#D0D0D0',
        },
      },
    },
    statistic: {
      title: {
        formatter: ({ percent }) => {
          if (percent < ticks[1]) {
            return '低';
          }

          if (percent < ticks[2]) {
            return '中';
          }

          return '高';
        },
        style: ({ percent }) => {
          return {
            fontSize: '24px',
            lineHeight: 1,
            color: percent < ticks[1] ? color[0] : percent < ticks[2] ? color[1] : color[2],
          };
        },
      },
      content: {
        offsetY: 36,
        style: {
          fontSize: '24px',
          color: '#4B535E',
        },
        formatter: () => '',
      },
    },
    onReady: (plot) => {
      graphRef.current = plot;
    },
  };

  const columns = [
    {
      title: '名称',
      dataIndex: 'name',
    },
    {
      title: '数据',
      dataIndex: 'address',
    },
  ];

  const data1 = [
    {
      key: '1',
      name: '主机名',
      address: 'ebs-75895',
    },
    {
      key: '2',
      name: '操作系统',
      address: 'centos',
    },
    {
      key: '3',
      name: '启动时间',
      address: '2022-05-11 10:38:51',
    },
    {
      key: '4',
      name: 'CPU核心',
      address: '4 core',
    },
    {
      key: '5',
      name: '内存容量',
      address: '3789MB',
    },
  ];

  // @ts-ignore
  return (
    <PageContainer content="">
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
              name={'ip'}
              label={'选择服务器'}
              rules={[{ required: true, message: '请选择服务器' }]}
            >
              <Select showSearch style={{ width: 260 }} placeholder="请选择服务器">
                {serverList && serverList.map(item => <Option key={item.ip} value={item.ip}>{item.ip}</Option>)}

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
        data.cpuPercentChartList && data.cpuPercentChartList.length > 0 &&
        <>
          <Row gutter={[16, 24]}>
            <Col xl={6} lg={12} sm={24} xs={24}>
              <Card
                title=""
                bodyStyle={{ textAlign: 'center', fontSize: 0 }}
                bordered={false}
                style={{ marginTop: 10 }}
              >
                <Table columns={columns} dataSource={data1} size="small" pagination={false} />
              </Card>
            </Col>
            <Col xl={6} lg={12} sm={24} xs={24}>
              <Card
                title="CPU使用率"
                bodyStyle={{ textAlign: 'center', fontSize: 0 }}
                bordered={false}
                style={{ marginTop: 10 }}
              >
                <div style={{ height: 168 }}>
                  <Gauge {...config} />
                </div>
              </Card>
            </Col>
            <Col xl={6} lg={12} sm={24} xs={24}>
              <Card
                title="内存使用率"
                bodyStyle={{ textAlign: 'center', fontSize: 0 }}
                bordered={false}
                style={{ marginTop: 10 }}
              >
                <div style={{ height: 168 }}>
                  <Gauge {...config} />
                </div>
              </Card>
            </Col>

            <Col xl={6} lg={12} sm={24} xs={24} >
              <Card
                title="硬盘最大使用率"
                bodyStyle={{ textAlign: 'center', fontSize: 0 }}
                bordered={false}
                style={{ marginTop: 10 }}
              >
                <div style={{ height: 168 }}>
                  <Gauge {...config} />
                </div>
              </Card>
            </Col>
          </Row>
          <Row gutter={[16, 24]}>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.loadChartList} unit="" />
                </div>
              </Card>
            </Col>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.cpuPercentChartList} unit="%" />
                </div>
              </Card>
            </Col>
          </Row>
          <Row gutter={[16, 24]}>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.memoryUsedPercentChartList} unit="" />
                </div>
              </Card>
            </Col>
            <Col span={12}>
              <Card style={{ marginTop: 10 }}>
                <div style={{ height: 300 }}>
                  <LineChart data={data.diskUsedPercentChartList} unit="%" />
                </div>
              </Card>
            </Col>
          </Row>
        </>
      }
    </PageContainer>
  );
};

export default ServerChart;
