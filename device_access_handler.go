package main

import (
	"zc-common-go/common"
	log "zc-common-go/glog"
	"zc-dm/device"
	"zc-service-go"
)

type DeviceAccessPointHandler struct {
	accessPoint *device.AccessRouter
}

func NewDeviceAccessPointHandler(access *device.AccessRouter) *DeviceAccessPointHandler {
	if access == nil {
		return nil
	}
	return &DeviceAccessPointHandler{accessPoint: access}
}

// check privelige and return the access device keys
func (this *DeviceAccessPointHandler) handleGetAccessPoint(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	uid := req.GetInt("uid")
	did := req.GetInt("did")
	if len(domain) <= 0 || uid <= 0 || did <= 0 {
		resp.SetErr(common.ErrInvalidRequest.Error())
		log.Warningf("check request invalid:domain[%s], uid[%d], did[%d]", domain, uid, did)
		return
	}
	subDomain, deviceId, err := this.accessPoint.GetAccessPoint(uid, domain, did)
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("get device ctrl access point failed:domain[%s], uid[%d], did[%d], err[%v]", domain, uid, did, err)
		return
	}
	resp.AddString("domain", domain)
	resp.AddString("submain", subDomain)
	resp.AddString("deviceid", deviceId)
	resp.SetAck()
}
