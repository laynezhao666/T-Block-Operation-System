// Package chd806d4 实现CHD806D4门禁控制器协议的驱动层。
package chd806d4

import (
	"dac/entity/consts"
	"dac/entity/model/driver"
	"fmt"
	"net"

	"dac/entity/model/rt"
	"dac/entity/utils"
	consts2 "dac/logic/collect/driver/chd806d4/consts"

	"dac/entity/utils/ttime"
)

// ============ 门状态操作 ============

// DoorWorkState 门工作状态（工作状态字节解析）
type DoorWorkState struct {
	ClockError   bool // D7: =1 实时钟IC不正常，需要重新设置时间
	HasEvent     bool // D6: =1 DCU有事件要SU处理
	PowerError   bool // D5: =1 工作电源不正常，电压低
	TamperSwitch bool // D4: 保留（防拆开关）
	MonitorIR    bool // D3: =1 监视红外入侵
	MonitorDoor  bool // D2: =1 监视门开关状态
	RelayOn      bool // D1: =1 门控电磁继电器加电驱动
	AlarmState   bool // D0: =1 处于报警状态
}

// DoorLineState 门线路状态（线路状态字节解析）
type DoorLineState struct {
	EmergencyInput bool // D7: =1 紧急驱动输入
	NormallyClose  bool // D6: =1 常闭门
	NormallyOpen   bool // D5: =1 常开门
	Duress         bool // D4: =1 胁迫
	DoorOpen       bool // D3: =1 门开的; =0 门闭合
	IRAlarm        bool // D2: =1 窃入红外报警
	ExitButton     bool // D1: =1 出门放行键按下
	Reserved       bool // D0: 保留
}

// DoorStateInfo 门状态信息
type DoorStateInfo struct {
	DoorNo    int           // 门编号
	WorkState DoorWorkState // 工作状态
	LineState DoorLineState // 线路状态
}

// parseWorkState 解析门工作状态字节
func parseWorkState(b byte) DoorWorkState {
	return DoorWorkState{
		ClockError:   (b & 0x80) != 0, // D7
		HasEvent:     (b & 0x40) != 0, // D6
		PowerError:   (b & 0x20) != 0, // D5
		TamperSwitch: (b & 0x10) != 0, // D4
		MonitorIR:    (b & 0x08) != 0, // D3
		MonitorDoor:  (b & 0x04) != 0, // D2
		RelayOn:      (b & 0x02) != 0, // D1
		AlarmState:   (b & 0x01) != 0, // D0
	}
}

// parseLineState 解析门线路状态字节
func parseLineState(b byte) DoorLineState {
	return DoorLineState{
		EmergencyInput: (b & 0x80) != 0, // D7
		NormallyClose:  (b & 0x40) != 0, // D6
		NormallyOpen:   (b & 0x20) != 0, // D5
		Duress:         (b & 0x10) != 0, // D4
		DoorOpen:       (b & 0x08) != 0, // D3
		IRAlarm:        (b & 0x04) != 0, // D2
		ExitButton:     (b & 0x02) != 0, // D1
		Reserved:       (b & 0x01) != 0, // D0
	}
}

