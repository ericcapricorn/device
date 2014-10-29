package device

import (
	"database/sql"
	"zc-common-go/mysql"
)

const (
	NORMAL = 0
	MASTER = 1
)

// device key(subdomain + deviceid)
type BasicInfo struct {
	subDomain  string
	deviceId   string
	publicKey  sql.NullString
	deviceType int
	status     int8
}

// invalid basic info
func NewBasicInfo() *BasicInfo {
	return &BasicInfo{status: INVALID}
}

func (this *BasicInfo) PublicKey() string {
	return this.publicKey.String
}

func (this *BasicInfo) IsMaster() bool {
	return this.deviceType == MASTER
}

func (this *BasicInfo) Validate() bool {
	if len(this.deviceId) == 0 || len(this.subDomain) == 0 {
		return false
	}
	if this.deviceType == MASTER {
		if this.publicKey.Valid && len(this.publicKey.String) > 0 {
			return true
		} else {
			return false
		}
	} else if this.publicKey.Valid || len(this.publicKey.String) > 0 {
		return false
	}
	return true
}

// device binding info
type BindingInfo struct {
	did        int64
	subDomain  string
	deviceId   string
	grantToken sql.NullString
	grantTime  mysql.NullTime
}

// invalid default binding info
func NewBindingInfo() *BindingInfo {
	return &BindingInfo{did: -1}
}

// if masterDid = did, it is master device, else it's normal device
type DeviceInfo struct {
	did        int64
	hid        int64
	deviceName string
	status     int8
	masterDid  int64
}

// default invalid device info
func NewDeviceInfo() *DeviceInfo {
	return &DeviceInfo{did: -1}
}

// device inner id
func (this *DeviceInfo) GetDid() int64 {
	return this.did
}

// device home id
func (this *DeviceInfo) GetHid() int64 {
	return this.hid
}

// device name for device user
func (this *DeviceInfo) GetDeviceName() string {
	return this.deviceName
}

// device status set by user
func (this *DeviceInfo) GetStatus() int8 {
	return this.status
}

// master device did = master did
func (this *DeviceInfo) IsMasterDevice() bool {
	return this.masterDid == this.did
}

func (this *DeviceInfo) GetMasterDid() int64 {
	return this.masterDid
}

// whole info
type Device struct {
	_ BasicInfo
	_ DeviceInfo
	_ BindingInfo
}
