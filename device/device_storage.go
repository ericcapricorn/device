package device

import (
	"database/sql"
	"fmt"
	"zc-common-go/common"
	log "zc-common-go/glog"
	_ "zc-common-go/mysql"
)

type DeviceStorage struct {
	db       *sql.DB
	database string
}

func NewDeviceStorage(host, user, password, database string) *DeviceStorage {
	dns := fmt.Sprintf("%s:%s@tcp(%s)/%s", user, password, host, database)
	driver, err := sql.Open("mysql", dns)
	if err != nil {
		log.Errorf("open MySQL Driver failed:dns[%s], err[%v]", dns, err)
		return nil
	}
	return &DeviceStorage{db: driver, database: database}
}

func (this *DeviceStorage) Destory() {
	if this.db != nil {
		this.db.Close()
	}
}

// just for unit test warning
func (this *DeviceStorage) Clean(domain, table string) error {
	if this.database != "device_test" {
		log.Error("clean interface only can be used by test")
		return common.ErrNotAllowed
	}
	SQL := fmt.Sprintf("DELETE FROM %s_%s", domain, table)
	stmt, err := this.db.Prepare(SQL)
	if err != nil {
		log.Errorf("prepare query failed:domain[%s], table[%s], err[%v]", domain, table, err)
		return err
	}
	defer stmt.Close()
	_, err = stmt.Exec()
	if err != nil {
		log.Errorf("delete failed:domain[%s], table[%s], err[%v]", domain, table, err)
		return err
	}
	return nil
}
