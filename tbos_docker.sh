#!/bin/bash
set -e
log() { echo "[$(date +'%Y-%m-%d %H:%M:%S')] $1"; }
PROJECT_ROOT=$(cd "$(dirname "$0")"; pwd)
# 确保 ~/bin 在 PATH 中（docker-compose 可能安装在此）
export PATH="$HOME/bin:$PATH"
mkdir -p target
TARGET_DIR="$PROJECT_ROOT/target"

# 兼容 docker compose (plugin) 和 docker-compose (standalone)
run_compose() {
    if docker compose version &>/dev/null; then
        docker compose "$@"
    elif command -v docker-compose &>/dev/null; then
        docker-compose "$@"
    else
        log "❌ 未找到 docker compose 或 docker-compose，请先运行: ./tbos_docker.sh deps"
        exit 1
    fi
}

# 读取配置文件
read_config() {
    local config_file="$PROJECT_ROOT/server.cfg"
    if [ ! -f "$config_file" ]; then
        log "配置文件 $config_file 不存在"
        return 1
    fi

    source "$config_file"

    # 设置默认值
    MYSQL_ADDR=${MYSQL_ADDR:-127.0.0.1:3306}
    MYSQL_ROOT_PASSWORD=${MYSQL_ROOT_PASSWORD:-tbos88888888}
    MYSQL_USER=${MYSQL_USER:-tbos}
    MYSQL_PASSWORD=${MYSQL_PASSWORD:-tbos88888888}
    MYSQL_DATABASE=${MYSQL_DATABASE:-tbos}

    REDIS_ADDR=${REDIS_ADDR:-127.0.0.1:6380}
    REDIS_PASSWORD=${REDIS_PASSWORD:-tbos88888888}

    KAFKA_ADDR=${KAFKA_ADDR:-127.0.0.1:9092}
    KAFKA_POINT_TOPIC=${KAFKA_POINT_TOPIC:-agent_points}
    KAFKA_ALARM_PRODUCE_TOPIC=${KAFKA_ALARM_PRODUCE_TOPIC:-tbos-alarm-produce}
    KAFKA_ALARM_VALID_TOPIC=${KAFKA_ALARM_VALID_TOPIC:-tbos-rule-valid}
    KAFKA_ALARM_PUSH_TOPIC=${KAFKA_ALARM_PUSH_TOPIC:-tbos-alarm-push}

    LOCAL_IP=${LOCAL_IP:-127.0.0.1}

    log "配置文件加载完成"
}

# 模块白名单
INCLUDE_DIRS=(
    "cmdb"
    "scheduler"
    "agent"
    "collector"
    "alarm-compute"
    "alarm-manage"
    "alarm-server"
    "data-cache"
    "data-compute"
    "data-query"
    "data-store"
    "cgi"
    "dac"
    "web"
)


# 中间件镜像列表
MIDDLEWARE_IMAGES=(
    "mysql:5.7"
    "redis:7.0"
    "apache/kafka:3.7.0"
)

# 拉取中间件镜像
pull_middleware_images() {
    log "开始拉取中间件镜像..."
    # 检测当前架构，arm64 上部分镜像需要指定 platform
    local PULL_PLATFORM=""
    if [[ "$(uname -m)" == "arm64" || "$(uname -m)" == "aarch64" ]]; then
        PULL_PLATFORM="--platform linux/amd64"
        log "检测到 ARM64 架构，将使用 linux/amd64 平台拉取镜像（通过 QEMU 模拟运行）"
    fi
    for image in "${MIDDLEWARE_IMAGES[@]}"; do
        log "拉取镜像: $image"
        docker pull $PULL_PLATFORM "$image" || { log "拉取镜像 $image 失败"; exit 1; }
    done
    log "✅ 中间件镜像拉取完成"
}

