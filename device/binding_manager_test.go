package device

import (
	"fmt"
	"testing"
)

func cleanAll(store *DeviceStorage) {
	store.Clean(domain, "device_warehouse")
	store.Clean(domain, "device_info")
	store.Clean(domain, "device_mapping")
	store.Clean(domain, "home_info")
	store.Clean(domain, "home_members")
}

// can binding one device more than one times
func TestBinding(t *testing.T) {
	// import the device basic info to warehouse
	store := NewDeviceStorage(host, user, password, database)
	if store == nil {
		t.Errorf("init storage failed")
	}
	defer store.Destory()

	// register master and slave device
	warehouse := NewDeviceWarehouse(store)
	subDomain := "flying"
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("20141017%d", i)
		key := fmt.Sprintf("publicKey%d", i)
		err := warehouse.Register(domain, subDomain, id, key, true)
		if err != nil {
			t.Error("register master device failed", err)
		}
	}
	for i := 0; i < 30; i++ {
		id := fmt.Sprintf("20151017%d", i)
		err := warehouse.Register(domain, subDomain, id, "", false)
		if err != nil {
			t.Error("register master device failed", err)
		}
	}

	// create home for account
	home := NewHomeManager(store)
	var uid int64 = 100
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("home%d", i)
		err := home.Create(domain, uid, name)
		if err != nil {
			t.Errorf("create home failed:name[%s], err[%v]", name, err)
		}
	}
	// check home count of user
	list, err := home.GetAllHome(domain, uid)
	if err != nil || len(list) != 5 {
		t.Errorf("get user all home failed:err[%v], count[%d]", err, len(list))
	}

	binding := NewBindingManager(store)
	// basic info not exist
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("20151017%d", i)
		name := fmt.Sprintf("master%d", i)
		err := binding.Binding(domain, subDomain, id, name, list[i%5].hid, -1)
		if err == nil {
			t.Errorf("binding master device failed:err[%v]", err)
		}
	}
	master := make([]int64, 0)
	// binding master device
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("20141017%d", i)
		name := fmt.Sprintf("master%d", i)
		err = binding.Binding(domain, subDomain, id, name, list[i%5].hid, -1)
		if err != nil {
			t.Errorf("binding master device failed:err[%v]", err)
		}
		var bind BindingInfo
		err = binding.getBindingInfo(domain, subDomain, id, &bind)
		if err != nil {
			t.Errorf("get binding info failed:err[%v]", err)
		} else {
			master = append(master, bind.did)
		}
		// binding again
		err = binding.Binding(domain, subDomain, id, name, list[i%5].hid, -1)
		if err != nil {
			t.Errorf("rebinding master device failed:err[%v]", err)
		}
		var bind2 BindingInfo
		err = binding.getBindingInfo(domain, subDomain, id, &bind2)
		if err != nil || bind2.did != bind.did {
			t.Errorf("get binding info failed:err[%v]", err)
		}
	}

	// binding slave devices
	var invalidDid int64 = 1000000
	// master not binding
	for i := 0; i < 30; i++ {
		id := fmt.Sprintf("20151017%d", i)
		name := fmt.Sprintf("slave%d", i)
		err := binding.Binding(domain, subDomain, id, name, list[i%5].hid, invalidDid)
		if err == nil {
			t.Errorf("binding master device failed:err[%v]", err)
		}
	}

	for i := 0; i < 30; i++ {
		id := fmt.Sprintf("20151017%d", i)
		name := fmt.Sprintf("slave%d", i)
		// bind as master
		err := binding.Binding(domain, subDomain, id, name, list[i%5].hid, -1)
		if err == nil {
			t.Errorf("binding master device failed:err[%v]", err)
		}
		err = binding.Binding(domain, subDomain, id, name, list[i%5].hid, master[i%10])
		if err != nil {
			t.Errorf("binding master device failed:err[%v]", err)
		}
		var bind BindingInfo
		err = binding.getBindingInfo(domain, subDomain, id, &bind)
		if err != nil {
			t.Errorf("get binding info failed:err[%v]", err)
		}
		// bind again
		err = binding.Binding(domain, subDomain, id, name, list[i%5].hid, master[i%10])
		if err != nil {
			t.Errorf("binding master device failed:err[%v]", err)
		}
		var bind2 BindingInfo
		err = binding.getBindingInfo(domain, subDomain, id, &bind2)
		if err != nil || bind2.did != bind.did {
			t.Errorf("get binding info failed:err[%v]", err)
		}
	}
	cleanAll(store)
}

