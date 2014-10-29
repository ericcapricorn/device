package device

import (
	"database/sql"
	"fmt"
	"zc-common-go/common"
	log "zc-common-go/glog"
)

type BindingProxy struct {
	cacheOn bool
	cache   *BindingCache
	store   *DeviceStorage
}

const MAX_BINDING_COUNT int64 = 10000

func newBindingProxy(store *DeviceStorage) *BindingProxy {
	cache := NewBindingCache(MAX_BINDING_COUNT)
	if cache == nil {
		log.Error("new binding cache failed")
		return nil
	}
	return &BindingProxy{cacheOn: true, cache: cache, store: store}
}

// get by device global key, if not exist return nil + nil
func (this *BindingProxy) GetBindingInfo(domain, subDomain, deviceId string) (*BindingInfo, error) {
	SQL := fmt.Sprintf("SELECT did, bind_token, expire_time FROM %s_device_mapping WHERE sub_domain = ? AND device_id = ?", domain)
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:domain[%s], device[%s:%s], err[%v]",
			domain, subDomain, deviceId, err)
		return nil, err
	}
	defer stmt.Close()
	bind := NewBindingInfo()
	err = stmt.QueryRow(subDomain, deviceId).Scan(&bind.did, &bind.grantToken, &bind.grantTime)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Warningf("query binding info failed:domain[%s], device[%s:%s], err[%v]",
				domain, subDomain, deviceId, err)
		} else {
			err = common.ErrEntryNotExist
		}
		return nil, err
	}
	bind.subDomain = subDomain
	bind.deviceId = deviceId
	if this.cacheOn {
		this.cache.Set(domain, bind)
	}
	return bind, nil
}

func (this *BindingProxy) GetBindingByDid(domain string, did int64) (*BindingInfo, error) {
	if this.cacheOn {
		bind, find := this.cache.Get(domain, did)
		if find {
			return bind, nil
		}
	}
	SQL := fmt.Sprintf("SELECT sub_domain, device_id, bind_token, expire_time FROM %s_device_mapping WHERE did = ?", domain)
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Warningf("prepare query failed:domain[%s], did[%d], err[%v]", domain, did, err)
		return nil, err
	}
	defer stmt.Close()
	bind := NewBindingInfo()
	err = stmt.QueryRow(did).Scan(&bind.subDomain, &bind.deviceId, &bind.grantToken, &bind.grantTime)
	if err != nil {
		if err != sql.ErrNoRows {
			log.Warningf("query and parse binding info failed:domain[%s], did[%d], err[%v]", domain, did, err)
		} else {
			err = common.ErrEntryNotExist
		}
		return nil, err
	}
	bind.did = did
	if this.cacheOn {
		this.cache.Set(domain, bind)
	}
	return bind, nil
}

func (this *BindingProxy) IsBindingExist(domain string, did int64, exist *bool) error {
	if this.cacheOn {
		_, find := this.cache.Get(domain, did)
		if find {
			*exist = true
			return nil
		}
	}
	SQL := fmt.Sprintf("SELECT did FROM %s_device_mapping WHERE did = ?", domain)
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:domain[%s], did[%d], err[%v]", domain, did, err)
		return err
	}
	defer stmt.Close()
	var value int64
	err = stmt.QueryRow(did).Scan(&value)
	if err != nil {
		if err == sql.ErrNoRows {
			*exist = false
		} else {
			log.Warningf("get binding info failed:domain[%s], did[%d], err[%v]", domain, did, err)
			return err
		}
	}
	*exist = true
	return nil
}

func (this *BindingProxy) ChangeDeviceBinding(did int64, domain, subDomain, deviceId string) error {
	if this.cacheOn {
		this.cache.Delete(domain, did)
	}
	SQL1 := fmt.Sprintf("UPDATE %s_device_mapping SET sub_domain = ?, device_id = ?, bind_token = NULL WHERE did = ?", domain)
	stmt, err := this.store.db.Prepare(SQL1)
	if err != nil {
		log.Errorf("prepare update mapping failed:domain[%s], device[%s:%s], err[%v]", domain, subDomain, deviceId, err)
		return err
	}
	defer stmt.Close()
	result, err := stmt.Exec(subDomain, deviceId, did)
	if err != nil {
		log.Errorf("execute update failed:domain[%s], device[%s:%s], err[%v]", domain, subDomain, deviceId, err)
		return err
	}
	affect, err := result.RowsAffected()
	if err != nil {
		log.Warningf("get affected rows failed:err[%v]", err)
		return err
	}
	if affect != 1 {
		log.Errorf("check affected rows failed:domain[%s], device[%s:%s], err[%v]", domain, subDomain, deviceId, err)
		return common.ErrEntryNotExist
	}
	return nil
}

