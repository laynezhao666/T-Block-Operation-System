// Package thttp 提供HTTP请求体解析工具函数。
package thttp

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http"
)

// GetBody 从HTTP请求中提取请求体内容。
// 支持普通请求体、multipart表单和URL编码表单三种格式。
// 对于multipart表单，会重新编码文件和字段数据，并更新Content-Type头。
func GetBody(r *http.Request, h http.Header) ([]byte, error) {
	var (
		b   []byte
		err error
	)
	// 优先读取原始请求体
	if r.Body != nil {
		b, err = io.ReadAll(r.Body)
		if err != nil {
			return nil, err
		}
		if len(b) > 0 {
			return b, nil
		}
	}

	// GET请求返回空body
	if r.Method == http.MethodGet {
		return make([]byte, 0), nil
	}

	// 尝试解析multipart表单（最大100MB）
	if err = r.ParseMultipartForm(100 * (1 << 20)); err != nil {
		return nil, err
	}

	// 处理multipart表单数据
	if r.MultipartForm != nil {
		var buff bytes.Buffer
		w := multipart.NewWriter(&buff)
		defer func() {
			_ = w.Close()
		}()

		// 写入文件字段
		for key, files := range r.MultipartForm.File {
			for i := range files {
				writer, err := w.CreateFormFile(key, files[i].Filename)
				if err != nil {
					return nil, err
				}

				fp, err := files[i].Open()
				if err != nil {
					return nil, err
				}
				if _, err = io.CopyN(writer, fp, files[i].Size); err != nil {
					return nil, err
				}
			}
		}

		// 写入普通字段
		for key, values := range r.MultipartForm.Value {
			for _, v := range values {
				if err = w.WriteField(key, v); err != nil {
					return nil, err
				}
			}
		}

		if err = w.Close(); err != nil {
			return nil, err
		}

		// 更新Content-Type为multipart格式
		h.Set("Content-Type", w.FormDataContentType())

		b = buff.Bytes()

		if len(b) > 0 {
			return b, nil
		}
	}

	// 最后尝试URL编码表单
	if r.Form != nil {
		b = []byte(r.Form.Encode())
	}

	return b, nil
}
