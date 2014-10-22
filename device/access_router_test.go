package device

import (
	"testing"
)

func TestGetAccessPoint(t *testing.T) {
	// import the device basic info to warehouse
	store := NewDeviceStorage(host, user, password, database)
	if store == nil {
		t.Errorf("init storage failed")
	}
	defer store.Destory()
	prepare(store)
	home := NewHomeManager(store)
	var uid int64 = 100
	list, err := home.GetAllHome(domain, uid)
	if err != nil || len(list) != 5 {
		t.Errorf("get user all home failed:err[%v], len[%d]", err, len(list))
	}

	router := NewAccessRouter(store)
	mapping := NewBindingManager(store)
	var invalidUid int64 = 10000000
	var validDid int64
	// get home all device
	// 5 home, 2 master/home, 3 slave/master
	device := NewDeviceManager(store)
	for _, home := range list {
		devList, err := device.GetAllDevices(domain, home.hid)
		if err != nil {
			t.Error("get all devices failed", err)
		} else if len(devList) != 8 {
			t.Error("check all devices count failed", len(devList))
		}
		for _, dev := range devList {
			subDomain, deviceId, err := router.GetAccessPoint(uid, domain, dev.did)
			if err != nil {
				t.Error("get access point failed", err)
			} else if dev.IsMasterDevice() {
				bind, err := mapping.Get(domain, dev.did)
				if err != nil {
					t.Error("get mapping failed")
				} else if bind == nil {
					t.Error("get bind failed")
				} else if bind.subDomain != subDomain || bind.deviceId != deviceId {
					t.Error("check master info failed")
				}
			}
			validDid = dev.did
		}
	}
	// invalid account
	_, _, err = router.GetAccessPoint(invalidUid, domain, validDid)
	if err == nil {
		t.Error("invalid user get access point succ")
	}
	// invalid did
	var invalidDid int64 = 1000000
	_, _, err = router.GetAccessPoint(uid, domain, invalidDid)
	if err == nil {
		t.Error("get access point failed")
	}
	cleanAll(store)
}
