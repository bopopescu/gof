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
 * created at 2018-06-06 22:39:55
 ******************************************************************************/

package gofutils

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/atcharles/gof/openssl"
)

//DefSecretString ...
var DefSecretString = "rYtY0RD5hvN2T0McxjNWfH1MM7PExE0w"

/*func initSecret() {
	str := random.String(32)
	key := "auth:secret"
	b, err := DefCache.Remember(key, func() error {
		return DefCache.Set(key, str, 0)
	})
	if err != nil {
		log.Fatalln(err.Error())
	}
	DefSecretString = string(b)
}*/

//Md5string ...
func Md5string(in string) string {
	md5s := md5.New().Sum([]byte(in))
	strMd5 := hex.EncodeToString(md5s)
	return strMd5
}

//EncryPassword ...
//加密一个密码
func EncryPassword(pwd string) (string, error) {
	o := openssl.New()
	md5str := Md5string(pwd)
	bt, err := o.EncryptString(DefSecretString, md5str)
	if err != nil {
		return "", err
	}
	return string(bt), nil
}

//VerifyPassword ...
//验证一个密码
func VerifyPassword(pwd string, encryptPwd string) (bool, error) {
	o := openssl.New()
	md5str := Md5string(pwd)
	bt, err := o.DecryptString(DefSecretString, encryptPwd)
	if err != nil {
		return false, err
	}
	return md5str == string(bt), nil
}
