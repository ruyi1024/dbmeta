export interface TableListItem {
  id: number;
  name: string;
  description: string;
  enable: number;
  mail_list: string;
  sms_list: string;
  phone_list: string;
  wechat_list: string;
  mail_enable: number;
  sms_enable: number;
  phone_enable: number;
  wechat_enable: number;
  webhook_enable: number;
  webhook_url: string;
  gmt_created: date;
  gmt_updated: date;
}

export interface TableListPagination {
  total: number;
  pageSize: number;
  current: number;
}

export interface TableListData {
  list: TableListItem[];
  pagination: Partial<TableListPagination>;
}

export interface TableListParams {
  id?: number;
  name?: string;
  mail_list?: string;
  sms_list?: string;
  wechat_list?: string;
  phone_list?: string;
  mail_enable: number;
  sms_enable: number;
  phone_enable: number;
  wechat_enable: number;
  webhook_enable: number;
  webhook_url?: string;
  enable: number;
  pageSize?: number;
  currentPage?: number;
  filter?: { [key: string]: any[] };
  sorter?: { [key: string]: any };
}
