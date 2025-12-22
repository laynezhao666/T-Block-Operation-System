package std

import (
	"agent/entity/consts"
	"agent/utils"
	"strings"
)

// 样式 A=test.BMTR_1.Ubat_1;B=test.BMTR_1.Ubat_1  => test.BMTR_1.Ubat_1
type mapping string

func (m mapping) getCollectPoints() ([]string, error) {
	mapStr := string(m)
	mpList := strings.Split(mapStr, consts.MpListSplitChar)

	var cp []string
	for _, v := range mpList {
		parts := strings.Split(v, consts.MpExpressionSplitChar)
		if len(parts) != 2 {
			continue
			//return nil, errors.New("mapping format err")
		}
		cp = append(cp, parts[1])
	}

	unique := utils.RemoveDuplicates(cp)
	return unique, nil
}

func (m mapping) getMappingParam() (map[string]string, error) {
	mapStr := string(m)
	mpList := strings.Split(mapStr, consts.MpListSplitChar)

	param := map[string]string{}
	for _, v := range mpList {
		parts := strings.Split(v, consts.MpExpressionSplitChar)
		if len(parts) != 2 {
			continue
			//return nil, errors.New("mapping format err")
		}
		param[parts[0]] = parts[1]
	}
	return param, nil
}
