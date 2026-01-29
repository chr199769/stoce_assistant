# 本地运行操作手册

本文档详细说明如何在本地环境中运行 Stock Assistant 的所有前后端服务。

## 1. 项目结构简介

本项目包含以下主要部分：

*   **Backend (后端)**: 基于 Golang 的微服务架构。
    *   `stock_service` (:8888): 提供股票数据相关服务。
    *   `ai_service` (:8889): 提供 AI 分析与对话服务。
    *   `gateway` (:8080): API 网关，统一对外提供 HTTP 接口。
    *   `knowledge_graph` (:8890，可选): 基于 Neo4j 的知识图谱服务。
*   **Mobile (前端)**: 基于 React Native 的移动端应用。
*   **Web Admin (管理台)**: 基于 React + Vite 的管理后台。

## 2. 环境前置条件

在开始之前，请确保您的开发环境已安装以下工具：

*   **Go**: 版本 1.24 或更高 (本项目使用 Go Workspace)。
*   **Node.js**: 版本 >= 20。
*   **Python**: 版本 >= 3.9（用于 AkShare 抓取宏观与新闻数据）。
*   **React Native 开发环境**:
    *   **Android**: Android Studio, Android SDK, 模拟器或真机。
    *   **iOS** (仅限 macOS): Xcode, CocoaPods。
*   **Docker 与 Docker Compose**: 如需一键启动基础设施或全栈服务。

## 3. 后端运行指南

后端服务之间通过直连方式通信（本地开发模式），无需额外的服务发现组件（如 Etcd）。

### 3.1 配置 AI 模型 Key

在启动 `ai_service` 之前，您需要配置大模型的 API Key。

