package xfreecache

import (
	"github.com/coocood/freecache"
	"github.com/zhengyansheng/jupiter/pkg/xlog"
	"go.uber.org/zap"
)

type localStorage[K comparable, V any] struct {
	config *Config
}

func (l *localStorage[K, V]) setCacheData(key string, data []byte) (err error) {
	err = l.config.Cache.Set([]byte(key), data, int(l.config.Expire.Seconds()))
	if err != nil {
		xlog.Jupiter().Error("cache SetCacheData", zap.String("data", string(data)), zap.Error(err))
		if err == freecache.ErrLargeEntry || err == freecache.ErrLargeKey {
			err = nil
		}
		return
	}
	return
}

func (l *localStorage[K, V]) getCacheData(key string) (data []byte, err error) {
	data, err = l.config.Cache.Get([]byte(key))
	if err != nil && err != freecache.ErrNotFound {
		xlog.Jupiter().Error("cache GetCacheData", zap.String("key", key), zap.Error(err))
		return
	}
	return
}
