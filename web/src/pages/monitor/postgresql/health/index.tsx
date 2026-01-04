import { Tooltip, Card, Row, Col, Table, Badge, Progress, Menu } from 'antd';
import { DashboardOutlined, AreaChartOutlined, QuestionCircleOutlined } from '@ant-design/icons';
import React, { useState, useEffect } from 'react';
import { queryHealthList } from './service';
import { PgListData } from './data';
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

const formatKByte = (num: number) => {
  if (num >= 1000) {
    return (num / 1000).toFixed(1) + "MB"
  }
  else if (num >= 0) {
    return num + "KB"
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

const PgHealthList: React.FC = () => {
  const [list, setList] = useState<PgListData[]>([]);
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
      width: 150,
      render: (text: string, value: any) => {
        const nodes = value.host + ":" + value.port
        return nodes
      }


    },
    {
      title: '连接',
      dataIndex: 'connect',
      width: 60,
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
          width: 130,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
        {
          title: '运行',
          dataIndex: 'uptime',
          width: 75,
          render: (text: number, _) => {
            return checkValue(formatUptime(text))
          },
        },
        {
          title: '角色',
          dataIndex: 'role',
          width: 75,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
      ],
    },
    {
      title: '连接资源',
      children: [
        {
          title: '当前',
          dataIndex: 'connections',
          width: 60,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
        {
          title: '上限',
          dataIndex: 'max_connections',
          width: 60,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
      ],
    },
    {
      title: 'SQL会话',
      children: [
        {
          title: '活动',
          dataIndex: 'active_sql',
          width: 60,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
        {
          title: '等待',
          dataIndex: 'wait_event',
          width: 60,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
        {
          title: '锁',
          dataIndex: 'locks',
          width: 60,
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
          title: '准备',
          dataIndex: 'prepared_xacts',
          width: 60,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
        {
          title: '提交',
          dataIndex: 'xact_commit',
          width: 60,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
        {
          title: '回滚',
          dataIndex: 'xact_rollback',
          width: 60,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
      ],
    },

    {
      title: '性能统计',
      children: [
        {
          title: '索引读',
          dataIndex: 'tup_fetched',
          key: 'tup_fetched',
          width: 70,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
        {
          title: '全表读',
          dataIndex: 'tup_returned',
          key: 'tup_returned',
          width: 70,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
        {
          title: '写入',
          dataIndex: 'tup_inserted',
          key: 'tup_inserted',
          width: 60,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
        {
          title: '删除',
          dataIndex: 'tup_deleted',
          key: 'tup_deleted',
          width: 60,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
        {
          title: '更新',
          dataIndex: 'tup_updated',
          key: 'tup_updated',
          width: 60,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
      ],
    },
    {
      title: 'CheckPoint',
      children: [
        {
          title: 'ReqPct',
          dataIndex: 'checkpoint_req_pct',
          key: 'checkpoint_req_pct',
          width: 60,
          render: (text: number, _) => {
            return checkValue(formatNum(text)) + '%'
          },
        },
        {
          title: 'WritePct',
          dataIndex: 'checkpoint_write_pct',
          key: 'checkpoint_write_pct',
          width: 70,
          render: (text: number, _) => {
            return checkValue(formatNum(text)) + '%'
          },
        },
        {
          title: 'BackendPct',
          dataIndex: 'checkpoint_backend_write_pct',
          key: 'checkpoint_backend_write_pct',
          width: 70,
          render: (text: number, _) => {
            return checkValue(formatNum(text)) + '%'
          },
        },
        {
          title: 'AvgWrite',
          dataIndex: 'checkpoint_avg_write',
          key: 'checkpoint_avg_write',
          width: 70,
          render: (text: number, _) => {
            return formatByte(text)
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
          <a href="/performance/postgresql/health" rel="noopener noreferrer">
            PostgreSQL健康大盘
          </a>
        </Menu.Item>
        <Menu.Item key="chart" icon={<AreaChartOutlined />}>
          <a href="/performance/postgresql/chart" rel="noopener noreferrer">
            PostgreSQL性能图表
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
                    标签：{record.tag}
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
export default PgHealthList;
