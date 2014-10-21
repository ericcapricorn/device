package device

import (
	"database/sql"
	"fmt"
	"zc-common-go/common"
	log "zc-common-go/glog"
	_ "zc-common-go/mysql"
)

type BindingManager struct {
	store *DeviceStorage
}

func NewBindingManager(store *DeviceStorage) *BindingManager {
	return &BindingManager{store: store}
}

//////////////////////////////////////////////////////////////////////////////
/// public interface can not delete the binding info in normal cases
//////////////////////////////////////////////////////////////////////////////
func (this *BindingManager) Domain() string {
	return this.store.GetDomain()
}

// binding one device to one home, if masterDid < 0 it's master device, otherwise it's slave device
func (this *BindingManager) Binding(subDomain, deviceId, deviceName string, hid, masterDid int64) error {
	// step 1. check the device basic info is valid
	err := this.checkDeviceInfo(subDomain, deviceId, masterDid < 0)
	if err != nil {
		log.Warningf("check the binding device failed:err[%v]", err)
		return err
	}
	// step 2. check the master device is binding
	err = this.checkMasterDevice(subDomain, deviceId, masterDid)
	if err != nil {
		log.Warningf("check the master device not active")
		return err
	}
	// step 3. build mapping device ids, if already exist return succ for rebinding...
	err = this.bindingDevice(subDomain, deviceId, deviceName, hid, masterDid)
	if err != nil {
		log.Warningf("binding device failed:domain[%s], device[%s:%s], master[%d]", this.Domain(), subDomain, deviceId, masterDid)
		return err
	}
	return nil
}

// change the slave or master device, the old must valid
func (this *BindingManager) ChangeBinding(did int64, subDomain, deviceId string) error {
	// step 1. check the old did must be register succ
	var device DeviceInfo
	deviceManager := NewDeviceManager(this.store)
	err := deviceManager.getDeviceInfo(did, &device)
	if err != nil {
		log.Warningf("get master device failed:domain[%s], did[%d], err[%v]", this.Domain(), did, err)
		return err
	}
	// step 2 check the old did must be binging ok
	var exist bool
	err = this.isBindingExist(did, &exist)
	if err != nil {
		log.Warningf("check binding exist failed:domain[%s], did[%d], err[%v]", this.Domain(), did, err)
		return err
	} else if !exist {
		log.Warningf("the old did not binding:domain[%s], did[%d]", this.Domain(), did)
		return common.ErrNotYetBinded
	}
	// step 3. check the new device basic info is valid according the old device type
	err = this.checkDeviceInfo(subDomain, deviceId, device.IsMasterDevice())
	if err != nil {
		log.Warningf("check device info failed:domain[%s], device[%s:%s], master[%d], err[%v]",
			this.Domain(), subDomain, deviceId, did, err)
		return err
	}
	// step 4. check the new device is not binding by others
	var binding BindingInfo
	err = this.getBindingInfo(subDomain, deviceId, &binding)
	if err != nil {
		if err != common.ErrEntryNotExist {
			log.Warningf("get device binding info failed:domain[%s], device[%s:%s], master[%d], err[%v]",
				this.Domain(), subDomain, deviceId, did, err)
			return err
		}
	} else {
		log.Warningf("check the device is already binded:domain[%s], device[%s:%s], oldDid[%d], newDid[%d]",
			this.Domain(), subDomain, deviceId, binding.did, did)
		return common.ErrAlreadyBinded
	}
	// step 5. change the mapping relation
	err = this.changeDeviceBinding(did, subDomain, deviceId)
	if err != nil {
		log.Warningf("do change the device mapping binding info failed:domain[%s], device[%s:%s], err[%v]",
			this.Domain(), subDomain, deviceId, err)
		return err
	}
	return nil
}

//////////////////////////////////////////////////////////////////////////////
/// private interface related to database
//////////////////////////////////////////////////////////////////////////////
// transaction rollback according to the error status
func rollback(err *error, tx *sql.Tx) {
	if *err != nil {
		log.Infof("error occured rollback:err[%v]", *err)
		newErr := tx.Rollback()
		if newErr != nil {
			log.Errorf("rollback failed:err[%v]", newErr)
		}
	}
}

