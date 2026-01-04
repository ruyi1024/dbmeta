export interface RuleListItem {
  id: number;
  ruleName: string;
  ruleType: string;
  ruleDesc: string;
  ruleConfig: string;
  threshold: number;
  severity: string;
  enabled: number;
  createdBy?: string;
  createdAt: string;
  updatedAt: string;
}

