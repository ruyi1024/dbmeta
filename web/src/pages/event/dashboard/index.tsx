import React, { useEffect, useState } from 'react';
import { Row, Col, Card, Alert, message, Tooltip, Table, Space } from 'antd';
import { InfoCircleOutlined, SmileTwoTone, PieChartOutlined, LineChartOutlined, ProfileTwoTone, SoundTwoTone } from '@ant-design/icons';
import styles from './index.less';
import { ChartCard, MiniArea, MiniBar, MiniProgress } from './components/Charts';
import Trend from './components/Trend';
import { Gauge } from '@ant-design/charts';
import PieChart from '@/components/Chart/PieChart';
import LineChart from '@/components/Chart/LineChart';
import moment from "moment";

//const wsAddr = `ws://127.0.0.1:8080/api/v1/ws/dashbaord/info`
const wsAddr = `ws://${window.location.hostname}${window.location.port === '' ? '' : ':8088'
  }/api/v1/monitor/dashbaord/websocket`;

export default (): React.ReactNode => {
  const [dashData, setDashData] = useState<any>([]);
  const [wsState, setWsState] = useState<boolean>(false);
  const [wsData, setWsData] = useState<any>([]);
  const [seconds, setSeconds] = useState<number>(1);
  const [lastTime, setLastTime] = useState<any>(new Date());
  const [loading, setLoading] = useState<boolean>(true);
  const [eventList, setEventList] = useState<any>([]);
  const [alarmList, setAlarmList] = useState<any>([]);
  const [alarmCount, setAlarmCount] = useState<number>(0);
  const [alarmMinuteData, setAlarmMinuteData] = useState<any>([]);
  const [eventCount, setEventCount] = useState<number>(0);
  const [eventHourCount, setEventHourCount] = useState<number>(0);
  const [eventMinuteData, setEventMinuteData] = useState<any>([]);
  const [eventHourData, setEventHourData] = useState<any>([]);
  const [lastEventTime, setLastEventTime] = useState<any>();
  const [lastAlarmTime, setLastAlarmTime] = useState<any>();
  const [taskCount, setTaskCount] = useState<number>(0);
  const [taskHourCount, setTaskHourCount] = useState<number>(0);
  const [taskMinuteData, setTaskMinuteData] = useState<any>([]);
  //const [alarmPieData, setAlarmPieData] = useState<any>([{ type: 'noData', value: 1 }]);
  const [nodeCount, setNodeCount] = useState<number>(1);
  const [taskNextCount, setTaskNextCount] = useState<number>(1);
  const [taskFailCount, setTaskFailCount] = useState<number>(0);
  const [disEventsCount, setDisEventsCount] = useState<number>(1);
  const [sqlModeUnSupportCount, setSqlModeUnSupportCount] = useState<number>(0);

  //健康仪表盘
  const [percent, setPercent] = useState(0);
  let ref;
  const ticks = [0, 75 / 100, 90 / 100, 99 / 100, 1];
  const color = ['#F4664A', '#FAAD14', '#30BF78'];
  const config = {
    percent,
    range: {
      ticks: [0, 100],
      color: ['l(0) 0:#F4664A 0.5:#FAAD14 1:#30BF78'],
    },
    indicator: {
      pointer: { style: { stroke: '#D0D0D0' } },
      pin: { style: { stroke: '#D0D0D0' } },
    },
    statistic: {
      title: {
        formatter: function formatter(_ref) {
          const percent = _ref.percent;
          if (percent < ticks[1]) {
            return '糟糕';
          }
          if (percent < ticks[2]) {
            return '中等';
          }
          if (percent < ticks[3]) {
            return '良好';
          }
          return '优秀';
        },
        style: function style(_ref2: { percent: any; }) {
          // eslint-disable-next-line @typescript-eslint/no-shadow
          const percent = _ref2.percent;
          return {
            fontSize: '36px',
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
      },
    },
  };

  useEffect(() => {
    const WS = new WebSocket(wsAddr);
    let intervalId: NodeJS.Timeout | null = null;

    const tick = () => {
      // 检查 WebSocket 状态，只有在 OPEN 状态下才发送
      if (WS.readyState === WebSocket.OPEN) {
        WS.send(JSON.stringify('ping', null));
        setSeconds(seconds + 1);
        setLastTime(new Date());
        setLoading(false);
      }
    };

    WS.onmessage = (evt) => {
      const data = JSON.parse(evt.data);
      setWsData(data);
      setEventList(data.eventList);
      setAlarmList(data.alarmList);
      setAlarmCount(data.alarmCount);
      setAlarmMinuteData(data.alarmMinuteData);
      setEventCount(data.eventCount);
      setEventHourCount(data.eventHourCount);
      setEventMinuteData(data.eventMinuteData);
      setEventHourData(data.eventHourData);
      setLastEventTime(data.lastEventTime);
      setLastAlarmTime(data.lastAlarmTime);
      setTaskCount(data.taskCount);
      setTaskHourCount(data.taskHourCount);
      setTaskMinuteData(data.taskMinuteData);
      setPercent(data.healthPct);
      //setAlarmPieData(data.alarmPieData);
      //console.info(data.alarmPieData)
      setNodeCount(data.nodeCount);
      setTaskNextCount(data.taskNextCount);
      setTaskFailCount(data.taskFailCount);
      setDisEventsCount(data.disEvents);
      setSqlModeUnSupportCount(data.sqlModeUnSupportCount)
    };
    WS.onopen = () => {
      setWsState(true);
      tick();
      // 启动定时器
      intervalId = setInterval(() => tick(), 5000);
    };
    WS.onclose = () => {
      setWsState(false);
      // 清除定时器
      if (intervalId) {
        clearInterval(intervalId);
        intervalId = null;
      }
    };
    WS.onerror = () => {
      setWsState(false);
      message.error('WebSocket通信失败！');
      // 清除定时器
      if (intervalId) {
        clearInterval(intervalId);
        intervalId = null;
      }
    };

    // 清理函数：组件卸载时关闭 WebSocket 和清除定时器
    return () => {
      if (intervalId) {
        clearInterval(intervalId);
      }
      if (WS.readyState === WebSocket.OPEN || WS.readyState === WebSocket.CONNECTING) {
        WS.close();
      }
    };
  }, []);

  const columns_event = [
    {
      title: '事件时间',
      dataIndex: 'event_time',
    },
    {
      title: '事件类型',
      dataIndex: 'event_type',
    },
    {
      title: '事件组',
      dataIndex: 'event_group',
    },
    {
      title: '事件实体',
      dataIndex: 'event_entity',
    },
    {
      title: '事件指标',
      dataIndex: 'event_key',
    },
    {
      title: '事件值',
      dataIndex: 'event_value',
      render: (_: any, record: any) => <>{record.event_value}</>,
    },
  ];

  const columns_alarm = [
    {
      title: '告警时间',
      dataIndex: 'event_time',
    },
    {
      title: '告警信息',
      dataIndex: 'alarm_title',
    },
    {
      title: '告警级别',
      dataIndex: 'alarm_level',
    },
    {
      title: '事件类型',
      dataIndex: 'event_type',
    },
    {
      title: '事件实体',
      dataIndex: 'event_entity',
    },
  ];

  useEffect(() => {
    try {
      fetch(`/api/v1/monitor/dashbaord/info`)
        .then((response) => response.json())
        .then((json) => {
          return (
            setDashData(json.data)
          );
        })
        .catch((error) => {
          console.log('fetch dashboard data failed', error);
        });
    } catch (e) {
      message.error(`get data error. ${e}`)
    }
  }, []);

  //console.info(dashData);
  return (
    <div>
      {!wsState && (
        <Alert type="error" message="WebSocket服务通信失败，请检查服务是否正常" banner />
      )}
      {wsState && (<Alert type={"success"} message={"WebSocket连接成功, 请求时间: " + (lastTime ? moment(lastTime).format('YYYY-MM-DD HH:mm:ss') : '-')} banner />)}
      {wsState && wsData.datasourceCount == 0 && (
        <Alert type="warning" message="未配置监控数据源信息，请先配置监控数据源信息" banner />
      )}
      {wsState && wsData.currentEventCount == 0 && (
        <Alert type="warning" message="近一分钟未产生新事件，请检查任务是否正常运行" banner />
      )}

      <Row gutter={[16, 24]} style={{ marginTop: '10px' }}>
        <Col span={6}>
          <ChartCard
            bordered={false}
            loading={loading}
            title="实时事件数量"
            action={
              <Tooltip title="近10分钟内每分钟产生的事件数量">
                <InfoCircleOutlined />
              </Tooltip>
            }
            total={eventCount}
            footer={
              <Trend flag="up" style={{ marginRight: 16 }}>
                最新事件: <span className={styles.trendText}>{lastEventTime}</span>
              </Trend>
            }
            contentHeight={46}
          >
            <MiniArea color="#1979C9" data={eventMinuteData} animate={false} />
          </ChartCard>
        </Col>
        <Col span={6}>
          <ChartCard
            bordered={false}
            loading={loading}
            title="小时事件数量"
            action={
              <Tooltip title="近60分钟内每5分钟产生的事件数量">
                <InfoCircleOutlined />
              </Tooltip>
            }
            total={eventHourCount}
            footer={
              <Trend flag="up" style={{ marginRight: 16 }}>
                最新事件: <span className={styles.trendText}>{lastEventTime}</span>
              </Trend>
            }
            contentHeight={46}
          >
            <MiniBar color="#D62A0D" data={eventHourData} forceFit={true} height={60} />
          </ChartCard>
        </Col>
        <Col span={6}>
          <ChartCard
            loading={loading}
            bordered={false}
            title="即时指标覆盖度"
            action={
              <Tooltip title="当前1分钟指标覆盖小时指标占比">
                <InfoCircleOutlined />
              </Tooltip>
            }
            total={wsData && wsData.eventKeyPctStr}
            footer={
              <div style={{ whiteSpace: 'nowrap', overflow: 'hidden' }}>
                <>
                  当前指标数：
                  <span className={styles.trendText}>{wsData && wsData.minKeyCount}</span>
                </>
                <>
                  &nbsp;&nbsp;&nbsp;&nbsp;小时指标数：
                  <span className={styles.trendText}>{wsData && wsData.hourKeyCount}</span>
                </>
              </div>
            }
            contentHeight={46}
          >
            <MiniProgress percent={wsData && wsData.eventKeyPct} strokeWidth={8} target={100} />
          </ChartCard>

        </Col>
        <Col span={6}>
          <ChartCard
            bordered={false}
            title={'异常数据源'}
            action={
              <Tooltip title="数据源监控情况统计">
                <InfoCircleOutlined />
              </Tooltip>
            }
            loading={loading}
            total={() => wsData.failDatasourceCount}
            footer={
              <>
                健康数据源占比: <span className={styles.trendText}>{wsData && wsData.healthPct2}%</span>
              </>
            }
            contentHeight={46}
          >
            数据源：<span className={styles.trendText}>{wsData && wsData.totalDatasourceCount && wsData.totalDatasourceCount} </span>&nbsp;&nbsp; 监控中：
            <span className={styles.trendText}>{wsData && wsData.monitorDatasourceCount} </span>&nbsp;&nbsp; 正常：
            <span className={styles.trendText}>{wsData && wsData.healthDatasourceCount} </span>&nbsp;&nbsp; 异常：
            <span className={styles.trendText}>{wsData && wsData.failDatasourceCount} </span>&nbsp;&nbsp;
          </ChartCard>
        </Col>
      </Row>

      <Row gutter={[16, 24]} style={{ marginTop: '15px' }}>
        <Col span={12}>
          <Card title={<span><SmileTwoTone />&nbsp;数据源健康度</span>} bordered={false} style={{ paddingBottom: '40px' }}>
            <Gauge
              {...config}
              loading={loading}
              chartRef={(chartRef) => {
                // eslint-disable-next-line @typescript-eslint/no-unused-vars
                ref = chartRef
              }}
              height={300}
            />
          </Card>
        </Col>
        <Col span={12}>

          <Card title={<span><ProfileTwoTone />&nbsp;实时事件</span>} bordered={false} style={{ paddingBottom: '8px' }}>
            <Table
              columns={columns_event}
              loading={loading}
              dataSource={eventList}
              size="small"
              pagination={false}
            />
          </Card>

        </Col>
      </Row>

      <Row gutter={[16, 24]} style={{ marginTop: '15px' }}>
        <Col span={12}>
          <Card title={<span><PieChartOutlined />&nbsp;事件采集分布</span>} bordered={false} style={{ paddingBottom: '70px' }}>
            {dashData.eventPieChartData && (
              <PieChart data={dashData.eventPieChartData} loading={loading} height={330} />
            )}

          </Card>
        </Col>
        <Col span={12}>
          <Card title={<span><LineChartOutlined />&nbsp;1小时事件趋势</span>} bordered={false}>
            {dashData.eventLineChartData && (
              <LineChart data={dashData.eventLineChartData} unit="" />
            )}

          </Card>
        </Col>
      </Row>
    </div>
  );
};