// check basic device info from device warehouse
// binding one device to one home, if masterDid < 0 it's master device, otherwise it's slave device
func (this *BindingManager) checkDeviceInfo(subDomain, deviceId string, isMaster bool) error {
	warehouse := NewDeviceWarehouse(this.store)
	if warehouse == nil {
		log.Errorf("new device warehouse failed:domain[%s], device[%s:%s]", this.Domain(), subDomain, deviceId)
		return common.ErrUnknown
	}
	dev, err := warehouse.Get(subDomain, deviceId)
	if err != nil {
		log.Warningf("get device failed:domain[%s], device[%s:%s], err[%v]",
			this.Domain(), subDomain, deviceId, err)
		return err
	} else if dev == nil {
		log.Warning("check the device not exist:domain[%s], device[%s:%s]", this.Domain(), subDomain, deviceId)
		return common.ErrInvalidDevice
	} else if !dev.Validate() {
		log.Errorf("device validate failed:domain[%s], device[%s:%s]", this.Domain(), subDomain, deviceId)
		return common.ErrInvalidDevice
	} else if isMaster && dev.deviceType != MASTER {
		log.Warningf("check the master device type failed:domain[%s], device[%s:%s]", this.Domain(), subDomain, deviceId)
		return common.ErrInvalidDevice
	} else if (!isMaster) && (dev.deviceType == MASTER) {
		log.Warningf("check the slave device type failed:domain[%s], device[%s:%s]", this.Domain(), subDomain, deviceId)
		return common.ErrInvalidDevice
	}
	return nil
}

// check the master is mapping ok
// if masterDid <= 0 it's master device, otherwise it's slave device
func (this *BindingManager) checkMasterDevice(subDomain, deviceId string, masterDid int64) error {
	if masterDid > 0 {
		SQL := fmt.Sprintf("SELECT sub_domain, device_id FROM %s_device_mapping WHERE did = ?", this.Domain())
		stmt, err := this.store.db.Prepare(SQL)
		if err != nil {
			log.Errorf("prepare select mapping failed:domain[%s], device[%d], err[%v]", this.Domain(), masterDid, err)
			return err
		}
		var masterDomain, masterDeviceId string
		err = stmt.QueryRow(masterDid).Scan(&masterDomain, &masterDeviceId)
		if err != nil {
			log.Warningf("query and parse device failed:domain[%s], device[%d], err[%v]", this.Domain(), masterDid, err)
			return err
		} else if masterDomain != subDomain || masterDeviceId == deviceId {
			log.Warningf("check the device info with master failed:domain[%s], slave[%s:%s], master[%s:%s], masterDid[%d]",
				this.Domain(), subDomain, deviceId, masterDomain, masterDeviceId, masterDid)
			return common.ErrInvalidDevice
		}
	}
	return nil
}

func (this *BindingManager) getBindingInfo(subDomain, deviceId string, bind *BindingInfo) error {
	SQL := fmt.Sprintf("SELECT did, bind_token, expire_time FROM %s_device_mapping WHERE sub_domain = ? AND device_id = ?", this.Domain())
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:domain[%s], device[%s:%s], err[%v]",
			this.Domain(), subDomain, deviceId, err)
		return err
	}
	defer stmt.Close()
	err = stmt.QueryRow(subDomain, deviceId).Scan(&bind.did, &bind.grantToken, &bind.grantTime)
	if err != nil {
		if err == sql.ErrNoRows {
			return common.ErrEntryNotExist
		} else {
			log.Warningf("query binding info failed:domain[%s], device[%s:%s], err[%v]",
				this.Domain(), subDomain, deviceId, err)
			return err
		}
	}
	return nil
}

func (this *BindingManager) isBindingExist(did int64, exist *bool) error {
	SQL := fmt.Sprintf("SELECT did FROM %s_device_mapping WHERE did = ?", this.Domain())
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:domain[%s], did[%d], err[%v]", this.Domain(), did, err)
		return err
	}
	defer stmt.Close()
	var value int64
	err = stmt.QueryRow(did).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			*exist = false
		} else {
			log.Warningf("get binding info failed:domain[%s], did[%d], err[%v]", this.Domain(), did, err)
			return err
		}
	}
	*exist = true
	return nil
}

