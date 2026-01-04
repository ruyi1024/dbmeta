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
        {
          title: '角色',
          dataIndex: 'role',
          width: 60,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
        {
          title: '只读',
          dataIndex: 'readonly',
          width: 60,
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
          dataIndex: 'threads_connected',
          width: 80,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
        {
          title: '上限',
          dataIndex: 'max_connections',
          width: 80,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
      ],
    },
    {
      title: '文件句柄',
      children: [
        {
          title: '当前',
          dataIndex: 'open_files',
          width: 80,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
        {
          title: '上限',
          dataIndex: 'open_files_limit',
          width: 80,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
      ],
    },
    {
      title: '表缓存',
      children: [
        {
          title: '当前',
          dataIndex: 'open_tables',
          width: 80,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
        {
          title: '上限',
          dataIndex: 'table_open_cache',
          width: 80,
          render: (text: number, _: any) => {
            return checkValue(text)
          },
        },
      ],
    },
    // {
    //   title: '复制',
    //   children: [
    //     {
    //       title: '状态',
    //       dataIndex: 'repl_status',
    //       key: 'repl_status',
    //       width: 50,
    //       render: (text: number,_) => {
    //         if(text==1){
    //           return <Badge status={"success"} />
    //         } else if(text==0){
    //           return <Badge status={"error"} />
    //         }else{
    //           return <Badge status={"default"} />
    //         }
    //       },
    //     },
    //     {
    //       title: '延迟',
    //       dataIndex: 'repl_delay',
    //       key: 'repl_delay',
    //       width: 50,
    //       render: (text: number,_) => {
    //         if(text==1){
    //           return <Badge status={"success"} />
    //         } else if(text==0){
    //           return <Badge status={"error"} />
    //         }else{
    //           return <Badge status={"default"} />
    //         }
    //       },
    //     },
    //   ],
    // },
    // {
    //   title: '资源',
    //   children: [
    //     {
    //       title: '连接池',
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
    //       title: '文件数',
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
    //     {
    //       title: '表缓存',
    //       dataIndex: 'table_open_cache',
    //       key: 'table_open_cache',
    //       width: 90,
    //       render: (text: string, value: any) => {
    //         if(value.connect==0){
    //           return <Badge status={"default"} />
    //         }
    //         const pct = (value.open_tables / value.table_open_cache).toFixed(1);
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
          title: '已连接',
          dataIndex: 'threads_connected',
          key: 'threads_connected',
          width: 60,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
        {
          title: '活动中',
          dataIndex: 'threads_running',
          key: 'threads_running',
          width: 60,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
        {
          title: '阻塞中',
          dataIndex: 'threads_wait',
          key: 'threads_wait',
          width: 60,
          render: (text: number, _) => {
            return checkValue(text)
          },
        },
      ],
    },
    {
      title: '性能',
      children: [
        {
          title: 'QPS',
          dataIndex: 'queries',
          key: 'queries',
          width: 50,
          render: (text: number, _) => {
            return checkValue(formatNum(text))
          },
        },
        {
          title: '查询',
          dataIndex: 'com_select',
          key: 'com_select',
          width: 50,
          render: (text: number, _) => {
            return checkValue(formatNum(text))
          },
        },
        {
          title: '写入',
          dataIndex: 'com_insert',
          key: 'com_insert',
          width: 50,
          render: (text: number, _) => {
            return checkValue(formatNum(text))
          },
        },
        {
          title: '更新',
          dataIndex: 'com_update',
          key: 'com_update',
          width: 50,
          render: (text: number, _) => {
            return checkValue(formatNum(text))
          },
        },
        {
          title: '删除',
          dataIndex: 'com_delete',
          key: 'com_delete',
          width: 50,
          render: (text: number, _) => {
            return checkValue(formatNum(text))
          },
        },
        {
          title: '提交',
          dataIndex: 'com_commit',
          key: 'com_commit',
          width: 50,
          render: (text: number, _) => {
            return checkValue(formatNum(text))
          },
        },
        {
          title: '回滚',
          dataIndex: 'com_rollback',
          key: 'com_rollback',
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
          dataIndex: 'bytes_received',
          key: 'bytes_received',
          width: 50,
          render: (text: number, _) => {
            return checkValue(formatByte(text))
          },
        },
        {
          title: '发送',
          dataIndex: 'bytes_sent',
          key: 'bytes_sent',
          width: 50,
          render: (text: number, _) => {
            return checkValue(formatByte(text))
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
    //                     history.push('/monitor/mysql/chart');
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
    //           更多 <DownOutlined />
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
          <a href="/monitor/greatsql/health" rel="noopener noreferrer">
            GreatSQL健康大盘
          </a>
        </Menu.Item>
        <Menu.Item key="chart" icon={<AreaChartOutlined />}>
          <a href="/monitor/greatsql/chart" rel="noopener noreferrer">
            GreatSQL性能图表
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
export default MySQLHealthList;
