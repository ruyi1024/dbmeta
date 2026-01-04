# 数据质量管理系统数据库表结构说明

## 表结构概览

### 核心表

1. **data_quality_assessment** - 数据质量评估主表
   - 存储每次质量评估的整体结果
   - 包含6个质量维度的评分
   - 记录评估时间和数据源信息

2. **data_quality_issue** - 质量问题详情表
   - 存储具体的质量问题记录
   - 包含问题类型、严重程度、描述等信息
   - 支持问题状态跟踪和处理记录

3. **data_quality_ai_analysis** - AI分析结果表
   - 存储AI分析评估的结果
   - 包含AI评分、等级、趋势分析
   - 记录AI模型版本信息

4. **data_quality_ai_insight** - AI智能洞察表
   - 存储AI生成的智能洞察信息
   - 支持多条洞察记录
   - 包含优先级排序

5. **data_quality_ai_recommendation** - AI优化建议表
   - 存储AI给出的优化建议
   - 包含建议类型、优先级、预期提升
   - 支持建议采纳状态跟踪

### 辅助表

6. **data_quality_distribution** - 质量指标分布数据表
   - 存储各种质量指标的分布数据
   - 用于生成饼图等可视化图表
   - 支持多种分布类型

7. **data_quality_rule** - 质量规则配置表
   - 存储数据质量检查规则配置
   - 支持规则启用/禁用
   - 包含规则阈值和严重程度

8. **data_quality_task** - 质量评估任务表
   - 存储质量评估任务的执行记录
   - 支持定时任务和手动任务
   - 记录任务执行状态和结果

9. **data_quality_history** - 质量评估历史表
   - 存储历史评估记录
   - 用于趋势分析和对比
   - 按日期聚合数据

## 表关系说明

```
data_quality_assessment (评估主表)
    ├── data_quality_issue (质量问题)
    ├── data_quality_ai_analysis (AI分析)
    │   ├── data_quality_ai_insight (AI洞察)
    │   └── data_quality_ai_recommendation (AI建议)
    ├── data_quality_distribution (分布数据)
    └── data_quality_history (历史记录)
```

## 主要字段说明

### 质量评分字段
- `overall_score`: 整体质量评分 (0-100)
- `field_completeness`: 字段完整性 (%)
- `field_accuracy`: 字段准确性 (%)
- `table_completeness`: 表完整性 (%)
- `data_consistency`: 数据一致性 (%)
- `data_uniqueness`: 数据唯一性 (%)
- `data_timeliness`: 数据及时性 (%)

### 问题级别
- `high`: 高优先级问题
- `medium`: 中优先级问题
- `low`: 低优先级问题

### 质量等级
- `优秀`: 评分 >= 90
- `良好`: 评分 >= 80
- `一般`: 评分 >= 70
- `较差`: 评分 < 70

## 使用示例

### 1. 创建一次质量评估
```sql
INSERT INTO data_quality_assessment (
    assessment_time, total_tables, total_columns, total_issues,
    overall_score, overall_level,
    field_completeness, field_accuracy, table_completeness,
    data_consistency, data_uniqueness, data_timeliness
) VALUES (
    NOW(), 1250, 15680, 342,
    88.2, '良好',
    87.5, 92.3, 89.2,
    85.6, 94.1, 88.7
);
```

### 2. 记录质量问题
```sql
INSERT INTO data_quality_issue (
    assessment_id, table_name, column_name,
    issue_type, issue_level, issue_desc, issue_count
) VALUES (
    1, 'user_info', 'email',
    '完整性', 'high', '空值率超过20%', 1250
);
```

### 3. 保存AI分析结果
```sql
INSERT INTO data_quality_ai_analysis (
    assessment_id, ai_score, ai_level,
    trend_analysis, trend_direction, trend_percentage
) VALUES (
    1, 88.2, '良好',
    '近30天数据质量呈上升趋势', '上升', 2.3
);
```

### 4. 查询最新评估结果
```sql
SELECT 
    a.*,
    ai.ai_score, ai.ai_level, ai.trend_analysis,
    COUNT(DISTINCT i.id) as issue_count
FROM data_quality_assessment a
LEFT JOIN data_quality_ai_analysis ai ON a.id = ai.assessment_id
LEFT JOIN data_quality_issue i ON a.id = i.assessment_id
WHERE a.status = 1
ORDER BY a.assessment_time DESC
LIMIT 1;
```

## 索引优化建议

1. **时间范围查询**: 在 `assessment_time`, `check_time`, `analysis_time` 上建立索引
2. **数据源查询**: 在 `datasource_id`, `database_name` 上建立索引
3. **问题筛选**: 在 `issue_type`, `issue_level`, `status` 上建立索引
4. **趋势分析**: 在 `assessment_date` 上建立索引用于历史数据查询

## 数据维护建议

1. **定期归档**: 建议定期将历史数据归档，保留最近1年的详细数据
2. **数据清理**: 定期清理已处理且超过保留期的质量问题记录
3. **性能优化**: 对于大数据量场景，考虑分区表或读写分离
4. **备份策略**: 建议每日备份，保留至少30天的备份数据

