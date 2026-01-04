export type PgListItem = {
  id: number;
  host: string;
  hostname: string;
  port: string;
  tag: string;
  connect: bigint;
  version: string;
  role: string;

};

export type PgListData = {
  list: PgListItem[];
};

export type PgValueType = {
  host: string;
  port: string;
} & Partial<PgListItem>;
