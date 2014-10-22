package device

import (
	"zc-common-go/common"
	log "zc-common-go/glog"
)

type AccessRouter struct {
	store *DeviceStorage
}

func NewAccessRouter(store *DeviceStorage) *AccessRouter {
	return &AccessRouter{store: store}
}

// give device inner id get the master device info(did)
func (this *AccessRouter) GetAccessPoint(uid int64, domain string, did int64) (string, string, error) {
	// step 0. TODO check the mapping is valid
	// step 1. get the device info check it is master or normal device
	var invalidString string
	deviceManager := NewDeviceManager(this.store)
	device, err := deviceManager.Get(domain, did)
	if err != nil {
		log.Warningf("get device info failed:domain[%s], did[%d], err[%v]", domain, did, err)
		return invalidString, invalidString, err
	} else if device == nil {
		log.Warningf("not find the device:domain[%s], did[%d]", domain, did)
		return invalidString, invalidString, common.ErrEntryNotExist
	} else if device.GetStatus() != ACTIVE {
		log.Warningf("the device status not invalid:domain[%s], did[%d], status[%d]", domain, did, device.GetStatus())
		return invalidString, invalidString, common.ErrInvalidStatus
	}
	// get device's master did and home id
	masterDid := device.GetMasterDid()
	hid := device.GetHid()
	if masterDid <= 0 || hid <= 0 {
		log.Warningf("check master did or home id failed:domain[%s], did[%d], master[%d], hid[%d]",
			domain, did, masterDid, hid)
		return invalidString, invalidString, common.ErrUnknown
	}

	// step 2. get master device subdomain + deviceid
	bindManager := NewBindingManager(this.store)
	bind, err := bindManager.Get(domain, masterDid)
	if err != nil {
		log.Warningf("get master mapping binding info failed:domain[%s], did[%d], err[%v]", domain, masterDid, err)
		return invalidString, invalidString, err
	} else if bind == nil {
		log.Warningf("master device mapping not exist:domain[%s], did[%d]", domain, masterDid)
		return invalidString, invalidString, err
	}

	// step 3. check the master status
	if !device.IsMasterDevice() {
		master, err := deviceManager.Get(domain, masterDid)
		if err != nil {
			log.Warningf("get master device info failed:domain[%s], did[%d], err[%v]", domain, masterDid, err)
			return invalidString, invalidString, err
		} else if master == nil {
			log.Warningf("master device not exist:domain[%s], did[%d]", domain, masterDid)
			return invalidString, invalidString, err
		} else if master.GetStatus() != ACTIVE {
			log.Warningf("master device status not active:domain[%s], did[%d], status[%d]", domain, masterDid, device.GetStatus())
			return invalidString, invalidString, common.ErrInvalidStatus
		}
	}

	// step 4. check the home status ok
	homeManager := NewHomeManager(this.store)
	home, err := homeManager.Get(domain, hid)
	if err != nil {
		log.Warningf("get home info failed:domain[%s], hid[%d], uid[%d], err[%v]", domain, hid, uid, err)
		return invalidString, invalidString, err
	} else if home == nil {
		log.Warningf("not find the home:domain[%s], hid[%d], uid[%d], err[%v]", domain, hid, uid, err)
		return invalidString, invalidString, common.ErrEntryNotExist
	} else if home.GetStatus() != ACTIVE {
		log.Warningf("the home status not active:domain[%s], hid[%d]", domain, hid)
		return invalidString, invalidString, common.ErrInvalidStatus
	}

	// step 5. check the uid in the same home and status ok
	memberManager := NewMemberManager(this.store)
	member, err := memberManager.Get(domain, hid, uid)
	if err != nil {
		log.Warningf("get user member failed:domain[%s], hid[%d], uid[%d], err[%v]", domain, hid, uid, err)
		return invalidString, invalidString, err
	} else if member == nil {
		log.Warningf("not find the user:domain[%s], hid[%d], uid[%d]", domain, hid, uid)
		return invalidString, invalidString, common.ErrNoPrivelige
	} else if member.GetStatus() != ACTIVE {
		log.Warningf("the user status not active:domain[%s], hid[%d], uid[%d]", domain, hid, uid)
		return invalidString, invalidString, common.ErrNotAllowed
	}

	return bind.subDomain, bind.deviceId, nil
}
