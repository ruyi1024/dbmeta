import { PageContainer } from '@ant-design/pro-components';
import { Card, Tabs, Form, Select, Button, Input, message, Table, Alert, Space,Radio, Transfer, Checkbox  } from 'antd';
import React, { useState, useEffect } from 'react';
import type { TransferDirection } from 'antd/lib/transfer';
import TextArea from 'antd/lib/input/TextArea';

const checkBoxOptionsDatabase=[
  {label:'查询数据',value:'select'},
  {label:'写入数据',value:'insert'},
  {label:'更新数据',value:'update'},
  {label:'删除数据',value:'delete'},
  {label:'创建结构',value:'create'},
  {label:'修改结构',value:'alter'},
]

const checkBoxOptionsTable=[
  {label:'查询数据',value:'select'},
  {label:'写入数据',value:'insert'},
  {label:'更新数据',value:'update'},
  {label:'删除数据',value:'delete'},
]

const maxQueryNumberOptions=[
  {value:'100',label:'100'},
  {value:'300',label:'300'},
  {value:'500',label:'500'},
  {value:'1000',label:'1000'},
  {value:'5000',label:'5000'},
  {value:'10000',label:'10000'},
]

const Index: React.FC = () => {
  const [form] = Form.useForm();

  const [loading,setLoading] = useState<boolean>(false);
  const [userlist, setUserlist] = useState([]);
  const [typeList, setTypeList] = useState<any[]>([{ id: 0, cluster_name: '' }]);
  const [datasourceList, setDatasourceList] = useState([]);
  const [databaseList, setDatabaseList] = useState([]);
  const [tableList, setTableList] = useState([]);
  const [targetKeys, setTargetKeys] = useState([]);
  const [selectedKeys, setSelectedKeys] = useState([]);

  const [type, setType] = useState<string>('');
  const [datasource, setDatasource] = useState<string>('');
  const [database, setDatabase] = useState<string>('');
  const [grantType, setGrantType] = useState<string>('');

  const [formValues, setFormValues] = useState({});

  useEffect(() => {

    //获取用户列表
    fetch('/api/v1/users/manager/lists?offset=0&limit=100')
      .then((response) => response.json())
      .then((json) => setUserlist(json.data))
      .catch((error) => {
        console.log('fetch userlist failed', error);
      });

    //获取数据源类型  
    fetch('/api/v1/datasource_type/list?enable=1')
      .then((response) => response.json())
      .then((json) => {
        setTypeList(json.data);
        const valueDict: { [key: number]: string } = {};
        json.data.forEach((record: { id: string | number; name: string; }) => {
          valueDict[record.id] = record.name;
        });
      })
      .catch((error) => {
        console.log('Fetch type list failed', error);
    });

  }, []);

  const didQueryDatasource = (val: React.SetStateAction<string>) => {
    form.setFieldsValue({"datasource": "", "database": "", "tables": []});
    setType(val);
    setDatabaseList([]);
    setTableList([]);
    setSelectedKeys([]);
    setTargetKeys([]);
    //const formValue = form.getFieldsValue();
    //const type = formValue.type;

    fetch('/api/v1/datasource/list?type=' + val)
      .then((response) => response.json())
      .then((json) => setDatasourceList(json.data))
      .catch((error) => {
        console.log('fetch datasource list failed', error);
      });
  };

  const didQueryDatabase = (val: string) => {
    form.setFieldsValue({"database": "", "tables": []});
    setDatasource(val);
    setTableList([]);
    setSelectedKeys([]);
    setTargetKeys([]);
    fetch('/api/v1/query/database?datasource=' + val+'&type='+type)
      .then((response) => response.json())
      .then((json) => setDatabaseList(json.data))
      .catch((error) => {
        console.log('fetch database list failed', error);
      });
  };

  const didSetGrantType = (val:string)=>{
    setGrantType(val);
  }

  const didQueryTable = (val: string) => {
    form.setFieldsValue({"tables": []});
    setDatabase(val);
    setTableList([]);
    setSelectedKeys([]);
    setTargetKeys([]);
    fetch('/api/v1/query/table?datasource=' + datasource+'&database='+val+'&type='+type)
      .then((response) => response.json())
      .then((json) => setTableList(json.data))
      .catch((error) => {
        console.log('fetch table list failed', error);
      });
  };

  const onChange=(nextTargetKeys: string[], direction:TransferDirection, moveKeys:string[])=>{
      // @ts-ignore
      setTargetKeys(nextTargetKeys);
  }

  const onSelectChange=(sourceSelectedKeys: string[],  targetSelectedKeys:string[])=>{
    // @ts-ignore
    setSelectedKeys([...sourceSelectedKeys,...targetSelectedKeys]);
  }


  const asyncFetch = (values: {}) => {
    setLoading(true);
    const params = { ...values };
    const headers = new Headers();
    headers.append('Content-Type', 'application/json');
    fetch('/api/v1/privilege/grant', {
      method: 'post',
      headers: headers,
      body: JSON.stringify(params),
    })
      .then((response) => response.json())
      .then((json) => {
        setLoading(false);
        if(json.success==true){
          message.success("执行授权成功")
        }else{
          message.error("执行授权失败: "+json.msg)
        }
        
      })
      .catch((error) => {
        console.log('do grant failed', error);
        setLoading(false);
      });
  };



  const onFinish = (fieldValue: []) => {
    console.info("start grant privilege.")
    const values = {
      username: fieldValue["username"],
      type: fieldValue["type"],
      datasource: fieldValue["datasource"],
      grant_type: fieldValue["grant_type"],
      database: fieldValue["database"],
      tables: fieldValue["tables"] && fieldValue["tables"].join(";"),
      privileges: fieldValue["privileges"] && fieldValue["privileges"].join(";"),
      max_select: fieldValue["max_select"],
      max_update: fieldValue["max_update"],
      max_delete: fieldValue["max_delete"],
      expire_day: fieldValue["expire_day"],
      reason: fieldValue["reason"],
      enable:"1",
    };
    // @ts-ignore
    setFormValues(values);
    asyncFetch(values);
  };

  const onFinishFailed = (errorInfo: any) => {
    console.info(errorInfo);
    setLoading(false);
    message.error('执行授权失败');
  };

 

  return (
    <PageContainer>
      <Card>
        <Form
          labelCol={{ span: 3 }}
          wrapperCol={{ span: 31 }}
          style={{ marginTop: 8 }}
          form={form}
          onFinish={onFinish}
          onFinishFailed={err=>onFinishFailed(err)}
          initialValues={{max_select:'500',max_update:'100',max_delete:'100',expire_day:'7'}}
          name={'Form'}
        >

        <Form.Item
            name={'username'}
            label="授权用户"
            rules={[{ required: true, message: '请选择' }]}
          >
            <Select
              showSearch style={{ width: 350 }}
              placeholder=""
            >
              {userlist && userlist.map(item => <Option key={item.chinese_name} value={item.username}>{item.chinese_name }</Option>)}
            </Select>
          </Form.Item>
       
        <Form.Item
            name={'type'}
            label="授权数据源类型"
            rules={[{ required: true, message: '请选择' }]}
          >
            <Select
              showSearch style={{ width: 350 }}
              placeholder=""
              onChange={(val) => {didQueryDatasource(val); }}
            >
              {typeList && typeList.map(item => <Option key={item.name} value={item.name}>{item.name }</Option>)}
            </Select>
          </Form.Item>

          <Form.Item
            name={'datasource'}
            label="选择数据源"
            rules={[{ required: true, message: '请选择数据源' }]}
          >
            <Select
              showSearch style={{ width: 350 }}
              placeholder="请选择数据源"
              onChange={(val) => {
                didQueryDatabase(val);
              }}
            >
              {datasourceList && datasourceList.map(item => <Option key={item.host + ":" + item.port} value={item.host + ":" + item.port}>{item.name+"["+item.host + ":" + item.port + "]" }</Option>)}
            </Select>
          </Form.Item>

          { (type=="MySQL" || type=="Oracle"  || type=="PostgreSQL" || type=="SQLServer" || type=="ClickHouse" || type=="TiDB" || type=="Doris" || type=="MongoDB") &&
          <>
          <Form.Item
            name={'grant_type'}
            label="授权范围"
            rules={[{ required: true, message: '请选择' }]}
          >
            <Select style={{ width: 200 }} placeholder="请选择" value={database}
              onChange={(val) => {
                didSetGrantType(val);
              }}
            >
              <Option key={"database"} value={"database"}>整库授权</Option>
              <Option key={"table"} value={"table"}>按表授权</Option>
            </Select>
          </Form.Item>

          <Form.Item
            name={'database'}
            label="授权数据库"
            rules={[{ required: true, message: '请选择' }]}
          >
            <Select
              showSearch style={{ width: 350 }}
              placeholder="请选择"
              onChange={(val) => {
                didQueryTable(val);
              }}
            >
              {databaseList && databaseList.map(item => <Option key={item.database_name} value={item.database_name}>{item.database_name}</Option>)}
            </Select>
          </Form.Item>

          {form.getFieldValue("grant_type")=="table" && 
          <Form.Item
            name={'tables'}
            label="授权数据表"
            rules={[{ required: true, message: '请选择' }]}
          >
           <Transfer
            rowKey={record=>record && record.table_name}
            dataSource={tableList}
            showSearch
            listStyle={{
              width:320,
              height:300,
            }}
            titles={['数据表','授权表']}
            targetKeys={targetKeys}
            selectedKeys={selectedKeys}
            onChange={onChange}
            onSelectChange={onSelectChange}
            //onScroll={onscroll}
            render={item=>item && item.table_name}
           />
          </Form.Item>
          }
          
          <Form.Item
            name={'privileges'}
            label="授权权限"
            rules={[{ required: true, message: '请选择' }]}
          >
            <Checkbox.Group
              options={grantType=="database" ? checkBoxOptionsDatabase : checkBoxOptionsTable}
              defaultValue={['select']}
            />
          </Form.Item>

          <Form.Item
            name={'max_select'}
            label="查询上限"
            rules={[{ required: true, message: '请选择' }]}
          >
            <Select
              defaultValue={"100"}
              style={{width:120}}
              options={maxQueryNumberOptions}
            />
          </Form.Item>

          <Form.Item
            name={'max_update'}
            label="更新上限"
            rules={[{ required: true, message: '请选择' }]}
          >
            <Select
              defaultValue={"100"}
              style={{width:120}}
              options={maxQueryNumberOptions}
            />
          </Form.Item>

          <Form.Item
            name={'max_delete'}
            label="删除上限"
            rules={[{ required: true, message: '请选择' }]}
          >
            <Select
              defaultValue={"100"}
              style={{width:120}}
              options={maxQueryNumberOptions}
            />
          </Form.Item>
          </>
          }

          <Form.Item
            name={'expire_day'}
            label="有效期限"
            rules={[{ required: true, message: '请选择' }]}
          >
            <Select
              defaultValue={"7"}
              style={{width:120}}
              options={[
                {
                  value: '7',
                  label: '7天'
                },
                {
                  value: '31',
                  label: '1月'
                },
                {
                  value: '92',
                  label: '3月'
                },
                {
                  value: '183',
                  label: '6月'
                },
                {
                  value: '365',
                  label: '1年'
                },
              ]}
            />
          </Form.Item>

          <Form.Item
            name={'reason'}
            label="授权原因"
            rules={[{ required: true, message: '请填写授权原因' }]}
          >
            <TextArea
              showCount
              maxLength={100}
              style={{height:80,resize:'none'}}
            />
          </Form.Item>

          <Form.Item wrapperCol={{ offset: 3, span: 16 }}>
              <Button type="primary" htmlType="submit" loading={loading}>
                执行授权
              </Button>
          </Form.Item>

        </Form>

      </Card>

    </PageContainer>
  );
};

export default Index;