1.  打开文件：`backend/ai_service/conf/llm_config.json`（可点击查看 [llm_config.json](file:///Users/bytedance/Projects/trae_projects/stock_assistant/backend/ai_service/conf/llm_config.json)）
2.  找到您想使用的 Provider (例如 `zhipu`, `openai`, `deepseek`)。
3.  将 `api_key` 替换为您自己的真实 Key。
4.  (可选) 修改 `current_provider` 字段为您选择的 Provider 名称。
5.  如需启用 Langfuse 评测与追踪，请确保 `langfuse` 配置中的 `public_key`、`secret_key` 与 `base_url` 正确。

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

#### （可选）第四步：启动知识图谱服务

知识图谱服务依赖 Neo4j。您可以本地安装 Neo4j，或使用 Docker 运行。

```bash
# 启动 Neo4j（Docker 示例）
docker run -d --name stock_neo4j -p 7474:7474 -p 7687:7687 -e NEO4J_AUTH=neo4j/neo4j_password neo4j:5

# 启动 knowledge_graph
cd backend/knowledge_graph
go run .
```
*   配置文件位置：`backend/knowledge_graph/conf/prod.json`（可点击查看 [prod.json](file:///Users/bytedance/Projects/trae_projects/stock_assistant/backend/knowledge_graph/conf/prod.json)）
*   成功启动后，服务将监听 `:8890` 端口。

### 3.3 验证后端

在浏览器或 Postman 中访问以下地址，确认网关已启动：
`http://localhost:8080/ping` (假设有 ping 接口，或者直接查看终端日志无报错)

### 3.4 一键启动脚本（可选）

项目提供一键启动脚本，自动在后台运行三大后端服务并输出日志：

```bash
./start_services.sh
```
*   日志位置：`logs/stock_service.log`、`logs/ai_service.log`、`logs/gateway.log`（脚本见 [start_services.sh](file:///Users/bytedance/Projects/trae_projects/stock_assistant/start_services.sh)）

### 3.5 使用 Docker Compose 启动基础设施（MySQL/Redis）

在项目根目录：

```bash
docker compose up -d
```
*   默认启动 `MySQL:3306` 与 `Redis:6379`，Compose 文件见 [docker-compose.yml](file:///Users/bytedance/Projects/trae_projects/stock_assistant/docker-compose.yml)。

### 3.6 使用 Docker Compose 启动全栈（含 Langfuse）

在 `infrastructure` 目录：

```bash
cd infrastructure
docker compose up -d
```
*   将自动构建并运行 `stock-service(8888)`、`ai-service(8889)`、`gateway(8080)`，以及 Langfuse 所需的 Postgres/ClickHouse/Redis/MinIO 与 Web/Worker（Compose 文件见 [infrastructure/docker-compose.yml](file:///Users/bytedance/Projects/trae_projects/stock_assistant/infrastructure/docker-compose.yml)）。
*   如需包含知识图谱与 Neo4j 的示例，请参考 [infrastructure/docker-compose.example.yml](file:///Users/bytedance/Projects/trae_projects/stock_assistant/infrastructure/docker-compose.example.yml)。

### 3.7 Python 环境与 AkShare（必备）

`stock_service` 会通过 Python 调用 AkShare 获取新闻与宏观数据（对应代码：[crawler.go](file:///Users/bytedance/Projects/trae_projects/stock_assistant/backend/stock_service/biz/provider/crawler/crawler.go)、[akshare_news.py](file:///Users/bytedance/Projects/trae_projects/stock_assistant/backend/stock_service/biz/provider/crawler/akshare_news.py)）。请按以下步骤准备本地 Python 环境：

1) 安装 Python 3（macOS 示例）：
```bash
# Homebrew
brew install python@3.11
```

2) 安装 AkShare（系统 Python，默认不使用虚拟环境）：
```bash
pip3 install -U akshare
# 如需国内镜像（可选）
pip3 install -U akshare -i https://pypi.tuna.tsinghua.edu.cn/simple
```

3) 配置 `python_bin`（如使用系统 Python，保持默认即可）：
- 配置文件位置：[prod.json](file:///Users/bytedance/Projects/trae_projects/stock_assistant/backend/stock_service/conf/prod.json)
- `akshare.python_bin` 保持为 `"python3"`，或填入你的系统 Python 绝对路径（可选）。

4) （可选）使用虚拟环境以隔离依赖：
```bash
python3 -m venv .venv
source .venv/bin/activate
pip install -U akshare
# 配置 akshare.python_bin 为 .venv/bin/python 绝对路径
```

5) 快速自检：
- 新闻抓取脚本（打印 JSON）：
```bash
cd backend/stock_service
python3 biz/provider/crawler/akshare_news.py conf/prod.json
```
- AkShare 直接调用测试（宏观函数示例）：
```bash
python3 -c 'import akshare as ak; print(ak.macro_china_cpi().tail().to_dict("records"))'
```

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

## 5. 管理台（Web Admin）运行指南

管理台位于 `web_admin` 目录下。

### 5.1 安装与开发启动
```bash
cd web_admin
npm install
npm run dev
```
*   默认通过 Vite 启动本地开发服务器，命令参考 [package.json](file:///Users/bytedance/Projects/trae_projects/stock_assistant/web_admin/package.json)。

### 5.2 构建与预览
```bash
npm run build
npm run preview
```

## 5. 常见问题与故障排除

### Q1: 端口冲突 (Address already in use)
*   **现象**: 启动服务时报错 `bind: address already in use`。
*   **解决**: 检查是否有其他程序占用了 8080、8888、8889、8890、3306、6379、3000 等端口。可以使用 `lsof -i :8080` (macOS/Linux) 查看占用进程并关闭它。

### Q2: Android 模拟器无法连接后端
*   **现象**: App 显示网络错误，后端没有收到请求。
*   **原因**: Android 模拟器中 `localhost` 指向模拟器本身，而不是电脑主机。
*   **解决**: 代码中已默认配置 `10.0.2.2:8080` 适配 Android 模拟器。如果您使用真机，请将手机和电脑连接同一 WiFi，并将 `mobile/src/api/client.ts` 中的 `BASE_URL` 修改为电脑的局域网 IP (例如 `http://192.168.1.x:8080`)。（参考 [client.ts](file:///Users/bytedance/Projects/trae_projects/stock_assistant/mobile/src/api/client.ts)）

### Q3: 找不到 `go.work` 模块
*   **现象**: 运行 `go run .` 时提示包找不到。
*   **解决**: 确保您是在项目根目录或各服务目录下运行，且 `go.work` 文件存在于项目根目录。Go 1.18+ 会自动识别 `go.work`。

### Q4: 缺少 API Key 导致 AI 功能不可用
*   **现象**: AI 对话或分析功能返回错误。
*   **解决**: 请确保按照 3.1 节正确配置了 `llm_config.json` 中的 API Key，并正确设置 Langfuse 的 `public_key`/`secret_key` 与 `base_url`（如启用评测）。

### Q5: 知识图谱服务连接失败
*   **现象**: `knowledge_graph` 服务启动时报 Neo4j 连接错误，或网关调用图谱接口失败。
*   **解决**: 确认 Neo4j 已启动并使用了与 `backend/knowledge_graph/conf/prod.json` 一致的地址与账号。Docker 模式下可参考 `infrastructure/docker-compose.example.yml` 的 `neo4j` 服务。

## 6. 运行验证与日志

*   **网关健康检查**: 访问 `http://localhost:8080/ping`。
*   **后端日志**: 使用一键脚本时，查看 `logs/stock_service.log`、`logs/ai_service.log`、`logs/gateway.log`。
*   **Langfuse Web**: 使用全栈 Compose 时，访问 `http://localhost:3000` 查看任务与追踪。
