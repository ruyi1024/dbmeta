import type { FC } from 'react';
import { Avatar, Card, Col, List, Alert, Row, Statistic, message, Tooltip } from 'antd';
import { InfoCircleOutlined, SmileTwoTone, PieChartOutlined, LineChartOutlined, ProfileTwoTone, SoundTwoTone } from '@ant-design/icons';
import React, { useEffect, useState } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { Link, useRequest } from 'umi';
import moment from 'moment';
import EditableLinkGroup from './components/EditableLinkGroup';
import styles from './style.less';
import type { ActivitiesType, CurrentUser } from './data.d';
//import { queryProjectNotice, queryActivities, fakeChartData } from './service';
import { ChartCard, MiniArea, MiniBar, MiniProgress } from './components/Charts';
//import Trend from './components/Trend';
import { Gauge } from '@ant-design/charts';
import PieChart from '@/components/Chart/PieChart';
import LineChart from '@/components/Chart/LineChart';




const wsAddr = `ws://${window.location.hostname}${window.location.port === '' ? '' : ':8088'
  }/api/v1/dashbaord/websocket`;


const links = [
  {
    title: 'SQL查询',
    href: '/execute/',
  },
  {
    title: '查询审计',
    href: '/execute/',
  },
  {
    title: '实例查询',
    href: '/meta/instance',
  },
  {
    title: '数据库查询',
    href: '/meta/instance',
  },
  {
    title: '数据表查询',
    href: '/meta/instance',
  },
  {
    title: '监控大盘',
    href: '/monitor/dashboard',
  },
  {
    title: '监控图表',
    href: '/monitor/event/',
  },
  {
    title: '告警信息',
    href: '/alarm/event',
  },
];


const PageHeaderContent: FC<{ currentUser: Partial<CurrentUser> }> = ({ currentUser }) => {
  const [currentUserinfo, setCurrentUserinfo] = useState({ "chineseName": "来宾", "username": "guest" });
  const [currentDate, setCurrentDate] = useState<string>("");

  useEffect(() => {
    const currentDate = moment().format("YYYY年MM月DD日");
    setCurrentDate(currentDate)
    //获取登录用户信息
    fetch('/api/v1/currentUser')
      .then((response) => response.json())
      .then((json) => {
        setCurrentUserinfo(json.data);
      })
      .catch((error) => {
        console.log('Fetch current userinfo failed', error);
      });
  });
  return (
    <>
      <div className={styles.pageHeaderContent}>
        <div className={styles.avatar}>
          <Avatar size="large" src="/avatar.jpg" />
        </div>
        <div className={styles.content}>
          <div className={styles.contentTitle}>
            欢迎您，{currentUserinfo.chineseName}，今天是{currentDate}，祝您工作开心！
          </div>
          <div>
            数据库是企业的核心数字资产，让Lepus和您一起来守护数据库的稳定和安全。
          </div>
        </div>
      </div>
    </>
  );
};

const ExtraContent: FC<Record<string, any>> = (data: any) => (
  <div className={styles.extraContent}>
    <div className={styles.statItem}>
      <Statistic title="数据源" value={data.datasourceCount ? data.datasourceCount : 0} />
    </div>
    <div className={styles.statItem}>
      <Statistic title="今日事件" value={data.todayEventCount ? data.todayEventCount : 0} />
    </div>
    <div className={styles.statItem}>
      <Statistic title="今日告警" value={data.todayAlarmCount ? data.todayAlarmCount : 0} />
    </div>
    <div className={styles.statItem}>
      <Statistic title="今日SQL查询" value={data.todaySqlQueryCount ? data.todaySqlQueryCount : 0} />
    </div>
    <div className={styles.statItem}>
      <Statistic title="今日SQL拦截" value={data.todaySqlQueryInterceptCount ? data.todaySqlQueryInterceptCount : 0} suffix="" />
    </div>
    <div className={styles.statItem}>
      <Statistic title="数据保护次数" value={data.sensitiveQueryCount ? data.sensitiveQueryCount : 0} />
    </div>
  </div>
);

