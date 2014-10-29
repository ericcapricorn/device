package device

import (
	"zc-common-go/common"
	log "zc-common-go/glog"
)

type DeviceWarehouse struct {
	proxy *WarehouseProxy
}

func NewDeviceWarehouse(store *DeviceStorage) *DeviceWarehouse {
	proxy := newWarehouseProxy(store)
	if proxy == nil {
		log.Error("new WarehouseProxy failed")
		return nil
	}
	return &DeviceWarehouse{proxy: proxy}
}

func (this *DeviceWarehouse) Clear() {
	this.proxy.Clear()
}

// import a new device
func (this *DeviceWarehouse) Register(domain, subDomain, deviceId, publicKey string, master bool) error {
	if len(subDomain) <= 0 || len(deviceId) <= 0 {
		log.Warningf("check domain[%s] device[%s:%s] failed", domain, subDomain, deviceId)
		return common.ErrInvalidParam
	} else if master && len(publicKey) <= 0 {
		log.Warningf("check public key length failed:domain[%s], device[%s:%s]", domain, subDomain, deviceId)
		return common.ErrInvalidParam
	}
	err := this.proxy.InsertDeviceInfo(domain, subDomain, deviceId, publicKey, master)
	if err != nil {
		log.Warningf("insert device info failed:domain[%s], device[%s:%s], master[%t]",
			domain, subDomain, deviceId, master)
		return err
	}
	return nil
}

// if find the record return basic info + nil, if not exist, return nil + nil
func (this *DeviceWarehouse) Get(domain, subDomain, deviceId string) (*BasicInfo, error) {
	device, err := this.proxy.GetDeviceInfo(domain, subDomain, deviceId)
	if err != nil {
		log.Warningf("get device info failed:domain[%s], device[%s:%s]", domain, subDomain, deviceId)
		return nil, err
	} else if device == nil {
		log.Warningf("not find the device info:domain[%s], device[%s:%s]", domain, subDomain, deviceId)
		return nil, nil
	}
	if !device.Validate() {
		log.Errorf("validate device info failed:domain[%s], device[%s:%s]", domain, subDomain, deviceId)
		return nil, common.ErrInvalidDevice
	}
	return device, nil
}

// WARNING: must be cautious for using this interface
// delete the device only for not online device
func (this *DeviceWarehouse) Delete(domain, subDomain, deviceId string) error {
	err := this.proxy.DeleteDeviceInfo(domain, subDomain, deviceId)
	if err != nil {
		log.Warningf("delete the device failed:domain[%s], device[%s:%s]", domain, subDomain, deviceId)
		return err
	}
	return nil
}
