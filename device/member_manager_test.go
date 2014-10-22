package device

import (
	"fmt"
	"testing"
)

func TearDown(store *DeviceStorage) {
	store.Clean(domain, "home_info")
	store.Clean(domain, "home_members")
	store.Destory()
}

var fakeHid int64 = 100
var fakeUid int64 = 100

func CreateHome(uid int64, store *DeviceStorage) (int64, error) {
	// create a home at first
	home := NewHomeManager(store)
	err := home.Create(domain, int64(uid), "home")
	if err != nil {
		fmt.Print("create home failed", err)
		return -1, err
	}
	list, err := home.GetAllHome(domain, uid)
	if err != nil || len(list) != 1 {
		fmt.Print("get user all home failed", err)
		return -1, err
	}
	return list[0].hid, nil
}

func TestAddMember(t *testing.T) {
	store := NewDeviceStorage(host, user, password, database)
	if store == nil {
		t.Errorf("init storage failed")
	}
	manager := NewMemberManager(store)
	defer TearDown(store)

	// add owner
	err := manager.AddOwner(domain, "owner", fakeHid, fakeUid)
	if err != nil {
		t.Error("add owner failed", err)
	}
	// home not exist
	for i := 0; i < 10; i++ {
		err = manager.AddMember(domain, "guest", fakeHid, fakeUid)
		if err == nil {
			t.Error("add guest failed", err)
		}
	}
	for i := 0; i < 10; i++ {
		err = manager.AddMember(domain, "guest", fakeHid, int64(i+1))
		if err == nil {
			t.Error("add guest failed", err)
		}
	}

	// create home
	validHid, err := CreateHome(fakeUid, store)
	if err != nil {
		t.Error("create new home failed", err)
	}

	// valid home id add owner as member return succ
	err = manager.AddMember(domain, "guest", validHid, fakeUid)
	if err != nil {
		t.Error("add owner oneself failed", err)
	}

	for i := 0; i < 10; i++ {
		err = manager.AddMember(domain, "guest", validHid, int64(i+1))
		if err != nil {
			t.Error("add guest failed", err)
		}
	}
}

func TestDeleteMember(t *testing.T) {
	store := NewDeviceStorage(host, user, password, database)
	if store == nil {
		t.Error("init storage failed")
	}
	manager := NewMemberManager(store)
	defer TearDown(store)
	// home not exist, member not exist too
	err := manager.Delete(domain, fakeHid, fakeUid)
	if err == nil {
		t.Error("delete not exist hid failed")
	}
	// create home
	validHid, err := CreateHome(fakeUid, store)
	if err != nil {
		t.Error("create new home failed", err)
	}
	// add member to the valid hid
	for i := 0; i < 10; i++ {
		err = manager.AddMember(domain, "guest", validHid, int64(i+1))
		if err != nil {
			t.Error("add guest failed", err)
		}
	}
	// delete exist uid
	for i := 0; i < 10; i++ {
		err = manager.Delete(domain, validHid, int64(i+1))
		if err != nil {
			t.Error("delete the user failed", err)
		}
		// home exist, member not exist
		err = manager.Delete(domain, validHid, int64(i+1))
		if err != nil {
			t.Error("delete the user failed", err)
		}
	}
}

func TestGetMemberInfo(t *testing.T) {
	store := NewDeviceStorage(host, user, password, database)
	if store == nil {
		t.Error("init storage failed")
	}
	manager := NewMemberManager(store)
	defer TearDown(store)
	// create home
	validHid, err := CreateHome(fakeUid, store)
	if err != nil {
		t.Error("create new home failed", err)
	}
	var member *Member
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("guest%d", i)
		err = manager.AddMember(domain, name, validHid, int64(i+1))
		if err != nil {
			t.Error("add guest failed", err)
		}
		member, err = manager.Get(domain, validHid, int64(i+1))
		if err != nil {
			t.Error("get guest member failed", err)
		} else if member == nil {
			t.Error("check member not exist")
		} else if member.memberName != name {
			t.Error("check name failed", member.memberName, name)
		} else if member.uid != int64(i+1) {
			t.Error("check member uid failed", member.uid, i+1)
		} else if member.memberType != NORMAL {
			t.Error("check normal member failed", member.memberType)
		} else if member.hid != validHid {
			t.Error("check member hid failed", member.hid)
		} else if member.status != ACTIVE {
			t.Error("check member status failed")
		}
	}
	// invalid hid
	member, err = manager.Get(domain, fakeHid, member.uid)
	if err != nil {
		t.Error("get home member failed")
	} else if member != nil {
		t.Error("should not get not exist home's member")
	}
	// valid hid not exist member info
	member, err = manager.Get(domain, validHid, 10000000)
	if err != nil {
		t.Error("get not exist member failed")
	} else if member != nil {
		t.Error("should not get not exist member")
	}
}