const Workplace: FC = () => {

  const renderActivities = (item: ActivitiesType) => {
    const events = item.template.split(/@\{([^ {}] *)\}/gi).map((key) => {
      if (item[key]) {
        return (
          <a href={item[key].link} key={item[key].name}>
            {item[key].name}
          </a>
        );
      }
      return key;
    });
    return (
      <List.Item key={item.id}>
        <List.Item.Meta
          avatar={<Avatar src={item.user.avatar} />}
          title={
            <span>
              <a className={styles.username}>{item.user.name}</a>
              &nbsp;
              <span className={styles.event}>{events}</span>
            </span>
          }
          description={
            <span className={styles.datetime} title={item.updatedAt}>
              {moment(item.updatedAt).fromNow()}
            </span>
          }
        />
      </List.Item>
    );
  };

  const [dashData, setDashData] = useState<any>([]);
  const [wsState, setWsState] = useState<boolean>(false);
  const [wsData, setWsData] = useState<any>([]);
  const [seconds, setSeconds] = useState<number>(1);
  const [lastTime, setLastTime] = useState<any>(new Date());
  const [loading, setLoading] = useState<boolean>(true);
  //const [alarmPieData, setAlarmPieData] = useState<any>([{ type: 'noData', value: 1 }]);


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
    console.info(wsData)
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
      setPercent(data.healthPct);
    };
    WS.onopen = () => {
      setWsState(true);
      tick();
      // 启动定时器
      intervalId = setInterval(() => tick(), 3000);
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

  useEffect(() => {
    try {
      fetch(`/api/v1/dashbaord/info`)
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


  return (

    <PageContainer
      content={
        <PageHeaderContent
          currentUser={{
            avatar: 'https://gw.alipayobjects.com/zos/rmsportal/BiazfanxmamNRoxxVxka.png',
            name: '吴彦祖',
            userid: '00000001',
            email: 'antdesign@alipay.com',
            signature: '海纳百川，有容乃大',
            title: '交互专家',
            group: '蚂蚁金服－某某某事业群－某某平台部－某某技术部－UED',
          }}
        />
      }
      //extraContent={<ExtraContent data={dashData} />}
      extraContent={<><div className={styles.extraContent}>
        <div className={styles.statItem}>
          <Statistic title="今日监控事件" value={wsData.todayEventCount ? wsData.todayEventCount : 0} />
        </div>
        <div className={styles.statItem}>
          <Statistic title="今日事件告警" value={wsData.todayAlarmCount ? wsData.todayAlarmCount : 0} />
        </div>
        <div className={styles.statItem}>
          <Statistic title="今日SQL查询" value={wsData.todaySqlQueryCount ? wsData.todaySqlQueryCount : 0} suffix="" />
        </div>
        <div className={styles.statItem}>
          <Statistic title="今日SQL拦截" value={wsData.todaySqlQueryInterceptCount ? wsData.todaySqlQueryInterceptCount : 0} suffix="" />
        </div>
        <div className={styles.statItem}>
          <Statistic title="累计数据保护" value={wsData.sensitiveQueryCount ? wsData.sensitiveQueryCount : 0} suffix="" />
        </div>
        <div className={styles.statItem}>
          <Statistic title="数据源数量" value={wsData.datasourceCount ? wsData.datasourceCount : 0} suffix="" />
        </div>
      </div></>}
    >
      {!wsState && (
        <Alert type="error" message="WebSocket服务通信失败，请检查服务是否正常" banner />
      )}
      {wsState && (<Alert type={"success"} message={"WebSocket服务连接成功, 请求时间: " + (lastTime ? moment(lastTime).format('YYYY-MM-DD HH:mm:ss') : '-')} banner />)}
      {wsState && wsData.datasourceCount == 0 && (
        <Alert type="warning" message="检测到未配置数据源信息，请先添加配置数据源" banner closable />
      )}
      {wsState && wsData.currentEventCount == 0 && (
        <Alert type="warning" message="检测到15分钟内未产生新事件，请检查任务运行状态" banner closable />
      )}
      <Row gutter={24} style={{ marginTop: 8 }}>

        <Col xl={16} lg={24} md={24} sm={24} xs={24}>
          <Card
            className={styles.projectList}
            style={{ marginBottom: 8 }}
            title="实时数据库监控概览"
            bordered={false}
            extra={<Link to="/monitor/dashboard">进入监控面板</Link>}
            loading={loading}
            bodyStyle={{ padding: 0 }}
          >

            <Card.Grid className={styles.projectGrid} key="1">
              <ChartCard
                bordered={false}
                loading={false}
                title="15分钟事件数"
                action={
                  <Tooltip title="当前15分钟事件总数和每分钟趋势">
                    <InfoCircleOutlined />
                  </Tooltip>
                }
                total={wsData.currentEventCount ? wsData.currentEventCount : 0}
                footer={wsData.lastEventTime ? "最新事件时间：" + wsData.lastEventTime : ""}
                contentHeight={46}
              >
                <MiniArea color="#1979C9" data={wsData.eventMinuteData ? wsData.eventMinuteData : []} animate={true} />
              </ChartCard>
            </Card.Grid>
            <Card.Grid className={styles.projectGrid} key="1">
              <ChartCard
                bordered={false}
                loading={false}
                title="今日告警数"
                action={
                  <Tooltip title="今日告警总数和每分钟趋势">
                    <InfoCircleOutlined />
                  </Tooltip>
                }
                total={wsData.todayAlarmCount ? wsData.todayAlarmCount : 0}
                footer={wsData.lastAlarmTime ? "最新告警时间：" + wsData.lastAlarmTime : ""}
                contentHeight={46}
              >
                <MiniArea color="#1979C9" data={wsData.alarmMinuteData ? wsData.alarmMinuteData : []} />
              </ChartCard>
            </Card.Grid>
            <Card.Grid className={styles.projectGrid} key="3">
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
            </Card.Grid>
          </Card>

          <Card
            className={styles.projectList}
            style={{ marginBottom: 8 }}
            title="实时SQL查询安全概览"
            bordered={false}
            extra={<Link to="/safe/dashboard">进入安全面板</Link>}
            loading={loading}
            bodyStyle={{ padding: 0 }}
          >

            <Card.Grid className={styles.projectGrid} key="1">
              <ChartCard
                bordered={false}
                loading={false}
                title="今日SQL查询量"
                action={
                  <Tooltip title="今日SQL执行次数和趋势">
                    <InfoCircleOutlined />
                  </Tooltip>
                }
                total={wsData.todaySqlQueryCount ? wsData.todaySqlQueryCount : 0}
                footer={wsData.lastQueryTime ? "最新查询时间：" + wsData.lastQueryTime : ""}
                contentHeight={46}
              >
                <MiniArea color="#1979C9" data={wsData.queryNumberTodayData ? wsData.queryNumberTodayData : []} animate={true} />
              </ChartCard>
            </Card.Grid>
            <Card.Grid className={styles.projectGrid} key="1">
              <ChartCard
                bordered={false}
                loading={false}
                title="今日SQL拦截量"
                action={
                  <Tooltip title="今日SQL执行拦截次数和趋势">
                    <InfoCircleOutlined />
                  </Tooltip>
                }
                total={wsData.todaySqlQueryInterceptCount ? wsData.todaySqlQueryInterceptCount : 0}
                footer={wsData.lastQueryInterceptTime ? "最新拦截时间：" + wsData.lastQueryInterceptTime : ""}
                contentHeight={46}
              >
                <MiniBar color="#1979C9" data={wsData.queryInterceptTodayData ? wsData.queryInterceptTodayData : []} />
              </ChartCard>
            </Card.Grid>
            <Card.Grid className={styles.projectGrid} key="3">
              <ChartCard
                loading={loading}
                bordered={false}
                title="SQL执行拦截占比"
                action={
                  <Tooltip title="SQL执行拦截比例">
                    <InfoCircleOutlined />
                  </Tooltip>
                }
                total={wsData && wsData.queryInterceptPctStr}
                footer={
                  <div style={{ whiteSpace: 'nowrap', overflow: 'hidden' }}>
                    <>
                      拦截数：
                      <span className={styles.trendText}>{wsData && wsData.todaySqlQueryInterceptCount}</span>
                    </>
                    <>
                      &nbsp;&nbsp;&nbsp;&nbsp;查询数：
                      <span className={styles.trendText}>{wsData && wsData.todaySqlQueryCount}</span>
                    </>
                  </div>
                }
                contentHeight={46}
              >
                <MiniProgress percent={wsData && wsData.queryInterceptPct} strokeWidth={8} target={100} />
              </ChartCard>
            </Card.Grid>
          </Card>


          <Card title={<span>&nbsp;1小时事件趋势图表</span>} bordered={false} style={{ marginBottom: 8 }}>
            {dashData.eventLineChartData && (
              <LineChart data={dashData.eventLineChartData ? dashData.eventLineChartData : []} unit="" />
            )}
          </Card>
          <Card title={<span>&nbsp;事件采集分布图</span>} bordered={false} style={{ marginBottom: 8 }}>
            {dashData.eventPieChartData && (
              <PieChart data={dashData.eventPieChartData ? dashData.eventPieChartData : []} loading={loading} height={420} />
            )}
          </Card>





        </Col>
        <Col xl={8} lg={24} md={24} sm={24} xs={24}>
          <Card
            style={{ marginBottom: 12 }}
            title="快速开始 / 便捷导航"
            bordered={false}
            bodyStyle={{ padding: 0 }}
          >
            <EditableLinkGroup onAdd={() => { }} links={links} linkElement={Link} />
          </Card>
          <Card
            style={{ marginBottom: 10, paddingBottom: 8 }}
            bordered={false}
            title="数据源健康度"
            loading={loading}
          >
            <div className={styles.chart}>
              <Gauge
                {...config}
                loading={loading}
                chartRef={(chartRef) => {
                  // eslint-disable-next-line @typescript-eslint/no-unused-vars
                  ref = chartRef
                }}
                height={240}
              />
            </div>
          </Card>
          <Card
            bodyStyle={{ paddingTop: 12, paddingBottom: 12 }}
            bordered={false}
            title="数据源类型分布"
            loading={loading}
          >

            <PieChart data={dashData.datasourcePieDataList ? dashData.datasourcePieDataList : []} loading={false} height={330} />

          </Card>
        </Col>
      </Row>
    </PageContainer>
  );
};

export default Workplace;
