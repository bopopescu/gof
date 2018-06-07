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
 * created at 2018-06-05 15:16:55
 ******************************************************************************/

package gofconf

import (
	"time"
)

//The configuration program uses viper;see: github.com/spf13/viper
//The structure tag uses `mapstructure`
//When writing a file, the tag changes according to the file type.
//For example, json, yaml...
const (
	//GlobalFileName is the name of the global program configuration file
	GlobalFileName = "__global.yaml"
)

var (
	DefaultProcess = Process{
		ListenPort: 80,
		Mode:       "release",
		CacheType:  "memory",
	}
	DefaultDatabaseSet = DatabaseSet{
		Type:     "postgres",
		Addr:     "127.0.0.1:5432",
		User:     "root",
		Password: "",
		DB:       "demo",
	}
	DefaultRedis = Redis{
		Addr:     "127.0.0.1:6379",
		Password: "",
	}
	DefaultLog = Log{
		ConsoleEnable: true,
		FileEnable:    true,
		FilePath:      "logs/web",
		Level:         "info",
	}
)

type (
	//Process Program global configuration items
	Process struct {
		//Key        string `mapstructure:"-"`                //the name of config key
		ListenPort   int    //server listen port
		Mode         string //program run mode,debug or release
		CacheType    string //redis or memory
		Secret       string //program secret , use to jwt
		ReadTimeOut  time.Duration
		WriteTimeOut time.Duration
	}
	//DatabaseSet Database setup configuration items
	DatabaseSet struct {
		Type     string //mysql,postgres,sqlite
		Addr     string //example:127.0.0.1:5432
		User     string
		Password string
		DB       string
	}
	//Redis set;need cacheType = `redis`
	Redis struct {
		Addr     string //redis server address;example:127.0.0.1:6379
		Password string
		DB       int
	}
	//Corss Cross-domain access Settings
	Corss struct {
		// AllowOrigin defines a list of origins that may access the resource.
		// Optional. Default value []string{"*"}.
		AllowOrigins []string `mapstructure:"allow_origins" yaml:"allow_origins"`
		// AllowMethods defines a list methods allowed when accessing the resource.
		// This is used in response to a preflight request.
		// Optional. Default value DefaultCORSConfig.AllowMethods.
		AllowMethods []string `mapstructure:"allow_methods" yaml:"allow_methods"`
		// AllowHeaders defines a list of request headers that can be used when
		// making the actual request. This in response to a preflight request.
		// Optional. Default value []string{}.
		AllowHeaders []string `mapstructure:"allow_headers" yaml:"allow_headers"`
		// AllowCredentials indicates whether or not the response to the request
		// can be exposed when the credentials flag is true. When used as part of
		// a response to a preflight request, this indicates whether or not the
		// actual request can be made using credentials.
		// Optional. Default value false.
		AllowCredentials bool `mapstructure:"allow_credentials" yaml:"allow_credentials"`
		// ExposeHeaders defines a whitelist headers that clients are allowed to
		// access.
		// Optional. Default value []string{}.
		ExposeHeaders []string `mapstructure:"expose_headers" yaml:"expose_headers"`
		// MaxAge indicates how long (in seconds) the results of a preflight request
		// can be cached.
		// Optional. Default value 0.
		MaxAge int `mapstructure:"max_age" yaml:"max_age"`
	}
	//Log Log system Settings
	//The server log system is placed in the "logs/web" directory.
	//Whenever the log file size exceeds 1MB, the system will automatically backup the log file.
	//The system will only back up the log file for the last 3 days.
	Log struct {
		ConsoleEnable bool
		FileEnable    bool
		FilePath      string //Program current directory;`logs/web`
		Level         string //error|warning|info|debug
	}
)
