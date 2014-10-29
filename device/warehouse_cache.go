package device

import (
	"sync"
	"zc-common-go/common"
	log "zc-common-go/glog"
)

// device key
type DeviceKey struct {
	domain    string
	subDomain string
	deviceId  string
}

type WarehouseCache struct {
	lock  sync.RWMutex
	cache *common.LRUCache
}

func newWarehouseCache(count int64) *WarehouseCache {
	cache := common.NewLRUCache(count)
	if cache != nil {
		return &WarehouseCache{cache: cache}
	}
	return nil
}

func (this *WarehouseCache) Clear() {
	this.lock.Lock()
	defer this.lock.Unlock()
	log.Infof("clear the cache:len[%d], hit_ratio[%f]", this.cache.Len(), this.cache.HitRatio())
}

func (this *WarehouseCache) Get(domain, subDomain, deviceId string) (*BasicInfo, bool) {
	this.lock.RLock()
	defer this.lock.RUnlock()
	basic, find := this.cache.Get(DeviceKey{domain: domain, subDomain: subDomain, deviceId: deviceId})
	if find {
		return basic.(*BasicInfo), true
	}
	return nil, false
}

func (this *WarehouseCache) Set(domain string, basic *BasicInfo) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.cache.Set(DeviceKey{domain: domain, subDomain: basic.subDomain, deviceId: basic.deviceId}, basic)
}

func (this *WarehouseCache) Delete(domain, subDomain, deviceId string) {
	this.lock.Lock()
	defer this.lock.Unlock()
	this.cache.Delete(DeviceKey{domain: domain, subDomain: subDomain, deviceId: deviceId})
}
