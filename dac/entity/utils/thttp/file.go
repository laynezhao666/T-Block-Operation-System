// Package thttp 提供HTTP请求工具函数，支持文件上传和JSON解析。
package thttp

import (
	"bytes"
	"io"
	"mime/multipart"
	"os"

	"dac/entity/utils/set"
)

// UploadFileJSON 上传单个文件并将响应解析为JSON
func UploadFileJSON(url, method, field, file string,
	params map[string][]string, timeout int,
	dataPointer interface{},
) error {
	return UploadFilesJSON(url, method, []string{field}, []string{file}, params, timeout, dataPointer)
}

// UploadFileData 上传单个文件的字节数据，返回原始响应
func UploadFileData(url, method string, field, file string,
	params map[string][]string, data []byte, timeout int,
) ([]byte, error) {
	var buff bytes.Buffer
	w := multipart.NewWriter(&buff)
	defer func() { _ = w.Close() }()

	var err error
	writer, err := w.CreateFormFile(field, file)
	if err != nil {
		return nil, err
	}

	if _, err = io.CopyN(writer, bytes.NewReader(data), int64(len(data))); err != nil {
		return nil, err
	}

	for k, values := range params {
		if field == k {
			continue
		}
		for _, v := range values {
			if err = w.WriteField(k, v); err != nil {
				return nil, err
			}
		}
	}

	if err = w.Close(); err != nil {
		return nil, err
	}

	return Request(url, method, map[string][]string{"Content-Type": {w.FormDataContentType()}}, &buff, timeout)
}

// UploadFiles 批量上传文件，返回原始响应
func UploadFiles(url, method string, fields, files []string,
	params map[string][]string, timeout int,
) ([]byte, error) {
	var buff bytes.Buffer
	w := multipart.NewWriter(&buff)
	defer func() { _ = w.Close() }()

	fieldNum := len(fields)
	fileNum := len(files)
	fieldsMap := set.NewStringSet()
	for i := fieldNum; i < fileNum; i++ {
		fields = append(fields, fields[i-1])
	}
	fieldsMap.AddSlice(fields)

	var err error
	for i, file := range files {
		writer, err := w.CreateFormFile(fields[i], file)
		if err != nil {
			return nil, err
		}

		b, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		if _, err = io.CopyN(writer, bytes.NewReader(b), int64(len(b))); err != nil {
			return nil, err
		}
	}

	for k, values := range params {
		if fieldsMap.Contain(k) {
			continue
		}
		for _, v := range values {
			if err = w.WriteField(k, v); err != nil {
				return nil, err
			}
		}
	}

	if err = w.Close(); err != nil {
		return nil, err
	}

	return Request(url, method, map[string][]string{"Content-Type": {w.FormDataContentType()}}, &buff, timeout)
}

// UploadFilesJSON 批量上传文件并将响应解析为JSON
func UploadFilesJSON(url, method string, fields, files []string,
	params map[string][]string, timeout int,
	dataPointer interface{},
) error {
	b, err := UploadFiles(url, method, fields, files, params, timeout)
	if err != nil {
		return err
	}

	return parseJSONResult(dataPointer, b)
}

// UploadFilesData 上传多个内存文件数据，返回原始响应
func UploadFilesData(url, method string, field string,
	files map[string][]byte, params map[string][]string, timeout int,
) ([]byte, error) {
	var buff bytes.Buffer
	w := multipart.NewWriter(&buff)
	defer func() { _ = w.Close() }()

	var err error
	for name, data := range files {
		writer, err := w.CreateFormFile(field, name)
		if err != nil {
			return nil, err
		}

		if _, err = io.CopyN(writer, bytes.NewReader(data), int64(len(data))); err != nil {
			return nil, err
		}
	}

	for k, values := range params {
		if k == field {
			continue
		}
		for _, v := range values {
			if err = w.WriteField(k, v); err != nil {
				return nil, err
			}
		}
	}

	if err = w.Close(); err != nil {
		return nil, err
	}

	return Request(url, method, map[string][]string{"Content-Type": {w.FormDataContentType()}}, &buff, timeout)
}

// UploadFilesDataJSON 上传多个内存文件数据并将响应解析为JSON
func UploadFilesDataJSON(url, method string, field string,
	files map[string][]byte, params map[string][]string,
	timeout int, dataPointer interface{},
) error {
	b, err := UploadFilesData(url, method, field, files, params, timeout)
	if err != nil {
		return err
	}

	return parseJSONResult(dataPointer, b)
}
