import { Tooltip, Card, Row, Col, Table, Badge, Progress, Menu, Dropdown } from 'antd';
import { DashboardOutlined, AreaChartOutlined, QuestionCircleOutlined } from '@ant-design/icons';
import React, { useState, useEffect } from 'react';
import { PageContainer } from '@ant-design/pro-layout';
import { queryHealthList } from './service';
import { MysqlListData } from './data';
import { UserListItem } from '@/pages/UserManager/data';
import { history } from 'umi';

// function formatDecimal(num, decimal) {
//   num = num.toString();
//   var index = num.indexOf('.');
//   if (index !== -1) {
//     num = num.substring(0, decimal + index + 1)
//   } else {
//     num = num.substring(0)
//   }
//   return parseFloat(num).toFixed(decimal)
// }

// const {Search} = Input;
// const handleSearchKeyword = async (val: string) => {
//   console.log('search val:', val);
//   try {
//     return await queryHealthList({'keyword':val});
//   } catch (e) {
//     return {success: false, msg: e}
//   }
//
// };

const query = async (params: string) => {
  try {
    return await queryHealthList(params);
  } catch (e) {
    return { success: false, msg: e };
  }
};

const goChart = async (params: string) => {
  history.push('/menu1/menu11/menu112/create');
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

const MsSQLHealthList: React.FC = () => {
  const [list, setList] = useState<MysqlListData[]>([]);
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
          width: 60,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
        {
          title: '运行',
          dataIndex: 'uptime',
          width: 60,
          render: (text: number, _) => {
            return checkValue(formatUptime(text))
          },
        },
      ],
    },
    {
      title: '进程/会话',
      children: [
        {
          title: '当前',
          dataIndex: 'processes',
          key: 'processes',
          width: 60,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
        {
          title: '运行',
          dataIndex: 'processes_running',
          key: 'processes_running',
          width: 60,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
        {
          title: '阻塞',
          dataIndex: 'processes_waits',
          key: 'processes_waits',
          width: 60,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
        {
          title: '上限',
          dataIndex: 'max_connections',
          key: 'max_connections',
          width: 60,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
      ],
    },
    {
      title: '当前请求',
      children: [
        {
          title: '读',
          dataIndex: 'current_read',
          key: 'current_read',
          width: 50,
          render: (text: number) => {
            return checkValue(text)
          },
        },
        {
          title: '写',
          dataIndex: 'current_write',
          key: 'current_write',
          width: 50,
          render: (text: number) => {
            return checkValue(text)
          },
        },
        {
          title: '错误',
          dataIndex: 'current_error',
          key: 'current_error',
          width: 50,
          render: (text: number) => {
            return checkValue(text)
          },
        },
      ],
    },
    {
      title: '数据包',
      children: [
        {
          title: '接收',
          dataIndex: 'pack_received',
          key: 'pack_received',
          width: 50,
          render: (text: number) => {
            return checkValue(text)
          },
        },
        {
          title: '发送',
          dataIndex: 'pack_sent',
          key: 'pack_sent',
          width: 50,
          render: (text: number) => {
            return checkValue(text)
          },
        },
        {
          title: '错误',
          dataIndex: 'packet_errors',
          key: 'packet_errors',
          width: 50,
          render: (text: number) => {
            return checkValue(text)
          },
        },
      ],
    },

    {
      title: '性能',
      children: [
        {
          title: '事务数',
          dataIndex: 'trancount',
          key: 'trancount',
          width: 60,
          render: (text: number) => {
            return checkValue(text)
          },
        },
        {
          title: '记录数',
          dataIndex: 'row_count',
          key: 'row_count',
          width: 60,
          render: (text: number) => {
            return checkValue(text)
          },
        },
        {
          title: 'CpuBusy',
          dataIndex: 'cpu_busy',
          key: 'cpu_busy',
          width: 70,
          render: (text: number) => {
            return checkValue(text)
          },
        },
        {
          title: 'IoBusy',
          dataIndex: 'io_busy',
          key: 'io_busy',
          width: 70,
          render: (text: number) => {
            return checkValue(text)
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
          <a href="/performance/sqlserver/health" rel="noopener noreferrer">
            SQLServer健康大盘
          </a>
        </Menu.Item>
        <Menu.Item key="chart" icon={<AreaChartOutlined />}>
          <a href="/performance/sqlserver/chart" rel="noopener noreferrer">
            SQLServer性能图表
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
export default MsSQLHealthList;
