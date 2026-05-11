package http

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"agent/entity/consts"
	"agent/utils"
	"agent/utils/osal"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"agent/entity/definition"
	model2 "agent/entity/model"
	"agent/logic/collector/device/model"
)

// quaField 配置qua字段名称在DriverExtend JSON中的key
const quaField = "qua_field"

type HTTPDevice struct {
	gid  definition.DeviceGidType
	name string

	timeout time.Duration // 默认 3s，可被 chanInfo.TimeoutMs 覆盖
	method  string        // "GET"|"POST"|"PUT"|"DELETE"
	path    string        // 以 '/' 开头
	host    string
	port    int
	quaName string

	client *http.Client
}

func NewHTTPDevice(gid definition.DeviceGidType, name string) *HTTPDevice {
	return &HTTPDevice{
		gid:     gid,
		name:    name,
		timeout: 3 * time.Second,
		method:  "GET",
		port:    80,
	}
}

func (d *HTTPDevice) Open(chanInfo model.ChannelInfo, _ model.ListCollectPackets) consts.Quality {
	if chanInfo.TimeoutMs > 0 {
		d.timeout = time.Duration(chanInfo.TimeoutMs) * time.Millisecond
	}
	d.method = chanInfo.Params

	d.path = strings.TrimSpace(chanInfo.Address)
	if d.path == "" {
		d.path = "/"
	}
	if !strings.HasPrefix(d.path, "/") {
		d.path = "/" + d.path
	}
	// 替换为tbos的接口地址
	if d.path == "/cgi/rtd" {
		d.path = "/cgi/rtdById"
	}

	host, port, err := parseHostPortFromName(chanInfo.Name)
	if err != nil {
		return consts.QualityConfigError
	}
	d.host, d.port = host, port

	d.client = &http.Client{Timeout: d.timeout}
	d.parseChanExtend(chanInfo)
	return consts.QualityOk
}

// parseChanExtend 从通道的DriverExtend JSON中解析qua字段名称
func (d *HTTPDevice) parseChanExtend(chanInfo model.ChannelInfo) {
	if len(chanInfo.DriverExtend) == 0 {
		return
	}

	var result map[string]interface{}
	if err := json.Unmarshal([]byte(chanInfo.DriverExtend), &result); err != nil {
		return
	}

	if v, ok := result[quaField]; ok {
		if s, ok := v.(string); ok {
			d.quaName = s
		}
	}
}

func (d *HTTPDevice) Close() consts.Quality {
	return consts.QualityOk
}

func (d *HTTPDevice) Request(ctx context.Context, packet *model.CollectProtocolPacket) (consts.Quality, model2.MessageStatistics) {
	if packet == nil {
		return consts.QualityUncertain, model2.MessageStatistics{}
	}

	var body map[string]any
	if strings.TrimSpace(packet.Command) != "" {
		if err := json.Unmarshal([]byte(packet.Command), &body); err != nil {
			return consts.QualityConfigError, model2.MessageStatistics{}
		}
	}

	reqURL, err := d.buildURL()
	if err != nil {
		return consts.QualityConfigError, model2.MessageStatistics{}
	}

	respJSON := make(map[string]any)
	if err := d.doJSON(ctx, d.method, reqURL, body, &respJSON); err != nil {
		return consts.QualityCmdRespError, model2.MessageStatistics{}
	}

	// 针对每个测点，用解析器键路径提取值并回填
	now := utils.GetNowUTCTimeStamp()
	for _, pt := range packet.Points {
		parser, _ := pt.Attr.ValParser.(*HTTPValueParser) // 由 CreateValParseObj 创建
		qua := d.parseAndFillPoint(pt, parser, respJSON, now)
		_ = qua // 后续需要统计不同质量可在此累加
	}

	return consts.QualityOk, model2.MessageStatistics{}
}

