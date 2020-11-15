package service

import (
	"github.com/hellojqk/refactor/src/core"
	"github.com/hellojqk/refactor/src/model"
)

// InsertProwyLog .
func InsertProwyLog(proxyLog *core.ProxyLog) (bool, error) {
	insertResult := core.DB.Create(proxyLog)
	return insertResult.RowsAffected > 0, insertResult.Error
}

// ListProxyLog .
func ListProxyLog(pageParam *model.PageParam) (result []core.ProxyLog, total int64, err error) {
	result = make([]core.ProxyLog, 0)
	err = core.DB.Model(&core.ProxyLog{}).Count(&total).Error
	if err != nil || total == 0 {
		return
	}
	err = core.DB.Debug().Limit(pageParam.PageSize).Offset((pageParam.PageIndex - 1) * pageParam.PageSize).Preload("Application").Preload("API").Find(&result).Error
	return
}
