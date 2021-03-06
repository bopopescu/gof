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
	"crypto/aes"
	"crypto/md5"
	"encoding/hex"
	"errors"
)

// Md5 returns the MD5 checksum string of the data.
func Md5(b []byte) string {
	checksum := md5.Sum(b)
	return hex.EncodeToString(checksum[:])
}

// AESEncrypt encrypts a piece of data.
// The cipherkey argument should be the AES key,
// either 16, 24, or 32 bytes to select
// AES-128, AES-192, or AES-256.
func AESEncrypt(cipherkey, src []byte) []byte {
	block, err := aes.NewCipher(cipherkey)
	if err != nil {
		panic(err)
	}
	bs := block.BlockSize()
	src = padData(src, bs)
	r := make([]byte, len(src))
	dst := r
	for len(src) > 0 {
		block.Encrypt(dst, src)
		src = src[bs:]
		dst = dst[bs:]
	}
	dst = make([]byte, hex.EncodedLen(len(r)))
	hex.Encode(dst, r)
	return dst
}

// AESDecrypt decrypts a piece of data.
// The cipherkey argument should be the AES key,
// either 16, 24, or 32 bytes to select
// AES-128, AES-192, or AES-256.
func AESDecrypt(cipherkey, ciphertext []byte) ([]byte, error) {
	block, err := aes.NewCipher(cipherkey)
	if err != nil {
		return nil, err
	}
	src := make([]byte, hex.DecodedLen(len(ciphertext)))
	_, err = hex.Decode(src, ciphertext)
	if err != nil {
		return nil, err
	}
	bs := block.BlockSize()
	r := make([]byte, len(src))
	dst := r
	for len(src) > 0 {
		block.Decrypt(dst, src)
		src = src[bs:]
		dst = dst[bs:]
	}
	return removePad(r)
}

func padData(d []byte, bs int) []byte {
	padedSize := ((len(d) / bs) + 1) * bs
	pad := padedSize - len(d)
	for i := len(d); i < padedSize; i++ {
		d = append(d, byte(pad))
	}
	return d
}

func removePad(r []byte) ([]byte, error) {
	l := len(r)
	if l == 0 {
		return []byte{}, errors.New("input []byte is empty")
	}
	last := int(r[l-1])
	pad := r[l-last : l]
	isPad := true
	for _, v := range pad {
		if int(v) != last {
			isPad = false
			break
		}
	}
	if !isPad {
		return r, errors.New("remove pad error")
	}
	return r[:l-last], nil
}
