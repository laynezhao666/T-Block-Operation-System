package base

import (
	"agent/entity/config"
	"agent/logic/distribution/distributor"
	"agent/logic/distribution/distributor/http"
	"agent/logic/distribution/distributor/tlink"
)

// GetDistributorList 入参为 "collect_change"等，返回 distributor列表
func GetDistributorList(dataType string) distributor.Distributors {
	list := distributor.Distributors{}
	if IsDistributorEnable(config.GetRB().Distributor.Http.Enable, dataType) {
		list = append(list, http.HttpDistributor())
	}
	if IsDistributorEnable(config.GetRB().Distributor.Tlink.Enable, dataType) {
		list = append(list, tlink.TLinkDistributor())
	}
	return list
}

// IsDistributorEnable 判断是否启用
func IsDistributorEnable(enableArray []string, dataType string) bool {
	for _, item := range enableArray {
		if item == dataType {
			return true
		}
	}
	return false
}
