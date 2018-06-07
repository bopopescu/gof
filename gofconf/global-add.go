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
	"time"

	"gitee.com/goframe/gof/gofutils"
)

// The configuration program uses viper;see: github.com/spf13/viper
// The structure tag uses `mapstructure`
// When writing a file, the tag changes according to the file type.
// For example, json, yaml...
const (
	// GlobalFileName is the name of the global program configuration file
	GlobalFileName = "__global.yaml"
)

var (
	DefaultProcess = Process{
		ListenPort: 8100,
		Mode:       "release",
		CacheType:  "memory",
		Secret:     gofutils.NewRandom(gofutils.Crs).RandomString(32),
	}
	DefaultRedis = Redis{
		Addr:     "127.0.0.1:6379",
		Password: "",
	}
	DefaultLog = Log{
		ConsoleEnable: true,
		FileEnable:    true,
		FilePath:      "logs/web" + gofutils.Delimiter,
		//Level:         "info",
	}
)

type (
	// Process Program global configuration items
	Process struct {
		// Key        string `mapstructure:"-"`                //the name of config key
		ListenPort   int    // server listen port
		Mode         string // program run mode,debug or release
		CacheType    string // redis or memory
		Secret       string // program secret , use to jwt
		ReadTimeOut  time.Duration
		WriteTimeOut time.Duration
	}
	// Redis set;need cacheType = `redis`
	Redis struct {
		Addr     string // redis server address;example:127.0.0.1:6379
		Password string
		DB       int
	}
	// Log Log system Settings
	// The server log system is placed in the "logs/web" directory.
	// Whenever the log file size exceeds 1MB, the system will automatically backup the log file.
	// The system will only back up the log file for the last 3 days.
	Log struct {
		ConsoleEnable bool
		FileEnable    bool
		FilePath      string // Program current directory;`logs/web`
		//Level         string // error|warn|info|debug
	}
)

//InitFunc ReadIn ...
func (p *Process) InitFunc() error {
	return ReadObjInformation(&DefaultProcess)
}

//InitFunc ReadIn ...
func (p *Redis) InitFunc() error {
	return ReadObjInformation(&DefaultRedis)
}

//InitFunc ReadIn ...
func (p *Log) InitFunc() error {
	return ReadObjInformation(&DefaultLog)
}
