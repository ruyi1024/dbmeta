import React, { useEffect, useState } from "react";
import { Drawer, message, Table, Tabs, Alert } from "antd";
import { AndroidOutlined, AppleOutlined } from "@ant-design/icons";
import styles from "@/pages/monitor/event/index.less";


const EventInfoView: React.FC<any> = (props: any) => {
  const { eventUuid, modalVisible } = props;
  const [eventDetail, setEventDetail] = useState<any>();
  const [detailTableColumn, setDetailTableColumns] = useState<any>();
  const [eventDescription, setEventDescription] = useState<string>('暂无介绍');
  console.info(eventUuid)
  useEffect(() => {
    try {
      // eslint-disable-next-line @typescript-eslint/no-use-before-define
      fetch(`/api/v1/event/detail?uuid=${eventUuid}`)
        .then((response) => response.json())
        .then((json) => {
          console.info(json.data);
          return (
            setEventDetail(json.data),
            setDetailTableColumns(json.columns),
            setEventDescription(json.description)
          );
        })
        .catch((error) => {
          console.log('fetch event detail failed', error);
        });
    } catch (e) {
      message.error(`get event error. ${e}`)
    }
  }, [eventUuid])

  return (<>
    <Drawer
      title={'事件告警AI自动化分析台'}
      width={1280}
      onClose={() => props.onCancel()}
      destroyOnClose
      open={modalVisible}
      placement="left"
      closable={false}
    >

      <Tabs defaultActiveKey="1">
        <Tabs.TabPane
          tab={
            <span>
              <AndroidOutlined />
              事件介绍
            </span>
          }
          key="1"
        >
          <p style={{ whiteSpace: 'pre-wrap' }}>
            <Alert
              message={eventDescription}
              type="info"
              showIcon
            />
          </p>
        </Tabs.TabPane>
      </Tabs>
      <Tabs defaultActiveKey="2">
        <Tabs.TabPane
          tab={
            <span>
              <AppleOutlined />
              定位分析
            </span>
          }
          key="2"
        >
          <Table
            className={styles.tableStyle}
            dataSource={eventDetail}
            columns={detailTableColumn}
            size={'small'}
            pagination={{
              defaultPageSize: 5,
              showSizeChanger: true,
              pageSizeOptions: ['5', '10', '20'],
              showQuickJumper: true,
              showTotal: (t, range) => `第 ${range[0]}-${range[1]}条， 共 ${t}条`,
            }}
          />
        </Tabs.TabPane>
      </Tabs>
    </Drawer>
  </>);
};
export default EventInfoView;
