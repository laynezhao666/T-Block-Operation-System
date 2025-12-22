package utils

import (
	"archive/zip"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"agent/entity/config"
	"agent/entity/consts"
	"agent/entity/errcode"
	logiccm "agent/logic/cm"
	"agent/repo/cm"
	tbosIo "agent/utils/file/io"

	"etrpc-go/log"

	"github.com/tealeg/xlsx/v3"
	"trpc.group/trpc-go/trpc-go"
	"trpc.group/trpc-go/trpc-go/codec"
	"trpc.group/trpc-go/trpc-go/errs"
	thttp "trpc.group/trpc-go/trpc-go/http"
)

const (
	tboxType             = 1
	notTboxType          = 3
	tboxSubDeviceType    = 2
	notTboxSubDeviceType = 4
)

// ProcessImportDevice 处理导入的设备，tbox或厂商设备
func ProcessImportDevice(fileHeaders []*multipart.FileHeader) error {
	if config.GetRB().Project.Source != "local" {
		time.Sleep(1 * time.Second)
		return fmt.Errorf("当前不是local模式，无法执行该操作")
	}
	devices := make([]deviceInfo, 0)
	for _, fileHeader := range fileHeaders {
		fileName := fileHeader.Filename
		file, err := fileHeader.Open()
		if err != nil {
			return fmt.Errorf("open flieHeader %s error: %s", fileName, err.Error())
		}
		xlsxFile, err := xlsx.OpenReaderAt(file, fileHeader.Size)
		if err != nil {
			return fmt.Errorf("open xlsx file %s error: %s", fileName, err.Error())
		}
		if xlsxFile == nil || len(xlsxFile.Sheets) == 0 {
			_ = file.Close()
			return fmt.Errorf("file %s is empty!", fileName)
		}
		for _, sheet := range xlsxFile.Sheets {
			fatherDevice, err := parseFatherDevice(sheet)
			if err != nil {
				log.Errorf("parse father device error: %s", err.Error())
				continue
			}
			subDevices, err := parseSubDevices(sheet)
			if err != nil {
				log.Errorf("parse sub device error: %s", err.Error())
			}
			subType := notTboxSubDeviceType
			if fatherDevice.CollectorType == tboxType {
				subType = tboxSubDeviceType
			}
			for idx, d := range subDevices {
				subDevices[idx].CollectorType = subType
				parts := strings.Split(d.Tpl.TplNm, "-") // 数据模板列格式：设备类型-厂商-设备编号.xlsx
				if len(parts) >= 3 {
					templateClass := parts[0]  // 设备类型
					templateVendor := parts[1] // 厂商
					subDevices[idx].Tpl.TplPath = fmt.Sprintf("%s/%s", templateClass, templateVendor)
				}
			}
			fatherDevice.SubDevices = subDevices
			devices = append(devices, *fatherDevice)
		}
		_ = file.Close()
	}
	if err := dumpDeviceToJson(devices); err != nil {
		log.Errorf("dump device to json error: %s", err.Error())
		return err
	}
	select {
	case cm.ConfigChangedChan() <- true:
	case <-time.After(time.Millisecond * 500):
		log.Warn("changed channel持续阻塞，暂不更新设备列表")
	}

	return nil
}

// 把设备信息写到json
func dumpDeviceToJson(devices []deviceInfo) (err error) {
	// 1. 加载本地文件
	var content *localFile
	filePath, err := getLatestDevicePath()
	if err != nil {
		log.Errorf("get latest device path error: %s", err.Error())
		return err
	}
	err = tbosIo.JSON.Read(filePath, &content)
	if err != nil {
		log.Errorf("load local devices file error: %s", err.Error())
		return err
	}
	// 2. 写device到json，再同步到文件
	deviceMap := map[string]deviceInfo{}
	if _, ok := content.Data.ConfigMap["default"]; !ok {
		content.Data.ConfigMap["default"] = deviceInfo{CollectorType: tboxType}
	}
	for _, d := range content.Data.ConfigMap["default"].SubDevices {
		deviceMap[d.DeviceCode] = d
	}
	for _, device := range devices {
		for _, d := range device.SubDevices {
			deviceMap[d.DeviceCode] = d
		}
	}
	deviceList := []deviceInfo{}
	for _, device := range deviceMap {
		deviceList = append(deviceList, device)
	}
	content.Data.ConfigMap["default"] = deviceInfo{
		CollectorType: tboxType,
		SubDevices:    deviceList,
		DeviceCode:    "default",
		DeviceName:    "default",
	}
	err = tbosIo.JSON.Write(filePath, content)
	if err != nil {
		log.Errorf("dump devices to json error: %s", err.Error())
		return err
	}
	return nil
}