# 生成中间件 docker-compose 配置文件
generate_middleware_compose() {
    log "生成中间件 docker-compose 配置文件..."

    # 创建中间件数据目录（存放在当前项目目录下）
    local DATA_DIR="$PROJECT_ROOT/.data"
    mkdir -p "$DATA_DIR/init-db-scripts"
    if [ -f "ddl.sql" ]; then
        cp ddl.sql "$DATA_DIR/init-db-scripts/"
    fi

    echo '
networks:
  shared_network:
    external: true
services:
  mysql:
    image: mysql:5.7
    platform: linux/amd64
    container_name: mysql
    command:
      - --character-set-server=utf8mb4
      - --collation-server=utf8mb4_unicode_ci
      - --default-authentication-plugin=mysql_native_password
    environment:
      MYSQL_ROOT_PASSWORD: '${MYSQL_ROOT_PASSWORD}'
      MYSQL_DATABASE: '${MYSQL_DATABASE}'
      MYSQL_USER: '${MYSQL_USER}'
      MYSQL_PASSWORD: '${MYSQL_PASSWORD}'
    ports:
      - "3306:3306"
    volumes:
      - '${PROJECT_ROOT}'/.data/mysql:/var/lib/mysql
      - '${PROJECT_ROOT}'/.data/init-db-scripts:/docker-entrypoint-initdb.d:ro
    user: "0:0"
    networks:
      - shared_network
    restart: unless-stopped

  redis:
    image: redis:7.0
    container_name: redis
    ports:
      - "6379:6379"
    volumes:
      - '${PROJECT_ROOT}'/.data/redis:/data
    networks:
      - shared_network
    command: redis-server --requirepass '${REDIS_PASSWORD}'
    restart: unless-stopped

  kafka:
    image: apache/kafka:3.7.0
    container_name: kafka
    ports:
      - "9092:9092"
    environment:
      KAFKA_NODE_ID: 1
      KAFKA_PROCESS_ROLES: broker,controller
      KAFKA_LISTENERS: PLAINTEXT://0.0.0.0:9092,CONTROLLER://0.0.0.0:9093
      KAFKA_ADVERTISED_LISTENERS: PLAINTEXT://kafka:9092
      KAFKA_CONTROLLER_LISTENER_NAMES: CONTROLLER
      KAFKA_LISTENER_SECURITY_PROTOCOL_MAP: CONTROLLER:PLAINTEXT,PLAINTEXT:PLAINTEXT
      KAFKA_CONTROLLER_QUORUM_VOTERS: 1@kafka:9093
      KAFKA_AUTO_CREATE_TOPICS_ENABLE: "true"
      KAFKA_NUM_PARTITIONS: 3
      KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
      CLUSTER_ID: MkU3OEVBNTcwNTJENDM2Qk
    volumes:
      - '${PROJECT_ROOT}'/.data/kafka:/tmp/kraft-combined-logs
    networks:
      - shared_network
    healthcheck:
      test: ["CMD-SHELL", "nc -z localhost 9092 || exit 1"]
      interval: 10s
      timeout: 5s
      retries: 15
    restart: unless-stopped' > docker-compose.yaml
    log "中间件 docker-compose 配置文件生成完成"
}