func TestAllMember(t *testing.T) {
	store := NewDeviceStorage(host, user, password, database)
	if store == nil {
		t.Error("init storage failed")
	}
	manager := NewMemberManager(store)
	defer TearDown(store)

	// not exist hid, return empty list
	list, err := manager.GetAllMembers(domain, fakeHid)
	if err != nil {
		t.Error("get fake hid members failed")
	} else if list == nil || len(list) != 0 {
		t.Error("get fake hid members succ")
	}

	// not exist but return succ
	err = manager.DeleteAllMembers(domain, fakeHid)
	if err != nil {
		t.Error("delete not exist home members failed", err)
	}

	// create home return the valid hid
	validHid, err := CreateHome(fakeUid, store)
	if err != nil {
		t.Error("create new home failed", err)
	}

	// no members return empty list
	list, err = manager.GetAllMembers(domain, validHid)
	if err != nil {
		t.Error("get all members failed", err)
	} else if list == nil {
		t.Error("check member list failed")
	}

	// add member
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("guest%d", i)
		err = manager.AddMember(domain, name, validHid, int64(i+1))
		if err != nil {
			t.Error("add guest failed", err)
		}
	}
	// get all
	list, err = manager.GetAllMembers(domain, validHid)
	if err != nil {
		t.Error("get all member failed", err)
	} else if list == nil || len(list) != 11 {
		t.Error("check member list failed")
	}

	// delete all
	err = manager.DeleteAllMembers(domain, validHid)
	if err != nil {
		t.Error("delete all members failed", err)
	}
	// get all again
	list, err = manager.GetAllMembers(domain, validHid)
	if err != nil {
		t.Error("get all member failed", err)
	} else if list == nil || len(list) != 0 {
		t.Error("check member list failed after delete all")
	}
}

func TestModifyMemberName(t *testing.T) {
	store := NewDeviceStorage(host, user, password, database)
	if store == nil {
		t.Error("init storage failed")
	}
	manager := NewMemberManager(store)
	defer TearDown(store)

	// hid not exist
	err := manager.ModifyName(domain, fakeHid, fakeUid, "fakeName")
	if err == nil {
		t.Error("modify not exist hid, uid succ")
	}

	// create home
	validHid, err := CreateHome(fakeUid, store)
	if err != nil {
		t.Error("create new home failed", err)
	}

	// hid exist but uid not exist
	err = manager.ModifyName(domain, validHid, fakeUid+1, "fakeName")
	if err == nil {
		t.Error("modify not exist uid succ")
	}

	// modify member name
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("guest%d", i)
		err = manager.AddMember(domain, name, validHid, int64(i+1))
		if err != nil {
			t.Error("add guest failed", err)
		}
		newName := fmt.Sprintf("%dguest", i)
		err = manager.ModifyName(domain, validHid, int64(i+1), newName)
		if err != nil {
			t.Error("modify name failed", err)
		}
		// check the new name
		member, err := manager.Get(domain, validHid, int64(i+1))
		if err != nil {
			t.Error("get member failed", err)
		} else if member == nil || member.GetMemberName() != newName {
			t.Error("check member or new name failed", member.GetMemberName())
		}
	}
}

func TestEnableMember(t *testing.T) {
	store := NewDeviceStorage(host, user, password, database)
	if store == nil {
		t.Error("init storage failed")
	}
	manager := NewMemberManager(store)
	defer TearDown(store)

	// hid not exist
	err := manager.Enable(domain, fakeHid, fakeUid)
	if err == nil {
		t.Error("enable not exist hid, uid succ")
	}

	err = manager.Disable(domain, fakeHid, fakeUid)
	if err == nil {
		t.Error("disable not exist hid, uid succ")
	}

	// create home
	validHid, err := CreateHome(fakeUid, store)
	if err != nil {
		t.Error("create new home failed", err)
	}

	// hid exist but uid not exist
	err = manager.Enable(domain, validHid, fakeUid+1)
	if err == nil {
		t.Error("enable not exist uid succ")
	}

	err = manager.Disable(domain, validHid, fakeUid+1)
	if err == nil {
		t.Error("disable not exist uid succ")
	}

	// add members
	for i := 0; i < 10; i++ {
		name := fmt.Sprintf("guest%d", i)
		err = manager.AddMember(domain, name, validHid, int64(i+1))
		if err != nil {
			t.Error("add guest failed", err)
		}
	}
	// disable member
	for i := 0; i < 10; i++ {
		err = manager.Disable(domain, validHid, int64(i+1))
		if err != nil {
			t.Error("disable member failed", err)
		}
		// check the new name
		member, err := manager.Get(domain, validHid, int64(i+1))
		if err != nil {
			t.Error("get member failed", err)
		} else if member == nil || member.GetStatus() == ACTIVE {
			t.Error("disable member failed")
		}
	}

	// enable member
	for i := 0; i < 10; i++ {
		err = manager.Enable(domain, validHid, int64(i+1))
		if err != nil {
			t.Error("disable member failed", err)
		}
		// check the new name
		member, err := manager.Get(domain, validHid, int64(i+1))
		if err != nil {
			t.Error("get member failed", err)
		} else if member == nil || member.GetStatus() != ACTIVE {
			t.Error("enable member failed")
		}
	}
}
