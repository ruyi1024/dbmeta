// eslint-disable-next-line @typescript-eslint/no-unused-vars
const chartOptionFull = (legend: [], xAxisData: [], series: any, unit: string) => {
  return {
    grid: {
      top: 60,
      bottom: 75,
      right: 20,
      left: 70,
    },
    legend: {
      type: 'scroll'
    },
    // toolbox: {
    //   top: 5,
    //   feature: {
    //     dataZoom: {
    //       yAxisIndex: 'none',
    //     },
    //     restore: {},
    //     saveAsImage: {},
    //   },
    // },
    xAxis: xAxisData,
    yAxis: [
      {
        type: 'value',
        show: true,
        axisLabel: {
          formatter:'{value} '+unit,
          color: "#696969",
        },
        lineStyle: {
          show: true
        },
        axisLine: {
          show: true,
          lineStyle: {
            color: "#696969"
          },
          textStyle: {
            color: '#696969'
          },
        },
        axisTick: {
          show:true
        },
        splitLine: {
          show: true
        },
      },

    ],
    tooltip: {
      trigger: 'axis',
    },
    dataZoom: [
      {
        show: true,
        realtime: true,
        start: 0,
        end: 100,
        xAxisIndex: [0, 1]
      },
      {
        type: 'inside',
        realtime: true,
        start: 0,
        end: 100,
        xAxisIndex: [0, 1]
      }
    ],
    series: series,
  };
};

const chartsData: any = {
  chartOptionFull,
};

export default chartsData;
