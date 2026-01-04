import { Tooltip, Card, Row, Col, Table, Badge, Progress, Menu } from 'antd';
import { DashboardOutlined, AreaChartOutlined, QuestionCircleOutlined } from '@ant-design/icons';
import React, { useState, useEffect } from 'react';
import { queryHealthList } from './service';
import { MysqlListData } from './data';
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

const OracleHealthList: React.FC = () => {
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
      width: 130,
      render: (text: string, value: any) => {
        const nodes = value.host + ":" + value.port + "/" + value.sid
        return nodes
      }


    },
    {
      title: '连接',
      dataIndex: 'connect',
      width: 50,
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
          width: 80,
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
        {
          title: '角色',
          dataIndex: 'database_role',
          width: 85,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
        {
          title: '状态',
          dataIndex: 'instance_status',
          width: 85,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
        {
          title: '模式',
          dataIndex: 'open_mode',
          width: 100,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
      ],
    },
    {
      title: '线程/会话',
      children: [
        {
          title: '当前',
          dataIndex: 'session_total',
          width: 50,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
        {
          title: '活动',
          dataIndex: 'session_active',
          width: 50,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
        {
          title: '等待',
          dataIndex: 'session_wait',
          width: 50,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
      ],
    },
    {
      title: '事务',
      children: [

        {
          title: '提交',
          dataIndex: 'user_commits_persecond',
          key: 'user_commits_persecond',
          width: 50,
          render: (text: number, _) => {
            return checkValue(formatNum(text))
          },
        },
        {
          title: '回滚',
          dataIndex: 'user_rollbacks_persecond',
          key: 'user_rollbacks_persecond',
          width: 50,
          render: (text: number, _) => {
            return checkValue(formatNum(text))
          },
        }
      ]
    },
    {
      title: '物理读写',
      children: [
        {
          title: '物理读',
          dataIndex: 'physical_read_persecond',
          key: 'physical_read_persecond',
          width: 60,
          render: (text: number, _) => {
            return checkValue(formatNum(text))
          },
        },
        {
          title: '物理写',
          dataIndex: 'physical_write_persecond',
          key: 'physical_write_persecond',
          width: 60,
          render: (text: number, _) => {
            return checkValue(formatNum(text))
          },
        }
      ]
    },
    {
      title: 'IO请求数',
      children: [

        {
          title: '读请求',
          dataIndex: 'physical_read_io_request_persecond',
          key: 'physical_read_io_request_persecond',
          width: 60,
          render: (text: number, _) => {
            return checkValue(formatNum(text))
          },
        },
        {
          title: '写请求',
          dataIndex: 'physical_write_io_request_persecond',
          key: 'physical_write_io_request_persecond',
          width: 60,
          render: (text: number, _) => {
            return checkValue(formatNum(text))
          },
        }
      ]
    },
    {
      title: '性能',
      children: [

        {
          title: '块变化',
          dataIndex: 'db_block_changes_persecond',
          key: 'db_block_changes_persecond',
          width: 60,
          render: (text: number, _) => {
            return checkValue(formatNum(text))
          },
        },
        {
          title: 'CPU等待',
          dataIndex: 'os_cpu_wait_time',
          key: 'os_cpu_wait_time',
          width: 70,
          render: (text: number, _) => {
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
          <a href="/performance/oracle/health" rel="noopener noreferrer">
            Oracle健康大盘
          </a>
        </Menu.Item>
        <Menu.Item key="chart" icon={<AreaChartOutlined />}>
          <a href="/performance/oracle/chart" rel="noopener noreferrer">
            Oracle性能图表
          </a>
        </Menu.Item>
      </Menu>


      <Card bodyStyle={{ padding: 15 }}>



        {/*<Alert*/}
        {/*  message={"MySQL监控最新数据上报时间：2021-02-18 22:20:30"}*/}
        {/*  type="success"*/}
        {/*  showIcon*/}
        {/*  banner*/}
        {/*  style={{*/}
        {/*    margin: -12,*/}
        {/*    marginBottom: 24,*/}
        {/*  }}*/}
        {/*/>*/}
        {/*<Row>*/}
        {/*  <Col flex="auto">*/}
        {/*    <Search*/}
        {/*      placeholder="IP or Hostname"*/}
        {/*      onSearch={handleSearchKeyword}*/}
        {/*      style={{width: 280}}*/}
        {/*    />*/}
        {/*    <Tooltip placement="top" title="重载并刷新表格数据">*/}
        {/*      <Button*/}
        {/*        type="link"*/}
        {/*        icon={<ReloadOutlined/>}*/}
        {/*        onClick={() => did('')}*/}
        {/*      />*/}
        {/*    </Tooltip>*/}
        {/*  </Col>*/}
        {/*</Row>*/}
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
export default OracleHealthList;