// getAllDoorStates 读取所有门状态（底层协议接口）
// 协议：COMMAND TYPE=0x8F, DATAF=0（1字节）
// 返回：DATAINFO（8字节）
//   - 第1字节：门1工作状态
//   - 第2字节：门1线路状态
//   - 第3字节：门2工作状态
//   - 第4字节：门2线路状态
//   - 第5字节：门3工作状态
//   - 第6字节：门3线路状态
//   - 第7字节：门4工作状态
//   - 第8字节：门4线路状态
func (c *Controller) getAllDoorStates() ([]DoorStateInfo, error) {
	// 构建 DATAF（1字节）= 0
	dataf := []byte{0x00}

	// 发送读取命令
	// CID2 = 0x4A (读取信息命令)
	// COMMAND GROUP = 0x03 (读取信息组)
	// COMMAND TYPE = 0x8F (读取门状态/远程监控)
	respInfo, err := c.Server.Request(
		consts2.CID2ReadInfo,
		consts2.GroupRead,
		consts2.TypeDoorStateGet,
		dataf,
		consts2.GetRRPCDoorGetState,
	)
	if err != nil {
		return nil, fmt.Errorf("读取门状态失败: %w", err)
	}

	// 解析响应（8字节）
	// 每个门占2字节：工作状态 + 线路状态
	if len(respInfo) < 8 {
		return nil, fmt.Errorf("门状态响应数据长度不足: 期望>=8字节, 实际=%d字节", len(respInfo))
	}

	states := make([]DoorStateInfo, 4)
	for i := 0; i < 4; i++ {
		workStateIdx := i * 2   // 工作状态字节索引: 0, 2, 4, 6
		lineStateIdx := i*2 + 1 // 线路状态字节索引: 1, 3, 5, 7
		states[i] = DoorStateInfo{
			DoorNo:    i + 1, // 门编号1-4
			WorkState: parseWorkState(respInfo[workStateIdx]),
			LineState: parseLineState(respInfo[lineStateIdx]),
		}
	}

	return states, nil
}

// GetDoorState 获取门状态（对外接口）
// 返回 map[门号]*rt.Point，Point 中包含门状态信息
func (c *Controller) GetDoorState(doors []int) (map[int]*rt.Point, error) {
	if c.Server == nil || !c.Server.IsConnected() {
		return nil, fmt.Errorf("未连接到门控器")
	}

	// 检查权限
	if err := c.checkAuth(); err != nil {
		return nil, err
	}

	// 一次性获取所有4个门的状态
	allStates, err := c.getAllDoorStates()
	if err != nil {
		return nil, fmt.Errorf("获取门状态失败: %w", err)
	}

	// 构建状态映射（门号 -> 状态）
	stateMap := make(map[int]DoorStateInfo)
	for _, state := range allStates {
		stateMap[state.DoorNo] = state
	}

	points := make(map[int]*rt.Point, len(doors))
	t := ttime.GetNowUTC()

	for _, doorNo := range doors {
		// 检查门号有效性
		if doorNo < 1 || doorNo > 4 {
			return nil, fmt.Errorf("无效的门编号: %d (有效范围: 1-4)", doorNo)
		}

		state, ok := stateMap[doorNo]
		if !ok {
			return nil, fmt.Errorf("未找到门%d的状态", doorNo)
		}

		// 将门状态转换为简单值：1=开，0=关
		// 根据 LineState.DoorOpen 判断门的开关状态
		var doorStatus int
		if state.LineState.DoorOpen {
			doorStatus = 1 // 门开
		} else {
			doorStatus = 0 // 门关
		}

		p := new(rt.Point)
		p.ID = utils.GenerateDoorStateID(c.baseInfo.ID, doorNo)
		p.SetValueWithTime(doorStatus, t.UnixMilli())
		points[doorNo] = p
	}

	return points, nil
}

// GetRawDoorState 获取原始门状态（对外接口）
// 返回 []DoorStateInfo，包含详细的门状态信息
func (c *Controller) GetRawDoorState(doors []int) (interface{}, error) {
	if c.Server == nil || !c.Server.IsConnected() {
		return nil, fmt.Errorf("未连接到门控器")
	}

	// 检查权限
	if err := c.checkAuth(); err != nil {
		return nil, err
	}

	// 一次性获取所有4个门的状态
	allStates, err := c.getAllDoorStates()
	if err != nil {
		return nil, fmt.Errorf("获取门状态失败: %w", err)
	}

	// 如果未指定门号，返回所有门的状态
	if len(doors) == 0 {
		return allStates, nil
	}

	// 构建状态映射
	stateMap := make(map[int]DoorStateInfo)
	for _, state := range allStates {
		stateMap[state.DoorNo] = state
	}

	// 筛选指定门的状态
	states := make([]DoorStateInfo, 0, len(doors))
	for _, doorNo := range doors {
		// 检查门号有效性
		if doorNo < 1 || doorNo > 4 {
			return nil, fmt.Errorf("无效的门编号: %d (有效范围: 1-4)", doorNo)
		}

		state, ok := stateMap[doorNo]
		if !ok {
			return nil, fmt.Errorf("未找到门%d的状态", doorNo)
		}
		states = append(states, state)
	}

	return states, nil
}