// parseFatherDevice 解析父设备, title第2行，数据第3行
func parseFatherDevice(sheet *xlsx.Sheet) (*deviceInfo, error) {
	titleRow, dataRow := 1, 2 // idx从0开始
	deviceID, err := getValue(sheet, "设备编号", titleRow, 0, dataRow, 0)
	if err != nil {
		return nil, fmt.Errorf("get device id error: %s", err.Error())
	}
	deviceType, err := getValue(sheet, "设备类型/名称", titleRow, 2, dataRow, 2)
	if err != nil {
		return nil, fmt.Errorf("get device type error: %s", err.Error())
	}
	channelType, err := getValue(sheet, "通道类型", titleRow, 3, dataRow, 3)
	if err != nil {
		return nil, fmt.Errorf("get channel type error: %s", err.Error())
	}
	channelID, err := getValue(sheet, "通道ID", titleRow, 4, dataRow, 4)
	if err != nil {
		return nil, fmt.Errorf("get channel id error: %s", err.Error())
	}
	addr, err := getValue(sheet, "从机地址", titleRow, 5, dataRow, 5)
	if err != nil {
		return nil, fmt.Errorf("get channel addr error: %s", err.Error())
	}
	channelParams, err := getValue(sheet, "通道参数", titleRow, 6, dataRow, 6)
	if err != nil {
		return nil, fmt.Errorf("get channel params error: %s", err.Error())
	}
	templateName, err := getValue(sheet, "驱动模板", titleRow, 7, dataRow, 7)
	if err != nil {
		return nil, fmt.Errorf("get template name error: %s", err.Error())
	}
	waitTime, err := getValue(sheet, "等待时间", titleRow, 8, dataRow, 8)
	if err != nil {
		return nil, fmt.Errorf("get wait time error: %s", err.Error())
	}
	cmdInterval, err := getValue(sheet, "命令间隔", titleRow, 9, dataRow, 9)
	if err != nil {
		return nil, fmt.Errorf("get cmd interval error: %s", err.Error())
	}
	timeout, err := getValue(sheet, "请求超时", titleRow, 10, dataRow, 10)
	if err != nil {
		return nil, fmt.Errorf("get timeout error: %s", err.Error())
	}
	maxFailCount, err := getValue(sheet, "最大失败次数", titleRow, 11, dataRow, 11)
	if err != nil {
		return nil, fmt.Errorf("get max fail count error: %s", err.Error())
	}
	maxFailTime, err := getValue(sheet, "最大失败时间", titleRow, 12, dataRow, 12)
	if err != nil {
		return nil, fmt.Errorf("get max fail time error: %s", err.Error())
	}
	concurrentNum, _ := getValue(sheet, "并发数量", titleRow, 13, dataRow, 13)
	maxPointNum, _ := getValue(sheet, "请求最大测点数", titleRow, 14, dataRow, 14)
	extendParams, _ := getValue(sheet, "扩展参数", titleRow, 15, dataRow, 15)
	isInstantiation, _ := getValue(sheet, "是否实例化", titleRow, 16, dataRow, 16)
	// var dtype int = tboxType
	// if !isTbox(deviceType) {
	// 	dtype = notTboxType
	// }
	device := formatDevice(deviceID, deviceType, deviceID, channelType, channelID, addr, channelParams, templateName,
		waitTime, cmdInterval, timeout, maxFailCount, maxFailTime, tboxType, 0, concurrentNum, maxPointNum,
		extendParams, isInstantiation)
	return device, nil
}

