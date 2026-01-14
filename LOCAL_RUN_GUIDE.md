# 本地运行操作手册

本文档详细说明如何在本地环境中运行 Stock Assistant 的所有前后端服务。

## 1. 项目结构简介

本项目包含以下主要部分：

*   **Backend (后端)**: 基于 Golang 的微服务架构。
    *   `stock_service` (:8888): 提供股票数据相关服务。
    *   `ai_service` (:8889): 提供 AI 分析与对话服务。
    *   `gateway` (:8080): API 网关，统一对外提供 HTTP 接口。
*   **Mobile (前端)**: 基于 React Native 的移动端应用。

## 2. 环境前置条件

在开始之前，请确保您的开发环境已安装以下工具：

*   **Go**: 版本 1.24 或更高 (本项目使用 Go Workspace)。
*   **Node.js**: 版本 >= 20。
*   **React Native 开发环境**:
    *   **Android**: Android Studio, Android SDK, 模拟器或真机。
    *   **iOS** (仅限 macOS): Xcode, CocoaPods。

## 3. 后端运行指南

后端服务之间通过直连方式通信（本地开发模式），无需额外的服务发现组件（如 Etcd）。

### 3.1 配置 AI 模型 Key

在启动 `ai_service` 之前，您需要配置大模型的 API Key。

1.  打开文件：`backend/ai_service/conf/llm_config.json`
2.  找到您想使用的 Provider (例如 `zhipu`, `openai`, `deepseek`)。
3.  将 `api_key` 替换为您自己的真实 Key。
4.  (可选) 修改 `current_provider` 字段为您选择的 Provider 名称。

### 3.2 启动服务

建议按照以下顺序启动服务。请打开 **3 个独立的终端窗口** 分别运行。

#### 第一步：启动 Stock Service

```bash
cd backend/stock_service
go run .
```
*   成功启动后，服务将监听 `:8888` 端口。

#### 第二步：启动 AI Service

```bash
cd backend/ai_service
go run .
```
*   成功启动后，服务将监听 `:8889` 端口。

#### 第三步：启动 Gateway

```bash
cd backend/gateway
go run .
```
*   成功启动后，服务将监听 `:8080` 端口。

### 3.3 验证后端

在浏览器或 Postman 中访问以下地址，确认网关已启动：
`http://localhost:8080/ping` (假设有 ping 接口，或者直接查看终端日志无报错)

## 4. 前端运行指南

前端项目位于 `mobile` 目录下。

### 4.1 安装依赖

打开一个新的终端窗口，进入 `mobile` 目录：

```bash
cd mobile
npm install
# 或者使用 yarn
# yarn install
```

**iOS 特别步骤 (macOS Only)**:
如果您要运行 iOS 版本，需要安装 CocoaPods 依赖：

```bash
cd ios
pod install
cd ..
```

### 4.2 启动 Metro 服务

Metro 是 React Native 的打包工具，需要一直运行。

```bash
# 在 mobile 目录下
npm start
```

### 4.3 运行应用

打开**另一个新的终端窗口** (保持 Metro 窗口运行)，执行以下命令安装并启动 App：

**Android**:
确保已启动 Android 模拟器或连接了开启 USB 调试的真机。
```bash
# 在 mobile 目录下
npm run android
```

**iOS**:
```bash
# 在 mobile 目录下
npm run ios
```

## 5. 常见问题与故障排除

### Q1: 端口冲突 (Address already in use)
*   **现象**: 启动服务时报错 `bind: address already in use`。
*   **解决**: 检查是否有其他程序占用了 8080, 8888 或 8889 端口。可以使用 `lsof -i :8080` (macOS/Linux) 查看占用进程并关闭它。

### Q2: Android 模拟器无法连接后端
*   **现象**: App 显示网络错误，后端没有收到请求。
*   **原因**: Android 模拟器中 `localhost` 指向模拟器本身，而不是电脑主机。
*   **解决**: 代码中已默认配置 `10.0.2.2:8080` 适配 Android 模拟器。如果您使用真机，请将手机和电脑连接同一 WiFi，并将 `mobile/src/api/client.ts` 中的 `BASE_URL` 修改为电脑的局域网 IP (例如 `http://192.168.1.x:8080`)。

### Q3: 找不到 `go.work` 模块
*   **现象**: 运行 `go run .` 时提示包找不到。
*   **解决**: 确保您是在项目根目录或各服务目录下运行，且 `go.work` 文件存在于项目根目录。Go 1.18+ 会自动识别 `go.work`。

### Q4: 缺少 API Key 导致 AI 功能不可用
*   **现象**: AI 对话或分析功能返回错误。
*   **解决**: 请确保按照 3.1 节正确配置了 `llm_config.json` 中的 API Key。
