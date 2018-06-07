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
 * created at 2018-06-06 08:18:28
 ******************************************************************************/

package gofconf

import (
	"log"
	"os"
	"runtime"

	"gitee.com/goframe/gof/gofutils"
	"github.com/fsnotify/fsnotify"
	"github.com/ivpusic/grpool"
	"github.com/spf13/viper"
)

var (
	//Queue External configuration, you must pass the queue in the use method.
	Queue          = make(chan func())
	Job            = grpool.NewPool(100, runtime.NumCPU())
	innerFuncGroup = make([]Init, 0)
)

// Init ...
type (
	Init interface {
		InitFunc() error
	}
)

//AddDefaultInformation ...
func AddDefaultInformation(obj ...Init) {
	innerFuncGroup = append(innerFuncGroup, obj...)
}

// ReadObjInformation Read information from the configuration file into a global variable,
// and if there is no information about the object in the configuration file,
// write the initial properties of the object to the configuration file.
func ReadObjInformation(ptr Init) error {
	key := gofutils.SnakeString(gofutils.ObjectName(ptr))
	if viper.IsSet(key) {
		return viper.UnmarshalKey(key, ptr)
	}
	viper.Set(key, ptr)
	// Execute the initialization event,Write to the configuration file.
	if err := viper.WriteConfig(); err != nil {
		return err
	}
	return nil
}

func initConfig() error {
	fileName := gofutils.SelfDir() + "conf/" + GlobalFileName
	if err := gofutils.TouchFile(fileName); err != nil {
		panic(err.Error())
	}
	viper.SetConfigType("yaml")
	viper.SetConfigFile(fileName)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	viper.WatchConfig()
	viper.OnConfigChange(func(e fsnotify.Event) {
		viper.ReadInConfig()
		for _, c := range innerFuncGroup {
			if err := c.InitFunc(); err != nil {
				log.Println(err.Error())
			}
		}
	})
	return nil
}

// Initialize ...
// This method needs to be referenced when the configuration file needs to be initialized
func Initialize() {
	if err := initConfig(); err != nil {
		panic(err.Error())
	}

	innerFuncGroup = append(innerFuncGroup,
		&DefaultProcess,
		&DefaultRedis,
		&DefaultLog,
	)
	for _, c := range innerFuncGroup {
		if err := c.InitFunc(); err != nil {
			panic(err.Error())
		}
	}

	Job.JobQueue <- func() {
		defer func() {
			if p := recover(); p != nil {
				log.SetFlags(log.LstdFlags)
				log.Println(string(gofutils.PanicTrace(4)))
			}
		}()
		for {
			select {
			case obFunc := <-Queue:
				obFunc()
			}
		}
	}

	log.SetFlags(log.LstdFlags)
	log.Printf("The PID of the current process is: %d \n", os.Getpid())
}
