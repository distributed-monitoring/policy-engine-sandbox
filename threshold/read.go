/*
 * Copyright 2018 NEC Corporation
 *
 *   Licensed under the Apache License, Version 2.0 (the "License");
 *   you may not use this file except in compliance with the License.
 *   You may obtain a copy of the License at
 *
 *       http://www.apache.org/licenses/LICENSE-2.0
 *
 *   Unless required by applicable law or agreed to in writing, software
 *   distributed under the License is distributed on an "AS IS" BASIS,
 *   WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 *   See the License for the specific language governing permissions and
 *   limitations under the License.
 */

package threshold

import (
	"fmt"
	"github.com/distributed-monitoring/policy-engine-sandbox/policyexpr"
	"github.com/go-redis/redis"
	"os"
	"strconv"
	"strings"
	"time"
)

// e.g. collectd/instance-00000001/virt/if_octets-tapd21acb51-35
// const redisKey = "collectd/*/virt/if_octets-*"

func zrangebyscore(client *redis.Client, key string, index int) []float64 {

	unixNow := int(time.Now().Unix())
	interval := 60

	val, err := client.ZRangeByScore(key, redis.ZRangeBy{
		Min: strconv.Itoa(unixNow - interval),
		Max: strconv.Itoa(unixNow),
	}).Result()

	datalist := []float64{}

	if err == redis.Nil {
		fmt.Println("this key is not exist")
		os.Exit(1)
	} else if err != nil {
		panic(err)
	} else {
		for _, strVal := range val {
			split := strings.Split(strVal, ":")
			txVal := split[index+1] // First elem is time
			floatVal, err := strconv.ParseFloat(txVal, 64)
			if err != nil {
				os.Exit(1)
			}
			datalist = append(datalist, floatVal)
		}
	}
	return datalist
}

func Read(p *policyexpr.Parser) []rawData {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})

	field := p.Left.ExprVar
	redisKey := strings.Replace(field, "vm.", "virt/", 1)

	index := -1
	if strings.HasSuffix(redisKey, ".rx") {
		redisKey = strings.TrimSuffix(redisKey, ".rx")
		index = 0
	} else if strings.HasSuffix(redisKey, ".tx") {
		redisKey = strings.TrimSuffix(redisKey, ".tx")
		index = 1
	} else {
		index = 0
	}

	fmt.Printf("debug: %#v\n", redisKey)

	keys, err := client.Keys("*" + redisKey + "*").Result()
	if err != nil {
		panic(err)
	}

	rdlist := []rawData{}

	for _, key := range keys {
		datalist := zrangebyscore(client, key, index)
		subkeys := strings.Split(key, "/")
		subsubkeys := strings.SplitN(subkeys[3], "-", 2)
		if strings.HasPrefix(subsubkeys[0], "if_") {
			rdlist = append(rdlist, rawData{key: ResourceLabel{VM: subkeys[1], IF: subsubkeys[1]}, datalist: datalist})
		} else {
			rdlist = append(rdlist, rawData{key: ResourceLabel{VM: subkeys[1], IF: ""}, datalist: datalist})
		}
	}

	fmt.Printf("debug: %#v\n", rdlist)
	return rdlist
}