// parseSubDevices 解析子设备, title第6行，数据第7行
func parseSubDevices(sheet *xlsx.Sheet) (devices []deviceInfo, err error) {
	devices = make([]deviceInfo, 0)
	var p devicePoint
	startRow := 6
	for dr := startRow; dr < sheet.MaxRow; dr++ {
		row, err := sheet.Row(dr)
		if err != nil {
			return nil, fmt.Errorf("get device info error: %s, row: %d", err.Error(), dr)
		}
		if err = row.ReadStruct(&p); err != nil {
			return nil, err
		}
		if p.DeviceID == "" {
			continue // 避免读到空行
		}
		// 设备类型和gid在后续补充，其他字段从p读取
		d := formatDevice(p.DeviceID, p.DeviceName, p.DeviceID, p.ChannelType, p.ChannelID, p.ChannelAddr,
			p.ChannelParams, p.ValueTemplate, p.WaitTime, p.CmdInterval, p.Timeout, p.MaxFailCount, p.MaxFailTime,
			0, 0, p.ConcurrentNum, p.MaxPointNum, p.ExtendParams, p.IsInstantiation)
		devices = append(devices, *d)
	}
	return devices, nil
}

func formatDevice(deviceID, deviceName, deviceNumber, chanType, chanID, chanAddr, chanParams, tplName, waitTime,
	cmdInterval, timeout, maxFailCount, maxFailTime string, deviceType, mozuID int, concurrentNum, maxPointNum,
	extendParams, isInstantiation string) *deviceInfo {
	device := &deviceInfo{
		Channel: channel{
			Addr:         chanAddr,
			ChID:         chanID,
			ChParams:     chanParams,
			ChType:       chanType,
			CmdInterval:  cmdInterval,
			MaxFailCount: maxFailCount,
			MaxFailTime:  maxFailTime,
			Timeout:      timeout,
			WaitTime:     waitTime,
		},
		CollectorType: deviceType,
		DeviceCode:    deviceID,
		DeviceGid:     string(logiccm.Worker().GetNextDeviceGid()),
		DeviceName:    deviceName,
		DeviceNumber:  deviceNumber,
		DeviceTypeEn:  "",
		DeviceTypeZh:  "",
		MozuID:        mozuID,
		Tpl: tpl{
			TplNm:   tplName,
			TplPath: "",
		},
		SubDevices:      nil,
		ConcurrentNum:   concurrentNum,
		MaxPointNum:     maxPointNum,
		ExtendParams:    extendParams,
		IsInstantiation: isInstantiation,
	}
	return device
}

func isTbox(deviceType string) bool {
	return strings.HasPrefix(deviceType, "TBOX")
}

// ExportAllDevices 导出所有设备，并响应下载
func ExportAllDevices(ctx context.Context) (err error) {
	// 1. 加载所有设备
	devices := make(map[string]deviceInfo)
	var content *localFile
	filePath, err := getLatestDevicePath()
	if err != nil {
		return errs.New(errcode.ErrMissLocalFile, fmt.Sprintf("读取本地设备失败：%s", err.Error()))
	}
	err = tbosIo.JSON.Read(filePath, &content)
	if err != nil {
		return errs.New(errcode.ErrMissLocalFile, fmt.Sprintf("读取本地设备失败：%s", err.Error()))
	}
	for dk, device := range content.Data.ConfigMap {
		devices[dk] = device
	}

	// 2. 遍历local json，写入xlsx
	tmpDir, err := os.MkdirTemp("", "devices-*")
	if err != nil {
		return errs.New(errcode.ErrMissLocalFile, fmt.Sprintf("创建临时目录失败：%s", err.Error()))
	}
	defer os.RemoveAll(tmpDir)
	tplFilePath := filepath.Join(consts.ProjectPath, consts.EmptyDevicesXlsx)

	// tplFilePath := filepath.Join(consts.DeployPath, "config", "local", consts.ElvdbTemplateXlsx)
	tplFile, err := os.Open(tplFilePath)
	if err != nil {
		return errs.New(errcode.ErrMissLocalFile, fmt.Sprintf("open template xlsx file failed: %v", err))
	}
	tplData, err := io.ReadAll(tplFile)
	if err != nil {
		log.Errorf("load local elvdb template file failed: %s", err.Error())
		return errs.New(errcode.ErrMissLocalFile, fmt.Sprintf("load template xlsx file failed: %v", err))
	}
	_ = tplFile.Close()
	for dk, d := range devices {
		newFilePath := filepath.Join(tmpDir, fmt.Sprintf("%s.xlsx", dk))
		err = os.WriteFile(newFilePath, tplData, 0666)
		if err != nil {
			log.Errorf("复制模板excel文件失败：%s", err.Error())
			continue
		}
		xlsxFile, err := xlsx.OpenFile(newFilePath)
		if err != nil {
			log.Errorf("open local device file failed: %s", err.Error())
			continue
		}
		err = formatXlsx(&d, xlsxFile)
		if err != nil {
			log.Errorf("format device file failed: %s", err.Error())
			continue
		}
		err = xlsxFile.Save(newFilePath)
		if err != nil {
			log.Errorf("save device file failed: %s", err.Error())
			continue
		}
	}

	// 3. 压缩后响应
	targetFile := filepath.Join(os.TempDir(), "devices.zip")
	defer os.Remove(targetFile)
	err = zipFolder(tmpDir, targetFile)
	if err != nil {
		return errs.New(errcode.ErrServerLogic, fmt.Sprintf("zip device file failed: %s", err.Error()))
	}
	f, err := os.Open(targetFile)
	if err != nil {
		return errs.New(errcode.ErrMissLocalFile, fmt.Sprintf("读取设备列表失败：%s", err.Error()))
	}
	defer f.Close()
	msg := trpc.Message(ctx)
	msg.WithSerializationType(codec.SerializationTypeUnsupported)
	head := thttp.Head(ctx)
	head.Response.Header().Set("Content-Type", "application/zip")
	head.Response.Header().Set("Content-Disposition", "attachment; filename="+"devices.zip")
	_, _ = io.Copy(head.Response, f)
	return nil
}