# 初始化数据库（DDL 使用 IF NOT EXISTS，可重复执行）
init_database() {
    if [ ! -f "ddl.sql" ]; then
        log "警告：ddl.sql文件不存在，跳过数据库初始化"
        return
    fi

    log "等待MySQL服务就绪..."
    local retries=60
    local MYSQL_CMD=""
    while [ $retries -gt 0 ]; do
        # 先尝试带密码连接
        if docker exec mysql mysql -uroot -p${MYSQL_ROOT_PASSWORD} -e "SELECT 1" &>/dev/null; then
            MYSQL_CMD="docker exec mysql mysql -uroot -p${MYSQL_ROOT_PASSWORD}"
            break
        fi
        # 再尝试无密码连接（user:0:0 模式下可能初始化为空密码）
        if docker exec mysql mysql -uroot -e "SELECT 1" &>/dev/null; then
            MYSQL_CMD="docker exec mysql mysql -uroot"
            log "MySQL 以空密码启动，正在设置密码和用户..."
            docker exec mysql mysql -uroot -e "
                ALTER USER 'root'@'localhost' IDENTIFIED BY '${MYSQL_ROOT_PASSWORD}';
                CREATE USER IF NOT EXISTS 'root'@'%' IDENTIFIED BY '${MYSQL_ROOT_PASSWORD}';
                GRANT ALL PRIVILEGES ON *.* TO 'root'@'%' WITH GRANT OPTION;
                CREATE DATABASE IF NOT EXISTS ${MYSQL_DATABASE};
                CREATE USER IF NOT EXISTS '${MYSQL_USER}'@'%' IDENTIFIED BY '${MYSQL_PASSWORD}';
                GRANT ALL PRIVILEGES ON ${MYSQL_DATABASE}.* TO '${MYSQL_USER}'@'%';
                FLUSH PRIVILEGES;
            " 2>/dev/null
            MYSQL_CMD="docker exec mysql mysql -uroot -p${MYSQL_ROOT_PASSWORD}"
            break
        fi
        retries=$((retries - 1))
        sleep 3
    done

    if [ $retries -eq 0 ]; then
        log "警告：MySQL服务未就绪，跳过数据库初始化"
        return
    fi
    log "MySQL服务已就绪"

    log "执行数据库DDL..."
    docker exec -i mysql mysql -uroot -p${MYSQL_ROOT_PASSWORD} < ddl.sql
    if [ $? -eq 0 ]; then
        log "数据库初始化完成"
    else
        log "警告：数据库初始化失败，请手动执行：docker exec -i mysql mysql -uroot -p${MYSQL_ROOT_PASSWORD} < ddl.sql"
    fi

    # 确保业务用户可以正常连接
    log "验证业务用户连接..."
    local user_retries=10
    while [ $user_retries -gt 0 ]; do
        if docker exec mysql mysql -u${MYSQL_USER} -p${MYSQL_PASSWORD} -e "USE ${MYSQL_DATABASE}" &>/dev/null; then
            log "业务用户连接验证通过"
            break
        fi
        user_retries=$((user_retries - 1))
        sleep 2
    done
    if [ $user_retries -eq 0 ]; then
        log "警告：业务用户 ${MYSQL_USER} 连接验证失败，业务服务可能无法正常启动"
    fi
}

# ============================================================
# 命令实现
# ============================================================

