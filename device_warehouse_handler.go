package main

import (
	"zc-common-go/common"
	log "zc-common-go/glog"
	"zc-dm/device"
	"zc-service-go"
)

type DeviceWarehouseHandler struct {
	warehouse *device.DeviceWarehouse
}

func NewDeviceWarehouseHandler(warehouse *device.DeviceWarehouse) *DeviceWarehouseHandler {
	if warehouse == nil {
		return nil
	}
	return &DeviceWarehouseHandler{warehouse: warehouse}
}

////////////////////////////////////////////////////////////////////////////////////////////
/// DEVICE BASIC INFO MANAGER
////////////////////////////////////////////////////////////////////////////////////////////
// import regist all devices one by one
func (this *DeviceWarehouseHandler) handleRegistDevice(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	subDomain := req.GetString("submain")
	deviceId := req.GetString("deviceid")
	master := req.GetBool("master")
	var publicKey string
	if master {
		publicKey = req.GetString("publickey")
	}
	err := this.warehouse.Register(domain, subDomain, deviceId, publicKey, master)
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("register device failed:domain[%s], device[%s:%s], key[%s], master[%t], err[%v]",
			domain, subDomain, deviceId, publicKey, master, err)
		return
	}
	log.Infof("register device succ:domain[%s], device[%s:%s], key[%s], master[%t]", domain, subDomain, deviceId, publicKey, master)
	resp.SetAck()
}

func (this *DeviceWarehouseHandler) handleGetPublicKey(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	subDomain := req.GetString("submain")
	deviceId := req.GetString("deviceid")
	// get master public key
	basic, err := this.warehouse.Get(domain, subDomain, deviceId)
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("get device basic info failed:domain[%s], device[%s:%s], err[%v]", domain, subDomain, deviceId, err)
		return
	} else if basic == nil || !basic.IsMaster() {
		resp.SetErr(common.ErrMasterNotExist.Error())
		log.Warningf("get device basic info failed:domain[%s], device[%s:%s]", domain, subDomain, deviceId)
		return
	}
	resp.PutString("publickey", basic.PublicKey())
	resp.SetAck()
}
