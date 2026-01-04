import { Card, Col, Row, Typography } from 'antd';
import React from 'react';
import ReactECharts from 'echarts-for-react';
import chartsData from '@/pages/monitor/event/chartsData';
import moment from 'moment';
import { InfoCircleOutlined } from '@ant-design/icons';

const { Text } = Typography;

const EventChart: React.FC<any> = (props: any) => {
  const { chartData, loading, eventInfo } = props;

  const chartHandle = (d: any) => {
    const series: { name: string; type: string; smooth: boolean; symbol: string; data: any }[] = [];
    const legend: string[] = [];
    const xAxis: { type: string; boundaryGap: boolean; data: string[] }[] = [];
    let unit: string = "";

    Object.keys(d).forEach((k: any) => {
      legend.push(`${k}`);

      const timeData: any[] = [];
      const valueData: any[] = [];

      d[k].forEach((s: any) => {
        //const timestamp = moment(s['time']).unix();
        const timestamp = moment(s['time']).format("YYYY-MM-DD HH:mm");
        //timeData.push(`${s['time']}`);
        timeData.push(timestamp);
        valueData.push(s['number']);
        unit = s["unit"]
      });


      xAxis.push({
        type: 'category',
        // @ts-ignore
        show: true,
        boundaryGap: false,
        position: 'bottom',
        nameTextStyle: {
          fontSize: 8,
          overflow: 'breakAll'
        },
        // 坐标轴名字旋转，角度值。
        axisLabel: {
          show: true,
          lineStyle: {
            show: true,
            color: "#696969"
          },
          textStyle: {
            color: '#696969'
          },
          height: 30,
          //interval: 5, //显示所有X轴信息
          //rotate: 35, //倾斜角度
          margin: 10, //刻度标签与轴线之间的距离
          padding: [0, 0, 8, 6], //表示 [上, 右, 下, 左] 的边距。
          hideOverlap: true, //隐藏重叠标签
          splitNumber: 30,
        },
        // @ts-ignore
        data: [...timeData],
      });
      const sItem = {
        name: `${k}`,
        type: 'line',
        smooth: true,
        symbol: 'none',
        // @ts-ignore
        // areaStyle: {},
        // @ts-ignore
        emphasis: {
          focus: 'series'
        },
        lineStyle: {
          width: 1
        },
        // stack: 'Total',
        // xAxisIndex: index > 0 ? 0 : 1,
        data: valueData,
      }
      series.push(sItem);
    });
    return chartsData.chartOptionFull(legend, xAxis, series, unit);
  };

  // eslint-disable-next-line @typescript-eslint/no-shadow
  const chartFunc = () => {
    return Object.keys(chartData).map((item: any) => {
      console.log("item -->", item, chartData[item])
      return (
        <>
          <Col span={12}>
            <Card size="small" title={item} extra={<Text type="secondary"><InfoCircleOutlined />  {(item && item.indexOf(":") >= 0) ? eventInfo(item.split(":")[0]) : eventInfo(item)}</Text>}>
              <ReactECharts
                key={item}
                lazyUpdate={false}
                notMerge={true}
                showLoading={loading}
                option={chartHandle(chartData[item])}
                shouldSetOption={() => true}
              />
            </Card>
          </Col>
        </>
      );
    });
  };

  return (
    <>
      <Row gutter={[16, 20]} style={{ background: '#ececec', paddingTop: 12, paddingBottom: 12 }}>
        {chartData !== undefined && chartFunc()}
      </Row>
    </>
  );
};
export default EventChart;
