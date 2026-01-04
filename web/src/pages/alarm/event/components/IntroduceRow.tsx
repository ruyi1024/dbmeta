import { InfoCircleOutlined } from '@ant-design/icons';
import { Progress } from '@ant-design/charts';
import { Col, Row, Tooltip } from 'antd';

import numeral from 'numeral';
import { ChartCard, Field } from './Charts';

import Trend from './Trend';
import styles from '../style.less';

import React from "react";
import {MiniBar} from "@/components/Charts";

const topColResponsiveProps = {
  xs: 24,
  sm: 12,
  md: 12,
  lg: 12,
  xl: 6,
  style: { marginBottom: 24 },
};


// @ts-ignore
const IntroduceRow = ({ loading, analysisData }: { loading: boolean; analysisData: {
    alarmTodayData: { x: string | number; y: number; }[];
    alarmLast7DayData: { x: string | number; y: number; }[];
    alarmLastTime: any;
    alarmHourCount: number;
    alarmTodayCount: number;
    alarmCount: React.ReactElement<any, string | React.JSXElementConstructor<any>> | string | number | {} | Iterable<React.ReactNode> | React.ReactPortal | boolean | null | undefined | (() => (React.ReactNode | number));
  } }) => (
  <Row gutter={24}>
    <Col {...topColResponsiveProps}>
      <ChartCard
        bordered={false}
        title="告警总数"
        action={
          <Tooltip title="图表为近14日每日告警数量">
            <InfoCircleOutlined />
          </Tooltip>
        }
        loading={loading}
        total={() => analysisData.alarmCount}
        footer={<Field label="最近告警时间" value={analysisData.alarmLastTime} />}
        contentHeight={46}
      >
        <MiniBar color="#1979C9" data={analysisData.alarmLast7DayData} />
      </ChartCard>
    </Col>

    <Col {...topColResponsiveProps}>
      <ChartCard
        bordered={false}
        loading={loading}
        title="今日告警数"
        action={
          <Tooltip title="图表为今日小时告警数量">
            <InfoCircleOutlined />
          </Tooltip>
        }
        total={analysisData.alarmTodayCount}
        footer={<Field label="处理数据" value={"处理中：35，处理完成：10，完成率：85%"} />}
        contentHeight={46}
      >
        <MiniBar color="#1979C9" data={analysisData.alarmTodayData} />
      </ChartCard>
    </Col>
    <Col {...topColResponsiveProps}>
      <ChartCard
        bordered={false}
        loading={loading}
        title="近1小时告警量"
        action={
          <Tooltip title="指标说明">
            <InfoCircleOutlined />
          </Tooltip>
        }
        total={analysisData.alarmHourCount}
        footer={
          <div style={{ whiteSpace: 'nowrap', overflow: 'hidden' }}>
          <Trend flag="up" style={{ marginRight: 16 }}>
            环比增长
            <span className={styles.trendText}>12%</span>
          </Trend>
          <Trend flag="down">
          同比增长
          <span className={styles.trendText}>11%</span>
          </Trend>
          </div>
        }
        contentHeight={46}
      >
        <Trend flag="up" style={{ marginRight: 16 }}>
          周同比
          <span className={styles.trendText}>12%</span>
        </Trend>
        <Trend flag="down">
          日同比
          <span className={styles.trendText}>11%</span>
        </Trend>

      </ChartCard>
    </Col>
    <Col {...topColResponsiveProps}>
      <ChartCard
        loading={loading}
        bordered={false}
        title="告警处理完成率"
        action={
          <Tooltip title="指标说明">
            <InfoCircleOutlined />
          </Tooltip>
        }
        total="78%"
        footer={
          <div style={{ whiteSpace: 'nowrap', overflow: 'hidden' }}>
            <Trend flag="up" style={{ marginRight: 16 }}>
              处理中
              <span className={styles.trendText}>12%</span>
            </Trend>
            <Trend flag="down">
              未处理
              <span className={styles.trendText}>11%</span>
            </Trend>
          </div>
        }
        contentHeight={46}
      >
        <Progress
          height={46}
          percent={0.78}
          color="#13C2C2"
          forceFit
          size={8}
          marker={[
            {
              value: 0.8,
              style: {
                stroke: '#13C2C2',
              },
            },
          ]}
        />
      </ChartCard>
    </Col>
  </Row>
);

export default IntroduceRow;
