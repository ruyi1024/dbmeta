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

const MySQLHealthList: React.FC = () => {
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
          dataIndex: 'redis_version',
          width: 60,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
        {
          title: '运行',
          dataIndex: 'uptime_in_seconds',
          width: 60,
          render: (text: number, _) => {
            return checkValue(formatUptime(text))
          },
        },
        {
          title: '模式',
          dataIndex: 'redis_mode',
          width: 60,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
        {
          title: '角色',
          dataIndex: 'role',
          width: 60,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
      ],
    },
    // {
    //   title: '资源使用率',
    //   children: [
    //     {
    //       title: '连接',
    //       dataIndex: 'max_connections',
    //       width: 80,
    //       render: (text: string, value: any) => {
    //         if(value.connect==0){
    //           return <Badge status={"default"} />
    //         }
    //         const pct = (value.threads_connected / value.max_connections).toFixed(2);
    //         const pct100 = pct * 100;
    //         if (pct100>80){
    //           return <Progress type="circle"  width={50} percent={pct100} status={"exception"} />;
    //         }else{
    //           return <Progress type="circle"  width={50} percent={pct100} />;
    //         }
    //       },
    //     },
    //     {
    //       title: '内存',
    //       dataIndex: 'open_files_limit',
    //       key: 'open_files_limit',
    //       width: 90,
    //       render: (text: string, value: any) => {
    //         if(value.connect==0){
    //           return <Badge status={"default"} />
    //         }
    //         const pct = (value.open_files / value.open_files_limit).toFixed(2);
    //         const pct100 = pct * 100;
    //         if (pct100>80){
    //           return <Progress type="circle"  width={50} percent={pct100} status={"exception"} />;
    //         }else{
    //           return <Progress type="circle"  width={50} percent={pct100} />;
    //         }
    //       },
    //     },
    //   ],
    // },
    {
      title: '线程/会话',
      children: [
        {
          title: '当前',
          dataIndex: 'connected_clients',
          key: 'connected_clients',
          width: 60,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
        {
          title: '阻塞',
          dataIndex: 'blocked_clients',
          key: 'blocked_clients',
          width: 60,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
        {
          title: '上限',
          dataIndex: 'max_connection',
          key: 'max_connection',
          width: 60,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
      ],
    },
    {
      title: '内存',
      children: [
        {
          title: '占用',
          dataIndex: 'used_memory_human',
          key: 'used_memory_human',
          width: 50,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
        {
          title: '峰值',
          dataIndex: 'used_memory_peak_human',
          key: 'used_memory_peak_human',
          width: 50,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
        {
          title: '系统',
          dataIndex: 'used_memory_rss_human',
          key: 'used_memory_rss_human',
          width: 50,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
        {
          title: '进程',
          dataIndex: 'used_memory_lua_human',
          key: 'used_memory_lua_human',
          width: 50,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
        {
          title: '分配器',
          dataIndex: 'mem_allocator',
          key: 'mem_allocator',
          width: 50,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
      ],
    },

    {
      title: '性能',
      children: [
        {
          title: 'OPS',
          dataIndex: 'instantaneous_ops_per_sec',
          key: 'instantaneous_ops_per_sec',
          width: 50,
          render: (text: number, _) => {
            return checkValue(formatNum(text))
          },
        },
      ],
    },
    // {
    //   title: '操作',
    //   dataIndex: 'tag',
    //   sorter: false,
    //   key: 'id',
    //   fixed: 'right',
    //   width: 75,
    //   render: (text: number, record: UserListItem) => (
    //     <>
    //       <Dropdown
    //         overlay={
    //           <Menu>
    //             <Menu.Item>
    //               <Tooltip title={`查看性能图表【${record.tag}】`}>
    //                 <a
    //                   onClick={() => {
    //                     history.push('/monitor/mysql/chart/219.234.1.11/3306');
    //                   }}
    //                 >
    //                   性能图表
    //                 </a>
    //               </Tooltip>
    //             </Menu.Item>
    //           </Menu>
    //         }
    //       >
    //         <a className="ant-dropdown-link" onClick={(e) => e.preventDefault()}>
    //           操作 <DownOutlined />
    //         </a>
    //       </Dropdown>
    //     </>
    //   ),
    // },
  ];

  useEffect(() => {
    did('');
  }, []);

  // @ts-ignore
  return (
    <>
      <Menu mode="horizontal" selectedKeys={[current]} >
        <Menu.Item key="dashboard" icon={<DashboardOutlined />}>
          <a href="#/performance/redis/health" rel="noopener noreferrer">
            Redis健康大盘
          </a>
        </Menu.Item>
        <Menu.Item key="chart" icon={<AreaChartOutlined />}>
          <a href="#/performance/redis/chart" rel="noopener noreferrer">
            Redis性能图表
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
export default MySQLHealthList;
