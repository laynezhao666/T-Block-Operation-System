# -*- coding: utf-8 -*-
"""门禁服务(DAC)构建脚本。

支持以下功能：
- 编译Go二进制文件
- 构建Docker镜像
- 推送Docker镜像到仓库
- 保存镜像到本地文件

使用示例：
    python build.py -d          # 编译镜像
    python build.py -d -p       # 编译并推送镜像
    python build.py -d -o a.tar # 编译并保存镜像到本地
"""
import argparse
import os
import subprocess
import sys
import time


def run(cmd):
    """执行shell命令，失败时打印输出并退出。"""
    r = subprocess.run(cmd)
    if r.returncode == 0:
        return
    print("stdout", r.stdout)
    print("stderr", r.stderr)
    sys.exit(r.returncode)


def main():
    """解析命令行参数并执行构建流程。"""
    parser = argparse.ArgumentParser()
    parser.add_argument("-p", action="store_true", help="推送镜像")
    parser.add_argument("-d", action="store_true", help="编译镜像")
    parser.add_argument("-o", default="", help="保存镜像至本地")
    parser.add_argument("-v", help="版本", default="")
    parser.add_argument("-r", help="镜像仓库地址", default="")
    args = parser.parse_args()

    args.d |= args.p

    tag_version = args.v
    if len(tag_version) == 0:
        tag_version = time.strftime("%Y%m%d%H%M%S", time.localtime(int(time.time())))
    tag_name = "dcos.tdac"

    if len(args.r) > 0:
        tag = args.r.rstrip("/") + "/" + tag_name + ":" + tag_version
    else:
        tag = tag_name + ":" + tag_version

    # 设置交叉编译环境为Linux AMD64
    os.putenv("GOOS", "linux")
    os.putenv("GOARCH", "amd64")

    # 编译Go二进制文件
    run(["go", "build", "-v", "-o", "tdac", "main.go"])

    while True:
        if not args.d:
            break
        # 构建Docker镜像
        print("build image:", tag)
        run(["docker", "build", "-f", "./Dockerfile", "-t", "{tag}".format(tag=tag), "."])

        if len(args.o) > 0:
            # 保存镜像到本地文件
            run(["docker", "save", tag, "-o", args.o])

        if not args.p:
            break
        # 推送镜像到仓库
        print("push image:", tag)
        run(["docker", "push", "{tag}".format(tag=tag)])
        break


if __name__ == "__main__":
    main()
