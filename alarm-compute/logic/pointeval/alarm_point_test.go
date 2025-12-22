package pointeval

import (
	"encoding/json"
	"fmt"
	"testing"

	"alarm-compute/entity/epoint"
	"alarm-compute/utils/common"
)

// TestUpdatePointFetchList 函数
func TestAdd(t *testing.T) {
	pointTypeMap := PointTypeMap{
		Express: "A==0&&JP(B,1,60)&&C==0",
		PMap: map[string][]string{
			"A": {"3458875016162127872.Uc", "3458875570979344384.Uc"},
			"B": {"3458875570979344384.PowerVoltAlarm"},
			"C": {"3458875016162127872.Ua"},
			"D": {"3458875016162127872.Ub"},
			"E": {"3458875016162127872.Uc"},
			"F": {"3458875016430563328.Ua"},
			"G": {"3458875016430563328.Ub"},
		},
		Engine:         "govaluate",
		PointFetchList: make(map[string][]PointFetchInfo),
		JPRangeSec:     30,
	}
	err := pointTypeMap.UpdatePointFetchList()
	fmt.Println(err)
	fmt.Println(common.JSONMarshalNoErr(pointTypeMap.PointFetchList))
}

// DelayEQ(A,1,45) && (B<=1)
func TestEval(t *testing.T) {
	pointTypeMap := PointTypeMap{
		Express: "CountRateGT(A,32,0)>=0.5",
		PMap: map[string][]string{
			"A": {"3689398347740742983.Temp"},
			"B": {"3458877364677644288.IbatGroup"},
			"C": {"3458877364677644288.UbatGroup"},
		},
	}
	intervalPointValueMap := map[string]epoint.IntervalMap{
		"3689398347740742978.Temp": {
			0: float64(31),
		},
		"3689398347740742983.Temp": {
			0: float64(33),
		},
		"3689398347740742992.Temp": {
			0:  float64(13.6),
			60: float64(13.6),
		},
	}
	result, err := pointTypeMap.EvalWithIntervalPointData(intervalPointValueMap)
	if err != nil {
		fmt.Printf("Error: %s", err)
	} else {
		fmt.Println(result)
	}
}

func TestPerformance(t *testing.T) {
	type Test struct {
		Id   int    `json:"id"`
		Name string `json:"name"`
	}
	list := []Test{
		{1, "test1"},
		{2, "test2"},
		{3, "test3"},
	}
	fmt.Println(list)
	res, err := json.Marshal(list)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(string(res))
	}
}
