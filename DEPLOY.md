# 股票助手 (Stock Assistant) 部署指南

本指南将指导你如何在云服务器（特别是**火山引擎 ECS**）上部署股票助手服务，包括后端服务环境搭建、部署流程及移动端配置。

## 目录

- [方案一：云服务器部署 (火山引擎 ECS)](#方案一-云服务器部署-火山引擎-ecs)
- [方案二：本地电脑部署 (省钱方案)](#方案二-本地电脑部署-省钱方案)
- [移动端应用配置与构建](#移动端应用配置与构建)
- [常见问题与故障排查](#常见问题与故障排查)

---

## 方案一：云服务器部署 (火山引擎 ECS)

### 1. 购买/准备 ECS 实例

建议配置如下：
- **操作系统**: Ubuntu 20.04 LTS 或 22.04 LTS (推荐) 或 CentOS 7.9+
- **CPU/内存**: 至少 2 vCPU / 4 GiB 内存 (因为运行了 Langfuse 全家桶和 AI 服务，建议 4 vCPU / 8 GiB 以获得更佳体验)
- **公网 IP**: 必须分配公网 IP (EIP) 以供移动端访问
- **云盘**: 建议至少 40GB 系统盘

### 2. 配置安全组 (防火墙)

在火山引擎控制台的安全组规则中，务必放行以下端口：

| 协议 | 端口范围 | 描述 |
| :--- | :--- | :--- |
| TCP | **8080** | API 网关 (移动端主要访问入口) |
| TCP | **3000** | Langfuse 控制台 (查看 AI 监控) |
| TCP | **22** | SSH 远程连接 |
| TCP | **8888** | (可选) Stock Service 直连调试 |
| TCP | **8889** | (可选) AI Service 直连调试 |

### 3. 安装运行环境 (Docker)

登录到你的 ECS 服务器，执行以下命令安装 Docker 和 Docker Compose。

**Ubuntu 用户:**

```bash
# 更新软件包索引
sudo apt-get update

# 安装依赖
sudo apt-get install -y apt-transport-https ca-certificates curl software-properties-common

# 安装 Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh

# 启动 Docker 并设置开机自启
sudo systemctl start docker
sudo systemctl enable docker

# 安装 Docker Compose (如果 docker compose 命令不可用)
# 较新版 Docker 已内置 'docker compose' 命令，可跳过此步，直接通过 'docker compose version' 验证
```

**配置国内镜像加速 (推荐):**
为了加快镜像拉取速度，建议配置 Docker 镜像加速。

```bash
sudo mkdir -p /etc/docker
sudo tee /etc/docker/daemon.json <<-'EOF'
{
  "registry-mirrors": [
    "https://docker.m.daocloud.io",
    "https://mirror.baidubce.com"
  ]
}
EOF
sudo systemctl daemon-reload
sudo systemctl restart docker
```

### 4. 部署服务

**4.1 获取代码**

```bash
# 安装 git
sudo apt-get install -y git

# 克隆代码
git clone <你的仓库地址> stock_assistant
cd stock_assistant
```

**4.2 启动服务**

```bash
cd infrastructure
docker compose up -d --build
```

**4.3 验证**

确保所有容器状态为 `Up`，并通过 `http://<服务器公网IP>:8080` 访问。

---

## 方案二：本地电脑部署 (省钱方案)

如果你想节省云服务器费用，可以将自己的电脑 (macOS/Windows/Linux) 作为服务器。

### 方法 A：局域网访问 (仅限家中 WiFi)

如果你的手机和电脑连接的是同一个 WiFi，可以直接通过电脑的局域网 IP 访问。

1.  **启动后端服务**: 在电脑上运行 `cd infrastructure && docker-compose up -d --build`。
2.  **获取电脑 IP**:
    *   **macOS**: 系统设置 -> 网络 -> Wi-Fi -> 详细信息 -> IP 地址 (通常是 `192.168.x.x`)。或者终端输入 `ipconfig getifaddr en0`。
    *   **Windows**: 命令行输入 `ipconfig`，查看 IPv4 地址。
3.  **配置手机端**: 将 `mobile/src/api/client.ts` 中的 `SERVER_IP` 修改为你电脑的局域网 IP。

**优点**: 零成本，速度快。
**缺点**: 手机断开 WiFi 或出门后无法使用。

### 方法 B：内网穿透 (推荐，支持外网访问)

通过内网穿透工具，将你本地电脑的 `8080` 端口映射到公网，生成一个临时域名的公网地址。

推荐工具：**cpolar** (国内稳定) 或 **ngrok** / **Cloudflare Tunnel**。

#### 使用 cpolar (示例)

1.  **注册并安装**: 访问 [cpolar 官网](https://www.cpolar.com/) 注册账号并下载安装包。
2.  **启动隧道**:
    在终端运行以下命令，将本地 8080 端口暴露到公网：
    ```bash
    cpolar http 8080
    ```
3.  **获取公网地址**:
    cpolar 会输出一个公网地址，例如 `http://1a2b3c4d.cpolar.cn`。
4.  **配置手机端**:
    修改 `mobile/src/api/client.ts`：
    ```typescript
    const BASE_URL = 'http://1a2b3c4d.cpolar.cn'; // 使用 cpolar 生成的地址
    ```

**注意**: 免费版 cpolar/ngrok 的域名通常是临时的，重启后会变，且有带宽限制。

---

## 移动端应用配置与构建

现在后端已在云端运行，需要修改移动端应用以连接到云服务器。

### 1. 修改 API 地址

在本地开发环境中，打开 `mobile/src/api/client.ts`：

```typescript
// mobile/src/api/client.ts

// 替换为你的火山引擎 ECS 公网 IP
const SERVER_IP = '123.45.67.89'; // <--- 修改这里为实际公网IP

const BASE_URL = `http://${SERVER_IP}:8080`;

// ...
```

### 2. 构建与运行

#### Android

连接真机或启动模拟器：

```bash
cd mobile
npm install
npm run android
```

或者构建发布版 APK：

```bash
cd android
./gradlew assembleRelease
```
APK 输出路径: `mobile/android/app/build/outputs/apk/release/app-release.apk`。

#### iOS (仅 macOS)

```bash
cd mobile/ios
pod install
cd ..
npm run ios
```

---

## 常见问题与故障排查

### 1. 无法访问公网 IP 的 8080 端口
- **检查安全组**: 确认火山引擎控制台的安全组规则已放行 TCP 8080。
- **检查防火墙**: 确认服务器内部防火墙 (如 `ufw` 或 `iptables`) 未拦截。
  ```bash
  sudo ufw allow 8080/tcp
  ```
- **检查服务监听地址**: 容器内部应监听 `0.0.0.0` 而非 `127.0.0.1` (Gateway 默认配置已正确设置)。

### 2. 数据库连接失败
- 如果修改了 `docker-compose.yml` 中的数据库密码，请务必同步修改 `backend/stock_service` 和 `backend/ai_service` 代码中的连接配置，或使用环境变量覆盖。

### 3. 内存不足 / 服务 OOM (Out of Memory)
- Langfuse 组件较多 (ClickHouse, Postgres 等)，如果服务器内存只有 2GB，可能会导致服务崩溃。
- **解决方案**: 增加 Swap 分区。
  ```bash
  # 创建 4GB Swap
  sudo fallocate -l 4G /swapfile
  sudo chmod 600 /swapfile
  sudo mkswap /swapfile
  sudo swapon /swapfile
  echo '/swapfile none swap sw 0 0' | sudo tee -a /etc/fstab
  ```

### 4. 镜像拉取超时
- 确保配置了国内镜像加速源 (见[环境准备](#环境准备-火山引擎-ecs)一节)。
