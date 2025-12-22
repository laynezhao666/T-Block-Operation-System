package network

const (
	enableSwitchShellFile = "/opt/tbbox/shells/enable_switch_network.sh"
	networkModeSwitch     = "switch"
	networkModeDefault    = ""
	bridge0Interface      = "br0"
	bond0Interface        = "bond0"
)

const (
	commandTemplate = `#!/usr/bin/env sh

set -e

bridge_lans="lan1 lan2 lan5 lan6 lan7 lan8 lan9 lan10"
# 桥接设置：
# 创建网桥
if ip link show br0 &> /dev/null; then
  echo "br0 existed"
else
  echo "create br0"
  ip link add name br0 type bridge
fi
for l in ${bridge_lans}; do
  # 启动对应网口
  ip link set ${l} up
  ip addr flush dev ${l}
  # 将网口加入网桥
  ip link set dev ${l} master br0
done

# 配置网桥参数
echo "flush br0 addr"
ip addr flush dev br0
echo "add br0 address"
ip addr add %v/%v dev br0
# 添加辅助 IP
ip addr add %v/%v dev br0
# 启动网桥
echo "set br0 up"
ip link set dev br0 up

# bond 设置
# 删除 bonding 模块
modprobe -r bonding
# 加载 bonding 模块，设置模式为 802.3ad
modprobe bonding mode=4 miimon=100 lacp_rate=fast
# 清空 bond0 ip
echo "flush bond0 addr"
ip addr flush dev bond0
# 添加 bond0 ip 和掩码
ip addr add %v/%v dev bond0

ip link set dev bond0 down

bond_lans="lan3 lan4"
# $l down 掉后加到 bond0
for l in ${bond_lans}; do
  ip addr flush dev ${l}
  ip link set dev ${l} down
  ip link set ${l} master bond0
done

ip link set dev bond0 up


# 修改默认路由
echo "replace bond0 default route"
ip route replace default via %v dev bond0
`
)
