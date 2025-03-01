package xfreecache

import (
	"github.com/coocood/freecache"
	"github.com/zhengyansheng/jupiter/pkg/xlog"
	"go.uber.org/zap"
)

type LocalCache struct {
	cache
}

type localStorage struct {
	cache *freecache.Cache
	req   Config
}

func (l *localStorage) SetCacheData(key string, data []byte) (err error) {
	err = l.cache.Set([]byte(key), data, int(l.req.Expire.Seconds()))
	if err != nil {
		xlog.Jupiter().Error("cache SetCacheData", zap.String("data", string(data)), zap.Error(err))
		if err == freecache.ErrLargeEntry || err == freecache.ErrLargeKey {
			err = nil
		}
		return
	}
	return
}

func (l *localStorage) GetCacheData(key string) (data []byte, err error) {
	data, err = l.cache.Get([]byte(key))

	if err == freecache.ErrNotFound || data == nil {
		err = nil
		return
	}
	if err != nil {
		xlog.Jupiter().Error("cache GetCacheData", zap.String("key", key), zap.Error(err))
		return
	}
	return
}
