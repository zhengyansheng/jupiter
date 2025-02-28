package xfreecache

import (
	"github.com/coocood/freecache"
	prome "github.com/douyu/jupiter/pkg/core/metric"
	"github.com/douyu/jupiter/pkg/xlog"
	"go.uber.org/zap"
)

type localStorage[K comparable, V any] struct {
	config *Config
}

func (l *localStorage[K, V]) setCacheData(key string, data []byte) (err error) {
	err = l.config.Cache.Set([]byte(key), data, int(l.config.Expire.Seconds()))
	// metric report
	if !l.config.DisableMetric {
		if err != nil {
			prome.CacheHandleCounter.WithLabelValues(prome.TypeLocalCache, l.config.Name, "SetFail", err.Error()).Inc()
		} else {
			prome.CacheHandleCounter.WithLabelValues(prome.TypeLocalCache, l.config.Name, "SetSuccess", "").Inc()
		}
	}
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
	// metric report
	if !l.config.DisableMetric {
		if err != nil {
			prome.CacheHandleCounter.WithLabelValues(prome.TypeLocalCache, l.config.Name, "MissCount", err.Error()).Inc()
		} else {
			prome.CacheHandleCounter.WithLabelValues(prome.TypeLocalCache, l.config.Name, "HitCount", "").Inc()
		}
	}
	if err != nil && err != freecache.ErrNotFound {
		xlog.Jupiter().Error("cache GetCacheData", zap.String("key", key), zap.Error(err))
		return
	}
	return
}
