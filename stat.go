// Copyright 2011 Google Inc.
// 
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
// 
//      http://www.apache.org/licenses/LICENSE-2.0
// 
//      Unless required by applicable law or agreed to in writing, software
//      distributed under the License is distributed on an "AS IS" BASIS,
//      WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
//      See the License for the specific language governing permissions and
//      limitations under the License.

package stat

import (
	"log"
	"net/rpc"
	"time"
)

var (
	In      = make(chan string, 10000)
	Process = "default"
)

const period = 1e9

type Point struct {
	Process string
	Series  string
	Value   int64
}

type counter struct {
	total, period, cycles int64
}

var (
	counters = make(map[string]*counter)
	client   *rpc.Client
)

func Monitor(addr string) {
	if addr != "" {
		var err error
		client, err = rpc.DialHTTP("tcp", addr)
		if err != nil {
			log.Println(err)
		}
	}
	t := time.NewTicker(period)
	for {
		select {
		case <-t.C:
			output()
		case s := <-In:
			c, ok := counters[s]
			if !ok {
				c = &counter{}
				counters[s] = c
			}
			c.period++
		}
	}
}

func output() {
	for s, c := range counters {
		c.total += c.period
		c.cycles++
		if client != nil {
			update(s, c.period)
		} else {
			log.Printf("%s: %d total, %d last, %d avg",
				s,
				c.total,
				c.period,
				c.total/c.cycles)
		}
		c.period = 0
	}
}

func update(series string, value int64) {
	err := client.Call("Server.Update", &Point{Process, series, value}, &struct{}{})
	if err != nil {
		log.Println("stat update:", err)
	}
}
