package device

import (
	"database/sql"
	"fmt"
	"zc-common-go/common"
	log "zc-common-go/glog"
	_ "zc-common-go/mysql"
)

type HomeManager struct {
	store *DeviceStorage
}

// not create the db instance
func NewHomeManager(store *DeviceStorage) *HomeManager {
	return &HomeManager{store: store}
}

////////////////////////////////////////////////////////////////////////////////////
// public interface
/////////////////////////////////////////////////////////////////////////////////////
// if find the record return home + nil, else if no record return nil + nil
func (this *HomeManager) Get(domain string, hid int64) (*Home, error) {
	common.CheckParam(this.store != nil)
	var home Home
	err := this.getHome(domain, hid, &home)
	if err != nil {
		if err == common.ErrEntryNotExist {
			return nil, nil
		} else {
			log.Warningf("get home failed:domain[%s], hid[%d]", domain, hid)
			return nil, err
		}
	}
	return &home, nil
}

// create a new home
func (this *HomeManager) Create(domain string, uid int64, name string) error {
	common.CheckParam(this.store != nil)
	if len(name) <= 0 {
		log.Warningf("check the home name failed:uid[%d], name[%s]", uid, name)
		return common.ErrInvalidParam
	}
	hid, err := this.insertHome(domain, uid, name)
	if err != nil {
		log.Warningf("insert home failed:domain[%s], createUid[%d], name[%s]", domain, uid, name)
		return err
	}
	// insert member as creator type to the home_members
	member := NewMemberManager(this.store)
	err = member.AddOwner(domain, "owner", hid, uid)
	if err != nil {
		log.Errorf("add owner to the home failed:domain[%s], createUid[%d], name[%s]", domain, uid, name)
		return err
	}
	return nil
}

// TODO do not really delete the home info from the storage
func (this *HomeManager) Delete(domain string, hid int64) error {
	common.CheckParam(this.store != nil)
	// step 1.delete all the devices not in a transaction
	device := NewDeviceManager(this.store)
	err := device.DeleteAllDevices(domain, hid)
	if err != nil {
		log.Warningf("delete the home devices failed:domain[%s], hid[%d], err[%v]", domain, hid, err)
		return err
	}
	// step 2. delete all the members not in a transaction
	member := NewMemberManager(this.store)
	err = member.DeleteAllMembers(domain, hid)
	if err != nil {
		log.Errorf("delete the home all members failed:domain[%s], hid[%d], err[%v]", domain, hid, err)
		return err
	}
	// step 3. delete the home info from
	err = this.deleteHome(domain, hid)
	if err != nil {
		log.Warningf("delete home info failed:domain[%s], hid[%d], err[%v]", domain, hid, err)
		return err
	}
	return nil
}

// if no home, return empty list not nil
func (this *HomeManager) GetAllHome(domain string, uid int64) ([]Home, error) {
	common.CheckParam(this.store != nil)
	member := NewMemberManager(this.store)
	homeIds, err := member.GetAllHomeIds(domain, uid)
	if err != nil {
		log.Warningf("get all home id list failed:domain[%s], uid[%d], err[%v]", domain, uid, err)
		return nil, err
	}
	list := make([]Home, 0)
	for _, hid := range homeIds {
		home, err := this.Get(domain, hid)
		if err != nil {
			log.Warningf("get home info failed:domain[%s], uid[%d], hid[%d]", domain, uid, hid)
			return nil, err
		} else if home == nil {
			log.Warningf("check home not exist:domain[%s], uid[%d], hid[%d]", domain, uid, hid)
		} else {
			list = append(list, *home)
		}
	}
	return list, nil
}

// enable/disable home member control
func (this *HomeManager) Disable(domain string, hid int64) error {
	common.CheckParam(this.store != nil)
	return this.modifyHome(domain, hid, "status", FROZEN)
}

func (this *HomeManager) Enable(domain string, hid int64) error {
	common.CheckParam(this.store != nil)
	return this.modifyHome(domain, hid, "status", ACTIVE)
}

func (this *HomeManager) ModifyName(domain string, hid int64, name string) error {
	common.CheckParam(this.store != nil)
	home, err := this.Get(domain, hid)
	if err != nil {
		log.Warningf("get home failed:domain[%s], hid[%d], err[%v]", domain, home, err)
		return err
	} else if home == nil {
		log.Warningf("home not exist:domain[%s], hid[%d]", domain, hid)
		return common.ErrEntryNotExist
	} else if home.GetStatus() != ACTIVE {
		log.Warningf("home is not active:domain[%s], hid[%d]", domain, hid)
		return common.ErrInvalidStatus
	}
	return this.modifyHome(domain, hid, "name", name)
}

////////////////////////////////////////////////////////////////////////////////////
// database related private interface
////////////////////////////////////////////////////////////////////////////////////
func (this *HomeManager) modifyHome(domain string, hid int64, key string, value interface{}) error {
	SQL := fmt.Sprintf("UPDATE %s_home_info SET %s = ? WHERE hid = ?", domain, key)
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:err[%v]", err)
		return err
	}
	defer stmt.Close()
	result, err := stmt.Exec(value, hid)
	if err != nil {
		log.Errorf("execute update failed:sql[%s], err[%v]", SQL, err)
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Warningf("affected rows failed:err[%v]", err)
		return err
	}
	if affect < 1 {
		log.Warningf("check affected rows failed:domain[%s], hid[%d], row[%d]", domain, hid, affect)
		return common.ErrAccountNotExist
	} else if affect > 1 {
		log.Errorf("check affected rows failed:domain[%s], key[%s], value[%v], hid[%d], row[%d]",
			domain, key, value, hid, affect)
		return common.ErrUnknown
	}
	return nil
}

func (this *HomeManager) getHome(domain string, hid int64, home *Home) error {
	SQL := fmt.Sprintf("SELECT hid, name, status, create_uid FROM %s_home_info WHERE hid = ?", domain)
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:err[%v]", err)
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRow(hid).Scan(&home.hid, &home.name, &home.status, &home.createUid)
	if err != nil {
		if err == sql.ErrNoRows {
			return common.ErrEntryNotExist
		} else {
			log.Warningf("get home info failed:domain[%s], hid[%d]", domain, hid)
			return err
		}
	}
	return nil
}

func (this *HomeManager) deleteHome(domain string, hid int64) error {
	SQL := fmt.Sprintf("DELETE FROM %s_home_info WHERE hid = ?", domain)
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:err[%v]", err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(hid)
	if err != nil {
		log.Errorf("insert a new home failed:domain[%s], hid[%d], err[%v]", domain, hid, err)
		return err
	}
	return nil
}

func (this *HomeManager) insertHome(domain string, uid int64, name string) (int64, error) {
	// TODO modify this in one transaction
	SQL := fmt.Sprintf("INSERT INTO %s_home_info(name, status, create_uid) VALUES(?,?,?)", domain)
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:err[%v]", err)
		return -1, err
	}
	defer stmt.Close()
	result, err := stmt.Exec(name, ACTIVE, uid)
	if err != nil {
		log.Errorf("create new home failed:domain[%s], createUid[%d], name[%s], err[%v]", domain, uid, name, err)
		return -1, err
	}
	hid, err := result.LastInsertId()
	if err != nil {
		log.Errorf("get last insert id failed:domain[%s], createUid[%d], name[%s], err[%v]", domain, uid, name, err)
		return -1, err
	}
	return hid, nil
}