// 把device信息写入xlsx
func formatXlsx(device *deviceInfo, file *xlsx.File) (err error) {
	if device == nil {
		return fmt.Errorf("device is nil")
	}
	if len(file.Sheets) == 0 {
		log.Errorf("no sheet in file")
		return fmt.Errorf("no sheet in file")
	}
	sheet := file.Sheets[0]
	err = formatXlsxFatherDevice(device, sheet)
	if err != nil {
		return
	}
	err = formatXlsxSubDevices(device, sheet)
	return err
}

// formatXlsxFatherDevice 将父设备信息写入xlsx
func formatXlsxFatherDevice(device *deviceInfo, sheet *xlsx.Sheet) (err error) {
	if device == nil {
		return fmt.Errorf("device is nil")
	}
	if sheet == nil {
		return fmt.Errorf("sheet is nil")
	}

	dataRow := 2
	if err := writeBasicInfo(device, sheet, dataRow); err != nil {
		return err
	}
	if err := writeChannelInfo(device, sheet, dataRow); err != nil {
		return err
	}
	if err := writeTemplateInfo(device, sheet, dataRow); err != nil {
		return err
	}
	if err := writeTimingInfo(device, sheet, dataRow); err != nil {
		return err
	}
	if err := writeFailureInfo(device, sheet, dataRow); err != nil {
		return err
	}
	if err := writeAdditionalInfo(device, sheet, dataRow); err != nil {
		return err
	}

	return nil
}

// writeBasicInfo 写入设备基本信息
func writeBasicInfo(device *deviceInfo, sheet *xlsx.Sheet, row int) error {
	// 设备编号
	if err := setCellValue(sheet, row, 0, device.DeviceCode); err != nil {
		return fmt.Errorf("write device code error: %w", err)
	}
	// 设备类型/名称
	if err := setCellValue(sheet, row, 2, device.DeviceName); err != nil {
		return fmt.Errorf("write device name error: %w", err)
	}
	return nil
}

// writeChannelInfo 写入通道相关信息
func writeChannelInfo(device *deviceInfo, sheet *xlsx.Sheet, row int) error {
	// 通道类型
	if err := setCellValue(sheet, row, 3, device.Channel.ChType); err != nil {
		return fmt.Errorf("write channel type error: %w", err)
	}
	// 通道ID
	if err := setCellValue(sheet, row, 4, device.Channel.ChID); err != nil {
		return fmt.Errorf("write channel ID error: %w", err)
	}
	// 从机地址
	if err := setCellValue(sheet, row, 5, device.Channel.Addr); err != nil {
		return fmt.Errorf("write channel address error: %w", err)
	}
	// 通道参数
	if err := setCellValue(sheet, row, 6, device.Channel.ChParams); err != nil {
		return fmt.Errorf("write channel params error: %w", err)
	}
	return nil
}

