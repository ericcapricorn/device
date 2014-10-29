package device

import (
	"zc-common-go/common"
	log "zc-common-go/glog"
)

type BindingManager struct {
	store     *DeviceStorage
	warehouse *DeviceWarehouse
	proxy     *BindingProxy
}

func NewBindingManager(store *DeviceStorage) *BindingManager {
	warehouse := NewDeviceWarehouse(store)
	if warehouse == nil {
		log.Errorf("new device warehouse failed")
		return nil
	}
	proxy := newBindingProxy(store)
	if proxy == nil {
		log.Errorf("new binding proxy failed")
		return nil
	}
	return &BindingManager{store: store, warehouse: warehouse, proxy: proxy}
}

//////////////////////////////////////////////////////////////////////////////
/// public interface can not delete the binding info in normal cases
//////////////////////////////////////////////////////////////////////////////
// if not exist return nil + nil
func (this *BindingManager) Get(domain string, did int64) (*BindingInfo, error) {
	bind, err := this.proxy.GetBindingByDid(domain, did)
	if err != nil {
		log.Warningf("get binding info failed:domain[%s], did[%d], err[%v]", domain, did, err)
		return nil, err
	}
	return bind, nil
}

// if not exist return nil + nil
func (this *BindingManager) GetBindingInfo(domain, subDomain, deviceId string) (*BindingInfo, error) {
	common.CheckParam(this.proxy != nil && this.warehouse != nil)
	return this.proxy.GetBindingInfo(domain, subDomain, deviceId)
}

// binding one device to one home, if masterDid < 0 it's master device, otherwise it's slave device
func (this *BindingManager) Binding(domain, subDomain, deviceId, deviceName string, hid, masterDid int64) error {
	common.CheckParam(this.proxy != nil && this.warehouse != nil)
	// step 1. check the device basic info is valid
	err := this.checkDeviceInfo(domain, subDomain, deviceId, masterDid < 0)
	if err != nil {
		log.Warningf("check the binding device failed:err[%v]", err)
		return err
	}
	// step 2. check the master device is binding ok
	if masterDid > 0 {
		_, err = this.proxy.GetBindingByDid(domain, masterDid)
		if err != nil {
			log.Warningf("check the master device not active:domain[%s], did[%d], err[%v]", domain, masterDid, err)
			return err
		}
	}
	// step 3. build mapping device ids, if already exist return succ for rebinding...
	err = this.proxy.BindingDevice(domain, subDomain, deviceId, deviceName, hid, masterDid)
	if err != nil {
		log.Warningf("binding device failed:domain[%s], device[%s:%s], master[%d]", domain, subDomain, deviceId, masterDid)
		return err
	}
	return nil
}

// change the slave or master device, the old must valid
func (this *BindingManager) ChangeBinding(did int64, domain, subDomain, deviceId string) error {
	common.CheckParam(this.proxy != nil && this.warehouse != nil)
	// step 1. check the old did must be register succ
	var device DeviceInfo
	deviceManager := NewDeviceManager(this.store)
	err := deviceManager.getDeviceInfo(domain, did, &device)
	if err != nil {
		log.Warningf("get master device failed:domain[%s], did[%d], err[%v]", domain, did, err)
		return err
	}
	// step 2 check the old did must be binging ok
	var exist bool
	err = this.proxy.IsBindingExist(domain, did, &exist)
	if err != nil {
		log.Warningf("check binding exist failed:domain[%s], did[%d], err[%v]", domain, did, err)
		return err
	} else if !exist {
		log.Warningf("the old did not binding:domain[%s], did[%d]", domain, did)
		return common.ErrNotYetBinded
	}
	// step 3. check the new device basic info is valid according the old device type
	err = this.checkDeviceInfo(domain, subDomain, deviceId, device.IsMasterDevice())
	if err != nil {
		log.Warningf("check device info failed:domain[%s], device[%s:%s], master[%d], err[%v]",
			domain, subDomain, deviceId, did, err)
		return err
	}
	// step 4. check the new device is not binding by others
	_, err = this.proxy.GetBindingInfo(domain, subDomain, deviceId)
	if err != nil {
		if err != common.ErrEntryNotExist {
			log.Warningf("get device binding info failed:domain[%s], device[%s:%s], master[%d], err[%v]",
				domain, subDomain, deviceId, did, err)
			return err
		}
	} else {
		log.Warningf("check the device is already binded:domain[%s], device[%s:%s], did[%d]",
			domain, subDomain, deviceId, did)
		return common.ErrAlreadyBinded
	}
	// step 5. change the mapping relation
	err = this.proxy.ChangeDeviceBinding(did, domain, subDomain, deviceId)
	if err != nil {
		log.Warningf("do change the device mapping binding info failed:domain[%s], device[%s:%s], err[%v]",
			domain, subDomain, deviceId, err)
		return err
	}
	return nil
}

// check basic device info from device warehouse
// binding one device to one home, if masterDid < 0 it's master device, otherwise it's slave device
func (this *BindingManager) checkDeviceInfo(domain, subDomain, deviceId string, isMaster bool) error {
	dev, err := this.warehouse.Get(domain, subDomain, deviceId)
	if err != nil {
		log.Warningf("get device failed:domain[%s], device[%s:%s], err[%v]",
			domain, subDomain, deviceId, err)
		return err
	} else if dev == nil {
		log.Warningf("check the device not exist:domain[%s], device[%s:%s]", domain, subDomain, deviceId)
		return common.ErrInvalidDevice
	} else if !dev.Validate() {
		log.Errorf("device validate failed:domain[%s], device[%s:%s]", domain, subDomain, deviceId)
		return common.ErrInvalidDevice
	} else if isMaster && dev.deviceType != MASTER {
		log.Warningf("check the master device type failed:domain[%s], device[%s:%s]", domain, subDomain, deviceId)
		return common.ErrInvalidDevice
	} else if (!isMaster) && (dev.deviceType == MASTER) {
		log.Warningf("check the slave device type failed:domain[%s], device[%s:%s]", domain, subDomain, deviceId)
		return common.ErrInvalidDevice
	}
	return nil
}
