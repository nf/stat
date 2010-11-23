package stat

import (
	"log"
	"time"
)

var In = make(chan string, 10000)

type counter struct {
	total, period, cycles int64
}

var counters = make(map[string]*counter)

func Monitor(period int64) {
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
	if len(counters) > 0 {
		log.Println()
	}
	for s, c := range counters {
		c.total += c.period
		c.cycles++
		log.Printf("%s: %d total, %d last, %d avg", 
			s,
			c.total,
			c.period,
			c.total/c.cycles)
		c.period = 0
	}
}

