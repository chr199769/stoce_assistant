# Stock Assistant (股票助手)

Stock Assistant 是一款专为个人 A 股投资者设计的智能监控与分析工具。它结合了实时行情数据和先进的 AI 模型，帮助用户实时掌握市场动态，提供智能化的投资决策支持。

## 核心功能

*   **实时监控**: 支持 A 股实时行情查看，包括个股价格、涨跌幅、成交量等。
*   **AI 智能预测**: 结合多源新闻资讯和历史数据，利用 LLM（GPT-4o, Qwen, DeepSeek 等）预测短期走势。
*   **市场深度分析**:
    *   **每日总结**: 自动生成大盘及板块热点日报。
    *   **龙虎榜分析**: 追踪机构与游资动向，解析资金流向。
    *   **涨停池监控**: 实时捕捉涨停股票，分析涨停原因及封单强度。
*   **智能辅助工具**:
    *   **图片识别**: 支持拍照识别股票代码，快速添加自选。
    *   **财报分析**: 可视化展示营收、净利润等核心财务指标。
*   **数据同步**: 支持轻量级用户系统，实现多设备间自选股同步。

## 技术架构

本项目采用前后端分离的架构，后端基于 CloudWeGo 微服务体系构建。

### 架构图

```mermaid
graph TD
    Client[移动端 (React Native)] -->|HTTP/JSON| Gateway[API 网关 (Hertz)]
    
    subgraph Backend [微服务集群]
        Gateway -->|Thrift/RPC| StockService[股票服务 (Kitex)]
        Gateway -->|Thrift/RPC| AIService[AI 服务 (Kitex)]
    end
    
    subgraph External [外部依赖]
        StockService -->|HTTP| SinaAPI[新浪财经]
        StockService -->|HTTP| EastMoney[东方财富]
        AIService -->|API| LLM[大模型 (OpenAI/Qwen/DeepSeek)]
    end
    
    subgraph Storage [数据存储]
        StockService --> MySQL[(PostgreSQL)]
        StockService --> Redis[(Redis)]
    end
```

### 技术栈详情

*   **移动端**: 
    *   React Native (TypeScript)
    *   React Navigation, React Native Paper
*   **后端**:
    *   **语言**: Go (Golang)
    *   **Web 框架**: Hertz (HTTP 网关)
    *   **RPC 框架**: Kitex (微服务通信)
    *   **AI 框架**: LangChainGo
*   **基础设施**:
    *   Docker & Docker Compose

## 目录结构

```
.
├── backend/                # 后端微服务代码
│   ├── gateway/            # API 网关 (Hertz)
│   ├── stock_service/      # 股票数据服务 (Kitex)
│   └── ai_service/         # AI 分析服务 (Kitex)
├── mobile/                 # React Native 移动端代码
├── idl/                    # Thrift 接口定义
├── infrastructure/         # 基础设施配置 (Docker Compose)
└── openspec/               # 项目规范与变更记录
```

## 快速开始

### 环境要求

*   Go >= 1.24
*   Node.js >= 18
*   Docker & Docker Compose

### 启动服务

1.  **启动基础设施**:
    ```bash
    cd infrastructure
    docker-compose up -d
    ```

2.  **运行后端服务**:
    可以使用提供的脚本一键启动：
    ```bash
    ./start_services.sh
    ```

3.  **运行移动端**:
    ```bash
    cd mobile
    npm install
    npm start
    ```

## 贡献指南

请参考 [openspec/project.md](openspec/project.md) 了解详细的项目规范和开发流程。
