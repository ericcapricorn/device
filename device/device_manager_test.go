package device

import (
	"fmt"
	"testing"
)

func prepare(store *DeviceStorage) {
	// register master and slave device
	warehouse := NewDeviceWarehouse(store)
	subDomain := "flying"
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("20141017%d", i)
		key := fmt.Sprintf("publicKey%d", i)
		err := warehouse.Register(domain, subDomain, id, key, true)
		if err != nil {
			panic("register master device failed")
		}
	}
	for i := 0; i < 30; i++ {
		id := fmt.Sprintf("20151017%d", i)
		err := warehouse.Register(domain, subDomain, id, "", false)
		if err != nil {
			panic("register master device failed")
		}
	}

	// create home for account
	home := NewHomeManager(store)
	var uid int64 = 100
	// user have 5 homes
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("home%d", i)
		err := home.Create(domain, uid, name)
		if err != nil {
			panic("create home failed")
		}
	}
	// check home count of user
	list, err := home.GetAllHome(domain, uid)
	if err != nil || len(list) != 5 {
		panic("get user all home failed")
	}

	binding := NewBindingManager(store)
	master := make([]int64, 0)
	// every home with two master
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("20141017%d", i)
		name := fmt.Sprintf("master%d", i)
		err = binding.Binding(domain, subDomain, id, name, list[i%5].hid, -1)
		if err != nil {
			panic("binding master device failed")
		}
		bind, err := binding.GetBindingInfo(domain, subDomain, id)
		if err != nil {
			panic("get binding info failed")
		} else {
			master = append(master, bind.did)
		}
	}
	// every master with 3 slave
	for i := 0; i < 30; i++ {
		id := fmt.Sprintf("20151017%d", i)
		name := fmt.Sprintf("slave%d", i)
		err = binding.Binding(domain, subDomain, id, name, list[i%5].hid, master[i%10])
		if err != nil {
			panic("binding slave device failed")
		}
	}
}

func TestGetAllDevice(t *testing.T) {
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
	// invalid hid
	var invalidHid int64 = 10000000
	device := NewDeviceManager(store)
	devList, err := device.GetAllDevices(domain, invalidHid)
	if err != nil {
		t.Error("get invalid hid devices failed", err)
	} else if len(devList) != 0 {
		t.Error("should not exist devices")
	}
	// get home all device
	// 5 home, 2 master/home, 3 slave/master
	for _, home := range list {
		devList, err := device.GetAllDevices(domain, home.hid)
		if err != nil {
			t.Error("get all devices failed", err)
		} else if len(devList) != 8 {
			t.Error("check all devices count failed", len(devList))
		}
		for _, dev := range devList {
			temp, err := device.Get(domain, dev.GetDid())
			if err != nil {
				t.Error("get did failed", dev.GetDid(), err)
			} else if temp.did != dev.did || temp.hid != dev.hid || temp.deviceName != dev.deviceName || temp.masterDid != dev.masterDid || temp.status != dev.status {
				t.Error("check device info faild")
			}
		}
	}
	// invalid did
	var invalidDid int64
	dev, err := device.Get(domain, invalidDid)
	if err != nil {
		t.Error("get invalid did succ")
	} else if dev != nil {
		t.Error("did should not exist")
	}
	cleanAll(store)
}

