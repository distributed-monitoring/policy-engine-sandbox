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
//	"fmt"
//	"github.com/BurntSushi/toml"
//	"log"
)

// Config is ...
type Config struct {
	Common    CommonConfig
	Threshold ThresholdConfig
}

// CommonConfig is ...
type CommonConfig struct {
	RedisHost     string `toml:"redis_host"`
	RedisPort     string `toml:"redis_port"`
	RedisPassword string `toml:"redis_password"`
	RedisDB       int    `toml:"redis_db"`
}

// ThresholdConfig is ...
type ThresholdConfig struct {
	RedisHost     string `toml:"redis_host"`
	RedisPort     string `toml:"redis_port"`
	RedisPassword string `toml:"redis_password"`
	RedisDB       int    `toml:"redis_db"`

	Interval int `toml:"interval"`
	Min      int `toml:"min"`

	CollectdPlugin string `toml:"collectd_plugin"`
	CollectdType   string `toml:"collectd_type"`
}

type ResourceLabel struct {
	VM string
	IF string
}

type rawData struct {
	key      ResourceLabel
	datalist []float64
}

// func main(p *policyexpr.Parser) []string{
//	var config Config
//	_, err := toml.DecodeFile("/etc/barometer-dma/config.toml", &config)
//	if err != nil {
//		log.Fatalf("Read error of config: %s", err)
//	}
//
//	thresConfig := config.Threshold
//	log.Printf("Raw data redis config Addr:%s:%s DB:%d", thresConfig.RedisHost, thresConfig.RedisPort, thresConfig.RedisDB)
//	if thresConfig.RedisPassword == "" {
//		log.Printf("Raw data redis password is not set")
//	}
//
//				result1 := read(&config)
//				// analysis()
//				result2 := evaluate(&config, result1)
//				// transmit(&config, result2)
//
//	fmt.Println("End")
// }
