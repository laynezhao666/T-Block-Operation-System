#!/bin/bash
set -e
log() { echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"; }
PROJECT_ROOT=$(cd "$(dirname "$0")"; pwd)
mkdir -p target
TARGET_DIR="$PROJECT_ROOT/target"

# 只编译这些模块（白名单）
INCLUDE_DIRS=(
#    "cmdb"
#    "scheduler"
#    "agent"
#    "collector"
#    "alarm-compute"
#    "alarm-manage"
#    "alarm-server"
#    "data-cache"
#    "data-compute"
#    "data-query"
#    "data-store"
#    "cgi"
)

# 构建Docker镜像
build_docker_images() {
    for modname in "${INCLUDE_DIRS[@]}"; do
        echo "Building Docker image for $modname ..."
        (docker build -t "$modname" -t "$modname:latest" --platform linux/amd64 --build-arg MODULE="$modname" .)
        echo "Docker image for $modname built successfully."
    done
    if [ -d web ]; then
         echo "Building Docker image for web ..."
         (cd web && docker build -t web -t "web:latest" --platform linux/amd64 .)
         echo "Docker image for web built successfully."
    fi
}

# 打包为Docker镜像
package_docker_images() {
    build_docker_images
    # find all file to package
    files_to_pack=(tbos_docker.sh server.cfg ddl.sql)
    # save all backend module images
    for modname in "${INCLUDE_DIRS[@]}"; do
        docker save "$modname" -o "target/$modname.tar"
        files_to_pack+=("target/$modname.tar")
    done
    # save web image
    docker save web -o "target/web.tar"
    files_to_pack+=("target/web.tar")
    echo "Docker images packaged successfully."
    # create package
    tar -czvf tbos_image.tar.gz "${files_to_pack[@]}"
    echo "Package created: tbos_image.tar.gz"
}

# 导入Docker镜像到本地服务器
install_docker_images() {
    for modname in "${INCLUDE_DIRS[@]}"; do
        if [ -f "target/$modname.tar" ]; then
            echo "Loading Docker image for $modname ..."
            docker load -i "target/$modname.tar"
            echo "Docker image for $modname loaded successfully."
        else
            echo "Error: $modname.tar not found. Please run 'package image' first."
            exit 1
        fi
    done
    if [ -f "target/web.tar" ]; then
        echo "Loading Docker image for web..."
        docker load -i "web.tar"
        echo "Docker image for web loaded successfully."
    fi
}
# 获取操作系统信息
detect_os() {
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        OS=$ID
        VERSION_ID=${VERSION_ID:-$(grep VERSION_ID /etc/os-release | cut -d '"' -f2)}
    elif type lsb_release >/dev/null 2>&1; then
        OS=$(lsb_release -si | tr '[:upper:]' '[:lower:]')
    else
        OS=$(uname -s)
    fi
    echo "$OS"
}

# 在线安装 Docker
online_install() {
    local OS=$1
    echo "正在尝试在线安装 Docker..."

    case "$OS" in
        ubuntu|debian)
            echo "正在为 Ubuntu/Debian 安装 Docker..."
            sudo apt update -y || true
            sudo apt install -y apt-transport-https ca-certificates curl gnupg lsb-release
            curl -fsSL https://download.docker.com/linux/$OS/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
            echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/$OS $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
            sudo apt update -y
            sudo apt install -y docker-ce docker-ce-cli containerd.io
            ;;
        centos|rhel)
            echo "正在为 CentOS/RHEL 安装 Docker..."
            sudo yum install -y yum-utils
            sudo yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
            sudo yum install -y docker-ce docker-ce-cli containerd.io
            ;;
        *)
            echo "不支持的操作系统：$OS"
            exit 1
            ;;
    esac
}

# 离线安装 Docker
offline_install() {
    local TARBALL_PATH="$1"
    echo "正在进行离线安装 Docker from $TARBALL_PATH"

    if [ ! -f "$TARBALL_PATH" ]; then
        echo "离线安装包不存在: $TARBALL_PATH"
        exit 1
    fi

    echo "解压安装包..."
    sudo tar -xvf "$TARBALL_PATH" -C /tmp/
    cd /tmp/docker || exit 1

    echo "拷贝二进制文件到 /usr/bin/"
    sudo cp dockerd docker-proxy docker daemon.json /usr/bin/

    echo "创建 systemd 单元文件..."
    cat <<EOF | sudo tee /etc/systemd/system/docker.service > /dev/null
[Unit]
Description=Docker Application Container Engine
Documentation=https://docs.docker.com
After=network-online.target firewalld.service
Wants=network-online.target

[Service]
Type=notify
ExecStart=/usr/bin/dockerd
ExecReload=/bin/kill -s HUP \$MAINPID
TimeoutSec=0
RestartSec=2s
Restart=always
StartLimitBurst=3
StartLimitInterval=60s
LimitNOFILE=infinity
LimitNPROC=infinity
LimitCORE=infinity
TasksMax=infinity
Delegate=yes
KillMode=process

[Install]
WantedBy=multi-user.target
EOF
    echo "启动并启用 Docker 服务..."
    sudo systemctl daemon-reexec
    sudo systemctl daemon-reload
    sudo systemctl enable docker
    sudo systemctl start docker

    echo "离线安装完成。"
}