// writeTemplateInfo 写入模板信息
func writeTemplateInfo(device *deviceInfo, sheet *xlsx.Sheet, row int) error {
	// 驱动模板
	if err := setCellValue(sheet, row, 7, device.Tpl.TplNm); err != nil {
		return fmt.Errorf("write template name error: %w", err)
	}
	return nil
}

// writeTimingInfo 写入时间相关配置
func writeTimingInfo(device *deviceInfo, sheet *xlsx.Sheet, row int) error {
	// 等待时间
	if err := setCellValue(sheet, row, 8, device.Channel.WaitTime); err != nil {
		return fmt.Errorf("write wait time error: %w", err)
	}
	// 命令间隔
	if err := setCellValue(sheet, row, 9, device.Channel.CmdInterval); err != nil {
		return fmt.Errorf("write command interval error: %w", err)
	}
	// 超时时间
	if err := setCellValue(sheet, row, 10, device.Channel.Timeout); err != nil {
		return fmt.Errorf("write timeout error: %w", err)
	}
	return nil
}

// writeFailureInfo 写入失败处理配置
func writeFailureInfo(device *deviceInfo, sheet *xlsx.Sheet, row int) error {
	// 最大失败次数
	if err := setCellValue(sheet, row, 11, device.Channel.MaxFailCount); err != nil {
		return fmt.Errorf("write max fail count error: %w", err)
	}
	// 最大失败时间
	if err := setCellValue(sheet, row, 12, device.Channel.MaxFailTime); err != nil {
		return fmt.Errorf("write max fail time error: %w", err)
	}
	return nil
}

// writeAdditionalInfo 写入其他附加信息
func writeAdditionalInfo(device *deviceInfo, sheet *xlsx.Sheet, row int) error {
	// 并发数
	if err := setCellValue(sheet, row, 13, device.ConcurrentNum); err != nil {
		return fmt.Errorf("write concurrent num error: %w", err)
	}
	// 最大测点数
	if err := setCellValue(sheet, row, 14, device.MaxPointNum); err != nil {
		return fmt.Errorf("write max point num error: %w", err)
	}
	// 附加参数
	if err := setCellValue(sheet, row, 15, device.ExtendParams); err != nil {
		return fmt.Errorf("write extend params error: %w", err)
	}
	// 是否实例化
	if err := setCellValue(sheet, row, 16, device.IsInstantiation); err != nil {
		return fmt.Errorf("write instantiation flag error: %w", err)
	}
	return nil
}

// setCellValue 设置单元格值的通用函数
func setCellValue(sheet *xlsx.Sheet, row, col int, value interface{}) error {
	cell, err := sheet.Cell(row, col)
	if err != nil {
		return fmt.Errorf("cell row %d col %d error: %w", row, col, err)
	}
	cell.SetValue(value)
	return nil
}

// formatXlsxSubDevices 将子设备信息写入xlsx
func formatXlsxSubDevices(device *deviceInfo, sheet *xlsx.Sheet) (err error) {
	if device == nil {
		return fmt.Errorf("device is nil")
	}
	if sheet == nil {
		return fmt.Errorf("sheet is nil")
	}

	row := 6 // 子设备从第6行开始
	devices := device.SubDevices

	// 确保有足够的行
	for sheet.MaxRow < row+len(devices) {
		sheet.AddRow()
	}

	// 写入每个子设备信息
	for _, d := range devices {
		if err := writeSubDeviceInfo(d, sheet, row); err != nil {
			log.Errorf("write sub device info error: %v", err)
			continue
		}
		row++
	}
	return nil
}

// writeSubDeviceInfo 写入单个子设备信息
func writeSubDeviceInfo(d deviceInfo, sheet *xlsx.Sheet, row int) error {
	// 写入基本信息
	if err := setCellValue(sheet, row, 1, d.DeviceCode); err != nil {
		return fmt.Errorf("write device code error: %w", err)
	}
	if err := setCellValue(sheet, row, 2, d.DeviceName); err != nil {
		return fmt.Errorf("write device name error: %w", err)
	}

	// 写入通道信息
	if err := writeSubDeviceChannelInfo(d, sheet, row); err != nil {
		return fmt.Errorf("write channel info error: %w", err)
	}

	// 写入模板信息
	if err := setCellValue(sheet, row, 7, d.Tpl.TplNm); err != nil {
		return fmt.Errorf("write template name error: %w", err)
	}

	// 写入时间相关配置
	if err := writeSubDeviceTimingInfo(d, sheet, row); err != nil {
		return fmt.Errorf("write timing info error: %w", err)
	}

	// 写入失败处理配置
	if err := writeSubDeviceFailureInfo(d, sheet, row); err != nil {
		return fmt.Errorf("write failure info error: %w", err)
	}

	// 写入其他附加信息
	if err := writeSubDeviceAdditionalInfo(d, sheet, row); err != nil {
		return fmt.Errorf("write additional info error: %w", err)
	}

	return nil
}

