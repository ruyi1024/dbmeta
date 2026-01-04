export type MysqlListItem = {
  id: number;
  host: string;
  hostname: string;
  port: string;
  tag: string;
  connect: bigint;
  version: string;
  role: string;
  readonly: string;
  select: number;


};

export type MysqlListData = {
  list: MysqlListItem[];
};

export type MysqlsValueType = {
  ip: string;
  port: string;
} & Partial<MysqlListItem>;