# init: 初始化环境（安装 Docker/Compose + 拉取中间件镜像）
do_init() {
    log "开始初始化环境..."

    # ---- 1. 安装 Docker ----
    if command -v docker &>/dev/null; then
        log "✅ Docker 已安装: $(docker --version)"
    else
        log "🔧 开始安装 Docker..."
        if [ -f /etc/os-release ]; then
            source /etc/os-release
            case "$ID" in
                ubuntu|debian)
                    apt-get update
                    apt-get install -y ca-certificates curl gnupg lsb-release
                    install -m 0755 -d /etc/apt/keyrings
                    curl -fsSL "https://download.docker.com/linux/$ID/gpg" | gpg --dearmor -o /etc/apt/keyrings/docker.gpg
                    chmod a+r /etc/apt/keyrings/docker.gpg
                    echo "deb [arch=$(dpkg --print-architecture) signed-by=/etc/apt/keyrings/docker.gpg] https://download.docker.com/linux/$ID $(lsb_release -cs) stable" > /etc/apt/sources.list.d/docker.list
                    apt-get update
                    apt-get install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
                    ;;
                centos|rhel|fedora|rocky|almalinux)
                    yum install -y yum-utils
                    yum-config-manager --add-repo https://download.docker.com/linux/centos/docker-ce.repo
                    yum install -y docker-ce docker-ce-cli containerd.io docker-buildx-plugin docker-compose-plugin
                    ;;
                *)
                    log "⚠️  不支持的 Linux 发行版: $ID，尝试使用官方安装脚本..."
                    curl -fsSL https://get.docker.com | sh
                    ;;
            esac
        elif [[ "$(uname)" == "Darwin" ]]; then
            log "❌ macOS 请手动安装 Docker Desktop: https://www.docker.com/products/docker-desktop"
            exit 1
        else
            log "⚠️  未知系统，尝试使用官方安装脚本..."
            curl -fsSL https://get.docker.com | sh
        fi

        # 启动 Docker 服务
        if command -v systemctl &>/dev/null; then
            systemctl enable docker
            systemctl start docker
        fi

        if command -v docker &>/dev/null; then
            log "✅ Docker 安装成功: $(docker --version)"
        else
            log "❌ Docker 安装失败，请手动安装"
            exit 1
        fi
    fi

    # ---- 2. 安装 Docker Compose ----
    if docker compose version &>/dev/null; then
        log "✅ Docker Compose (plugin) 已安装: $(docker compose version --short)"
    elif command -v docker-compose &>/dev/null; then
        log "✅ Docker Compose (standalone) 已安装: $(docker-compose --version)"
    else
        log "🔧 开始安装 Docker Compose..."
        if [ -f /etc/os-release ]; then
            source /etc/os-release
            case "$ID" in
                ubuntu|debian)
                    apt-get install -y docker-compose-plugin 2>/dev/null
                    ;;
                centos|rhel|fedora|rocky|almalinux)
                    yum install -y docker-compose-plugin 2>/dev/null
                    ;;
            esac
        fi

        # 如果 plugin 安装失败，回退到 standalone 方式
        if ! docker compose version &>/dev/null; then
            log "Plugin 方式不可用，安装 standalone 版本..."
            local COMPOSE_VERSION="v2.29.2"
            local OS=$(uname -s | tr '[:upper:]' '[:lower:]')
            local ARCH=$(uname -m)
            # 构建下载 URL（macOS 为 darwin-aarch64/darwin-x86_64，Linux 为 linux-aarch64/linux-x86_64）
            local DOWNLOAD_URL="https://github.com/docker/compose/releases/download/${COMPOSE_VERSION}/docker-compose-${OS}-${ARCH}"
            local INSTALL_DIR="$HOME/bin"
            mkdir -p "$INSTALL_DIR"
            log "下载 docker-compose: $DOWNLOAD_URL"
            curl -fsSL "$DOWNLOAD_URL" -o "$INSTALL_DIR/docker-compose"
            chmod +x "$INSTALL_DIR/docker-compose"
            export PATH="$INSTALL_DIR:$PATH"
        fi

        if docker compose version &>/dev/null || command -v docker-compose &>/dev/null; then
            log "✅ Docker Compose 安装成功"
        else
            log "❌ Docker Compose 安装失败，请手动安装"
            exit 1
        fi
    fi

    # ---- 3. 拉取中间件镜像 ----
    pull_middleware_images

    log "✅ 环境初始化完成！"
    log "💡 后续步骤："
    log "   构建镜像: ./tbos_docker.sh build"
    log "   启动服务: ./tbos_docker.sh start"
}

