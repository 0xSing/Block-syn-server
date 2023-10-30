package etc

import (
	"io/ioutil"
	zlog "walletSynV2/utils/zlog_sing"

	"go.uber.org/zap"
	"gopkg.in/yaml.v2"
)

var Conf config

func InitConfig(file string) error {
	zlog.Zlog.Info("loading conf file", zap.String("file", file))
	bs, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}

	return yaml.Unmarshal(bs, &Conf)
}
