package session

import (
	"sync"
)

type LatencyCalc struct {
	sync.Mutex
	latencies [4]int64
	num       int64
}

func (p *LatencyCalc) AddLatency(us int64) {
	p.Lock()
	defer p.Unlock()
	p.latencies[p.num%4] = us
	p.num++
	//log.Println(fmt.Sprintf("Latency:%d", us))
}

func (p *LatencyCalc) CalcLatency() int64 {
	p.Lock()
	defer p.Unlock()
	if p.num < 4 {
		return -1
	}

	return (p.latencies[3] + p.latencies[2] + p.latencies[1] + p.latencies[0]) / 4
}