# 主程序逻辑
main() {
    # 判断系统类型并安装 Docker
    OS=$(detect_os)
    echo "检测到操作系统: $OS"

    # 检查是否已经安装了Docker
    if command -v docker &> /dev/null
    then
        echo "Docker 已安装，版本：$(docker --version)"
    else
        echo "Docker 未安装，正在开始安装..."

        echo "尝试在线安装 Docker..."
        if online_install "$OS"; then
            echo "在线安装成功！Docker 已就绪。"
            return 0
        else
            echo "在线安装失败，尝试离线安装..."
            OFFLINE_TARBALL="./offline/docker-23.0.6.tgz"  # 替换为你自己的离线包路径
            offline_install "$OFFLINE_TARBALL"
            echo "离线安装成功！Docker 已就绪。"
        fi

        echo "验证安装..."
        sudo docker info
    fi
}

install_go() {
  # 设置要安装的 Go 版本
  GO_VERSION="1.23.0"
  ARCH="amd64"

  # 根据系统架构自动判断
  case $(uname -m) in
  x86_64)
    ARCH="amd64"
    ;;
  aarch64 | arm64)
    ARCH="arm64"
    ;;
  *)
    echo "不支持的架构: $(uname -m)"
    #exit 1
    ;;
  esac

  # 下载地址
  GO_DOWNLOAD_URL="https://studygolang.com/dl/go${GO_VERSION}.linux-${ARCH}.tar.gz"

  echo "正在下载 Go ${GO_VERSION} for linux/${ARCH}..."
  wget -O go.tar.gz "$GO_DOWNLOAD_URL"

  if [ $? -ne 0 ]; then
    echo "下载失败"
    GOTARB_PATH="./offline/go${GO_VERSION}.linux-${ARCH}.tar.gz"  # 替换为你自己的离线包路径
    echo "正在进行离线安装 from $GOTARB_PATH"
    if [ ! -f "$GOTARB_PATH" ]; then
        echo "离线安装包不存在: $GOTARB_PATH"
        exit 1
    fi

    echo "解压安装包..."
    # 解压到 /usr/local
    echo "解压文件到 /usr/local/go"
    sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-${ARCH}.tar.gz
    # 配置环境变量（如果尚未配置）
    if ! grep -q 'export PATH=$PATH:/usr/local/go/bin' ~/.bashrc; then
      echo "配置环境变量..."
      echo 'export GOROOT=/usr/local/go' >>~/.bashrc
      echo 'export PATH=$PATH:$GOROOT/bin' >>~/.bashrc
      echo 'export GOPROXY=https://goproxy.io,direct' >>~/.bashrc
    fi
    # 应用环境变量
    source ~/.bashrc

    # 检查安装结果
    echo "验证安装..."
    go version

    if [ $? -eq 0 ]; then
      echo "Go 安装成功！当前版本：$(go version)"
    else
      echo "Go 安装失败，请检查权限或路径设置。"
      #exit 1
    fi
    #如果下载成功则安装
    else
        # 删除旧的 Go 安装目录（如有）
        sudo rm -rf /usr/local/go

        # 解压到 /usr/local
        echo "解压文件到 /usr/local/go"
        sudo tar -C /usr/local -zxvf go.tar.gz

        # 清理安装包
        rm go.tar.gz

        # 配置环境变量（如果尚未配置）
        if ! grep -q 'export PATH=$PATH:/usr/local/go/bin' ~/.bashrc; then
          echo "配置环境变量..."
          echo 'export GOROOT=/usr/local/go' >>~/.bashrc
          echo 'export PATH=$PATH:$GOROOT/bin' >>~/.bashrc
          echo 'export GOPROXY=https://goproxy.io,direct' >>~/.bashrc
        fi

        # 应用环境变量
        source ~/.bashrc

        # 检查安装结果
        echo "验证安装..."
        go version

        if [ $? -eq 0 ]; then
          echo "Go 安装成功！当前版本：$(go version)"
        else
          echo "Go 安装失败，请检查权限或路径设置。"
          #exit 1
        fi
  fi
}