// modify binding id not changed
func TestChangeBinding(t *testing.T) {
	// import the device basic info to warehouse
	store := NewDeviceStorage(host, user, password, database)
	if store == nil {
		t.Errorf("init storage failed")
	}
	defer store.Destory()

	// register master and slave device
	warehouse := NewDeviceWarehouse(store)
	subDomain := "flying"
	for i := 0; i < 10; i++ {
		id := fmt.Sprintf("20141017%d", i)
		key := fmt.Sprintf("publicKey%d", i)
		err := warehouse.Register(domain, subDomain, id, key, true)
		if err != nil {
			t.Error("register master device failed", err)
		}
	}
	for i := 0; i < 30; i++ {
		id := fmt.Sprintf("20151017%d", i)
		err := warehouse.Register(domain, subDomain, id, "", false)
		if err != nil {
			t.Error("register master device failed", err)
		}
	}

	// create home for account
	home := NewHomeManager(store)
	var uid int64 = 100
	for i := 0; i < 5; i++ {
		name := fmt.Sprintf("home%d", i)
		err := home.Create(domain, uid, name)
		if err != nil {
			t.Errorf("create home failed:name[%s], err[%v]", name, err)
		}
	}
	// check home count of user
	list, err := home.GetAllHome(domain, uid)
	if err != nil || len(list) != 5 {
		t.Errorf("get user all home failed:err[%v]", err)
	}

	binding := NewBindingManager(store)
	master := make([]int64, 0)
	// binding master device
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("20141017%d", i)
		name := fmt.Sprintf("master%d", i)
		err := binding.Binding(domain, subDomain, id, name, list[i%5].hid, -1)
		if err != nil {
			t.Errorf("binding master device failed:err[%v]", err)
		}
		var bind BindingInfo
		err = binding.getBindingInfo(domain, subDomain, id, &bind)
		if err != nil {
			t.Errorf("get binding info failed:err[%v]", err)
		} else {
			master = append(master, bind.did)
		}
		// binding again succ
		err = binding.Binding(domain, subDomain, id, name, list[i%5].hid, -1)
		if err != nil {
			t.Errorf("rebinding master device failed:err[%v]", err)
		}
		var bind2 BindingInfo
		err = binding.getBindingInfo(domain, subDomain, id, &bind2)
		if err != nil || bind2.did != bind.did {
			t.Errorf("get binding info failed:err[%v]", err)
		}
	}

	// change master device to new master
	var invalidDid int64 = 100000
	for i := 0; i < 5; i++ {
		id := fmt.Sprintf("20141017%d", i+5)
		// invalid id
		err := binding.ChangeBinding(invalidDid, domain, subDomain, id)
		if err == nil {
			t.Errorf("change binding failed:err[%v]", err)
		}
		// new device not valid
		id = fmt.Sprintf("20151017%d", i+5)
		err = binding.ChangeBinding(master[i], domain, subDomain, id)
		if err == nil {
			t.Errorf("change binding failed:err[%v]", err)
		}

		// succ
		id = fmt.Sprintf("20141017%d", i+5)
		err = binding.ChangeBinding(master[i], domain, subDomain, id)
		if err != nil {
			t.Errorf("change binding failed:err[%v]", err)
		}
		var bind BindingInfo
		err = binding.getBindingInfo(domain, subDomain, id, &bind)
		if err != nil {
			t.Errorf("get binding info failed:err[%v]", err)
		} else if bind.did != master[i] {
			t.Errorf("check binding info error:old[%d], new[%d]", master[i], bind.did)
		}
		// already binded failed
		err = binding.ChangeBinding(master[i], domain, subDomain, id)
		if err == nil {
			t.Errorf("change binding failed:err[%v]", err)
		}
	}
	cleanAll(store)
}
