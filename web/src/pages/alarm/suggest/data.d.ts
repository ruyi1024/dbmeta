export type SuggestListItem = {
  id: number;
  event_key?: string;
  event_type: string;
  content: string;
  gmt_created: Date;
  modify?: boolean;
};

export type SuggestListData = {
  list: SuggestListItem[];
};
