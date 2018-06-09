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
 * created at 2018-06-06 21:00:19
 ******************************************************************************/

package gofcache

import (
	"encoding/json"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/atcharles/gof/gofconf"
	"github.com/go-redis/redis"
	"github.com/spf13/viper"
	"github.com/tidwall/buntdb"
)

//DefCache ...
var (
	DefCache         CacheInterface
	RedisGlobalCache *RedisCache
	MeCache          *MemoryCache
)

//CacheInterface ... 缓存接口
type CacheInterface interface {
	Get(key string) ([]byte, error)
	//if cannot get value,return 0
	GetInt64(key string) (int64, error)
	//return "" if can't get value
	GetValue(key string) (string, error)
	//bind value to struct point
	Bind(key string, bean interface{}) error
	Set(key string, value interface{}, exp time.Duration) error
	Remember(key string, set func() error) ([]byte, error)
	RememberBind(key string, bean interface{}, set func() error) error
	Exists(key string) bool
	Del(key string) error
	DelAll() error
}

func init() {
	DefCache = NewMemoryCache()
	MeCache = NewMemoryCache()
}

func InitCache() {
	switch gofconf.DefaultProcess.CacheType {
	case "redis":
		DefCache = NewRedisCache()
	case "memory":
		DefCache = NewMemoryCache()
	}
	MeCache = NewMemoryCache()
}

//NewRedisCache ...
func NewRedisCache() *RedisCache {
	if RedisGlobalCache != nil {
		return RedisGlobalCache
	}
	cache := &RedisCache{
		mu: new(sync.Mutex),
	}
	op := &redis.Options{}
	err := viper.UnmarshalKey("redis", op)
	if err != nil {
		log.Fatalf("load redis err : %s\n", err.Error())
	}
	client := redis.NewClient(op)
	cache.Client = client
	RedisGlobalCache = cache
	return cache
}

//RedisCache ...
type RedisCache struct {
	mu     *sync.Mutex
	Client *redis.Client
}

//JSONSet ... 将一个对象序列化成 json 字符串,并进行存储
func (r *RedisCache) JSONSet(key string, value interface{}, exp time.Duration) error {
	beanValue := reflect.ValueOf(value)
	if beanValue.Kind() == reflect.String {
		value = value.(string)
	} else {
		b, err := json.Marshal(value)
		if err != nil {
			return err
		}
		value = string(b)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	_, err := r.Client.SetNX(key, value, exp).Result()
	if err != nil {
		return err
	}
	return nil
}

//Get ...
func (r *RedisCache) Get(key string) (b []byte, err error) {
	b, err = r.Client.Get(key).Bytes()
	return
}

//GetInt64 ...
func (r *RedisCache) GetInt64(key string) (int64, error) {
	return r.Client.Get(key).Int64()
}

//GetValue ...
func (r *RedisCache) GetValue(key string) (string, error) {
	b, err := r.Client.Get(key).Bytes()
	return string(b), err
}

//Bind ...
func (r *RedisCache) Bind(key string, bean interface{}) error {
	return r.Client.Get(key).Scan(bean)
}

//Set write operation,and need lock
func (r *RedisCache) Set(key string, value interface{}, exp time.Duration) error {
	return r.JSONSet(key, value, exp)
}

//Remember ...
func (r *RedisCache) Remember(key string, set func() error) (b []byte, err error) {
	if !r.Exists(key) {
		//between write operation,need lock
		err = set()
		if err != nil {
			return
		}
	}
	return r.Client.Get(key).Bytes()
}

//RememberBind ...
func (r *RedisCache) RememberBind(key string, bean interface{}, set func() error) error {
	if !r.Exists(key) {
		err := set()
		if err != nil {
			return err
		}
	}
	return r.Bind(key, bean)
}

//Exists ...
func (r *RedisCache) Exists(key string) bool {
	h := r.Client.Exists(key).Val()
	return h > 0
}

//Del ...
func (r *RedisCache) Del(key string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, err := r.Client.Del(key).Result()
	return err
}

//DelAll ...
func (r *RedisCache) DelAll() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, err := r.Client.Del(r.Client.Keys("*").Val()...).Result()
	return err
}

//NewMemoryCache ...
func NewMemoryCache() *MemoryCache {
	if MeCache != nil {
		return MeCache
	}
	MeCache = &MemoryCache{
		mu: new(sync.Mutex),
	}
	db, err := buntdb.Open(":memory:")
	if err != nil {
		log.Fatalf("failed to run memory db: %s", err.Error())
	}
	MeCache.Client = db
	return MeCache
}

//MemoryCache ...
type MemoryCache struct {
	mu     *sync.Mutex
	Client *buntdb.DB
}

//Get ...
func (m *MemoryCache) Get(key string) ([]byte, error) {
	val, err := m.GetValue(key)
	if err != nil {
		return nil, err
	}
	return []byte(val), nil
}

//GetInt64 ...
func (m *MemoryCache) GetInt64(key string) (i int64, err error) {
	val, err := m.GetValue(key)
	ita, err := strconv.Atoi(val)
	if err != nil {
		return 0, err
	}
	return int64(ita), nil
}

//GetValue ...
func (m *MemoryCache) GetValue(key string) (string, error) {
	var str string
	err := m.Client.View(func(tx *buntdb.Tx) error {
		val, err := tx.Get(key)
		if err != nil {
			return err
		}
		str = val
		return nil
	})
	if err != nil {
		err = fmt.Errorf("memory cache get err: %s", err.Error())
		return "", err
	}
	return str, nil
}

//Bind ...
func (m *MemoryCache) Bind(key string, bean interface{}) error {
	val, err := m.GetValue(key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), bean)
}

//Set ...
func (m *MemoryCache) Set(key string, value interface{}, exp time.Duration) error {
	beanValue := reflect.ValueOf(value)
	return m.Client.Update(func(tx *buntdb.Tx) error {
		var val string
		if beanValue.Kind() == reflect.String {
			val = value.(string)
		} else {
			bt, err := json.Marshal(value)
			if err != nil {
				return err
			}
			val = string(bt)
		}
		expires := exp > 0
		_, _, err := tx.Set(key, val, &buntdb.SetOptions{Expires: expires, TTL: exp})
		if err != nil {
			return err
		}
		return nil
	})
}

//Remember ...
func (m *MemoryCache) Remember(key string, set func() error) (b []byte, err error) {
	if !m.Exists(key) {
		err = set()
		if err != nil {
			return
		}
	}
	return m.Get(key)
}

//RememberBind ...
func (m *MemoryCache) RememberBind(key string, bean interface{}, set func() error) error {
	if !m.Exists(key) {
		err := set()
		if err != nil {
			return err
		}
	}
	return m.Bind(key, bean)
}

//Exists ...
func (m *MemoryCache) Exists(key string) bool {
	err := m.Client.View(func(tx *buntdb.Tx) error {
		_, err := tx.Get(key)
		if err != nil {
			return err
		}
		return nil
	})
	if err == buntdb.ErrNotFound {
		return false
	}
	return true
}

//Del ...
func (m *MemoryCache) Del(key string) error {
	return m.Client.Update(func(tx *buntdb.Tx) error {
		_, err := tx.Delete(key)
		if err != nil {
			err = fmt.Errorf("memory cache get err: %s", err.Error())
		}
		return err
	})
}

//DelAll ...
func (m *MemoryCache) DelAll() error {
	return m.Client.Update(func(tx *buntdb.Tx) error {
		return tx.DeleteAll()
	})
}