# build: 构建 Docker 镜像
# 用法: ./tbos_docker.sh build [模块名] [平台]
#   ./tbos_docker.sh build              构建所有模块
#   ./tbos_docker.sh build web          只构建 web 模块
#   ./tbos_docker.sh build cgi linux/amd64  构建 cgi 模块并指定平台
do_build() {
    local target_module=""
    local platform_arg=""

    # 解析参数：如果第一个参数不是平台格式（linux/xxx），则视为模块名
    if [[ -n "$1" && "$1" != linux/* ]]; then
        target_module="$1"
        platform_arg="$2"
    else
        platform_arg="$1"
    fi

    PLATFORM=${platform_arg:-"linux/$(uname -m | sed 's/x86_64/amd64/;s/aarch64/arm64/')"}

    # 如果指定了模块名，验证是否在白名单中
    if [ -n "$target_module" ]; then
        local found=false
        for modname in "${INCLUDE_DIRS[@]}"; do
            if [ "$modname" = "$target_module" ]; then
                found=true
                break
            fi
        done
        if [ "$found" = false ]; then
            log "❌ 未知模块: $target_module"
            log "可用模块: ${INCLUDE_DIRS[*]}"
            exit 1
        fi
    fi

    local modules_to_build=()
    if [ -n "$target_module" ]; then
        modules_to_build=("$target_module")
        log "开始构建 Docker 镜像: $target_module（目标平台: $PLATFORM）..."
    else
        modules_to_build=("${INCLUDE_DIRS[@]}")
        log "开始构建所有 Docker 镜像（目标平台: $PLATFORM）..."
    fi

    for modname in "${modules_to_build[@]}"; do
        log "Building Docker image for $modname ..."
        if [ "$modname" = "web" ]; then
            if [ -d web ]; then
                (cd web && docker build -t web -t "web:latest" --platform "$PLATFORM" .)
            else
                log "Warning: web directory not found, skipping web module build"
            fi
        else
            (docker build -t "$modname" -t "$modname:latest" --platform "$PLATFORM" --build-arg MODULE="$modname" .)
        fi
        log "Docker image for $modname built successfully."
    done
    log "✅ 镜像构建完成"
}

# package: 将已构建的镜像导出为 tar 并打包为发布包
do_package() {
    log "开始打包 Docker 镜像..."
    files_to_pack=(tbos_docker.sh server.cfg ddl.sql)
    for modname in "${INCLUDE_DIRS[@]}"; do
        log "Saving image: $modname ..."
        docker save "$modname" -o "target/$modname.tar"
        files_to_pack+=("target/$modname.tar")
    done
    log "Docker images saved successfully."
    # 创建发布包
    tar -czvf tbos_image.tar.gz "${files_to_pack[@]}"
    log "✅ 发布包创建完成: tbos_image.tar.gz"
}

# load: 从 tar 文件导入 Docker 镜像
do_load() {
    log "开始导入 Docker 镜像..."
    local images_dir="./target"
    if [ ! -d "$images_dir" ]; then
        log "镜像目录 $images_dir 不存在，请确认已解压发布包。"
        exit 1
    fi

    for tar_file in "$images_dir"/*.tar; do
        if [ -f "$tar_file" ]; then
            log "正在加载镜像: $tar_file"
            docker load < "$tar_file" || { log "加载镜像 $tar_file 失败。"; exit 1; }
        fi
    done
    log "✅ Docker 镜像导入完成"
}

# 根据配置生成所有 docker-compose 文件
generate_compose_config() {
    # 创建共享网络
    docker network create shared_network 2>/dev/null || true

    # 生成 .env 文件（供容器 env_file 使用）
    log "生成 .env 文件..."
    cp "$PROJECT_ROOT/server.cfg" .env
    sed -i.bak '/^[[:space:]]*#/d;/^[[:space:]]*$/d' .env && rm -f .env.bak
    # 容器间通信需要使用容器名而非 127.0.0.1
    sed -i.bak 's/MYSQL_ADDR=127.0.0.1:/MYSQL_ADDR=mysql:/g' .env && rm -f .env.bak
    sed -i.bak 's/REDIS_ADDR=127.0.0.1:/REDIS_ADDR=redis:/g' .env && rm -f .env.bak
    sed -i.bak 's/KAFKA_ADDR=127.0.0.1:/KAFKA_ADDR=kafka:/g' .env && rm -f .env.bak
    sed -i.bak 's/INFLUXDB_ADDR=127.0.0.1:/INFLUXDB_ADDR=influxdb:/g' .env && rm -f .env.bak
    log ".env 文件生成完成"

    # 生成中间件 docker-compose 配置
    generate_middleware_compose

    # 生成业务服务 docker-compose 配置
    log "生成业务服务 docker-compose 配置文件..."
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
      - "'${PORT_CMDB}':'${PORT_CMDB}'"
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
      - "'${PORT_SCHEDULER}':'${PORT_SCHEDULER}'"
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
      - "'${PORT_COLLECTOR}':'${PORT_COLLECTOR}'"
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
      - "'${PORT_ALARM_COMPUTE}':'${PORT_ALARM_COMPUTE}'"
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
      - "'${PORT_ALARM_MANAGE}':'${PORT_ALARM_MANAGE}'"
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
      - "'${PORT_ALARM_SERVER}':'${PORT_ALARM_SERVER}'"
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
      - "'${PORT_DATA_CACHE}':'${PORT_DATA_CACHE}'"
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
      - "'${PORT_DATA_COMPUTE}':'${PORT_DATA_COMPUTE}'"
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
      - "'${PORT_DATA_QUERY}':'${PORT_DATA_QUERY}'"
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
      - "'${PORT_DATA_STORE}':'${PORT_DATA_STORE}'"
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
      - "'${PORT_CGI}':'${PORT_CGI}'"
    networks:
      - shared_network
    volumes:
      - /etc/localtime:/etc/localtime:ro
      - /etc/timezone:/etc/timezone:ro
    restart: unless-stopped

  dac:
    image: docker.io/library/dac:latest
    container_name: dac
    env_file:
      - .env
    ports:
      - "'${PORT_DAC}':'${PORT_DAC}'"
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
    ports:
      - "'${PORT_WEB}':'${PORT_WEB}'"
    networks:
      - shared_network
    restart: unless-stopped' > tbos-docker-compose.yaml
    log "docker-compose 配置文件生成完成"
}

# start: 启动系统
# 用法: ./tbos_docker.sh start [模块名]
#   ./tbos_docker.sh start          启动整个系统
#   ./tbos_docker.sh start cgi      只启动/重启 cgi 服务
do_start() {
    local target_module="$1"

    # 如果指定了模块名，只启动/重启该服务
    if [ -n "$target_module" ]; then
        # 验证模块名
        local found=false
        for modname in "${INCLUDE_DIRS[@]}"; do
            if [ "$modname" = "$target_module" ]; then
                found=true
                break
            fi
        done
        if [ "$found" = false ]; then
            log "❌ 未知模块: $target_module"
            log "可用模块: ${INCLUDE_DIRS[*]}"
            exit 1
        fi

        # 加载配置并确保 compose 文件存在
        read_config
        if [ ! -f "tbos-docker-compose.yaml" ]; then
            generate_compose_config
        fi

        log "重启服务: $target_module ..."
        run_compose -f tbos-docker-compose.yaml up -d --force-recreate "$target_module"
        log "✅ 服务 $target_module 已启动"
        return
    fi

    # 未指定模块名，启动整个系统
    log "启动TBOS系统..."

    # 1. 加载配置
    read_config

    # 2. 根据配置生成 docker-compose 文件（中间件 + 业务服务）
    generate_compose_config

    # 3. 启动中间件
    log "启动中间件服务..."
    run_compose -f docker-compose.yaml up -d

    # 4. 初始化数据库（DDL 使用 IF NOT EXISTS，幂等安全）
    init_database

    # 5. 启动业务服务
    log "启动业务服务..."
    run_compose -f tbos-docker-compose.yaml up -d
    
    log "等待服务启动..."
    sleep 10
    
    log "检查服务状态..."
    run_compose -f tbos-docker-compose.yaml ps
    
    log "✅ 部署完成！"
    log "🌐 你可以通过以下地址访问你的 Web 应用："
    log ""
    log "    http://<你的服务器IP>:${PORT_WEB:-8080}"
    log ""
    log "💡 注意事项："
    log "1. 如果你使用的是云服务器，请确保安全组/防火墙开放了对应端口。"
    log "2. 如果你配置了域名，请替换 IP 为你的域名访问。"
    log "3. 如果服务未启动，请检查服务状态：docker ps"
}

# stop: 停止系统
# 用法: ./tbos_docker.sh stop [模块名]
#   ./tbos_docker.sh stop          停止整个系统
#   ./tbos_docker.sh stop cgi      只停止 cgi 服务
do_stop() {
    local target_module="$1"

    # 如果指定了模块名，只停止该服务
    if [ -n "$target_module" ]; then
        # 验证模块名
        local found=false
        for modname in "${INCLUDE_DIRS[@]}"; do
            if [ "$modname" = "$target_module" ]; then
                found=true
                break
            fi
        done
        if [ "$found" = false ]; then
            log "❌ 未知模块: $target_module"
            log "可用模块: ${INCLUDE_DIRS[*]}"
            exit 1
        fi

        if [ ! -f "tbos-docker-compose.yaml" ]; then
            log "❌ tbos-docker-compose.yaml 不存在，请先执行 start"
            exit 1
        fi

        log "停止服务: $target_module ..."
        run_compose -f tbos-docker-compose.yaml stop "$target_module"
        run_compose -f tbos-docker-compose.yaml rm -f "$target_module"
        log "✅ 服务 $target_module 已停止"
        return
    fi

    # 未指定模块名，停止整个系统
    log "停止TBOS服务..."
    run_compose -f tbos-docker-compose.yaml down -v

    if [ -f "docker-compose.yaml" ]; then
        log "停止中间件服务（保留数据）..."
        run_compose -f docker-compose.yaml down
    fi

    log "✅ 服务已停止"
}

# clean: 清除所有数据（停止服务 + 删除数据目录 + 删除生成的配置文件）
do_clean() {
    log "⚠️  即将清除所有数据（包括数据库、Redis、Kafka 数据）..."
    read -p "确认清除所有数据？(y/N): " confirm
    if [[ "$confirm" != "y" && "$confirm" != "Y" ]]; then
        log "已取消"
        exit 0
    fi

    # 先停止所有服务
    if [ -f "tbos-docker-compose.yaml" ]; then
        log "停止业务服务..."
        run_compose -f tbos-docker-compose.yaml down -v 2>/dev/null || true
    fi
    if [ -f "docker-compose.yaml" ]; then
        log "停止中间件服务..."
        run_compose -f docker-compose.yaml down -v 2>/dev/null || true
    fi

    # 删除数据目录
    local DATA_DIR="$PROJECT_ROOT/.data"
    if [ -d "$DATA_DIR" ]; then
        log "删除数据目录: $DATA_DIR"
        rm -rf "$DATA_DIR"
    fi

    # 删除生成的配置文件
    rm -f "$PROJECT_ROOT/docker-compose.yaml"
    rm -f "$PROJECT_ROOT/tbos-docker-compose.yaml"
    rm -f "$PROJECT_ROOT/.env"

    log "✅ 所有数据已清除"
}

# ============================================================
# 入口
# ============================================================
case "$1" in
    "init")
        do_init
        ;;
    "build")
        do_build "$2" "$3"
        ;;
    "package")
        do_package
        ;;
    "load")
        do_load
        ;;
    "start")
        do_start "$2"
        ;;
    "stop")
        do_stop "$2"
        ;;
    "clean")
        do_clean
        ;;
    *)
        echo "Usage: $0 {init|build|package|load|start|stop|clean}"
        echo ""
        echo "  init              初始化环境（安装Docker/Compose + 拉取中间件镜像）"
        echo "  build [模块名] [platform]  构建 Docker 镜像（不指定模块名则构建全部，可选指定平台）"
        echo "  package           将已构建的镜像导出为 tar 发布包"
        echo "  load              从 tar 文件导入 Docker 镜像到本地"
        echo "  start [模块名]    启动系统（不指定模块名则启动全部，指定则只启动/重启该服务）"
        echo "  stop  [模块名]    停止系统（不指定模块名则停止全部，指定则只停止该服务）"
        echo "  clean             清除所有数据（停止服务 + 删除数据库/缓存/消息队列数据）"
        exit 1
        ;;
esac