package thttp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"

	"agent/utils/parse"
)

func getReader(requestBody interface{}) (io.Reader, error) {
	if requestBody == nil {
		return bytes.NewReader([]byte{}), nil
	}
	switch body := requestBody.(type) {
	case *bytes.Reader:
		return body, nil
	case bytes.Reader:
		return &body, nil
	case *bytes.Buffer:
		return body, nil
	case bytes.Buffer:
		return &body, nil
	default:
		break
	}

	var b []byte
	var err error
	switch body := requestBody.(type) {
	case string:
		b = []byte(body)
	case []byte:
		b = body
	default:
		if b, err = json.Marshal(body); err != nil {
			return nil, err
		}
	}
	return bytes.NewReader(b), nil
}

func parseJSONResult(dataPointer interface{}, responseBody []byte) error {
	var temp responseType
	if err := json.Unmarshal(responseBody, &temp); err != nil {
		return err
	}
	if temp.Code != 0 {
		return fmt.Errorf("code != 0, response: %+v", temp)
	}

	return parse.JSON(dataPointer, temp.Data)
}
