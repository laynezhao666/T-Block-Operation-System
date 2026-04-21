// Package room 提供机房信息的查询功能。
package room

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"dac/entity/config"
	"dac/entity/model/cgi"
	"dac/repo/dac"

	"dac/entity/utils/thttp"
)

// GetAllRoomsByBuildingContainingMozu 获取模组所在楼栋的所有机房信息
func GetAllRoomsByBuildingContainingMozu(ctx context.Context, mozuIDStr string) ([]cgi.Rooms, error) {
	var (
		mozuID int
		err    error
	)
	if len(mozuIDStr) > 0 {
		mozuID, err = strconv.Atoi(mozuIDStr)
		if err != nil {
			return nil, err
		}
	}

	mozus, err := dac.GetRW().GetMozuWithSameBuildings(ctx, mozuID)
	if err != nil {
		return nil, err
	}

	allRooms := make([]cgi.Rooms, len(mozus))
	j := 1

	for i := range mozus {
		m := &mozus[i]
		r := cgi.Rooms{
			MozuID:   m.ID,
			MozuName: m.Name,
			Rooms:    nil,
		}

		h := make(http.Header)
		h.Add("mozuid", fmt.Sprintf("%v", m.ID))

		if err = thttp.GetJSONWithHeader(config.C.GIDMapping.URL.Rooms, h, 30000, &r.Rooms); err != nil {
			return nil, fmt.Errorf("request %v error: %w", config.C.GIDMapping.URL.Rooms, err)
		}

		if m.ID == mozuID {
			allRooms[0] = r
		} else {
			allRooms[j] = r
			j++
		}
	}

	return allRooms, nil
}
