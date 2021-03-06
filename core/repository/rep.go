/**
 * Copyright 2015 @ z3q.net.
 * name : rep
 * author : jarryliu
 * date : 2016-05-24 10:14
 * description :
 * history :
 */
package repository

import (
	"github.com/jsix/gof/log"
	"github.com/jsix/gof/storage"
	"go2o/core/infrastructure/domain"
	"sync"
)

var (
	mux                 sync.Mutex
	DefaultCacheSeconds int64 = 3600
)

// 处理错误
func handleError(err error) error {
	return domain.HandleError(err, "rep")
	//if err != nil && gof.CurrentApp.Debug() {
	//	gof.CurrentApp.Log().Println("[ Go2o][ Rep][ Error] -", err.Error())
	//}
	//return err
}

// 删除指定前缀的缓存
func PrefixDel(sto storage.Interface, prefix string) error {
	rds := sto.(storage.IRedisStorage)
	_, err := rds.PrefixDel(prefix)
	if err != nil {
		log.Println("[ Cache][ Clean]: clean by prefix ", prefix, " error:", err)
	}
	return err
}
