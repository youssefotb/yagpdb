package common

import (
	"encoding/json"
	"github.com/mediocregopher/radix.v3"
)

// GetRedisJson executes a get redis command and unmarshals the value into out
func GetRedisJson(key string, out interface{}) error {
	var resp []byte
	err := RedisPool.Do(radix.Cmd(&resp, "GET", key))
	if err != nil {
		return err
	}

	if len(resp) == 0 {
		return nil
	}

	err = json.Unmarshal(resp, out)
	return err
}

// SetRedisJson marshals the value and runs a set redis command for key
func SetRedisJson(key string, value interface{}) error {
	serialized, err := json.Marshal(value)
	if err != nil {
		return err
	}

	err = RedisPool.Do(radix.Cmd(nil, "SET", key, string(serialized)))
	return err
}

// func RedisBool(resp *redis.Resp) (b bool, err error) {
// 	if resp.Err != nil {
// 		return false, resp.Err
// 	}

// 	if resp.IsType(redis.Nil) {
// 		return false, nil
// 	}

// 	if resp.IsType(redis.Int) {
// 		i, err := resp.Int()
// 		return i > 0, err
// 	}

// 	if resp.IsType(redis.Str) {
// 		s, err := resp.Str()
// 		return (s != "" && s != "false" && s != "0"), err
// 	}

// 	panic("Unknown redis reply type")
// }
