package device

import (
	"database/sql"
	"fmt"
	log "zc-common-go/glog"
)

type WarehouseProxy struct {
	cacheOn bool
	cache   *WarehouseCache
	store   *DeviceStorage
}

const MAX_DEVICE_COUNT int64 = 100000

func newWarehouseProxy(store *DeviceStorage) *WarehouseProxy {
	cache := newWarehouseCache(MAX_DEVICE_COUNT)
	if cache == nil {
		log.Error("new device warehouse Cache failed")
		return nil
	}
	return &WarehouseProxy{cacheOn: true, cache: cache, store: store}
}

// switch the cache on/off
func (this *WarehouseProxy) SwitchCache(on bool) {
	this.cacheOn = on
	this.cache.Clear()
}

// clear the cache
func (this *WarehouseProxy) Clear() {
	this.cache.Clear()
}

// if not find in database return nil + nil
func (this *WarehouseProxy) GetDeviceInfo(domain, subDomain, deviceId string) (*BasicInfo, error) {
	if this.cacheOn {
		basic, find := this.cache.Get(domain, subDomain, deviceId)
		if find {
			log.Infof("get device basic info from cache:domain[%s], device[%s:%s]", domain, subDomain, deviceId)
			return basic, nil
		}
	}
	SQL := fmt.Sprintf("SELECT device_type, public_key, status FROM %s_device_warehouse WHERE sub_domain = ? AND device_id = ?", domain)
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:err[%v]", err)
		return nil, err
	}
	defer stmt.Close()
	basic := NewBasicInfo()
	err = stmt.QueryRow(subDomain, deviceId).Scan(&basic.deviceType, &basic.publicKey, &basic.status)
	if err != nil {
		if err == sql.ErrNoRows {
			log.Warningf("no find the device:domain[%s], device[%s:%s]", domain, subDomain, deviceId)
			return nil, nil
		}
		log.Errorf("query failed:domain[%s], device[%s:%s]", domain, subDomain, deviceId)
		return nil, err
	}
	basic.subDomain = subDomain
	basic.deviceId = deviceId
	if this.cacheOn {
		this.cache.Set(domain, basic)
	}
	return basic, nil
}

func (this *WarehouseProxy) InsertDeviceInfo(domain, subDomain, deviceId, publicKey string, master bool) error {
	var SQL string
	if master {
		SQL = fmt.Sprintf("INSERT INTO %s_device_warehouse(sub_domain, device_id, device_type, public_key, status) VALUES(?,?,?,?,?)", domain)
	} else {
		SQL = fmt.Sprintf("INSERT INTO %s_device_warehouse(sub_domain, device_id, device_type, status) VALUES(?,?,?,?)", domain)
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
		log.Warningf("execute insert device[%s:%s] failed:domain[%s], err[%v]", subDomain, deviceId, domain, err)
		return err
	}
	if this.cacheOn {
		this.cache.Delete(domain, subDomain, deviceId)
	}
	return nil
}

func (this *WarehouseProxy) DeleteDeviceInfo(domain, subDomain, deviceId string) error {
	if this.cacheOn {
		this.cache.Delete(domain, subDomain, deviceId)
	}
	SQL := fmt.Sprintf("DELETE FROM %s_device_warehouse WHERE sub_domain = ? AND device_id = ?", domain)
	stmt, err := this.store.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:err[%v]", err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec(subDomain, deviceId)
	if err != nil {
		log.Errorf("delete the device failed:domain[%s], subDomain[%s], deviceId[%s], err[%s]",
			domain, subDomain, deviceId, err)
		return err
	}
	return nil
}
