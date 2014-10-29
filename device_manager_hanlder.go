package main

import (
	log "zc-common-go/glog"
	"zc-dm/device"
	"zc-service-go"
)

type DeviceManagerHandler struct {
	device *device.DeviceManager
	bind   *device.BindingManager
}

func NewDeviceManagerHandler(device *device.DeviceManager, bind *device.BindingManager) *DeviceManagerHandler {
	if device == nil || bind == nil {
		return nil
	}
	return &DeviceManagerHandler{device: device, bind: bind}
}

////////////////////////////////////////////////////////////////////////////////////////////
/// DEVICE MANAGER
////////////////////////////////////////////////////////////////////////////////////////////
// list all devices of one home
func (this *DeviceManagerHandler) handleListDevices(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	hid := req.GetInt("hid")
	list, err := this.device.GetAllDevices(domain, hid)
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("list all home devices failed:domain[%s], hid[%d], err[%v]", domain, hid, err)
		return
	}
	for _, device := range list {
		resp.AddObject("devices", zc.ZObject{"id": device.GetDid(), "hid": device.GetHid(), "name": device.GetDeviceName(), "master": device.GetMasterDid()})
	}
	log.Warningf("list all home devices succ:domain[%s], hid[%d], count[%d]", domain, hid, len(list))
	resp.SetAck()
}

// binding device
func (this *DeviceManagerHandler) handleBindDevice(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	subDomain := req.GetString("submain")
	deviceId := req.GetString("deviceid")
	name := req.GetString("dname")
	hid := req.GetInt("hid")
	master := req.GetInt("master")
	err := this.bind.Binding(domain, subDomain, deviceId, name, hid, master)
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("bind device to home failed:domain[%s], device[%s:%s], dname[%s], hid[%d], master[%d], err[%v]",
			domain, subDomain, deviceId, name, hid, master, err)
		return
	}
	log.Infof("bind device to home succ:domain[%s], device[%s:%s], dname[%s], hid[%d], master[%d]",
		domain, subDomain, deviceId, name, hid, master)
	resp.SetAck()
}

// change device
func (this *DeviceManagerHandler) handleChangeDevice(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	subDomain := req.GetString("submain")
	deviceId := req.GetString("deviceid")
	did := req.GetInt("hid")
	err := this.bind.ChangeBinding(did, domain, subDomain, deviceId)
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("change device failed:old[%d], domain[%s], device[%s:%s], err[%v]",
			did, domain, subDomain, deviceId, err)
		return
	}
	log.Infof("change device failed:old[%d], domain[%s], device[%s:%s]",
		did, domain, subDomain, deviceId)
	resp.SetAck()
}

// delete device
func (this *DeviceManagerHandler) handleDeleteDevice(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	hid := req.GetInt("hid")
	did := req.GetInt("did")
	err := this.device.DeleteDevice(domain, hid, did)
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("delete device from home failed:domain[%s], hid[%d], did[%d], err[%v]", domain, hid, did, err)
		return
	}
	log.Infof("delete device from home succ:domain[%s], hid[%d], did[%s]", domain, hid, did)
	resp.SetAck()
}

// modify device
func (this *DeviceManagerHandler) handleModifyDevice(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	did := req.GetInt("did")
	name := req.GetString("dname")
	err := this.device.ChangeDeviceName(domain, did, name)
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("modify device name failed:domain[%s], did[%d], name[%s], err[%v]", domain, did, name, err)
		return
	}
	log.Infof("modify device name succ:domain[%s], did[%s], name[%s]", domain, did, name)
	resp.SetAck()
}

// frozen/defrozen device
func (this *DeviceManagerHandler) handleFrozenDevice(req *zc.ZMsg, resp *zc.ZMsg) {
	domain := req.GetString("domain")
	did := req.GetInt("did")
	frozen := req.GetBool("frozen")
	var err error
	if frozen {
		err = this.device.Disable(domain, did)
	} else {
		err = this.device.Enable(domain, did)
	}
	if err != nil {
		resp.SetErr(err.Error())
		log.Warningf("frozen/defrozen device failed:domain[%s], did[%d], frozen[%t], err[%v]", domain, did, frozen, err)
		return
	}
	log.Infof("frozen/defrozen device succ:domain[%s], did[%d], frozen[%t]", domain, did, frozen)
	resp.SetAck()
}
