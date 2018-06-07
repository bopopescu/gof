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
 * created at 2018-06-06 08:19:05
 ******************************************************************************/

package gof

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gitee.com/goframe/gof/gofutils"
	"github.com/robfig/cron"
	"github.com/sirupsen/logrus"
)

//ByteSize ...
type ByteSize uint64

//kb,mb,gb
const (
	_           = iota             // ignore first value by assigning to blank identifier
	KB ByteSize = 1 << (10 * iota) // 1 << (10*1)
	MB                             // 1 << (10*2)
	GB                             // 1 << (10*3)
	TB                             // 1 << (10*4)
	// PB                             // 1 << (10*5)
	// EB                             // 1 << (10*6)
	// ZB                             // 1 << (10*7)
	// YB                             // 1 << (10*8)
)

var (
	//fileMap 全局文件 map
	fileMap = make(map[string]*File)
	//n 秒 处理一次 文件
	freshTime = 1 * time.Second
	//文件 备份 最大文件大小
	maxFile = 1 * MB
	//保留 n 天
	dirDays = 3
)

var (
	Cron *cron.Cron
)

func init() {
	Cron = cron.New()
	Cron.Start()
}

//Logger ...
type Logger struct {
	*logrus.Logger
}

//File ...
type File struct {
	Logger *Logger
	f      *os.File
	mu     sync.RWMutex
}

//GetFile ...
func GetFile(fName string) *File {
	return getFile(fName)
}

//getFile ... 获取一个初始化文件指针,文件名可以带路径,如果文件不存在,则创建
func getFile(fName string) *File {
	if fileMap[fName] == nil {
		fb := &File{mu: sync.RWMutex{}}
		if err := fb.innerFile(fName); err != nil {
			panic(err)
		}
		l := &logrus.Logger{
			Out:       fb.f,
			Formatter: new(logrus.TextFormatter),
			Hooks:     make(logrus.LevelHooks),
			Level:     logrus.DebugLevel,
		}
		fb.Logger = &Logger{l}
		go fb.backPack()
		//每日凌晨1点执行
		//0 0 1 * * *
		Cron.AddFunc("0 0 1 * * *", func() {
			fb.delDirForDays()
		})
		fileMap[fName] = fb
	}
	return fileMap[fName]
}

func (fl *File) delDirForDays() {
	dirSlice := make([]string, 0)    //不被删除的目录
	dir := filepath.Dir(fl.f.Name()) //当前文件目录
	for i := 0; i < dirDays; i++ {
		dStr := time.Now().Local().AddDate(0, 0, i*-1).Format("20060102")
		dirStr := fmt.Sprintf("%s%s%s", dir, gofutils.Delimiter, dStr)
		dirSlice = append(dirSlice, dirStr)
	}

	//当前文件目录下的所有目录
	filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			return nil
		}
		if path == dir {
			return nil
		}
		for _, v := range dirSlice {
			if v == path {
				return nil
			}
		}
		return os.RemoveAll(path)
	})
}

func (fl *File) packAction() error {

	//获取文件大小
	size, err := gofutils.GetFileSize(fl.f)
	if err != nil {
		return err
	}
	if size >= int64(maxFile) {
		//文件加锁
		fl.mu.Lock()
		defer fl.mu.Unlock()
		//获取文件当日目录
		dir := gofutils.TodayDir(fl.f)
		//获取文件名
		stat, err := fl.f.Stat()
		if err != nil {
			return err
		}
		var fullName string
		name := stat.Name()
		n := 0
		for {
			fullName = fmt.Sprintf("%s%s%s.%06d.log", dir, gofutils.Delimiter, name, n)
			if fl.fileExist(fullName) {
				n++
			} else {
				break
			}
		}
		if err := os.MkdirAll(filepath.Dir(fullName), 0755); err != nil {
			return err
		}
		_, err = gofutils.CopyFile(fullName, fl.f.Name())
		if err != nil {
			return err
		}
		if err := os.Truncate(fl.f.Name(), 0); err != nil {
			return err
		}
		fl.f.Sync()
	}
	return nil
}

func (fl *File) backPack() {
	tk := time.NewTicker(freshTime)
	defer func() {
		tk.Stop()
		fl.f.Close()
	}()
	for {
		select {
		case <-tk.C:
			if err := fl.packAction(); err != nil {
				log.Printf("err:%s\n", err.Error())
				return
			}
		}
	}
}

//GetFile ...
func (fl *File) GetFile() *os.File {
	return fl.f
}

//GetLogger ...
func (fl *File) GetLogger() *Logger {
	return fl.Logger
}

func (fl *File) fileExist(fName string) bool {
	if _, err := os.Stat(fName); os.IsNotExist(err) {
		return false
	}
	return true
}

func (fl *File) innerFile(fName string) error {
	if !fl.fileExist(fName) {
		os.MkdirAll(filepath.Dir(fName), 0755)
	}
	file, err := os.OpenFile(fName, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0666)
	if err != nil {
		return err
	}
	fl.f = file
	return nil
}

//Reload reload file pointer
func (fl *File) Reload() {
	fl.mu.Lock()
	defer fl.mu.Unlock()
	os.Truncate(fl.f.Name(), 0)
}
