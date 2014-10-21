package device

import (
	"database/sql"
	"fmt"
	"zc-common-go/common"
	log "zc-common-go/glog"
	_ "zc-common-go/mysql"
)

type DeviceWarehouse struct {
	store *DeviceStorage
}

func NewDeviceWarehouse(store *DeviceStorage) *DeviceWarehouse {
	return &DeviceWarehouse{store: store}
}

//////////////////////////////////////////////////////////////////////////////
/// public interface
//////////////////////////////////////////////////////////////////////////////
func (this *DeviceWarehouse) Domain() string {
	return this.store.GetDomain()
}

// if find the record return basic info + nil
func (this *DeviceWarehouse) Get(subDomain, deviceId string) (*BasicInfo, error) {
	var device BasicInfo
	err := this.getDeviceinfo(subDomain, deviceId, &device)
	if err != nil {
		if err == common.ErrEntryNotExist {
			return nil, nil
		} else {
			log.Warningf("get device info failed:domain[%s], device[%s:%s]", this.Domain(), subDomain, deviceId)
			return nil, err
		}
	}
	if device.Validate() {
		return &device, nil
	} else {
		log.Errorf("check device info failed:domain[%s], device[%s:%s]", this.Domain(), subDomain, deviceId)
		return nil, common.ErrInvalidDevice
	}
	return &device, nil
}

// WARNING: must be cautious for using this interface
// delete the device only for not online device
func (this *DeviceWarehouse) Delete(subDomain, deviceId string) error {
	err := this.deleteDeviceInfo(subDomain, deviceId)
	if err != nil {
		log.Warningf("delete the device failed:domain[%s], device[%s:%s]", this.Domain(), subDomain, deviceId)
		return err
	}
	return nil
}

// import a new device
func (this *DeviceWarehouse) Register(subDomain, deviceId, publicKey string, master bool) error {
	if len(subDomain) <= 0 || len(deviceId) <= 0 {
		log.Warningf("check domain[%s] device[%s:%s] failed", this.Domain(), subDomain, deviceId)
		return common.ErrInvalidParam
	} else if master && len(publicKey) <= 0 {
		log.Warningf("check public key length failed:domain[%s], device[%s:%s]", this.Domain(), subDomain, deviceId)
		return common.ErrInvalidParam
	}
	err := this.insertDeviceInfo(subDomain, deviceId, publicKey, master)
	if err != nil {
		log.Warningf("insert device info failed:domain[%s], device[%s:%s], master[%t]",
			this.Domain(), subDomain, deviceId, master)
		return err
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////
// database related private interface
//////////////////////////////////////////////////////////////////////////////
func (this *DeviceWarehouse) getDeviceinfo(subDomain, deviceId string, basic *BasicInfo) error {
	SQL := fmt.Sprintf("SELECT sub_domain, device_id, device_type, public_key, status FROM %s_device_warehouse WHERE sub_domain = ? AND device_id = ?", this.Domain())
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:err[%v]", err)
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRow(subDomain, deviceId).Scan(&basic.subDomain, &basic.deviceId, &basic.deviceType, &basic.publicKey, &basic.status)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Errorf("query failed:domain[%s], device[%s:%s]", this.Domain(), subDomain, deviceId)
			return err
		} else {
			return common.ErrEntryNotExist
		}
	}
	return nil
}

func (this *DeviceWarehouse) insertDeviceInfo(subDomain, deviceId, publicKey string, master bool) error {
	var SQL string
	if master {
		SQL = fmt.Sprintf("INSERT INTO %s_device_warehouse(sub_domain, device_id, device_type, public_key, status) VALUES(?,?,?,?,?)", this.Domain())
	} else {
		SQL = fmt.Sprintf("INSERT INTO %s_device_warehouse(sub_domain, device_id, device_type, status) VALUES(?,?,?,?)", this.Domain())
	}
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:err[%v]", err)
		return err
	}
	defer stmt.Close()
	if master {
		_, err = stmt.Exec(subDomain, deviceId, MASTER, publicKey, ACTIVE)
	} else {
		_, err = stmt.Exec(subDomain, deviceId, NORMAL, ACTIVE)
	}
	if err != nil {
		log.Warningf("execute insert device[%s:%s] failed:domain[%s], err[%v]", subDomain, deviceId, this.Domain(), err)
		return err
	}
	return nil
}

func (this *DeviceWarehouse) deleteDeviceInfo(subDomain, deviceId string) error {
	SQL := fmt.Sprintf("DELETE FROM %s_device_warehouse WHERE sub_domain = ? AND device_id = ?", this.Domain())
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:err[%v]", err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(subDomain, deviceId)
	if err != nil {
		log.Errorf("delete the device failed:domain[%s], subDomain[%s], deviceId[%s], err[%s]",
			this.Domain(), subDomain, deviceId, err)
		return err
	}
	return nil
}
