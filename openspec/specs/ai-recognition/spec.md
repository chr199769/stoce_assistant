# ai-recognition 规范

## Purpose
待定 - 由归档变更 add-stock-image-recognition 创建。归档后更新目的。

## Requirements
### Requirement: 股票图片识别
系统必须提供一个接口来上传图片并识别其中包含的股票信息。

#### Scenario: 从截图中成功识别股票
- **WHEN** 用户上传包含股票市场数据（例如，K线图、股票列表）的截图
- **THEN** 系统返回识别出的股票列表及其代码和名称
- **AND** 系统过滤掉非股票信息

#### Scenario: 未发现股票
- **WHEN** 用户上传没有可识别股票信息的图片
- **THEN** 系统返回空列表或表示无匹配的特定代码
