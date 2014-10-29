package device

import (
	"sync"
	"zc-common-go/common"
)

type BindingCacheKey struct {
	domain string
	did    int64
}

type BindingCache struct {
	lock  sync.RWMutex
	cache *common.LRUCache
}

func NewBindingCache(count int64) *BindingCache {
	cache := common.NewLRUCache(count)
	if cache != nil {
		return &BindingCache{cache: cache}
	}
	return nil
}

func (this *BindingCache) Get(domain string, did int64) (*BindingInfo, bool) {
	this.lock.RLock()
	this.lock.RUnlock()
	bind, find := this.cache.Get(BindingCacheKey{domain: domain, did: did})
	if find {
		return bind.(*BindingInfo), true
	}
	return nil, false
}

func (this *BindingCache) Set(domain string, bind *BindingInfo) {
	this.lock.Lock()
	this.lock.Unlock()
	this.cache.Set(BindingCacheKey{domain: domain, did: bind.did}, bind)
}

func (this *BindingCache) Delete(domain string, did int64) {
	this.lock.Lock()
	this.lock.Unlock()
	this.cache.Delete(BindingCacheKey{domain: domain, did: did})
}
