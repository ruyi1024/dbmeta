export type MongodbListItem = {
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

export type MongodbListData = {
  list: MongodbListItem[];
};

export type MongodbValueType = {
  ip: string;
  port: string;
} & Partial<MongodbListItem>;
