package thttp

import (
	"bytes"
	"io"
	"mime/multipart"
	"os"

	"agent/utils/datastructure/set"
)

// UploadFileJSON 上传文件
func UploadFileJSON(url, method, field, file string, params map[string][]string, timeout int,
	dataPointer interface{}) error {
	return UploadFilesJSON(url, method, []string{field}, []string{file}, params, timeout, dataPointer)
}

// UploadFiles 上传文件
func UploadFiles(url, method string, fields, files []string, params map[string][]string, timeout int) ([]byte, error) {
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

// UploadFilesJSON 上传文件
func UploadFilesJSON(url, method string, fields, files []string, params map[string][]string, timeout int,
	dataPointer interface{}) error {
	b, err := UploadFiles(url, method, fields, files, params, timeout)
	if err != nil {
		return err
	}

	return parseJSONResult(dataPointer, b)
}