func (d *HTTPDevice) buildURL() (string, error) {
	if d == nil || strings.TrimSpace(d.host) == "" {
		return "", fmt.Errorf("empty device or host")
	}
	port := d.port
	if port <= 0 || port > 65535 {
		port = 80
	}
	p := strings.TrimSpace(d.path)
	if p == "" || p == "." {
		p = "/"
	}
	if !strings.HasPrefix(p, "/") && !strings.HasPrefix(p, "?") {
		p = "/" + p
	}
	// 若 path 误传了完整 URL，则直接用它
	lp := strings.ToLower(p)
	if strings.HasPrefix(lp, "http://") || strings.HasPrefix(lp, "https://") {
		return p, nil
	}
	return "http://" + net.JoinHostPort(d.host, strconv.Itoa(port)) + p, nil
}

func (d *HTTPDevice) RequestPing(ctx context.Context, packet model.CollectProtocolPacket) consts.Quality {
	// 最小化指令发送包：GET无body；POST/PUT 发送空 JSON {}
	u := url.URL{
		Scheme: "http",
		Host:   net.JoinHostPort(d.host, strconv.Itoa(d.port)),
		Path:   d.path,
	}

	var body map[string]any
	m := strings.ToUpper(d.method)
	if m == "POST" || m == "PUT" {
		body = map[string]any{}
	}

	var sink map[string]any
	if err := d.doJSON(ctx, d.method, u.String(), body, &sink); err != nil {
		return consts.QualityCmdRespError
	}
	return consts.QualityOk
}

func (d *HTTPDevice) Control(_ *model.ControlProtocolPacket, _ string) consts.Quality {
	return consts.QualityOk
}

// --- helpers ---

func (d *HTTPDevice) doJSON(ctx context.Context, method, urlStr string, body map[string]any, out any) error {
	var req *http.Request
	var err error

	switch strings.ToUpper(method) {
	case "GET", "DELETE":
		req, err = http.NewRequestWithContext(ctx, method, urlStr, nil)
	case "POST", "PUT":
		var b []byte
		if body != nil {
			b, err = json.Marshal(body)
			if err != nil {
				return err
			}
		} else {
			b = []byte("{}")
		}
		req, err = http.NewRequestWithContext(ctx, method, urlStr, strings.NewReader(string(b)))
		if err != nil {
			return err
		}
		req.Header.Set("Content-Type", "application/json")
	default:
		return fmt.Errorf("unsupported method: %s", method)
	}
	if err != nil {
		return err
	}

	resp, err := d.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(out); err != nil {
		return err
	}
	return nil
}

func (d *HTTPDevice) parseAndFillPoint(pt *model.PointInfo, parser *HTTPValueParser, root map[string]any, now int64) consts.Quality {
	ok := false
	var sub any
	qua := "0"
	if parser != nil {
		sub, qua, ok = getByKeys(root, parser.Keys, d.quaName)
	}

	if !ok {
		// 未命中：配置错误
		pt.RtVal.Qua = consts.QualityConfigError
		return pt.RtVal.Qua
	}

	switch v := sub.(type) {
	case string:
		pt.RtVal.Pv = osal.NewVariantWithValue(v)
	default:
		// 兼容C++版本实现 将子 JSON 转字符串（非字符串则 dump 成 JSON 字符串）
		b, _ := json.Marshal(v)
		pt.RtVal.Pv = osal.NewVariantWithValue(b)
	}
	pt.RtVal.Qua = consts.QualityOk
	if qua != "0" {
		q, err := strconv.Atoi(qua)
		if err == nil {
			pt.RtVal.Qua = consts.Quality(q)
		}
	}
	pt.RtVal.Tms = now
	return pt.RtVal.Qua
}

func parseHostPortFromName(name string) (string, int, error) {
	// name 可能带 "http://" 前缀；没写端口默认 80；端口非法时报配置错误
	s := strings.TrimSpace(name)
	if s == "" {
		return "", 0, errors.New("empty host")
	}
	// 去掉 scheme
	if i := strings.Index(s, "//"); i >= 0 {
		s = s[i+2:]
	}
	// 解析端口
	host := s
	port := 80
	if j := strings.LastIndex(s, ":"); j > 0 {
		host = s[:j]
		pstr := s[j+1:]
		p, err := strconv.Atoi(pstr)
		if err != nil || p <= 0 || p > 65535 {
			return "", 0, fmt.Errorf("invalid port: %s", pstr)
		}
		port = p
	}
	return host, port, nil
}

