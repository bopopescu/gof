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
	"os"

	"github.com/gin-gonic/gin"
	"github.com/heqiawen/gof/gof-middleware"
	"github.com/labstack/echo"
	"github.com/robfig/cron"
)

//kb,mb,gb
const (
	_         = iota             // ignore first value by assigning to blank identifier
	KB uint64 = 1 << (10 * iota) // 1 << (10*1)
	MB                           // 1 << (10*2)
	GB                           // 1 << (10*3)
	TB                           // 1 << (10*4)
	// PB                             // 1 << (10*5)
	// EB                             // 1 << (10*6)
	// ZB                             // 1 << (10*7)
	// YB                             // 1 << (10*8)
)

var (
	global *Frame

	//Cron ... 默认初始化一个 cron, 后台运行定时器任务
	//如果需要定时任务,直接添加进来即可
	//example:	...
	//每日凌晨1点执行
	//0 0 1 * * *
	/*
		Cron.AddFunc("0 0 1 * * *", func() {
			//do something
		})
	*/
	Cron *cron.Cron
)

//Frame implement gin IRouter interface,extends gin engine
type Frame struct {
	*gin.Engine
	logger  log.Logger
	Routers *gin.RouterGroup
}

//Run ...
func Run(addr ...string) {
	if len(addr) == 0 {
		addr = append(addr, fmt.Sprintf(":%d", process.Port))
	}
	log.Printf("The PID of the current process is: %d \n", os.Getpid())
	if err := global.Run(addr...); err != nil {
		log.Fatalln(err.Error())
	}
}

func initialization() {
	//初始化配置项内容
	initConfig()
	//gin run mode
	gin.SetMode(process.GinMode)
	global = new(Frame)
	global.Engine = gin.New()
	global.Routers = global.Engine.Group("/", func(c *gin.Context) {
		server := fmt.Sprintf("GOF/%s:/%s", echo.Version, gin.Version)
		c.Writer.Header().Set(echo.HeaderServer, server)
		fmt.Println(server)
		c.Next()
	}, gin.Logger(), gin.Recovery(), gof_middleware.CORS())
	//初始化定时器
	Cron = cron.New()
	Cron.Start()
}

//New ...
func New() *Frame {
	initialization()
	return global
}
