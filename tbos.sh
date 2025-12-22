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
)

# 启动模块
start_modules() {
    if [ ! -f "server.cfg" ]; then
        echo "server.cfg not found."
        exit 1
    fi
    for modname in "${INCLUDE_DIRS[@]}"; do
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
        (export $(cat server.cfg | grep -v '^#' | xargs) && cd "$TARGET_DIR/$modname" && nohup ./$modname > "server.log" 2>&1 &)
        # 启动日志检查进程
        (sleep 3 && if grep -q "panic" "$TARGET_DIR/$modname/server.log"; then
            echo "ERROR: $modname failed to start with panic:"
            cat "$TARGET_DIR/$modname/server.log"
            exit 1
        fi)
        echo "$modname started successfully."
    done
}

stop_modules() {
    for modname in "${INCLUDE_DIRS[@]}"; do
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
            (cd "$dir" && go mod tidy && CGO_ENABLED=0 go build -o "$TARGET_DIR/$modname/$modname")
            if [[ -f "$dir/trpc_go.yaml" ]]; then
                cp "$dir/trpc_go.yaml" "$TARGET_DIR/$modname/"
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
        files_to_pack+=("target/$modname/$modname" "target/$modname/trpc_go.yaml")
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