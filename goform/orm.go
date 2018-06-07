/*******************************************************************************
 * Copyright (c) 2018  charles
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NON INFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 * -------------------------------------------------------------------------
 * created at 2018-06-06 12:08:33
 ******************************************************************************/

package goform

import (
	"fmt"
	"time"

	"gitee.com/goframe/gof/gofconf"
	"gitee.com/goframe/gof/gofutils"
	"gitee.com/goframe/gof/gofutils/errors"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/mysql"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/spf13/viper"
)

var (
	Slave       *gorm.DB
	Master      *gorm.DB
	configName  = "database.yaml"
	vp          = viper.New()
	defaultConf = Conf{
		ShowSQL: true,
		Master: Database{
			Type:     "postgres",
			User:     "puser",
			Password: "123",
			DB:       "db1",
			Address:  "127.0.0.1",
			Port:     5432,
		},
		Slave: Database{Address: "127.0.0.1", Port: 5432},
	}
	dbs = make([]*gorm.DB, 0)
)

type (
	//Conf  All Database configuration Settings
	Conf struct {
		ShowSQL bool
		Master  Database
		Slave   Database `yaml:",flow"`
	}
	//Database  configuration Settings
	Database struct {
		Type     string `yaml:",omitempty"`
		Address  string
		Port     int
		User     string `yaml:",omitempty"`
		Password string `yaml:",omitempty"`
		DB       string `yaml:",omitempty"`
	}
)

func Initialize() {
	if err := settingDatabase(); err != nil {
		panic(err.Error())
	}
}

//readConf ... Read configuration information
func readConf() error {
	fileName := gofutils.SelfDir() + "conf/" + configName
	if err := gofutils.TouchFile(fileName); err != nil {
		return err
	}
	vp.SetConfigFile(fileName)
	if err := vp.ReadInConfig(); err != nil {
		return err
	}
	ptr := &defaultConf
	key := gofutils.SnakeString(gofutils.ObjectName(ptr))
	if vp.IsSet(key) {
		return vp.UnmarshalKey(key, ptr)
	}
	vp.Set(key, ptr)
	if err := vp.WriteConfig(); err != nil {
		return err
	}
	return nil
}

//createDatabaseEngine  The database connection is created
func createDatabaseEngine() error {
	if err := readConf(); err != nil {
		return err
	}
	dbType := defaultConf.Master.Type
	switch dbType {
	case "mysql":
		sdn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
			defaultConf.Master.User,
			defaultConf.Master.Password,
			defaultConf.Master.Address,
			defaultConf.Master.Port,
			defaultConf.Master.DB,
		)
		db, err := gorm.Open(dbType, sdn)
		if err != nil {
			return err
		}
		Master = db
		dbs = append(dbs, db)
		sdnSlave := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8&parseTime=True&loc=Local",
			defaultConf.Master.User,
			defaultConf.Master.Password,
			defaultConf.Slave.Address,
			defaultConf.Slave.Port,
			defaultConf.Master.DB,
		)
		dbSlave, err := gorm.Open(dbType, sdnSlave)
		if err != nil {
			return err
		}
		Slave = db
		dbs = append(dbs, dbSlave)
	case "postgres":
		sdn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			defaultConf.Master.User,
			defaultConf.Master.Password,
			defaultConf.Master.Address,
			defaultConf.Master.Port,
			defaultConf.Master.DB,
		)
		db, err := gorm.Open(dbType, sdn)
		if err != nil {
			return err
		}
		Master = db
		dbs = append(dbs, db)
		sdnSlave := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
			defaultConf.Master.User,
			defaultConf.Master.Password,
			defaultConf.Slave.Address,
			defaultConf.Slave.Port,
			defaultConf.Master.DB,
		)
		dbSlave, err := gorm.Open(dbType, sdnSlave)
		if err != nil {
			return err
		}
		Slave = db
		dbs = append(dbs, dbSlave)
	default:
		return errors.Errorf("Unsupported database type:%s!", dbType)
	}
	return nil
}

func settingDatabase() error {
	if err := createDatabaseEngine(); err != nil {
		return err
	}
	for _, db := range dbs {
		db.SingularTable(true)
		db.DB().SetConnMaxLifetime(60 * time.Second)
		db.DB().SetMaxIdleConns(10)
		db.DB().SetMaxOpenConns(100)
		db.LogMode(defaultConf.ShowSQL)

		db.SetLogger(defaultLogger)
	}
	gofconf.Job.JobQueue <- func() {
		tk := time.NewTicker(30 * time.Second)
		for {
			select {
			case <-tk.C:
				for _, db := range dbs {
					db.DB().Ping()
				}
			}
		}
	}
	return nil
}