// ============ 门状态控制 ============

// getOperatorInfo 获取操作员信息（5字节）
// 使用系统IP的后4字节 + 填充，方便追溯操作来源
func getOperatorInfo() []byte {
	operatorInfo := make([]byte, 5)

	// 尝试获取系统IP
	if consts.ServiceIP != "" {
		ip := net.ParseIP(consts.ServiceIP)
		if ip != nil {
			ipv4 := ip.To4()
			if ipv4 != nil {
				// 使用 IPv4 的 4 字节
				copy(operatorInfo, ipv4)
				operatorInfo[4] = 0x00 // 填充最后一字节
				return operatorInfo
			}
		}
	}

	// 如果无法获取IP，使用默认标识 "PLAT\x00"
	copy(operatorInfo, []byte{'P', 'L', 'A', 'T', 0x00})
	return operatorInfo
}

// SetDoorState 设置门状态
// 根据 driver.DoorStateType 选择不同的协议命令：
//   - StateOpen (开门): 协议 4.6.2 COMMAND TYPE=0x8B, DATAF=6字节
//   - StateClose (常闭门): 协议 4.10.1 COMMAND TYPE=0x90, DATAF=8字节
//   - StateNormallyOpen (常开门): 协议 4.10.2 COMMAND TYPE=0x91, DATAF=8字节
//   - StateNormallyClose (常闭门): 同 StateClose，使用 0x90
//
// 操作员信息使用系统IP，方便追溯远程开门操作来源
func (c *Controller) SetDoorState(doorStates driver.SetDoorStateRequest) error {
	if c.Server == nil || !c.Server.IsConnected() {
		return fmt.Errorf("未连接到门控器")
	}

	// 检查权限
	if err := c.checkAuth(); err != nil {
		return err
	}

	// 获取操作员信息（系统IP）
	operatorInfo := getOperatorInfo()

	// 遍历每个门的状态设置请求
	for doorNo, stateType := range doorStates {
		// 检查门号有效性
		if doorNo < 1 || doorNo > 4 {
			return fmt.Errorf("无效的门编号: %d (有效范围: 1-4)", doorNo)
		}

		var err error
		switch stateType {
		case driver.StateOpen:
			// 开门：协议 4.6.2 远程放行（带系统操作员信息）
			// COMMAND TYPE=0x8B, DATAF（6字节）= 门号(1字节) + 操作员(5字节)
			err = c.remoteOpenDoor(byte(doorNo), operatorInfo)

		case driver.StateClose:
			// 恢复正常状态：需要同时解除常开或常闭状态
			if err = c.setDoorNormallyOpen(byte(doorNo), 0, operatorInfo); err != nil {
				return fmt.Errorf("设置门%d解除常开失败: %w", doorNo, err)
			}
			// 再解除常闭
			if err = c.setDoorNormallyClose(byte(doorNo), 0, operatorInfo); err != nil {
				return fmt.Errorf("设置门%d解除常闭失败: %w", doorNo, err)
			}

		case driver.StateNormallyOpen:
			// 常开门：协议 4.10.2 远程常开门
			// COMMAND TYPE=0x91, DATAF（8字节）= 门号(1字节) + 延时(2字节) + 操作员(5字节)
			// 延时不等于0时为常开，这里使用较大值表示持续常开
			err = c.setDoorNormallyOpen(byte(doorNo), 0xFFFF, operatorInfo) // 延时=0xFFFF（约45天）

		case driver.StateNormallyClose:
			// 常闭门：协议 4.10.1 远程常闭门
			// COMMAND TYPE=0x90, DATAF（8字节）= 门号(1字节) + 延时(2字节) + 操作员(5字节)
			err = c.setDoorNormallyClose(byte(doorNo), 0xFFFF, operatorInfo) // 延时=0xFFFF（约45天）

		default:
			return fmt.Errorf("门%d: 不支持的门状态类型: %d", doorNo, stateType)
		}

		if err != nil {
			return fmt.Errorf("设置门%d状态失败: %w", doorNo, err)
		}
	}

	return nil
}

