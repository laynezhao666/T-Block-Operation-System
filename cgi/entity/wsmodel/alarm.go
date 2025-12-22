package wsmodel

// ResponseData ResponseData
type ResponseData struct {
	Code      int         `json:"code"`
	Data      interface{} `json:"data"`
	Message   string      `json:"message"`
	Cmd       string      `json:"cmd"`
	Timestamp int64       `json:"timestamp"`
}

// DataMozu DataMozu
type DataMozu struct {
	MozuId int32  `json:"mozu_id"`
	Data   []byte `json:"data"`
}