// writeSubDeviceChannelInfo 写入子设备通道信息
func writeSubDeviceChannelInfo(d deviceInfo, sheet *xlsx.Sheet, row int) error {
	if err := setCellValue(sheet, row, 3, d.Channel.ChType); err != nil {
		return fmt.Errorf("write channel type error: %w", err)
	}
	if err := setCellValue(sheet, row, 4, d.Channel.ChID); err != nil {
		return fmt.Errorf("write channel ID error: %w", err)
	}
	if err := setCellValue(sheet, row, 5, d.Channel.Addr); err != nil {
		return fmt.Errorf("write channel address error: %w", err)
	}
	if err := setCellValue(sheet, row, 6, d.Channel.ChParams); err != nil {
		return fmt.Errorf("write channel params error: %w", err)
	}
	return nil
}

// writeSubDeviceTimingInfo 写入子设备时间相关配置
func writeSubDeviceTimingInfo(d deviceInfo, sheet *xlsx.Sheet, row int) error {
	if err := setCellValue(sheet, row, 8, d.Channel.WaitTime); err != nil {
		return fmt.Errorf("write wait time error: %w", err)
	}
	if err := setCellValue(sheet, row, 9, d.Channel.CmdInterval); err != nil {
		return fmt.Errorf("write command interval error: %w", err)
	}
	if err := setCellValue(sheet, row, 10, d.Channel.Timeout); err != nil {
		return fmt.Errorf("write timeout error: %w", err)
	}
	return nil
}

// writeSubDeviceFailureInfo 写入子设备失败处理配置
func writeSubDeviceFailureInfo(d deviceInfo, sheet *xlsx.Sheet, row int) error {
	if err := setCellValue(sheet, row, 11, d.Channel.MaxFailCount); err != nil {
		return fmt.Errorf("write max fail count error: %w", err)
	}
	if err := setCellValue(sheet, row, 12, d.Channel.MaxFailTime); err != nil {
		return fmt.Errorf("write max fail time error: %w", err)
	}
	return nil
}

// writeSubDeviceAdditionalInfo 写入子设备附加信息
func writeSubDeviceAdditionalInfo(d deviceInfo, sheet *xlsx.Sheet, row int) error {
	if err := setCellValue(sheet, row, 13, d.ConcurrentNum); err != nil {
		return fmt.Errorf("write concurrent num error: %w", err)
	}
	if err := setCellValue(sheet, row, 14, d.MaxPointNum); err != nil {
		return fmt.Errorf("write max point num error: %w", err)
	}
	if err := setCellValue(sheet, row, 15, d.ExtendParams); err != nil {
		return fmt.Errorf("write extend params error: %w", err)
	}
	if err := setCellValue(sheet, row, 16, d.IsInstantiation); err != nil {
		return fmt.Errorf("write instantiation flag error: %w", err)
	}
	return nil
}

// zipFolder 压缩文件夹
func zipFolder(srcDir, targetFile string) (err error) {
	if srcDir == "" || targetFile == "" {
		return fmt.Errorf("srcDir or targetFile is empty")
	}
	srcDir, _ = filepath.Abs(srcDir) // 保证是绝对路径
	zipFile, err := os.Create(targetFile)
	if err != nil {
		log.Errorf("create zip file error: %s", err.Error())
		return err
	}
	defer zipFile.Close()
	writer := zip.NewWriter(zipFile)
	defer writer.Close()

	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		header.Name = relPath
		if info.IsDir() {
			header.Name += "/"
		} else {
			header.Method = zip.Deflate
		}
		w, err := writer.CreateHeader(header)
		if err != nil {
			return err
		}
		if !info.IsDir() {
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			_, err = io.Copy(w, file)
			if err != nil {
				return err
			}
		}
		return err
	})
	if err != nil {
		log.Errorf("Error walking throuth the directory: %s", err.Error())
		return err
	}
	return nil
}