func TestDeleteAll(t *testing.T) {
	// import the device basic info to warehouse
	store := NewDeviceStorage(host, user, password, database)
	if store == nil {
		t.Errorf("init storage failed")
	}
	defer store.Destory()
	// prepare data
	prepare(store)
	// invalid hid
	var invalidHid int64 = 10000000
	device := NewDeviceManager(store)
	devList, err := device.GetAllDevices(domain, invalidHid)
	if err != nil {
		t.Error("get invalid hid devices failed", err)
	} else if len(devList) != 0 {
		t.Error("should not exist devices")
	}
	// get home all device
	home := NewHomeManager(store)
	var uid int64 = 100
	list, err := home.GetAllHome(domain, uid)
	if err != nil || len(list) != 5 {
		t.Errorf("get user all home failed:err[%v], len[%d]", err, len(list))
	}
	// 5 home, 2 master, 3 slave
	for _, home := range list {
		devList, err := device.GetAllDevices(domain, home.hid)
		if err != nil {
			t.Error("get all devices failed", err)
		} else if len(devList) != 8 {
			t.Error("check all devices count failed", len(devList))
		}
		for _, dev := range devList {
			// only delete the master device and all the normal binding deleted
			if dev.IsMasterDevice() {
				err := device.DeleteDevice(domain, home.hid, dev.GetDid())
				if err != nil {
					t.Error("delete did failed", dev.GetDid(), err)
				}
			}
		}
	}
	// check result all device deleted
	for _, home := range list {
		devList, err := device.GetAllDevices(domain, home.hid)
		if err != nil {
			t.Error("get all devices failed", err)
		} else if len(devList) != 0 {
			t.Error("check all devices count failed", len(devList))
		}
	}
	cleanAll(store)

	// recreate home
	prepare(store)
	list, err = home.GetAllHome(domain, uid)
	if err != nil || len(list) != 5 {
		t.Errorf("get user all home failed:err[%v], len[%d]", err, len(list))
	}
	// 5 home, 2 master, 3 slave
	for _, home := range list {
		devList, err := device.GetAllDevices(domain, home.hid)
		if err != nil {
			t.Error("get all devices failed", err)
		} else if len(devList) != 8 {
			t.Error("check all devices count failed", len(devList))
		}
		for _, dev := range devList {
			// only delete the normal device
			if !dev.IsMasterDevice() {
				err := device.DeleteDevice(domain, home.hid, dev.GetDid())
				if err != nil {
					t.Error("delete did failed", dev.GetDid(), err)
				}
			}
		}
	}
	// check result all slave device deleted
	for _, home := range list {
		devList, err := device.GetAllDevices(domain, home.hid)
		if err != nil {
			t.Error("get all devices failed", err)
		} else if len(devList) != 2 {
			t.Error("check master devices count failed", len(devList))
		}
	}

	// invalid did
	var invalidDid int64
	err = device.DeleteDevice(domain, invalidHid, invalidDid)
	if err != nil {
		t.Error("delete invalid did succ")
	}
	cleanAll(store)
}

func TestChangeDeviceInfo(t *testing.T) {
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
	device := NewDeviceManager(store)

	// change device name
	for _, home := range list {
		devList, err := device.GetAllDevices(domain, home.hid)
		if err != nil {
			t.Error("get all devices failed", err)
		} else if len(devList) != 8 {
			t.Error("check all devices count failed", len(devList))
		}
		for _, dev := range devList {
			if !dev.IsMasterDevice() {
				name := fmt.Sprintf("slave%dmaster%d", dev.GetDid(), dev.GetMasterDid())
				err = device.ChangeDeviceName(domain, dev.GetDid(), name)
				if err != nil {
					t.Error("change device info failed", dev.GetDid(), err)
				}
			} else {
				// be frozen
				err = device.Disable(domain, dev.GetDid())
				if err != nil {
					t.Error("disable device status failed", dev.GetDid(), err)
				}
				// can not change the device name
				name := fmt.Sprintf("newmastername%d", dev.GetDid())
				err = device.ChangeDeviceName(domain, dev.GetDid(), name)
				if err == nil {
					t.Error("change frozen device info succ", dev.GetDid(), err)
				}

				// be defrozen
				err = device.Enable(domain, dev.GetDid())
				if err != nil {
					t.Error("disable device status failed", dev.GetDid(), err)
				}
				err = device.ChangeDeviceName(domain, dev.GetDid(), name)
				if err != nil {
					t.Error("change normal device info failed", dev.GetDid(), err)
				}
			}
		}
	}

	// check the new name
	for _, home := range list {
		devList, err := device.GetAllDevices(domain, home.hid)
		if err != nil {
			t.Error("get all devices failed", err)
		} else if len(devList) != 8 {
			t.Error("check all devices count failed", len(devList))
		}
		for _, dev := range devList {
			name := fmt.Sprintf("slave%dmaster%d", dev.GetDid(), dev.GetMasterDid())
			if !dev.IsMasterDevice() {
				if name != dev.GetDeviceName() {
					t.Error("check device name failed", dev.GetDid(), err)
				}
			} else {
				name := fmt.Sprintf("newmastername%d", dev.GetDid())
				if name == dev.GetDeviceName() || dev.GetStatus() != FROZEN {
					t.Error("check the frozen device name changed", dev.GetDid())
				}
			}
		}
	}
	// invalid did
	var invalidDid int64 = 10000000
	err = device.ChangeDeviceName(domain, invalidDid, "xxxx")
	if err == nil {
		t.Error("change not exist device succ", invalidDid, err)
	}
	cleanAll(store)
}
