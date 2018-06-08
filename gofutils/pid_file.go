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
 * created at 2018-06-06 08:18:29
 ******************************************************************************/

package gofutils

import (
	"fmt"
	"os"
	"path/filepath"
)

// DEFAULT_PID_FILE the default PID file name
var DefaultPidFile = "logs/PID"

// WritePidFile writes the current PID to the specified file.
func WritePidFile(pidFile ...string) {
	flame := DefaultPidFile
	if len(pidFile) > 0 {
		flame = pidFile[0]
	}
	abs, err := filepath.Abs(flame)
	if err != nil {
		panic(err)
	}
	dir := filepath.Dir(abs)
	os.MkdirAll(dir, 0777)
	f, err := os.OpenFile(abs, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0666)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.WriteString(fmt.Sprintf("%d\n", os.Getpid()))
}
