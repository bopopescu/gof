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
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/tidwall/buntdb"
)

//Cache ...
var (
	Cache    CacheInterface
	RdiCache *RedisCache
	MemCache *MemoryCache
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
	Exists(key string) bool
	Del(key string) error
	DelAll() error
}

func initCache() error {
	var err error
	if process.UseRedis {
		Cache, err = NewRedisCache()
	} else {
		Cache, err = NewMemoryCache()
	}
	MemCache, err = NewMemoryCache()
	return err
}

//NewRedisCache ...
func NewRedisCache() (*RedisCache, error) {
	if RdiCache != nil {
		return RdiCache, nil
	}
	cache := &RedisCache{
		mu: new(sync.Mutex),
	}
	client := redis.NewClient(&redis.Options{
		Addr:     redisOption.Addr,
		Password: redisOption.Password,
		DB:       redisOption.DB,
	})
	cache.Client = client
	if err := cache.Client.Ping().Err(); err != nil {
		return nil, err
	}
	RdiCache = cache
	return cache, nil
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
func NewMemoryCache() (*MemoryCache, error) {
	if MemCache != nil {
		return MemCache, nil
	}
	MemCache = &MemoryCache{
		mu: new(sync.Mutex),
	}
	db, err := buntdb.Open(":memory:")
	if err != nil {
		return nil, err
	}
	MemCache.Client = db
	return MemCache, nil
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
