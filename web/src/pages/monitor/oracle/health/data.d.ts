export type MysqlListItem = {
  id: number;
  host: string;
  hostname: string;
  port: string;
  tag: string;
  connect: bigint;
  version: string;
  role: string;


};

export type MysqlListData = {
  list: MysqlListItem[];
};

export type MysqlsValueType = {
  host: string;
  port: string;
} & Partial<MysqlListItem>;
