package device

import (
	"fmt"
	"testing"
)

func TestImportDevice(t *testing.T) {
	store := NewDeviceStorage(host, user, password, domain, database)
	if store == nil {
		t.Errorf("init storage failed")
	}
	manager := NewDeviceWarehouse(store)
	defer store.Destory()
	subDomain := "flying"
	// register device
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("20141017%d", i)
		key := fmt.Sprintf("secret%d", i)
		err := manager.Register(subDomain, id, key, true)
		if err != nil {
			t.Error("register master device failed", err)
		}
		device, err := manager.Get(subDomain, id)
		if err != nil {
			t.Error("get device failed", err)
		} else if device == nil {
			t.Error("device should not exist")
		} else if !device.Validate() {
			t.Error("device invalid")
		}
	}

	// already exist
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("20141017%d", i)
		key := fmt.Sprintf("secret%d", i)
		err := manager.Register(subDomain, id, key, true)
		if err == nil {
			t.Error("register master device failed", err)
		}
	}

	// another subdomain
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("20141017%d", i)
		key := fmt.Sprintf("secret%d", i)
		err := manager.Register("housing", id, key, true)
		if err != nil {
			t.Error("register device failed", err)
		}
	}

	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("20141017%d", i)
		key := fmt.Sprintf("secret%d", i)
		// regist slave device
		err := manager.Register(subDomain, id, key, false)
		if err == nil {
			t.Error("register slave device failed", err)
		}
	}

	store.Clean("device_warehouse")
}

func TestDeleteDevice(t *testing.T) {
	store := NewDeviceStorage(host, user, password, domain, database)
	if store == nil {
		t.Errorf("init storage failed")
	}
	manager := NewDeviceWarehouse(store)
	defer store.Destory()
	subDomain := "flying"
	// register device
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("20141017%d", i)
		key := fmt.Sprintf("secret%d", i)
		err := manager.Register(subDomain, id, key, true)
		if err != nil {
			t.Error("register master device failed", err)
		}
	}

	// delete device
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("20141017%d", i)
		err := manager.Delete(subDomain, id)
		if err != nil {
			t.Error("delete device failed", err)
		}
		device, err := manager.Get(subDomain, id)
		if err != nil {
			t.Error("get device failed", err)
		} else if device != nil {
			t.Error("device should not exist")
		}
	}

	// delete not exist device return succ
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("20141017%d", i)
		err := manager.Delete(subDomain, id)
		if err != nil {
			t.Error("delete device failed", err)
		}
	}
	store.Clean("device_warehouse")
}

func TestGetDevice(t *testing.T) {
	store := NewDeviceStorage(host, user, password, domain, database)
	if store == nil {
		t.Errorf("init storage failed")
	}
	manager := NewDeviceWarehouse(store)
	defer store.Destory()
	subDomain := "flying"
	// register device
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("20141017%d", i)
		key := fmt.Sprintf("secret%d", i)
		err := manager.Register(subDomain, id, key, true)
		if err != nil {
			t.Error("register master device failed", err)
		}
		err = manager.Register("another", id, key, false)
		if err != nil {
			t.Error("register slave device failed", err)
		}
	}

	// get not exist device
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("20151017%d", i)
		device, err := manager.Get(subDomain, id)
		if err != nil || device != nil {
			t.Error("get device failed", err)
		}
		device, err = manager.Get("another", id)
		if err != nil || device != nil {
			t.Error("get device failed")
		}
	}

	// get exist device
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("20141017%d", i)
		key := fmt.Sprintf("secret%d", i)
		device, err := manager.Get(subDomain, id)
		if err != nil || device == nil {
			t.Error("get device failed", err)
		} else if !device.Validate() {
			t.Error("device invalid")
		}
		if device.subDomain != subDomain || device.deviceId != id || device.publicKey.String != key {
			t.Error("device info error")
		}
		// another subdomain
		device, err = manager.Get("another", id)
		if err != nil || device == nil {
			t.Error("get device failed")
		} else if !device.Validate() {
			t.Error("device invalid")
		}
		// slave device without key
		if device.subDomain != "another" || device.deviceId != id || device.publicKey.Valid == true {
			t.Error("device info error")
		}
	}
	store.Clean("device_warehouse")
}