install_middle() {
#中间件环境安装
# 导入镜像
log "开始导入 Docker 镜像..."
#export images_dir="./offline/image"
#if [ ! -d "$images_dir" ]; then
#  log "镜像目录 $images_dir 不存在。"
#  exit 1
#fi

#for tar_file in "$images_dir"/*.tar; do
#  if [ -f "$tar_file" ]; then
#    log "正在加载镜像: $tar_file"
#    docker load < "$tar_file" || { log "加载镜像 $tar_file 失败。"; exit 1; }
#  fi
#done
log "Docker 镜像导入完成。"

log "安装 Docker Compose..."
#sudo cp -r ./offline/docker-compose /usr/bin/
docker-compose --version

# 创建中间件目录
#sudo mkdir -p /data/middleware/{mysql,redis,zk,kafka}
#sudo chown 1000:1000 /data/middleware/* -R

echo '
networks:
  shared_network:
    external: true
services:
  mysql:
    image: mysql:5.7
    container_name: mysql
    environment:
      MYSQL_ROOT_PASSWORD: Tencent@123
      MYSQL_DATABASE: tbos
      MYSQL_USER: tbos
      MYSQL_PASSWORD: idc#2024!tbos28
    ports:
      - "3306:3306"
    volumes:
      - /data/middleware/mysql:/var/lib/mysql
      - /data/middleware/init-db-scripts:/docker-entrypoint-initdb.d
    networks:
      - shared_network
    restart: unless-stopped

  redis:
    image: redis:7.0
    container_name: redis
    ports:
      - "6379:6379"
    volumes:
      - /data/middleware/redis:/data
    networks:
      - shared_network
    command: redis-server --requirepass idc2024tbos28
    restart: unless-stopped

  zookeeper:
    image: bitnami/zookeeper:3.7
    container_name: zookeeper
    ports:
      - "2181:2181"
    environment:
      ALLOW_ANONYMOUS_LOGIN: "yes"
      ZOO_MAX_CLIENT_CNXNS: 100
      ZOO_ENABLE_ADMIN_SERVER: "no"
      ZOO_SESSION_TIMEOUT: 60000
    volumes:
      - /data/middleware/zk:/bitnami/zookeeper
    user: "0:0"
    networks:
      - shared_network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "echo", "ruok", "|", "nc", "localhost", "2181"]
      interval: 5s
      timeout: 10s
      retries: 10

  kafka:
    image: bitnami/kafka:2.8.1
    container_name: kafka
    ports:
      - "9092:9092"
    environment:
      KAFKA_BROKER_ID: 1
      KAFKA_CFG_LISTENERS: PLAINTEXT://:9092
      KAFKA_CFG_ADVERTISED_LISTENERS: PLAINTEXT://kafka:9092
      KAFKA_CFG_ZOOKEEPER_CONNECT: zookeeper:2181
      ALLOW_PLAINTEXT_LISTENER: "yes"
      KAFKA_CFG_AUTO_CREATE_TOPICS_ENABLE: true
      KAFKA_CFG_NUM_PARTITIONS: 3
      KAFKA_CFG_BROKER_ID: 1
      KAFKA_CFG_ZOOKEEPER_SESSION_TIMEOUT_MS: 60000
    volumes:
      - /data/middleware/kafka:/bitnami/kafka
      #- /data/middleware/init-scripts:/docker-entrypoint-initdb.d
    user: "0:0"
    depends_on:
      zookeeper:
        condition: service_healthy
    healthcheck:
      test: ["CMD", "nc", "-z", "localhost", "9092"]
      interval: 10s
      timeout: 5s
      retries: 15
    networks:
      - shared_network
    restart: unless-stopped' > docker-compose.yaml && docker-compose -f docker-compose.yaml up -d
    log "部署完成！"
}

# docker.io/library/cmdb:latest
# docker.io/library/data-store:latest
# docker.io/library/scheduler:latest
# docker.io/library/agent:latest
# docker.io/library/collector:latest
# docker.io/library/alarm-compute:latest
# docker.io/library/alarm-manage:latest
# docker.io/library/alarm-server:latest
# docker.io/library/data-cache:latest
# docker.io/library/data-compute:latest
# docker.io/library/data-query:latest
# docker.io/library/data-store:latest
# docker.io/library/cgi:latest
# docker.io/library/tbos-web:latest

