package db

import (
	"common/entity/consts"
	"context"
	tgorm "etrpc-go/client/gorm"
	"fmt"
)

// GetKeyDicFunc 获取Key配置项的函数类型定义
type GetKeyDicFunc = func(ctx context.Context, filter string, mozuId int32) ([]string, error)

// GetKvDicFunc 获取KV配置项的函数类型定义
type GetKvDicFunc = func(ctx context.Context, filter string, mozuId int32) (map[string]string, error)

// ICommonDao 通用接口定义
type ICommonDao interface {
	TableKeyDicFunc(tableName, colName string, mozuIdEnable bool) GetKeyDicFunc
	TableKvDicFunc(tableName, keyColName, valueColName string, mozuIdEnable bool) GetKvDicFunc
}

// NewCommonDao 创建一个通用接口实现类
func NewCommonDao() ICommonDao {
	return commonDaoImpl{}
}

type commonDaoImpl struct {
}

// TableKeyDicFunc 单字段distinct的类型定义
func (commonDaoImpl) TableKeyDicFunc(tableName, colName string, mozuIdEnable bool) GetKeyDicFunc {
	return func(ctx context.Context, filter string, mozuId int32) ([]string, error) {
		res := make([]string, 0)
		// 组装SQL
		sql := tgorm.GetDB(consts.TbosMysqlName).WithContext(ctx).Table(tableName).Distinct(colName)
		if len(filter) > 0 {
			sql.Where(fmt.Sprintf("%s like ?", colName), filter+"%")
		}
		if mozuIdEnable {
			if mozuId <= 0 {
				return nil, fmt.Errorf("mozu_id is invalid")
			} else {
				sql.Where("mozu_id = ?", mozuId)
			}
		}
		if err := sql.Find(&res).Error; err != nil {
			return nil, err
		}
		// 组装数据
		return res, nil
	}
}

// TableKvDicFunc KV字段distinct的类型定义
func (commonDaoImpl) TableKvDicFunc(tableName, keyColName, valueColName string, mozuIdEnable bool) GetKvDicFunc {
	return func(ctx context.Context, filter string, mozuId int32) (map[string]string, error) {
		// 组装SQL
		sql := tgorm.GetDB(consts.TbosMysqlName).WithContext(ctx).Table(tableName).Distinct(keyColName, valueColName)
		if len(filter) > 0 {
			sql.Where(fmt.Sprintf("%s like ?", keyColName), filter+"%")
		}
		if mozuIdEnable {
			if mozuId <= 0 {
				return nil, fmt.Errorf("mozu_id is invalid")
			} else {
				sql.Where("mozu_id = ?", mozuId)
			}
		}
		rows, err := sql.Rows()
		if err != nil {
			return nil, err
		}
		// 组装数据
		data := make(map[string]string)
		for rows.Next() {
			var key = ""
			var val = ""
			err := rows.Scan(&key, &val)
			if err != nil {
				return nil, err
			}
			data[key] = val
		}
		return data, nil
	}
}