// binding device main routine
func (this *BindingProxy) BindingDevice(domain, subDomain, deviceId, deviceName string, hid, masterDid int64) (err error) {
	// step 1. check the mapping exist or not
	var did int64
	var stmt1 *sql.Stmt
	binding, err := this.GetBindingInfo(domain, subDomain, deviceId)
	if err == nil {
		did = binding.did
		log.Infof("deivce already activated:domain[%s], device[%s:%s], did[%d]", domain, subDomain, deviceId, did)
	} else if err == common.ErrEntryNotExist {
		log.Infof("new deivce activated:domain[%s], device[%s:%s]", domain, subDomain, deviceId)
		SQL1 := fmt.Sprintf("INSERT INTO %s_device_mapping(sub_domain, device_id) VALUES(?,?)", domain)
		stmt1, err = this.store.db.Prepare(SQL1)
		if err != nil {
			log.Errorf("prepare insert mapping failed:domain[%s], device[%s:%s], err[%v]",
				domain, subDomain, deviceId, err)
			return err
		}
		defer stmt1.Close()
	} else {
		log.Warningf("get binding info failed:domain[%s], device[%s:%s], err[%v]",
			domain, subDomain, deviceId, err)
		return err
	}
	// WARNING: TODO device info cache should be updated(deleted) it at first
	// step 2. replace into the device info if exist replace, if not insert
	// the home->devices list cache should be updated....
	SQL2 := fmt.Sprintf("REPLACE INTO %s_device_info(did, hid, name, status, master_did) VALUES(?, ?, ?, ?, ?)", domain)
	stmt2, err := this.store.db.Prepare(SQL2)
	if err != nil {
		log.Errorf("prepare replace device info failed:domain[%s], device[%s:%s], err[%v]",
			domain, subDomain, deviceId, err)
		return err
	}
	defer stmt2.Close()

	// begin the transaction update mapping and device info table
	tx, err := this.store.db.Begin()
	if err != nil {
		log.Errorf("begin transaction failed:domain[%s], device[%s:%s], err[%v]", domain, subDomain, deviceId, err)
		return err
	}
	return func() error {
		defer rollback(&err, tx)
		var result sql.Result
		if did <= 0 {
			result, err = tx.Stmt(stmt1).Exec(subDomain, deviceId)
			if err != nil {
				log.Errorf("insert mapping failed:domain[%s], device[%s:%s], err[%v]", domain, subDomain, deviceId, err)
				return err
			}
			did, err = result.LastInsertId()
			if err != nil {
				log.Errorf("get insert id failed:domain[%s], device[%s:%s], err[%v]", domain, subDomain, deviceId, err)
				return err
			}
		}
		_, err = tx.Stmt(stmt2).Exec(did, hid, deviceName, ACTIVE, getMasterDid(masterDid, did))
		if err != nil {
			log.Errorf("replace device info failed:domain[%s], device[%s:%s], hid[%d], masterDid[%d], err[%v]",
				domain, subDomain, deviceId, hid, masterDid, err)
			return err
		}
		newErr := tx.Commit()
		if newErr != nil {
			log.Errorf("commit failed:domain[%s], device[%s:%s], hid[%d], masterDid[%d], err[%v]",
				domain, subDomain, deviceId, hid, masterDid, newErr)
			return newErr
		}
		log.Infof("binding the device succ:domain[%s], device[%s:%s], did[%d], name[%s], hid[%d], masterDid[%d]",
			domain, subDomain, deviceId, did, deviceName, hid, masterDid)
		return nil
	}()
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

func getMasterDid(master, did int64) int64 {
	if master > 0 {
		return master
	}
	return did
}