// 本地文件存的配置
type localFile struct {
	Data struct {
		ConfigMap map[string]deviceInfo `json:"config_map"`
	} `json:"data"`
}

type deviceInfo struct {
	Channel         channel      `json:"channel"`
	CollectorType   int          `json:"collector_type"`
	DeviceCode      string       `json:"device_code"`
	DeviceGid       string       `json:"device_gid"`
	DeviceName      string       `json:"device_name"`
	DeviceNumber    string       `json:"device_number"`
	DeviceTypeEn    string       `json:"device_type_en"`
	DeviceTypeZh    string       `json:"device_type_zh"`
	MozuID          int          `json:"mozu_id"`
	Tpl             tpl          `json:"tpl,omitempty"`
	SubDevices      []deviceInfo `json:"sub_devices,omitempty"`
	ConcurrentNum   string       `json:"concurrent_num"`
	MaxPointNum     string       `json:"max_point_num"`
	ExtendParams    string       `json:"extend_params"`
	IsInstantiation string       `json:"is_instantiation"`
}

type devicePoint struct {
	DeviceID        string `xlsx:"1"`
	DeviceName      string `xlsx:"2"`
	ChannelType     string `xlsx:"3"`
	ChannelID       string `xlsx:"4"`
	ChannelAddr     string `xlsx:"5"`
	ChannelParams   string `xlsx:"6"`
	ValueTemplate   string `xlsx:"7"`
	WaitTime        string `xlsx:"8"`
	CmdInterval     string `xlsx:"9"`
	Timeout         string `xlsx:"10"`
	MaxFailCount    string `xlsx:"11"`
	MaxFailTime     string `xlsx:"12"`
	ConcurrentNum   string `xlsx:"13"`
	MaxPointNum     string `xlsx:"14"`
	ExtendParams    string `xlsx:"15"`
	IsInstantiation string `xlsx:"16"`
}

// channel 设备的channel信息
// type channel utils.Channel
type channel struct {
	Addr         string `json:"addr"`
	ChID         string `json:"chid"`
	ChParams     string `json:"chparams"`
	ChType       string `json:"chtype"`
	CmdInterval  string `json:"cmd_interval"`
	MaxFailCount string `json:"max_fail_count"`
	MaxFailTime  string `json:"max_fail_time"`
	Timeout      string `json:"timeout"`
	WaitTime     string `json:"wait_time"`
}

type tpl struct {
	TplNm   string `json:"tplnm"`
	TplPath string `json:"tplpath"`
}

// 获取最新的设备路径
func getLatestDevicePath() (string, error) {
	// 获取项目路径
	projectPath := config.GetRB().GetProjectLocalPath()
	configType := consts.DeviceTag

	// 查找匹配的文件
	files, err := filepath.Glob(filepath.Join(projectPath, fmt.Sprintf("%s*.json", configType)))
	if err != nil {
		log.Errorf("failed to find %s*.json files: %v", configType, err)
		return "", fmt.Errorf("failed to find %s*.json files: %v", configType, err)
	}

	if len(files) == 0 {
		log.Errorf("no %s*.json files found", configType)
		return "", fmt.Errorf("no %s*.json files found", configType)
	}

	// 正则表达式匹配文件名中的时间戳
	re := regexp.MustCompile(configType + `@(\d+)\.json`)
	var maxTimestamp int64 = -1
	// 兼容没有时间戳的情况
	targetFile := files[0]

	// 遍历所有匹配的文件，找到时间戳最大的文件
	for _, file := range files {
		matches := re.FindStringSubmatch(file)
		if len(matches) == 2 { // 匹配成功
			timestamp, err := strconv.ParseInt(matches[1], 10, 64)
			if err != nil {
				log.Warnf("invalid timestamp in filename %s: %v", file, err)
				continue
			}

			// 更新最大时间戳的文件
			if timestamp > maxTimestamp {
				maxTimestamp = timestamp
				targetFile = file
			}
		}
	}

	if targetFile == "" {
		log.Errorf("no valid %s @<timestamp>.json file found", configType)
		return "", fmt.Errorf("no %s*.json files found", configType)
	}
	return targetFile, nil
}
