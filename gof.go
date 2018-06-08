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

package gof

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/atcharles/gof/gofcache"
	"github.com/atcharles/gof/gofconf"
	"github.com/atcharles/gof/goflogger"
	"github.com/atcharles/gof/goform"
	"github.com/atcharles/gof/gofutils"
	"github.com/gin-contrib/cors"
	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
)

var (
	DEFCorsConfig = CorsConfig{
		AllowAllOrigins:  true,
		AllowMethods:     []string{gofconf.GET, gofconf.POST, gofconf.PUT, gofconf.PATCH, gofconf.HEAD, gofconf.DELETE},
		AllowHeaders:     []string{},
		ExposeHeaders:    []string{"Content-Length", "Set-Cap-Key"},
		AllowCredentials: false,
		MaxAge:           12 * time.Hour,
	}
)

type CorsConfig struct {
	AllowAllOrigins  bool
	AllowOrigins     []string `yaml:",flow"`
	AllowMethods     []string `yaml:",flow"`
	AllowHeaders     []string `yaml:",flow"`
	ExposeHeaders    []string `yaml:",flow"`
	AllowCredentials bool
	MaxAge           time.Duration
}

func (p *CorsConfig) InitFunc() error {
	return gofconf.ReadObjInformation(p)
}

func New() *gin.Engine {
	gin.SetMode(gofconf.DefaultProcess.Mode)
	multi := make([]io.Writer, 0)
	if gofconf.DefaultLog.FileEnable {
		fl := goflogger.GetFile(gofutils.SelfDir() + gofconf.DefaultLog.FilePath + "info.log")
		multi = append(multi, fl.GetFile())
	}
	if gofconf.DefaultLog.ConsoleEnable {
		multi = append(multi, os.Stdout)
	}
	gin.DefaultWriter = io.MultiWriter(multi...)
	fl2 := goflogger.GetFile(gofutils.SelfDir() + gofconf.DefaultLog.FilePath + "error.log")
	gin.DefaultErrorWriter = io.MultiWriter(fl2.GetFile())
	corsConfig := cors.Config{
		AllowAllOrigins:  DEFCorsConfig.AllowAllOrigins,
		AllowOrigins:     DEFCorsConfig.AllowOrigins,
		AllowMethods:     DEFCorsConfig.AllowMethods,
		AllowHeaders:     DEFCorsConfig.AllowHeaders,
		ExposeHeaders:    DEFCorsConfig.ExposeHeaders,
		AllowCredentials: DEFCorsConfig.AllowCredentials,
		MaxAge:           DEFCorsConfig.MaxAge,
	}
	router := gin.New()
	router.Use(
		gin.Logger(),
		gin.Recovery(),
		limit.MaxAllowed(runtime.NumCPU()),
		cors.New(corsConfig),
		gzip.Gzip(9),
	)
	return router
}

func Run(router *gin.Engine, addr ...string) error {
	var thisAddr string
	if len(addr) == 0 {
		thisAddr = fmt.Sprintf(":%d", gofconf.DefaultProcess.ListenPort)
	} else {
		thisAddr = addr[0]
	}
	srv := http.Server{
		Addr:         thisAddr,
		ReadTimeout:  gofconf.DefaultProcess.ReadTimeOut,
		WriteTimeout: gofconf.DefaultProcess.WriteTimeOut,
		Handler:      router,
	}
	return srv.ListenAndServe()
}

func init() {
	gofconf.AddDefaultInformation(&DEFCorsConfig)

	gofconf.Initialize()
	gofcache.InitCache()
	goform.Initialize()
}
