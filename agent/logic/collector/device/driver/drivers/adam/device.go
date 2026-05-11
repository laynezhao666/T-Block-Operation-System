package adam

import (
	"agent/entity/consts"
	"agent/entity/definition"
	model3 "agent/entity/model"
	"agent/logic/collector/device/model"
	"agent/utils"
	"agent/utils/encoding"
	"agent/utils/osal"
	"context"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"trpc.group/trpc-go/trpc-go/log"
)

// Device 实现 IDevice
type Device struct {
	gid  definition.DeviceGidType
	name string

	addrASCII [2]byte
	port      *SerialPort
}

// Open 外部会把 chanInfo.Address 传 "01"
func (d *Device) Open(chanInfo model.ChannelInfo, _ model.ListCollectPackets) consts.Quality {
	baudRate, parity, dataBits, stopBits, err := model.ParseRTUParam(chanInfo.Params)
	if err != nil {
		log.Errorf("RTU params parse err=%v, fallback 9600:N:8:1", err)
		baudRate, parity, dataBits, stopBits = 9600, "N", 8, 1
	}
	// 地址：支持 "1"/"01"
	addr := strings.ToUpper(strings.TrimSpace(chanInfo.Address))
	if addr == "" {
		log.Errorf("adam address empty")
		return consts.QualityConfigError
	}
	if len(addr) == 1 {
		addr = "0" + addr
	}
	if len(addr) != 2 {
		log.Errorf("adam address must be 2 ASCII, got %q", addr)
		return consts.QualityConfigError
	}
	d.addrASCII[0], d.addrASCII[1] = addr[0], addr[1]

	sp, err := OpenSerial(SerialOptions{
		Port:       chanInfo.Name,
		Baud:       baudRate,
		DataBits:   dataBits,
		StopBits:   stopBits,
		Parity:     model.NormalizeParity(parity),
		ReadTO:     time.Duration(utils.FirstNonZero(chanInfo.TimeoutMs, 3000)) * time.Millisecond,
		WriteTO:    2 * time.Second,
		HardwareFC: false,
		SoftwareFC: false,
	})
	if err != nil {
		log.Errorf("adam open serial err=%v", err)
		return consts.QualityCannotOpen
	}
	d.port = sp
	return consts.QualityOk
}

func (d *Device) Close() consts.Quality {
	if d.port != nil {
		_ = d.port.Close()
		d.port = nil
	}
	return consts.QualityOk
}

// Request packet.Command
func (d *Device) Request(ctx context.Context, packet *model.CollectProtocolPacket) (consts.Quality, model3.MessageStatistics) {
	var stat model3.MessageStatistics
	if d.port == nil {
		return consts.QualityCommDisconnected, stat
	}

	cmdChar := byte('6')
	if c := strings.TrimSpace(packet.Command); c != "" {
		cmdChar = c[0]
	}
	req := buildAdamRequest(d.addrASCII, cmdChar)
	//stat.SendPackets = hexSp(req)

	if err := d.port.WriteAll(req); err != nil {
		log.Errorf("adam write err=%v", err)
		stat.ErrLog = err
		return consts.QualityCmdSendError, stat
	}

	rx, err := d.port.ReadUntilCR(ctx, 512)
	if err != nil {
		log.Errorf("adam read err=%v", err)
		stat.ErrLog = err
		return consts.QualityCmdRespError, stat
	}
	stat.RecvPackets = encoding.HexSp(rx)

	parsed, err := parseAdamResponse(rx)
	if err != nil {
		log.Errorf("adam parse err=%v", err)
		return consts.QualityCmdRespError, stat
	}

	// 映射到测点
	if err := fillPointsFromPayload(parsed.Data, packet); err != nil {
		log.Warnf("adam map points warn: %v", err)
	}

	return consts.QualityOk, stat
}

func (d *Device) RequestPing(ctx context.Context, pkt model.CollectProtocolPacket) consts.Quality {
	q, _ := d.Request(ctx, &pkt)
	return q
}

func (d *Device) Control(_ *model.ControlProtocolPacket, _ string) consts.Quality {
	return consts.QualityOk
}

// —— 内部 —— //

func buildAdamRequest(addr [2]byte, cmd byte) []byte {
	//例如 "$" + AA + '6' + CR
	return []byte{0x24, addr[0], addr[1], cmd, 0x0D}
}

type adamParsed struct {
	PayloadASCII string
	Data         []byte
}

func parseAdamResponse(rx []byte) (*adamParsed, error) {
	n := len(rx)
	if n < 2 || rx[0] != '!' || rx[n-1] != 0x0D {
		return nil, fmt.Errorf("bad frame: % X", rx)
	}
	ascii := string(rx[1 : n-1])
	if len(ascii)%2 != 0 {
		return nil, fmt.Errorf("odd ascii len=%d", len(ascii))
	}
	data := make([]byte, hex.DecodedLen(len(ascii)))
	if _, err := hex.Decode(data, []byte(ascii)); err != nil {
		return nil, fmt.Errorf("hex decode: %w", err)
	}
	return &adamParsed{PayloadASCII: ascii, Data: data}, nil
}

func fillPointsFromPayload(payload []byte, packet *model.CollectProtocolPacket) error {
	if packet == nil || len(packet.Points) == 0 {
		return nil
	}
	now := utils.GetNowUTCTimeStamp()

	for _, pt := range packet.Points {
		if pt == nil || pt.Attr.ValParser == nil {
			continue
		}
		vp, ok := pt.Attr.ValParser.(*ValueParser)
		if !ok {
			continue
		}
		val, err := vp.ReadFrom(payload)
		if err != nil {
			// 单点失败不影响其它点
			continue
		}

		// ext function: not..
		val = utils.Unary(vp.Extend, val)

		pt.RtVal.Pv = osal.NewVariantWithValue(val)
		pt.RtVal.Tms = now
		pt.RtVal.Qua = consts.QualityOk
	}
	return nil
}
