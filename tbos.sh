#!/bin/bash
set -e

PROJECT_ROOT=$(cd "$(dirname "$0")"; pwd)
mkdir -p target
TARGET_DIR="$PROJECT_ROOT/target"

# 只编译这些模块（白名单）
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

# 加载配置函数
load_config() {
    # 统一加载server.cfg配置到环境变量
    if [ -f "server.cfg" ]; then
        echo "Loading server.cfg configuration to env..."
        # 读取server.cfg中的所有配置到环境变量
        while IFS='=' read -r key value; do
            # 跳过注释行和空行
            [[ "$key" =~ ^# ]] || [[ -z "$key" ]] && continue
            # 设置环境变量（如果值为空字符串，则不设置，让后续的默认值生效）
            if [ -n "$value" ]; then
                export "$key"="$value"
            fi
        done < "server.cfg"
    else
        echo "server.cfg not found, using default configuration."
    fi
    
    # 设置所有配置的默认值（如果server.cfg中没有配置或配置为空字符串）
    export LOCAL_IP="${LOCAL_IP:-127.0.0.1}"
    export PORT_AGENT="${PORT_AGENT:-9100}"
    export PORT_COLLECTOR="${PORT_COLLECTOR:-9101}"
    export PORT_CMDB="${PORT_CMDB:-9102}"
    export PORT_SCHEDULER="${PORT_SCHEDULER:-9103}"
    export PORT_DATA_COMPUTE="${PORT_DATA_COMPUTE:-9104}"
    export PORT_DATA_CACHE="${PORT_DATA_CACHE:-9105}"
    export PORT_DATA_STORE="${PORT_DATA_STORE:-9106}"
    export PORT_DATA_QUERY="${PORT_DATA_QUERY:-9107}"
    export PORT_ALARM_COMPUTE="${PORT_ALARM_COMPUTE:-9108}"
    export PORT_ALARM_MANAGE="${PORT_ALARM_MANAGE:-9109}"
    export PORT_ALARM_SERVER="${PORT_ALARM_SERVER:-9110}"
    export PORT_CGI="${PORT_CGI:-9111}"
    export PORT_DAC="${PORT_DAC:-9113}"
    export PORT_WEB="${PORT_WEB:-8080}"
    export POD_IP="${POD_IP:-0.0.0.0}"
}

# 启动模块
start_modules() {
    load_config
    for modname in "${INCLUDE_DIRS[@]}"; do
        # 特殊处理web前端模块（使用nginx启动以支持路由规则）
        if [ "$modname" = "web" ]; then
            echo "Starting web frontend with nginx..."
            if [ ! -d "$TARGET_DIR/web/dist" ]; then
                echo "ERROR: Web frontend dist directory not found. Please run './tbos.sh build' first."
                exit 1
            fi
            
            # 检查nginx是否安装
            if ! command -v nginx &> /dev/null; then
                echo "ERROR: nginx is not installed. Cannot start web frontend."
                echo "Please install nginx to run the web frontend with routing support:"
                echo "  - macOS: brew install nginx"
                echo "  - Ubuntu/Debian: sudo apt install nginx"
                echo "  - CentOS/RHEL: sudo yum install nginx"
                exit 1
            fi
            
            # 检查是否已有nginx在运行
            running_pids=$(ps aux | grep "nginx:" | grep -v grep | awk '{print $2}' || true)
            if [ -n "$running_pids" ]; then
                echo "Found running nginx with PIDs: $running_pids. Stopping nginx..."
                nginx -s stop
                sleep 2
            fi
            
            # 复制nginx配置模板
            cp "$PROJECT_ROOT/web/nginx.conf" "$TARGET_DIR/web/"
            
            # 使用正则表达式替换所有__VAR__格式的变量
            # 获取所有环境变量，匹配__VAR__格式并进行替换
            while IFS='=' read -r var_name var_value; do
                # 跳过非环境变量行和空行
                [[ -z "$var_name" ]] || [[ "$var_name" =~ ^[[:space:]]*# ]] && continue
                # 替换__VAR_NAME__格式的占位符
                sed -i.bak "s|\${${var_name}}|${var_value}|g" "$TARGET_DIR/web/nginx.conf"
            done < <(env | grep -E '^(PORT_WEB|PORT_CGI|PORT_CMDB|PORT_DAC)=')
            
            # 配置本机nginx静态页面
            cp -r "$TARGET_DIR/web/dist/main/adaptor.html" "$TARGET_DIR/web/dist/main/index.html"

            if [ -d "/usr/share/nginx/html/tnebula" ]; then
                rm -rf "/usr/share/nginx/html/tnebula"
            fi
            mkdir -p /usr/share/nginx/html/tnebula/main
            cp -r "$TARGET_DIR/web/dist/main/." "/usr/share/nginx/html/tnebula/main/"

            # 启动nginx
            nginx -c "$TARGET_DIR/web/nginx.conf"
            
            # 等待服务器启动并检查
            sleep 3
            if curl -s http://localhost:$PORT_WEB > /dev/null; then
                echo "Web frontend started successfully on http://localhost:$PORT_WEB with nginx routing"
            else
                echo "WARNING: Web frontend may not have started properly. Check nginx error logs for details."
            fi
            continue
        fi
        
        # 精确匹配进程的命令行参数
        running_pids=$(ps aux | grep "[.]/$modname" | grep -v grep | awk '{print $2}' || true)
        if [ -n "$running_pids" ]; then
            echo "Found running $modname with PIDs: $running_pids. Killing all running instances..."
            kill $running_pids
            sleep 3
        fi
        echo "starting $modname ..."
        if [ ! -f "$TARGET_DIR/$modname/$modname" ]; then
            echo "$modname binary in $TARGET_DIR/$modname not found."
            exit 1
        fi
        (cd "$TARGET_DIR/$modname" && nohup ./$modname > "server.log" 2>&1 &)
        # 启动日志检查进程
        (sleep 3 && if grep -q "panic" "$TARGET_DIR/$modname/server.log"; then
            echo "ERROR: $modname failed to start with panic:"
            cat "$TARGET_DIR/$modname/server.log"
            exit 1
        fi)
        echo "$modname started successfully."
    done
}
# 停止模块
stop_modules() {
    for modname in "${INCLUDE_DIRS[@]}"; do
        # 特殊处理web前端模块（停止nginx进程）
        if [ "$modname" = "web" ]; then
            echo "Stopping web frontend nginx server..."
            # 停止nginx
            if command -v nginx &> /dev/null; then
                nginx -s stop 2>/dev/null || true
                sleep 2
                # 确保nginx进程完全停止
                running_pids=$(ps aux | grep "nginx:" | grep -v grep | awk '{print $2}' || true)
                if [ -n "$running_pids" ]; then
                    echo "Force killing remaining nginx processes with PIDs: $running_pids"
                    kill -9 $running_pids 2>/dev/null || true
                fi
                echo "Web frontend nginx server stopped."
            else
                echo "nginx is not installed, skipping web frontend stop."
            fi
            continue
        fi
        
        # 精确匹配进程的命令行参数
        running_pids=$(ps aux | grep "[.]/$modname" | grep -v grep | awk '{print $2}' || true)
        if [ -n "$running_pids" ]; then
            echo "Found running $modname with PIDs: $running_pids. Killing all running instances..."
            kill $running_pids
            echo "$modname stopped."
        else
            echo "$modname is not running."
        fi
    done
}

# 构建模块
build_modules() {
    mkdir -p "$TARGET_DIR"
    for modname in "${INCLUDE_DIRS[@]}"; do
        dir="$PROJECT_ROOT/$modname"
        if [ -d "$dir" ]; then
            echo "begin build $modname ..."
            mkdir -p "$TARGET_DIR/$modname"
            
            # 特殊处理web前端模块
            if [ "$modname" = "web" ]; then
                echo "Building web frontend (Node.js project)..."
                (cd "$dir" && npm install && npm run public-tedge)
                cp -r "$dir/dist" "$TARGET_DIR/web/"
            else
                # 后端Go模块构建
                (cd "$dir" && go mod tidy && CGO_ENABLED=0 go build -o "$TARGET_DIR/$modname/$modname")
                if [[ -f "$dir/trpc_go.yaml" ]]; then
                    cp "$dir/trpc_go.yaml" "$TARGET_DIR/$modname/"
                fi
            fi
            
            echo "success build $modname ..."
        fi
    done
    echo "all modules build successfully in $TARGET_DIR"
}

# 打包模块
package_modules() {
    echo "Building modules ..."
    build_modules
    echo "Packaging all files into tbos.tar.gz ..."
    files_to_pack=(tbos.sh server.cfg ddl.sql)
    for modname in "${INCLUDE_DIRS[@]}"; do
        # 特殊处理web前端模块（打包dist目录）
        if [ "$modname" = "web" ]; then
            if [ -d "target/web/dist" ]; then
                files_to_pack+=("target/web/dist")
            fi
        else
            # 后端模块打包可执行文件和配置文件
            files_to_pack+=("target/$modname/$modname")
            if [[ -f "target/$modname/trpc_go.yaml" ]]; then
                files_to_pack+=("target/$modname/trpc_go.yaml")
            fi
        fi
    done
    tar -czvf tbos.tar.gz "${files_to_pack[@]}"
    echo "Package created: tbos.tar.gz"
}

case "$1" in
    "build")
        build_modules
        ;;
    "package")
        package_modules
        ;;
    "start")
        start_modules
        ;;
    "stop")
        stop_modules
        ;;
    *)
        echo "Usage: $0 {build|start|stop|package}"
        exit 1
        ;;
esac
