package device

import (
	"testing"
)

func TestValidate(t *testing.T) {
	var device BasicInfo
	device.deviceType = MASTER
	device.publicKey.Valid = false
	device.publicKey.String = ""
	device.status = ACTIVE

	// no subDomain no device id
	if device.Validate() {
		t.Error("domain or id failed")
	}
	device.subDomain = "domain"
	if device.Validate() {
		t.Error("domain or id failed")
	}
	device.deviceId = "device_id"

	// right normal
	device.deviceType = NORMAL

	device.publicKey.Valid = false
	device.publicKey.String = ""
	if !device.Validate() {
		t.Error("normal device validate failed")
	}
	// error normal
	device.publicKey.Valid = true
	device.publicKey.String = ""
	if device.Validate() {
		t.Error("normal device should no public key")
	}
	device.publicKey.Valid = true
	device.publicKey.String = "key"
	if device.Validate() {
		t.Error("normal device should no public key")
	}
	device.publicKey.Valid = false
	device.publicKey.String = "key"
	if device.Validate() {
		t.Error("normal device should no public key")
	}

	// error master
	device.deviceType = MASTER

	device.publicKey.Valid = false
	device.publicKey.String = "key"
	if device.Validate() {
		t.Error("master device should no public key")
	}

	device.publicKey.Valid = true
	device.publicKey.String = ""
	if device.Validate() {
		t.Error("master device should no public key")
	}

	device.publicKey.Valid = false
	device.publicKey.String = ""
	if device.Validate() {
		t.Error("master device validate failed")
	}

	// right master
	device.publicKey.Valid = true
	device.publicKey.String = "key"
	if !device.Validate() {
		t.Error("master device validate failed")
	}

	device.deviceId = ""
	if device.Validate() {
		t.Error("master device validate failed")
	}
}
