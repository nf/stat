package stat

import (
	"log"
	"os"
	"rpc"
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
		var err os.Error
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