install_tbos() {
#中间件环境安装
# 导入镜像
log "开始导入 Docker 镜像..."
export images_dir="./target"
if [ ! -d "$images_dir" ]; then
  log "镜像目录 $images_dir 不存在。"
  exit 1
fi

for tar_file in "$images_dir"/*.tar; do
  if [ -f "$tar_file" ]; then
    log "正在加载镜像: $tar_file"
    docker load < "$tar_file" || { log "加载镜像 $tar_file 失败。"; exit 1; }
  fi
done
log "Docker 镜像导入完成。"

echo '
networks:
  shared_network:
    external: true
services:
  cmdb:
    image: docker.io/library/cmdb:latest
    container_name: cmdb
    env_file:
      - .env
    ports:
      - "9102:9102"
    networks:
      - shared_network
    volumes:
      - /etc/localtime:/etc/localtime:ro
      - /etc/timezone:/etc/timezone:ro
    restart: unless-stopped

  scheduler:
    image: docker.io/library/scheduler:latest
    container_name: scheduler
    env_file:
      - .env
    ports:
      - "9103:9103"
    networks:
      - shared_network
    volumes:
      - /etc/localtime:/etc/localtime:ro
      - /etc/timezone:/etc/timezone:ro
    restart: unless-stopped

  collector:
    image: docker.io/library/collector:latest
    container_name: collector
    env_file:
      - .env
    ports:
      - "9101:9101"
    networks:
      - shared_network
    volumes:
      - /etc/localtime:/etc/localtime:ro
      - /etc/timezone:/etc/timezone:ro
    restart: unless-stopped

  alarm-compute:
    image: docker.io/library/alarm-compute:latest
    container_name: alarm-compute
    env_file:
      - .env
    ports:
      - "9108:9108"
    networks:
      - shared_network
    volumes:
      - /etc/localtime:/etc/localtime:ro
      - /etc/timezone:/etc/timezone:ro
    restart: unless-stopped

  alarm-manage:
    image: docker.io/library/alarm-manage:latest
    container_name: alarm-manage
    env_file:
      - .env
    ports:
      - "9109:9109"
    networks:
      - shared_network
    volumes:
      - /etc/localtime:/etc/localtime:ro
      - /etc/timezone:/etc/timezone:ro
    restart: unless-stopped

  alarm-server:
    image: docker.io/library/alarm-server:latest
    container_name: alarm-server
    env_file:
      - .env
    ports:
      - "9110:9110"
    networks:
      - shared_network
    volumes:
      - /etc/localtime:/etc/localtime:ro
      - /etc/timezone:/etc/timezone:ro
    restart: unless-stopped

  data-cache:
    image: docker.io/library/data-cache:latest
    container_name: data-cache
    env_file:
      - .env
    ports:
      - "9105:9105"
    networks:
      - shared_network
    volumes:
      - /etc/localtime:/etc/localtime:ro
      - /etc/timezone:/etc/timezone:ro
    restart: unless-stopped

  data-compute:
    image: docker.io/library/data-compute:latest
    container_name: data-compute
    env_file:
      - .env
    ports:
      - "9104:9104"
    networks:
      - shared_network
    volumes:
      - /etc/localtime:/etc/localtime:ro
      - /etc/timezone:/etc/timezone:ro
    restart: unless-stopped

  data-query:
    image: docker.io/library/data-query:latest
    container_name: data-query
    env_file:
      - .env
    ports:
      - "9107:9107"
    networks:
      - shared_network
    volumes:
      - /etc/localtime:/etc/localtime:ro
      - /etc/timezone:/etc/timezone:ro
    restart: unless-stopped

  data-store:
    image: docker.io/library/data-store:latest
    container_name: data-store
    env_file:
      - .env
    ports:
      - "9106:9106"
    networks:
      - shared_network
    volumes:
      - /etc/localtime:/etc/localtime:ro
      - /etc/timezone:/etc/timezone:ro
    restart: unless-stopped

  cgi:
    image: docker.io/library/cgi:latest
    container_name: cgi
    env_file:
      - .env
    ports:
      - "9111:9111"
      - "8081:8081"
    networks:
      - shared_network
    volumes:
      - /etc/localtime:/etc/localtime:ro
      - /etc/timezone:/etc/timezone:ro
    restart: unless-stopped

  web:
    image: docker.io/library/web:latest
    container_name: web
    env_file:
      - .env
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
    ports:
      - "8080:8080"
    networks:
      - shared_network
    restart: unless-stopped' > tbos-docker-compose.yaml && docker-compose -f tbos-docker-compose.yaml up -d
    log "部署完成！"
    # 部署完成后输出访问提示
    log "✅ 部署完成！"
    log "🌐 你可以通过以下地址访问你的 Web 应用："
    log ""
    log "    http://<你的服务器IP>:8080"
    log ""
    log "💡 注意事项："
    log "1. 如果你使用的是云服务器，请确保安全组/防火墙开放了 8080 端口。"
    log "2. 如果你配置了域名，请替换 IP 为你的域名访问。"
    log "3. 如果服务未启动，请检查服务状态：docker ps"
}

case "$1" in
    "build")
        build_docker_images
        ;;
    "package")
        package_docker_images
        ;;
    "install")
        #install_docker_images
        install_tbos
        ;;
    "strat")
        docker-compose -f tbos-docker-compose.yaml up -d
        ;;
    "stop")
        docker-compose -f tbos-docker-compose.yaml down -v
        ;; 
    *)
        echo "Usage: $0 {build|package|install|start|stop} image"
        exit 1
        ;;
esac