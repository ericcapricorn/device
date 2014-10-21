package device

import (
	"database/sql"
	"fmt"
	"zc-common-go/common"
	log "zc-common-go/glog"
	_ "zc-common-go/mysql"
)

type DeviceManager struct {
	store *DeviceStorage
}

func NewDeviceManager(store *DeviceStorage) *DeviceManager {
	return &DeviceManager{store: store}
}

//////////////////////////////////////////////////////////////////////////////
/// public interface
//////////////////////////////////////////////////////////////////////////////
func (this *DeviceManager) Domain() string {
	return this.store.GetDomain()
}

// if find the record return device + nil, else if no record return nil + nil
func (this *DeviceManager) Get(did int64) (*DeviceInfo, error) {
	var device DeviceInfo
	err := this.getDeviceInfo(did, &device)
	if err != nil {
		if err == common.ErrEntryNotExist {
			return nil, nil
		}
		log.Warningf("get device info failed:domain[%s], did[%d], err[%v]", this.Domain(), did, err)
		return nil, err
	}
	return &device, nil
}

// get all devices from one home, if no one return empty list not nil
func (this *DeviceManager) GetAllDevices(hid int64) ([]DeviceInfo, error) {
	return this.getAllDevices(hid)
}

// delete one device from home, if it is master device delete all the related slave devices from the home
func (this *DeviceManager) DeleteDevice(hid int64, did int64) error {
	return this.deleteDeviceInfo(hid, did)
}

// delete all devices from one home
func (this *DeviceManager) DeleteAllDevices(hid int64) error {
	return this.deleteAllDevices(hid)
}

// no matter the device status
func (this *DeviceManager) ChangeDeviceStatus(did int64, status int8) error {
	return this.modifyDeviceInfo(false, did, "status", status)
}

// TODO the primary key info and status should not changed by this interface
func (this *DeviceManager) ChangeDeviceInfo(did int64, key string, value interface{}) error {
	return this.modifyDeviceInfo(true, did, key, value)
}

////////////////////////////////////////////////////////////////////////////////////
// private interface database related
////////////////////////////////////////////////////////////////////////////////////
func (this *DeviceManager) getAllDevices(hid int64) ([]DeviceInfo, error) {
	SQL := fmt.Sprintf("SELECT did, hid, name, status, master_did FROM %s_device_info WHERE hid = ?", this.Domain())
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Warningf("prepare query all home devices failed:domain[%s], hid[%d], err[%v]", this.Domain(), hid, err)
		return nil, err
	}
	rows, err := stmt.Query(hid)
	if err != nil {
		log.Warningf("query the device info of one home failed:domain[%s], hid[%d], err[%v]", this.Domain(), hid, err)
		return nil, err
	}
	var device DeviceInfo
	list := make([]DeviceInfo, 0)
	for rows.Next() {
		err = rows.Scan(&device.did, &device.hid, &device.deviceName, &device.status, &device.masterDid)
		if err != nil {
			log.Warningf("parse the result failed:domain[%s], hid[%d], err[%v]", this.Domain(), hid, err)
			return nil, err
		}
		list = append(list, device)
	}
	return list, nil
}

// delete all device info in home
func (this *DeviceManager) deleteAllDevices(hid int64) error {
	SQL := fmt.Sprintf("DELETE FROM %s_device_info WHERE hid = ?", this.Domain())
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Warningf("prepare delete all devices of home failed:domain[%s], hid[%d], err[%v]", this.Domain(), hid, err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(hid)
	if err != nil {
		log.Errorf("delete all device of home failed:domain[%s], hid[%d], err[%v]", this.Domain(), hid, err)
		return err
	}
	return nil
}

// get device info
func (this *DeviceManager) getDeviceInfo(did int64, device *DeviceInfo) error {
	SQL := fmt.Sprintf("SELECT did, hid, name, status, master_did FROM %s_device_info WHERE did = ?", this.Domain())
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:domain[%s], did[%d], err[%v]", this.Domain(), did, err)
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRow(did).Scan(&device.did, &device.hid, &device.deviceName, &device.status, &device.masterDid)
	if err != nil {
		if err == sql.ErrNoRows {
			return common.ErrEntryNotExist
		} else {
			log.Errorf("get deive info failed:domain[%s], did[%d]", this.Domain(), did)
			return err
		}
	}
	return nil
}

// modify device info all column can be modified
func (this *DeviceManager) modifyDeviceInfo(checkStatus bool, did int64, key string, value interface{}) error {
	common.CheckParam(len(key) != 0)
	var SQL string
	if checkStatus {
		SQL = fmt.Sprintf("UPDATE %s_device_info SET %s = ? WHERE did = ? AND status = %d", this.Domain(), key, ACTIVE)
	} else {
		SQL = fmt.Sprintf("UPDATE %s_device_info SET %s = ? WHERE did = ?", this.Domain(), key)
	}
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:err[%v]", err)
		return err
	}
	defer stmt.Close()
	result, err := stmt.Exec(value, did)
	if err != nil {
		log.Errorf("execute update failed:domain[%s], key[%s], value[%v], err[%v]", this.Domain(), key, value, err)
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Warningf("get affected rows failed:err[%v]", err)
		return err
	}
	if affect != 1 {
		log.Warningf("update device info failed:domain[%s], key[%s], value[%v], err[%v]", this.Domain(), key, value, err)
		return common.ErrEntryNotExist
	}
	return nil
}

// delete master or normal device
func (this *DeviceManager) deleteDeviceInfo(hid, did int64) error {
	// at first delete the device from the device info
	SQL1 := fmt.Sprintf("DELETE FROM %s_device_info WHERE did = ? AND hid = ?", this.Domain())
	stmt1, err := this.store.db.Prepare(SQL1)
	if err != nil {
		log.Errorf("prepare delete device failed:domain[%s], hid[%d], did[%d], err[%v]", this.Domain(), hid, did, err)
		return err
	}
	defer stmt1.Close()
	// delete all the related normal devices if not master step 1 must do
	SQL2 := fmt.Sprintf("DELETE FROM %s_device_info WHERE hid = ? AND master_did = ?", this.Domain())
	stmt2, err := this.store.db.Prepare(SQL2)
	if err != nil {
		log.Errorf("prepare delete all normal device failed:domain[%s], hid[%d], did[%d], err[%v]", this.Domain(), hid, did, err)
		return err
	}
	defer stmt2.Close()

	// begin in a transaction
	tx, err := this.store.db.Begin()
	if err != nil {
		log.Errorf("begin transaction failed:domain[%s], hid[%d], did[%d], err[%v]", this.Domain(), hid, did, err)
		return err
	}
	defer rollback(&err, tx)

	_, err = tx.Stmt(stmt1).Exec(did, hid)
	if err != nil {
		log.Errorf("delete did failed:domain[%s], hid[%d], did[%d], err[%v]", this.Domain(), hid, did, err)
		return err
	}
	_, err = tx.Stmt(stmt2).Exec(hid, did)
	if err != nil {
		log.Errorf("delete all the device related to this device failed:domain[%s], hid[%d], did[%d], err[%v]",
			this.Domain(), hid, did, err)
		return err
	}
	err = tx.Commit()
	if err != nil {
		log.Errorf("commit failed:domain[%s], hid[%d], did[%d], err[%v]", this.Domain(), hid, did, err)
		return err
	}
	return nil
}
