package utils

import (
	"agent/entity/model"
	"regexp"
	"strings"
)
// FilterTree 过滤树
func FilterTree(tree *model.TreeNode, filterList map[string]bool) *model.TreeNode {
	if tree == nil {
		return nil
	}

	// Check if the current node's device_gid is in the filter list
	if _, ok := filterList[tree.DeviceGid]; !ok {
		return nil
	}

	// Create a new node with the filtered children
	newNode := &model.TreeNode{
		DeviceGid:         tree.DeviceGid,
		DeviceNumber:      tree.DeviceNumber,
		DeviceNumberShow:  tree.DeviceNumberShow,
		DeviceNo:          tree.DeviceNo,
		DeviceTypeEn:      tree.DeviceTypeEn,
		DeviceTypeZh:      tree.DeviceTypeZh,
		ApplicationTypeEn: tree.ApplicationTypeEn,
		ApplicationTypeZh: tree.ApplicationTypeZh,
		DeviceCount:       tree.DeviceCount,
	}

	for _, child := range tree.Children {
		filteredChild := FilterTree(child, filterList)
		if filteredChild != nil {
			newNode.Children = append(newNode.Children, filteredChild)
		}
	}

	return newNode
}

// FilterStdDevice 筛选采集相关的设备
func FilterStdDevice(deviceList []model.StdDevice, filterList map[string]bool) []model.StdDevice {
	var filterlist []model.StdDevice
	for _, d := range deviceList {
		if _, ok := filterList[d.DeviceGid]; !ok {
			continue
		}
		filterlist = append(filterlist, d)
	}

	return filterlist
}

// AddShowDeviceNumber 展示编号处理
func AddShowDeviceNumber(deviceList []model.StdDevice) []model.StdDevice {
	var resultList []model.StdDevice
	for _, d := range deviceList {
		parts := strings.Split(d.DeviceNumber, "-")
		d.DeviceNumberShow = strings.Join(parts[3:], "-")
		resultList = append(resultList, d)
	}
	return resultList
}

// GetConciseCodeMap 短编号处理
func GetConciseCodeMap(deviceList []model.StdDevice) (map[string]string, []model.StdDevice) {
	resMap := make(map[string]string, len(deviceList))
	reg := regexp.MustCompile(`^ITM([A-Z]*)?$`) // 正则匹配以ITM开头，后跟0个或多个大写字母，且不包含其他字符的模式
	var resList []model.StdDevice
	for _, d := range deviceList {
		parts := strings.Split(d.DeviceNumber, "-")
		if len(parts) < 5 {
			continue // 防止索引越界
		}
		var conciseCode string

		conciseCode = strings.Join(parts[4:], "-")

		// 使用正则表达式匹配"ITM*"模式（如ITM、ITMS、ITME等），但不匹配"ITMA-HVDC-REC04"
		if reg.MatchString(conciseCode) {
			conciseCode = "ITM"
		}
		reg1 := regexp.MustCompile(`^ITM[A-Z]*-`)
		// 如果 conciseCode 以 "ITM*-" 开头，则去掉这部分
		if reg1.MatchString(conciseCode) {
			parts := strings.Split(conciseCode, "-")
			conciseCode = strings.Join(parts[1:], "-")
		}
		resMap[conciseCode] = d.DeviceGid
		d.ConciseCode = conciseCode
		resList = append(resList, d)
	}
	return resMap, resList
}
