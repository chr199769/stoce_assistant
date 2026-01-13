# mobile-watchlist 规范

## Purpose
待定 - 由归档变更 add-stock-image-recognition 创建。归档后更新目的。

## Requirements
### Requirement: 本地自选股持久化
移动应用必须在本地持久化用户的自选股列表，以便在应用重启后保持可用。

#### Scenario: 保存自选股
- **WHEN** 用户将新股票添加到自选股
- **THEN** 更新后的自选股列表保存到本地存储

#### Scenario: 启动时加载自选股
- **WHEN** 用户启动应用程序
- **THEN** 从本地存储加载自选股并显示在主屏幕上

### Requirement: 从图片导入股票
移动应用必须允许用户通过从设备图库中选择图片来导入股票。

#### Scenario: 导入流程
- **WHEN** 用户点击“导入图片”按钮
- **THEN** 系统打开设备图片选择器
- **WHEN** 用户选择一张图片
- **THEN** 图片被发送到服务器进行识别
- **AND** 识别出的股票呈现给用户进行确认
