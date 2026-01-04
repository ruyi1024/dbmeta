import { Tooltip, Card, Row, Col, Table, Badge, Progress, Menu } from 'antd';
import { DashboardOutlined, AreaChartOutlined, QuestionCircleOutlined } from '@ant-design/icons';
import React, { useState, useEffect } from 'react';
import { queryHealthList } from './service';
import { MongodbListData } from './data';



const query = async (params: string) => {
  try {
    return await queryHealthList(params);
  } catch (e) {
    return { success: false, msg: e };
  }
};


const checkValue = (num: any) => {
  if (num == -1 || num == '-1') {
    return <Badge status={"default"} />
  }
  return num
}

const formatNum = (num: number) => {
  if (num >= 10000) {
    return (num / 10000).toFixed(1) + "W"
  }
  else if (num >= 1000) {
    return (num / 1000).toFixed(1) + "K"
  }
  return num
}

const formatByte = (num: number) => {
  if (num >= 1000000) {
    return (num / 1000000).toFixed(1) + "Mb"
  }
  else if (num >= 1000) {
    return (num / 1000).toFixed(1) + "Kb"
  }
  else if (num >= 0) {
    return num + "b"
  }
  return num
}

const formatMB = (num: number) => {
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + "GB"
  }
  else if (num >= 0) {
    return num + "MB"
  }
  return num
}

const formatUptime = (num: number) => {
  if (num >= 86400) {
    return (num / 86400).toFixed(1) + "天"
  }
  else if (num >= 3600) {
    return (num / 3600).toFixed(1) + "小时"
  }
  else if (num >= 60) {
    return (num / 60).toFixed(1) + "分钟"
  }
  else if (num >= 0) {
    return num + "秒"
  }
  return num
}

