/**
 * Copyright 2015 @ S1N1 Team.
 * name : device
 * author : jarryliu
 * date : -- :
 * description :
 * history :
 */
package front

import (
	gutil "github.com/atnet/gof/util"
	"github.com/atnet/gof/web"
	"net/http"
	"time"
)

const (
	clientDeviceTypeCookieId string = "client_device_type"

	// PC设备
	DevicePC string = "1"
	// 手持设备
	DeviceMobile string = "2"
	// 触摸设备
	DeviceTouchPad string = "3"
)

// 获取浏览设备
func GetBrownerDevice(ctx *web.Context) string {
	ck, err := ctx.Request.Cookie(clientDeviceTypeCookieId)
	if err == nil && ck != nil {
		switch ck.Value {
		case "1":
			return DevicePC
		case "2":
			return DeviceMobile
		case "3":
			return DeviceTouchPad
		}
	}
	if gutil.IsMobileAgent(ctx.Request.UserAgent()) {
		return DeviceMobile
	}
	return DevicePC
}

// 设置浏览设备
func SetBrownerDevice(ctx *web.Context, deviceType string) {
	ck, err := ctx.Request.Cookie(clientDeviceTypeCookieId)
	isDefaultDevice := deviceType == "" || deviceType == "1"
	if err == nil && ck != nil {
		if isDefaultDevice {
			ck.Value = deviceType
			ck.Expires = time.Now().Add(-time.Hour * 48)
		} else {
			ck.Value = deviceType
			ck.Expires = time.Now().Add(time.Hour * 24)
		}
	} else if !isDefaultDevice {
		ck = &http.Cookie{
			Name:    clientDeviceTypeCookieId,
			Value:   deviceType,
			Expires: time.Now().Add(time.Hour * 24),
		}
	}

	if ck != nil {
		ck.HttpOnly = false
		ck.Path = "/"
		http.SetCookie(ctx.ResponseWriter, ck)
	}
}
