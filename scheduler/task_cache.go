package scheduler

import (
	"sync"
	"sync/atomic"

	"github.com/lonng/nano/internal/log"
)

type Task func()

var (
	taskBit    = 10 //task chan size 1024
	taskMax    = int32(1<<taskBit - 100)
	taskResume = 20
)

type TaskWithCache struct {
	chTasks   chan Task
	chanSize  int32
	chanFull  int32
	pending   []Task
	muPending sync.Mutex
}

func NewTaskWithCache() *TaskWithCache {
	return &TaskWithCache{
		chTasks:  make(chan Task, 1<<taskBit),
		chanSize: 0,
		chanFull: 0,
	}
}

func (p *TaskWithCache) _pushTask(task Task) {
	atomic.AddInt32(&p.chanSize, 1)
	p.chTasks <- task
}

func (p *TaskWithCache) _pushTasks(tasks []Task) {
	atomic.AddInt32(&p.chanSize, int32(len(tasks)))
	for _, v := range tasks {
		p.chTasks <- v
	}
}

func (p *TaskWithCache) pushTask(task Task) {
	c := atomic.LoadInt32(&p.chanSize)
	if c < taskMax {
		if atomic.LoadInt32(&p.chanFull) <= 0 { //fast path
			p._pushTask(task)
		} else { //pending queue has elem, resume it!
			p.muPending.Lock()
			l := len(p.pending)
			if l > taskResume {
				tmp := p.pending[:taskResume]
				p.pending = p.pending[taskResume:]
				p.muPending.Unlock()
				p._pushTasks(tmp)
			} else {
				tmp := p.pending
				p.pending = make([]Task, 0)
				p.muPending.Unlock()
				p._pushTasks(tmp)
				atomic.StoreInt32(&p.chanFull, 0)
				log.Println("tasks queue is empty now!")
			}
			p._pushTask(task)
		}
	} else {
		//chan is neary full, push to pending queue
		//task will post new task. if without pending queue, it will dead lock when chan is full.
		p.muPending.Lock()
		if len(p.pending) == 0 {
			atomic.StoreInt32(&p.chanFull, 1)
			log.Println("tasks queue is full!!!")
		}
		p.pending = append(p.pending, task)
		p.muPending.Unlock()
	}
}
