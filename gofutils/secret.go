package core

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/Luzifer/go-openssl"
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
