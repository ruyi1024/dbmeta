import { Card, Row, Col, Table, Badge, Progress, Menu } from 'antd';
import { PageContainer } from '@ant-design/pro-components';
import React, { useState, useEffect } from 'react';
import { queryHealthList } from './service';
import { MysqlListData } from './data';

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
    return (num / 1000000).toFixed(0) + "Mb"
  }
  else if (num >= 1000) {
    return (num / 1000).toFixed(0) + "Kb"
  }
  else if (num >= 0) {
    return num + "b"
  }
  return num
}

const formatMB = (num: number) => {
  if (num >= 1024) {
    return (num / 1024).toFixed(1) + "GB"
  }
  else if (num >= 0) {
    return num + "MB"
  }
  return num
}

const formatUptime = (num: number) => {
  if (num >= 86400) {
    return (num / 86400).toFixed(0) + "天"
  }
  else if (num >= 3600) {
    return (num / 3600).toFixed(0) + "小时"
  }
  else if (num >= 60) {
    return (num / 60).toFixed(0) + "分钟"
  }
  else if (num >= 0) {
    return num + "秒"
  }
  return num
}

const ServerHealthList: React.FC = () => {
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
      title: 'IP',
      dataIndex: 'ip',
      sorter: true,
      width: 100,
    },
    {
      title: '基本信息',
      children: [
        {
          title: '系统',
          dataIndex: 'os',
          width: 100,
          render: (text: number) => {
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
      ],
    },
    {
      title: 'CPU',
      children: [
        {
          title: '物理',
          dataIndex: 'cpu_physical_num',
          width: 65,
          render: (text: number) => {
            return text + "c"
          },
        },
        {
          title: '逻辑',
          dataIndex: 'cpu_logical_num',
          width: 65,
          render: (text: number) => {
            return text + "c"
          },
        },
        {
          title: '负载',
          dataIndex: 'cpu_load',
          width: 65,
          render: (text: number) => {
            return checkValue(text)
          },
        },
        {
          title: '使用率',
          dataIndex: 'cpu_percent',
          width: 80,
          render: (value: bigint) => {
            if (value > 50) {
              return <Progress type="circle" width={40} percent={value} status={"exception"} format={percent => `${percent}%`} />;
            } else {
              return <Progress type="circle" width={40} percent={value} status={"success"} format={percent => `${percent}%`} />;
            }
          },
        }
      ],
    },
    {
      title: '内存',
      children: [
        {
          title: '总计',
          dataIndex: 'memory_total',
          width: 60,
          render: (text: number) => {
            return checkValue(formatMB(text))
          },
        },
        {
          title: '已用',
          dataIndex: 'memory_used',
          width: 60,
          render: (text: number) => {
            return checkValue(formatMB(text))
          },
        },
        {
          title: '可用',
          dataIndex: 'memory_available',
          width: 60,
          render: (text: number) => {
            return checkValue(formatMB(text))
          },
        },
        {
          title: '使用率',
          dataIndex: 'memory_used_percent',
          width: 80,
          render: (value: bigint) => {
            if (value > 80) {
              return <Progress type="circle" width={40} percent={value} status={"exception"} format={percent => `${percent}%`} />;
            } else {
              return <Progress type="circle" width={40} percent={value} status={"success"} format={percent => `${percent}%`} />;
            }
          },
        }
      ],
    },

    {
      title: 'SWAP',
      children: [
        {
          title: '总计',
          dataIndex: 'swap_total',
          width: 60,
          render: (text: number) => {
            return checkValue(formatMB(text))
          },
        },
        {
          title: '缓存',
          dataIndex: 'swap_cached',
          width: 60,
          render: (text: number) => {
            return checkValue(formatMB(text))
          },
        },
        {
          title: '空闲',
          dataIndex: 'swap_free',
          width: 60,
          render: (text: number) => {
            return checkValue(formatMB(text))
          },
        },
      ],
    },

    {
      title: '硬盘',
      children: [
        {
          title: '已用',
          dataIndex: 'disk_used_percent',
          width: 80,
          render: (value: bigint) => {
            if (value > 80) {
              return <Progress type="circle" width={40} percent={value} status={"exception"} format={percent => `${percent}%`} />;
            } else {
              return <Progress type="circle" width={40} percent={value} status={"success"} format={percent => `${percent}%`} />;
            }
          },
        },
        {
          title: 'Inode',
          dataIndex: 'inodes_used_percent',
          width: 80,
          render: (value: bigint) => {
            if (value > 80) {
              return <Progress type="circle" width={40} percent={value} status={"exception"} format={percent => `${percent}%`} />;
            } else {
              return <Progress type="circle" width={40} percent={value} status={"success"} format={percent => `${percent}%`} />;
            }
          },

        },

      ],
    },

    {
      title: 'IO性能',
      children: [
        {
          title: '读取',
          dataIndex: 'diskio_read_bytes',
          width: 80,
          render: (text: number) => {
            return checkValue(formatByte(text))
          },
        },
        {
          title: '写入',
          dataIndex: 'diskio_write_bytes',
          width: 80,
          render: (text: number) => {
            return checkValue(formatByte(text))
          },
        },
        {
          title: '读延迟',
          dataIndex: 'diskio_read_time',
          width: 80,
          render: (text: number) => {
            return checkValue(text)
          },
        },
        {
          title: '写延迟',
          dataIndex: 'diskio_write_time',
          width: 80,
          render: (text: number) => {
            return checkValue(text)
          },
        },
      ],
    },

    {
      title: '网卡',
      children: [
        {
          title: '进入',
          dataIndex: 'network_bytes_recv',
          key: 'network_bytes_recv',
          width: 80,
          render: (text: number, _) => {
            return checkValue(formatByte(text))
          },
        },
        {
          title: '流出',
          dataIndex: 'network_bytes_sent',
          key: 'network_bytes_sent',
          width: 80,
          render: (text: number, _) => {
            return checkValue(formatByte(text))
          },
        },
        {
          title: '入包',
          dataIndex: 'network_packets_recv',
          width: 70,
          render: (text: number) => {
            return checkValue(formatNum(text))
          },
        },
        {
          title: '出包',
          dataIndex: 'network_packets_sent',
          width: 70,
          render: (text: number) => {
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
    <PageContainer content="">
      {/*<TabNav />*/}
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
                    标签：{record.tag}，操作系统：{record.platform} {record.version}，主机名：{record.hostname}，启动时间：{record.boot_time}
                  </p>
                ),
                rowExpandable: (record) => record.name !== 'Not Expandable',
              }}
            />
          </Col>
        </Row>
      </Card>
    </PageContainer>
  );
};
export default ServerHealthList;