func (this *BindingManager) changeDeviceBinding(did int64, subDomain, deviceId string) error {
	SQL1 := fmt.Sprintf("UPDATE %s_device_mapping SET sub_domain = ?, device_id = ?, bind_token = NULL WHERE did = ?", this.Domain())
	stmt, err := this.store.db.Prepare(SQL1)
	if err != nil {
		log.Errorf("prepare update mapping failed:domain[%s], device[%s:%s], err[%v]", this.Domain(), subDomain, deviceId, err)
		return err
	}
	defer stmt.Close()
	result, err := stmt.Exec(subDomain, deviceId, did)
	if err != nil {
		log.Errorf("execute update failed:domain[%s], device[%s:%s], err[%v]", this.Domain(), subDomain, deviceId, err)
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Warningf("get affected rows failed:err[%v]", err)
		return err
	}
	if affect != 1 {
		log.Errorf("check affected rows failed:domain[%s], device[%s:%s], err[%v]", this.Domain(), subDomain, deviceId, err)
		return common.ErrEntryNotExist
	}
	return nil
}

func getMasterDid(master, did int64) int64 {
	if master > 0 {
		return master
	}
	return did
}

// binding device main routine
func (this *BindingManager) bindingDevice(subDomain, deviceId, deviceName string, hid, masterDid int64) (err error) {
	// step 1. check the mapping exist or not
	var did int64
	var stmt1 *sql.Stmt
	var binding BindingInfo
	err = this.getBindingInfo(subDomain, deviceId, &binding)
	if err == nil {
		did = binding.did
		log.Infof("deivce already activated:domain[%s], device[%s:%s], did[%d]", this.Domain(), subDomain, deviceId, did)
	} else if err == common.ErrEntryNotExist {
		log.Infof("new deivce activated:domain[%s], device[%s:%s]", this.Domain(), subDomain, deviceId)
		SQL1 := fmt.Sprintf("INSERT INTO %s_device_mapping(sub_domain, device_id) VALUES(?,?)", this.Domain())
		stmt1, err = this.store.db.Prepare(SQL1)
		if err != nil {
			log.Errorf("prepare insert mapping failed:domain[%s], device[%s:%s], err[%v]",
				this.Domain(), subDomain, deviceId, err)
			return err
		}
		defer stmt1.Close()
	} else {
		log.Warningf("get binding info failed:domain[%s], device[%s:%s], err[%v]",
			this.Domain(), subDomain, deviceId, err)
		return err
	}
	// step 2. replace into the device info if exist replace, if not insert
	SQL2 := fmt.Sprintf("REPLACE INTO %s_device_info(did, hid, name, status, master_did) VALUES(?, ?, ?, ?, ?)", this.Domain())
	stmt2, err := this.store.db.Prepare(SQL2)
	if err != nil {
		log.Errorf("prepare replace device info failed:domain[%s], device[%s:%s], err[%v]",
			this.Domain(), subDomain, deviceId, err)
		return err
	}
	defer stmt2.Close()

	// begin the transaction update mapping and device info table
	tx, err := this.store.db.Begin()
	if err != nil {
		log.Errorf("begin transaction failed:domain[%s], device[%s:%s], err[%v]", this.Domain(), subDomain, deviceId, err)
		return err
	}
	return func() error {
		defer rollback(&err, tx)
		var result sql.Result
		if did <= 0 {
			result, err = tx.Stmt(stmt1).Exec(subDomain, deviceId)
			if err != nil {
				log.Errorf("insert mapping failed:domain[%s], device[%s:%s], err[%v]", this.Domain(), subDomain, deviceId, err)
				return err
			}
			did, err = result.LastInsertId()
			if err != nil {
				log.Errorf("get insert id failed:domain[%s], device[%s:%s], err[%v]", this.Domain(), subDomain, deviceId, err)
				return err
			}
		}
		_, err = tx.Stmt(stmt2).Exec(did, hid, deviceName, ACTIVE, getMasterDid(masterDid, did))
		if err != nil {
			log.Errorf("replace device info failed:domain[%s], device[%s:%s], hid[%d], masterDid[%d], err[%v]",
				this.Domain(), subDomain, deviceId, hid, masterDid, err)
			return err
		}
		newErr := tx.Commit()
		if newErr != nil {
			log.Errorf("commit failed:domain[%s], device[%s:%s], hid[%d], masterDid[%d], err[%v]",
				this.Domain(), subDomain, deviceId, hid, masterDid, newErr)
			return newErr
		}
		log.Infof("binding the device succ:domain[%s], device[%s:%s], did[%d], name[%s], hid[%d], masterDid[%d]",
			this.Domain(), subDomain, deviceId, did, deviceName, hid, masterDid)
		return nil
	}()
}
