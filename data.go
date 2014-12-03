package main

import (
	"encoding/json"
	"strings"
	//"strconv"
	"time"
	//"github.com/davecgh/go-spew/spew"
	"github.com/garyburd/redigo/redis"
)

const SLEEP_TIME = 30

func saveComment() {
	for {
		doSaveComment()
		time.Sleep(time.Second * SLEEP_TIME)
	}
}

func doSaveComment() bool {
	//Here:
	rd := rdPool.Get()
	//RedisExec(rd, "SELECT", "3")
	defer rd.Close()

	members, err := redis.Strings(RedisExec(rd, "ZRANGE", REDIS_SCREENSHOT_COMMENT_ZSET_KEY, 0, -1, "WITHSCORES"))
	if err != nil {
		log4e("err", err)
		//time.Sleep(time.Second * SLEEP_TIME)
		//goto Here
		return false
	}

	for i := 0; i < len(members); i = i + 2 {
		hash := members[i]
		key_suffix := members[i+1]
		//key := "sreenshot:comment:list:" + key_suffix
		key := strings.Replace(REDIS_SCREENSHOT_COMMENT_LIST_KEY, "IMGID", key_suffix, 1)

		data, err := redis.String(RedisExec(rd, "HGET", key, hash))
		if err != nil && err != redis.ErrNil {
			log4e("Redis HGET Error (%v),", err)
			continue

		}

		if data == "" {
			RedisExec(rd, "ZREM", REDIS_SCREENSHOT_COMMENT_ZSET_KEY, hash)
			continue
		}

		r := JsonString{}
		err = json.Unmarshal([]byte(data), &r)
		if err != nil {
			log4e("json.Unmarshal Error (%v)", err, data)
			RedisExec(rd, "ZREM", REDIS_SCREENSHOT_COMMENT_ZSET_KEY, hash)
			continue

		}

		if dbSaveComment(&r) == nil {
			RedisSend(rd, "MULTI")
			RedisExec(rd, "ZREM", REDIS_SCREENSHOT_COMMENT_ZSET_KEY, hash)
			RedisExec(rd, "HDEL", key, hash)
			_, err = RedisExec(rd, "EXEC")
			if err != nil {
				log4e("Redis MULTI EXEC(%v),", err)
			}
		}
	}

	return true
}
