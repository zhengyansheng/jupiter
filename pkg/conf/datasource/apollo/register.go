package apollo

import (
	"net/url"

	"github.com/philchia/agollo/v4"
	"github.com/zhengyansheng/jupiter/pkg/conf"
	"github.com/zhengyansheng/jupiter/pkg/flag"
	"github.com/zhengyansheng/jupiter/pkg/xlog"
)

// DataSourceApollo defines apollo scheme
const DataSourceApollo = "apollo"

func init() {
	conf.Register(DataSourceApollo, func() conf.DataSource {
		var (
			configAddr = flag.String("config")
		)
		if configAddr == "" {
			xlog.Jupiter().Panic("new apollo dataSource, configAddr is empty")
			return nil
		}
		// configAddr is a string in this format:
		// apollo://ip:port?appId=XXX&cluster=XXX&namespaceName=XXX&key=XXX&accesskeySecret=XXX&insecureSkipVerify=XXX&cacheDir=XXX
		urlObj, err := url.Parse(configAddr)
		if err != nil {
			xlog.Jupiter().Panic("parse configAddr error", xlog.FieldErr(err))
			return nil
		}
		apolloConf := agollo.Conf{
			AppID:              urlObj.Query().Get("appId"),
			Cluster:            urlObj.Query().Get("cluster"),
			NameSpaceNames:     []string{urlObj.Query().Get("namespaceName")},
			MetaAddr:           urlObj.Host,
			InsecureSkipVerify: true,
			AccesskeySecret:    urlObj.Query().Get("accesskeySecret"),
			CacheDir:           ".",
		}
		if urlObj.Query().Get("insecureSkipVerify") == "false" {
			apolloConf.InsecureSkipVerify = false
		}
		if urlObj.Query().Get("cacheDir") != "" {
			apolloConf.CacheDir = urlObj.Query().Get("cacheDir")
		}
		return NewDataSource(&apolloConf, urlObj.Query().Get("namespaceName"), urlObj.Query().Get("key"))
	})
}