// remoteOpenDoor 远程放行（带系统操作员信息）
// 协议 4.6.2：COMMAND TYPE=0x8B, DATAF（6字节）
//   - 第1字节: 门号 (=1 开门1; =2 开门2; =4 开门4; =0xFF 全开; =0 不操作)
//   - 后5字节: 操作员编号信息
func (c *Controller) remoteOpenDoor(doorNo byte, operatorInfo []byte) error {
	// 构建 DATAF（6字节）= 门号(1字节) + 操作员(5字节)
	dataf := make([]byte, 6)
	dataf[0] = doorNo
	copy(dataf[1:], operatorInfo)

	_, err := c.Server.Request(
		consts2.CID2SetParameter,
		consts2.GroupSet,
		consts2.TypeDoorRemoteOpenWithOp,
		dataf,
		consts2.GetRRPCDoorRemoteOpenWithOp,
	)
	return err
}

// setDoorNormallyClose 远程常闭门与解除（带系统操作员信息）
// 协议 4.10.1：COMMAND TYPE=0x90, DATAF（8字节）
//   - 第1字节: 门号 (=1 操作门1; =2 操作门2; =4 操作门4; =0xFF 操作所有门; =0 不操作)
//   - 第2-3字节: 延时时间（单位分钟，低位在前，高位在后；=0 表示解除常闭）
//   - 后5字节: 操作员编号信息
func (c *Controller) setDoorNormallyClose(doorNo byte, delayMinutes uint16, operatorInfo []byte) error {
	// 构建 DATAF（8字节）= 门号(1字节) + 延时(2字节) + 操作员(5字节)
	dataf := make([]byte, 8)
	dataf[0] = doorNo
	dataf[1] = byte(delayMinutes & 0xFF)        // 低位在前
	dataf[2] = byte((delayMinutes >> 8) & 0xFF) // 高位在后
	copy(dataf[3:], operatorInfo)

	_, err := c.Server.Request(
		consts2.CID2SetParameter,
		consts2.GroupSet,
		consts2.TypeDoorNormallyCloseWithOp,
		dataf,
		consts2.GetRRPCDoorNormallyCloseWithOp,
	)
	return err
}

// setDoorNormallyOpen 远程常开门与解除（带系统操作员信息）
// 协议 4.10.2：COMMAND TYPE=0x91, DATAF（8字节）
//   - 第1字节: 门号 (=1 操作门1; =2 操作门2; =4 操作门4; =0xFF 操作所有门; =0 不操作)
//   - 第2-3字节: 延时时间（单位分钟，低位在前，高位在后；=0 表示解除常开）
//   - 后5字节: 操作员编号信息
func (c *Controller) setDoorNormallyOpen(doorNo byte, delayMinutes uint16, operatorInfo []byte) error {
	// 构建 DATAF（8字节）= 门号(1字节) + 延时(2字节) + 操作员(5字节)
	dataf := make([]byte, 8)
	dataf[0] = doorNo
	dataf[1] = byte(delayMinutes & 0xFF)        // 低位在前
	dataf[2] = byte((delayMinutes >> 8) & 0xFF) // 高位在后
	copy(dataf[3:], operatorInfo)

	_, err := c.Server.Request(
		consts2.CID2SetParameter,
		consts2.GroupSet,
		consts2.TypeDoorNormallyOpenWithOp,
		dataf,
		consts2.GetRRPCDoorNormallyOpenWithOp,
	)
	return err
}

// GetDoorPositionState 获取门位置状态（CHD 协议暂不支持）
func (c *Controller) GetDoorPositionState() (interface{}, error) {
	return nil, fmt.Errorf("chd806d4协议不支持 GetDoorPositionState 获取门位置状态操作")
}
