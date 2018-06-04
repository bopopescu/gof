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
 * created at 2018-06-04 17:59:04
 ******************************************************************************/

package gof

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/Pallinder/go-randomdata"
	"github.com/fsnotify/fsnotify"
	"github.com/gin-gonic/gin"
	"github.com/heqiawen/gof/gof-errors"
	"github.com/heqiawen/gof/gof-middleware"
	"github.com/spf13/viper"
)

var (
	initFs  []InitFunc
	configs []Config
	process = Process{
		Key:      "process",
		Port:     8100,
		GinMode:  gin.ReleaseMode, //release
		UseRedis: false,
	}
	secret      string
	redisOption = RedisOption{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	}
)

//InitConfigKeyString ... 初始化配置函数
type InitFunc func() error

//AddInitFunc ... add one
func AddInitFunc(fn ...InitFunc) {
	initFs = append(initFs, fn...)
}

//innerAddFunc 初始化行为添加 +++++
func innerAddFunc() {
	//添加配置结构体
	configs = append(
		configs,
		&process,
		&redisOption,
		&gof_middleware.DefaultCORSConfig,
	)
	AddInitFunc(func() error {
		key := "secret"
		if viper.IsSet(key) {
			secret = viper.GetString(key)
		} else {
			rand.Seed(time.Now().UnixNano())
			secret = randomdata.RandStringRunes(32)
			viper.Set(key, secret)
			return gof_errors.ErrInitConfig
		}
		return nil
	}, func() error {
		return initCache()
	})
}

//Config ... 配置信息接口
type Config interface {
	ReadIn() error
}

//ConfigReadIn ... 配置信息读取
func ConfigReadIn(config Config) error {
	return config.ReadIn()
}

//ConfigReadInExecute ... 读取所有配置项,init 执行
func ConfigReadInExecute() {
	errs := make([]error, 0)
	innerAddFunc()
	for _, conf := range configs {
		if err := ConfigReadIn(conf); err != nil {
			if err == gof_errors.ErrInitConfig {
				errs = append(errs, err)
			} else {
				log.Fatalln(err.Error())
			}
		}
	}
	for _, fn := range initFs {
		if err := fn(); err != nil {
			if err == gof_errors.ErrInitConfig {
				errs = append(errs, err)
			} else {
				log.Fatalln(err.Error())
			}
		}
	}
	if len(errs) != 0 {
		if err := viper.WriteConfig(); err != nil {
			log.Printf("read in config: %s\n", err.Error())
		}
	}
}

//Process ... 程序基础配置
type Process struct {
	Key      string `mapstructure:"-"`
	Port     int    `mapstructure:"port"`
	GinMode  string `mapstructure:"gin_mode" yaml:"gin_mode"`
	UseRedis bool
}

//ReadIn ...
func (p *Process) ReadIn() error {
	if viper.IsSet(process.Key) {
		return viper.UnmarshalKey("process", p)
	}
	viper.Set("process", process)
	return gof_errors.ErrInitConfig
}

//RedisOption ...
type RedisOption struct {
	Addr     string
	Password string
	DB       int
}

//ReadIn ...
func (r *RedisOption) ReadIn() error {
	key := "redis"
	if viper.IsSet(key) {
		return viper.UnmarshalKey(key, r)
	}
	viper.Set(key, redisOption)
	return gof_errors.ErrInitConfig
}

//init execute
func initConfig() error {
	file := new(File)
	globalFile := fmt.Sprintf("%sconf/__global.yaml", MustGetCurrentPath())
	if err := file.innerFile(fmt.Sprintf(globalFile)); err != nil {
		log.Fatalln(err.Error())
	}

	//viper.SetConfigType("toml")
	viper.AddConfigPath(MustGetCurrentPath() + "conf")
	viper.SetConfigFile(globalFile)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Printf("read config failed . %s\n", err.Error())
		return err
	}

	//读取配置
	ConfigReadInExecute()

	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		viper.ReadInConfig()
		ConfigReadInExecute()
	})
	return nil
}
