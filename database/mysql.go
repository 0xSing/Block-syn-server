package database

import (
	"fmt"
	"time"
	"walletSynV2/etc"

	"github.com/astaxie/beego/orm"
	_ "github.com/go-sql-driver/mysql"
)

func InitMysql(options []*etc.Mysql, debug bool) error {
	orm.Debug = debug

	err := orm.RegisterDriver("mysql", orm.DRMySQL)
	if err != nil {
		return err
	}

	maxIdle := 30
	maxConn := 30
	auth := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&loc=Local",
		options[0].User, options[0].Password, options[0].IP, options[0].Port, options[0].Db)

	_, err = orm.GetDB("default")
	if err != nil {
		err = orm.RegisterDataBase("default", "mysql", auth, maxIdle, maxConn)
		if err != nil {
			return err
		}
	}

	database, err := orm.GetDB("default")
	if err != nil {
		return err
	}
	database.SetConnMaxLifetime(30 * time.Second)

	return orm.RunSyncdb("default", false, true)
}
