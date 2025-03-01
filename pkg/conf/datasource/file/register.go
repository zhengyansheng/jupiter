package file

import (
	"github.com/zhengyansheng/jupiter/pkg/conf"
	"github.com/zhengyansheng/jupiter/pkg/flag"
	"github.com/zhengyansheng/jupiter/pkg/xlog"
)

// DataSourceFile defines file scheme
const DataSourceFile = "file"

func init() {
	conf.Register(DataSourceFile, func() conf.DataSource {
		var (
			watchConfig = flag.Bool("watch")
			configAddr  = flag.String("config")
		)
		if configAddr == "" {
			xlog.Jupiter().Panic("new file dataSource, configAddr is empty")
			return nil
		}
		return NewDataSource(configAddr, watchConfig)
	})
}
