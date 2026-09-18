# EQS AOC Adapter 开发基线 v1

EQS 保持交易业务事实源。新增独立只读适配层，不改变现有订单、支付、用户和权限模型。

## 只读资源

- `/adapter/v1/health`
- `/adapter/v1/projects`
- `/adapter/v1/projects/{id}/summary`
- `/adapter/v1/alerts`
- `/adapter/v1/version`

所有响应带 `project_id=eqs`、`source=eqs`、`observed_at`、`schema_version`。手机号、证件、支付、Token 和文件原文默认脱敏或不返回。

## 门禁

AOC 只能读取；跨项目访问必须拒绝；断连返回 `unavailable`；接口测试、脱敏测试、越权测试、版本兼容测试和回滚证据由 EQS 项目群提交。