// getByKeys 支持：
// - map 取键：     a.b.c
// - 数组取下标：   data.0.value
// - 也支持括号式： data[0].value  或 items[1][2].name
func getByKeys(root any, keys []string, quaName string) (any, string, bool) {
	cur := root
	endIdx := len(keys) - 1
	qua := "0"
	for i, seg := range keys {
		seg = strings.TrimSpace(seg)
		if seg == "" {
			return nil, qua, false
		}

		// 允许诸如 "data[0][1]" 这种一段里带多级下标
		base, idxs, ok := parseIndexedSegment(seg)
		if !ok {
			return nil, qua, false
		}

		// 先在 map 上取 base（base 可能为空，表示直接在当前就是切片上取下标，如 "[0]"）
		if base != "" {
			m, mok := cur.(map[string]any)
			if !mok {
				return nil, qua, false
			}
			v, ex := m[base]
			if !ex {
				return nil, qua, false
			}
			cur = v
			// 如果配置了qua名且为最后一段，提取出来
			if len(quaName) > 0 && i == endIdx {
				quaStr, has := m[quaName]
				if has {
					if val, ok := quaStr.(string); ok {
						qua = val
					}
				}
			}
		}

		// 依次在切片上应用所有下标
		for _, idx := range idxs {
			arr, aok := cur.([]any)
			if !aok || idx < 0 || idx >= len(arr) {
				return nil, qua, false
			}
			cur = arr[idx]
		}
		// 如果本段既没有 base 也没有下标（理应不会发生），判失败
	}

	return cur, qua, true
}

// 解析一个段：支持 "key"、"key[0]"、"key[0][1]"、"[0]"、"0"（纯数字）
// 返回：base（可能为空）、索引链（可能为空）、是否解析成功
func parseIndexedSegment(seg string) (base string, idxs []int, ok bool) {
	s := seg

	// 纯数字：当作直接对当前切片取下标（等价于 base="" + idx）
	if i, e := strconv.Atoi(s); e == nil {
		return "", []int{i}, true
	}

	// 支持 [0] 开头的形式（等价于 base=""）
	if strings.HasPrefix(s, "[") {
		base = ""
		var rest = s
		for len(rest) > 0 {
			i, r, e := takeLeadingBracketIndex(rest)
			if e != nil {
				return "", nil, false
			}
			idxs = append(idxs, i)
			rest = r
			if rest == "" {
				break
			}
		}
		return base, idxs, true
	}

	// 常规的 "key" 或 "key[0][1]" 形式
	// 先取 base（直到第一个 '['）
	if p := strings.IndexByte(s, '['); p >= 0 {
		base = s[:p]
		rest := s[p:]
		for len(rest) > 0 {
			i, r, e := takeLeadingBracketIndex(rest)
			if e != nil {
				return "", nil, false
			}
			idxs = append(idxs, i)
			rest = r
			if rest == "" {
				break
			}
		}
		return base, idxs, true
	}

	// 单纯的 "key"
	return s, nil, true
}

// 取出前导的 "[number]"，返回 index、剩余字符串、错误
func takeLeadingBracketIndex(s string) (int, string, error) {
	if !strings.HasPrefix(s, "[") {
		return 0, "", fmt.Errorf("no leading '['")
	}
	end := strings.IndexByte(s, ']')
	if end <= 1 {
		return 0, "", fmt.Errorf("unclosed or empty bracket")
	}
	num := s[1:end]
	i, err := strconv.Atoi(strings.TrimSpace(num))
	if err != nil {
		return 0, "", err
	}
	return i, s[end+1:], nil
}
