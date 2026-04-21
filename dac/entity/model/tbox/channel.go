package tbox

type Channel struct {
	Type            string `json:"chtype"`
	ID              string `json:"chid"`
	Params          string `json:"chparams"`
	Address         string `json:"addr"`
	WaitTime        string `json:"waitTime"`     // 单位：毫秒
	CommandInterval string `json:"cmdInterval"`  // 单位：毫秒
	RequestTimeout  string `json:"timeout"`      // 单位：毫秒
	MaxFailCount    string `json:"maxFailCount"` // 单位：毫秒
	MaxFailTime     string `json:"maxFailTime"`  // 单位：毫秒
}

type ChannelRaw struct {
	Type            string `json:"chtype"`
	ID              string `json:"chid"`
	Params          string `json:"chparams"`
	Address         string `json:"addr"`
	WaitTime        string `json:"wait_time"`      // 单位：毫秒
	CommandInterval string `json:"cmd_interval"`   // 单位：毫秒
	RequestTimeout  string `json:"timeout"`        // 单位：毫秒
	MaxFailCount    string `json:"max_fail_count"` // 单位：毫秒
	MaxFailTime     string `json:"max_fail_time"`  // 单位：毫秒
}

func ChannelConvert(c Channel) ChannelRaw {
	return ChannelRaw{
		Type:            c.Type,
		ID:              c.ID,
		Params:          c.Params,
		Address:         c.Address,
		WaitTime:        c.WaitTime,
		CommandInterval: c.CommandInterval,
		RequestTimeout:  c.RequestTimeout,
		MaxFailCount:    c.MaxFailCount,
		MaxFailTime:     c.MaxFailTime,
	}
}
