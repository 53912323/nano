// Copyright (c) nano Authors. All Rights Reserved.
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
// SOFTWARE.

package scheduler

import (
	"fmt"
	"reflect"
	"runtime"
	"runtime/debug"
	"sync/atomic"
	"time"

	"github.com/lonng/nano/internal/env"
	"github.com/lonng/nano/internal/log"
	"github.com/smallnest/chanx"
)

const (
	messageQueueBacklog = 1 << 10
	sessionCloseBacklog = 1 << 8
)

type Task func()

// LocalScheduler schedules task to a customized goroutine
type LocalScheduler interface {
	Schedule(Task)
}

type Hook func()

var (
	chDie    = make(chan struct{})
	chExit   = make(chan struct{})
	chTask   = chanx.NewUnboundedChan(messageQueueBacklog)
	started  int32
	closed   int32
	lastTick time.Time
)

func try(f func()) {
	defer func() {
		if err := recover(); err != nil {
			errInfo := fmt.Sprintf("Handle message panic: %+v\n%s", err, debug.Stack())
			log.Error(errInfo)
			fmt.Println("--panic--", errInfo)
		}
	}()
	lastTick = time.Now()
	f()
	if lastTick.Add(time.Second / 4).Before(time.Now()) {
		v := reflect.ValueOf(f).Pointer()
		vf := runtime.FuncForPC(v)
		file, line := vf.FileLine(v)
		log.Println(fmt.Sprintf("task time out,name:\n%s\nsource:%s:%d", vf.Name(), file, line))
	}
}

func Sched() {
	if atomic.AddInt32(&started, 1) != 1 {
		return
	}

	ticker := time.NewTicker(env.TimerPrecision)
	defer func() {
		ticker.Stop()
		close(chExit)
	}()

	for {
		select {
		case <-ticker.C:
			cron()

		case f := <-chTask.Out:
			try(f.(Task))

		case <-chDie:
			return
		}
	}
}

func Close() {
	if atomic.AddInt32(&closed, 1) != 1 {
		return
	}
	close(chDie)
	<-chExit
	log.Println("Scheduler stopped")
}

func PushTask(task Task) {
	chTask.In <- task
}

func TaskChanLen() int {
	return len(chTask.Out)
}
