package device

import (
	"fmt"
	"testing"
)

func TestCreatemanager(t *testing.T) {
	store := NewDeviceStorage(host, user, password, domain, database)
	if store == nil {
		t.Errorf("init storage failed")
	}
	manager := NewHomeManager(store)
	defer store.Destory()
	// create a manager for one user
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("home%d", i)
		err := manager.Create(int64(i+1), name)
		if err != nil {
			t.Errorf("create home failed:name[%s], err[%v]", name, err)
		}
	}
	// create by one user uid = 1
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("home%d", i)
		err := manager.Create(1, name)
		if err != nil {
			t.Errorf("create home failed:uid[%d], name[%s], err[%v]", 1, name, err)
		}
	}

	// check home count of user 1
	list, err := manager.GetAllHome(1)
	if err != nil || len(list) != 11 {
		t.Errorf("get user all home failed:err[%v]", err)
	}

	// get home info
	for _, home := range list {
		temp, err := manager.Get(home.GetHid())
		if err != nil {
			t.Errorf("get home failed:hid[%d], err[%v]", home.GetHid(), err)
		}
		if home.GetName() != temp.GetName() || home.GetStatus() != temp.GetStatus() ||
			home.GetCreateUid() != temp.GetCreateUid() || home.GetHid() != temp.GetHid() {
			t.Errorf("check home info not right")
		}
	}
	store.Clean("home_info")
	store.Clean("home_members")
}

func TestDeletemanager(t *testing.T) {
	store := NewDeviceStorage(host, user, password, domain, database)
	if store == nil {
		t.Errorf("init storage failed")
	}
	manager := NewHomeManager(store)
	defer store.Destory()
	// create a manager for one user
	var uid int64 = 1
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("home%d", i)
		err := manager.Create(uid, name)
		if err != nil {
			t.Errorf("create home failed:name[%s], err[%v]", name, err)
		}
	}
	// get the user all manager info
	list, err := manager.GetAllHome(uid)
	if err != nil || len(list) != 10 {
		t.Errorf("get user all home failed:err[%v]", err)
	}

	// delete all managers
	for _, home := range list {
		err = manager.Delete(home.hid)
		if err != nil {
			t.Errorf("delete home failed:hid[%d], err[%v]", home.hid, err)
		}

		// delete not exist return nil
		err = manager.Delete(home.hid)
		if err != nil {
			t.Errorf("delete not exist home failed:hid[%d], err[%v]", home.hid, err)
		}
	}
	// get the user all manager info
	list, err = manager.GetAllHome(uid)
	if err != nil || len(list) != 0 {
		t.Errorf("get user all home failed:err[%v]", err)
	}
	store.Clean("home_info")
	store.Clean("home_members")
	store.Clean("device_info")
}

func TestDisable(t *testing.T) {
	store := NewDeviceStorage(host, user, password, domain, database)
	if store == nil {
		t.Errorf("init storage failed")
	}
	manager := NewHomeManager(store)
	defer store.Destory()
	// create a manager for one user
	var uid int64 = 1
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("home%d", i)
		err := manager.Create(uid, name)
		if err != nil {
			t.Errorf("create home failed:name[%s], err[%v]", name, err)
		}
	}
	// get the user all manager info
	list, err := manager.GetAllHome(uid)
	if err != nil || len(list) != 10 {
		t.Errorf("get user all home failed:err[%v]", err)
	}
	// disable enable modify name
	for _, home := range list {
		err = manager.Disable(home.hid)
		if err != nil {
			t.Errorf("disable home failed:hid[%d], err[%v]", home.hid, err)
		}
		name := fmt.Sprintf("Myhome%d", home.hid)
		err = manager.ModifyName(home.hid, name)
		if err == nil {
			t.Errorf("modify disable home succ:hid[%d]", home.hid)
		}
		temp, err := manager.Get(home.hid)
		if err != nil {
			t.Errorf("get home failed:hid[%d], err[%v]", home.hid, err)
		} else if temp.GetName() == name {
			t.Errorf("check home name modified:hid[%d]", home.hid)
		}
	}

	for _, home := range list {
		err = manager.Enable(home.hid)
		if err != nil {
			t.Errorf("enable home failed:hid[%d], err[%v]", home.hid, err)
		}
		name := fmt.Sprintf("Myhome%d", home.hid)
		err = manager.ModifyName(home.hid, name)
		if err != nil {
			t.Errorf("modify home name succ:hid[%d]", home.hid)
		}
		temp, err := manager.Get(home.hid)
		if err != nil {
			t.Errorf("get home failed:hid[%d], err[%v]", home.hid, err)
		} else if temp.GetName() != name {
			t.Errorf("home name not modified:hid[%d]", home.hid)
		}
	}
	store.Clean("home_info")
	store.Clean("home_members")
}