const MongodbHealthList: React.FC = () => {
  const [list, setList] = useState<MongodbListData[]>([]);
  const [total, setTotal] = useState<number>(0);
  const [loading, setLoading] = useState<boolean>(false);
  const [current, setCurrent] = useState<string>("");


  const did = (params: string) => {
    setLoading(true);
    setCurrent("dashboard");
    query(params).then((res) => {
      if (res.success) {
        setList(res.data);
        setTotal(res.total);
      }
      setLoading(false);
    });
  };

  const columns = [
    {
      title: '实例节点',
      dataIndex: 'tag',
      sorter: true,
      width: 140,
      render: (text: string, value: any) => {
        const nodes = value.host + ":" + value.port
        return nodes
      }


    },
    {
      title: '连接',
      dataIndex: 'connect',
      width: 70,
      render: (text: string, value: any) => {
        if (text == '1') {
          return <Badge status={"success"} />
        } else {
          return <Tooltip title={value.error_info} ><Badge status={"error"} /><QuestionCircleOutlined /></Tooltip>
        }
      }
    },

    {
      title: '基本信息',
      children: [
        {
          title: '版本',
          dataIndex: 'version',
          width: 100,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
        {
          title: '运行',
          dataIndex: 'uptime',
          width: 80,
          render: (text: number, _) => {
            return checkValue(formatUptime(text))
          },
        },
        // {
        //   title: '角色',
        //   dataIndex: 'role',
        //   width: 60,
        //   render: (text: number,_) => {
        //     return checkValue(text)
        //   },
        // },
        // {
        //   title: '只读',
        //   dataIndex: 'readonly',
        //   width: 60,
        //   render: (text: number,_) => {
        //     return checkValue(text)
        //   },
        // },
      ],
    },

    {
      title: '连接数',
      children: [
        {
          title: '当前',
          dataIndex: 'connections_current',
          width: 80,

        },
        {
          title: '可用',
          dataIndex: 'connections_available',
          width: 80,
        },
        {
          title: '使用率',
          dataIndex: 'connections_available',
          width: 80,
          render: (text: string, value: any) => {
            if (value.connect == 0) {
              return <Badge status={"default"} />
            }
            const pct = (value.connections_current / (value.connections_current + value.connections_available)).toFixed(1);
            const pct100 = pct * 100;
            return pct100 + "%";
          },
        },
      ],
    },

    {
      title: '内存',
      children: [
        {
          title: '位数',
          dataIndex: 'mem_bits',
          key: 'mem_bits',
          width: 50,
        },
        {
          title: '物理',
          dataIndex: 'mem_resident',
          key: 'mem_resident',
          width: 50,
          render: (text: number, _: any) => {
            return checkValue(formatMB(text))
          },
        },
        {
          title: '虚拟',
          dataIndex: 'mem_virtual',
          key: 'mem_virtual',
          width: 50,
          render: (text: number, _: any) => {
            return checkValue(formatMB(text))
          },
        },
        {
          title: '映射',
          dataIndex: 'mem_mapped',
          key: 'mem_mapped',
          width: 50,
          render: (text: number, _: any) => {
            return checkValue(formatMB(text))
          },
        },
      ],
    },

    {
      title: '性能',
      children: [
        {
          title: '查询',
          dataIndex: 'opcounters_query',
          key: 'opcounters_query',
          width: 50,
          render: (text: number, _) => {
            return checkValue(formatNum(text))
          },
        },
        {
          title: '写入',
          dataIndex: 'opcounters_insert',
          key: 'opcounters_insert',
          width: 50,
          render: (text: number, _) => {
            return checkValue(formatNum(text))
          },
        },
        {
          title: '更新',
          dataIndex: 'opcounters_update',
          key: 'opcounters_update',
          width: 50,
          render: (text: number, _) => {
            return checkValue(formatNum(text))
          },
        },
        {
          title: '删除',
          dataIndex: 'opcounters_delete',
          key: 'opcounters_delete',
          width: 50,
          render: (text: number, _) => {
            return checkValue(formatNum(text))
          },
        },
        {
          title: '命令',
          dataIndex: 'opcounters_command',
          key: 'opcounters_command',
          width: 50,
          render: (text: number, _) => {
            return checkValue(formatNum(text))
          },
        },
      ],
    },

    {
      title: '网络',
      children: [
        {
          title: '接收',
          dataIndex: 'network_bytesIn',
          key: 'network_bytesIn',
          width: 50,
          render: (text: number, _: any) => {
            return checkValue(formatByte(text))
          },
        },
        {
          title: '发送',
          dataIndex: 'network_bytesOut',
          key: 'network_bytesOut',
          width: 50,
          render: (text: number, _: any) => {
            return checkValue(formatByte(text))
          },
        },
        {
          title: '请求',
          dataIndex: 'network_numRequests',
          key: 'network_numRequests',
          width: 50,
          render: (text: number, _: any) => {
            return checkValue(formatNum(text))
          },
        },
      ],
    },
  ];



  useEffect(() => {
    did('');
  }, []);

  // @ts-ignore
  return (
    <>
      <Menu mode="horizontal" selectedKeys={[current]} >
        <Menu.Item key="dashboard" icon={<DashboardOutlined />}>
          <a href="/performance/mongodb/health" rel="noopener noreferrer">
            MongoDB健康大盘
          </a>
        </Menu.Item>
        <Menu.Item key="chart" icon={<AreaChartOutlined />}>
          <a href="/performance/mongodb/chart" rel="noopener noreferrer">
            MongoDB性能图表
          </a>
        </Menu.Item>
      </Menu>


      <Card bodyStyle={{ padding: 15 }}>
        <Row style={{ paddingTop: 10 }}>
          <Col span={24}>
            <Table
              size="small"
              rowKey="id"
              bordered
              loading={loading}
              columns={columns}
              dataSource={list}
              pagination={false}
              expandable={{
                expandedRowRender: (record) => (
                  <p style={{ margin: 0 }}>
                    标签：{record.tag}，主机名：{record.hostname}
                  </p>
                ),
                rowExpandable: (record) => record.name !== 'Not Expandable',
              }}
            />
          </Col>
        </Row>
      </Card>
    </>
  );
};
export default MongodbHealthList;
