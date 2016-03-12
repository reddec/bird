package bird

import (
	"log"
	"strconv"
	"sync"
	"sync/atomic"
	"time"
)

// Gun is a channel with stop signal. Any value from this channel indicates that bird must die
type Gun <-chan int

//Handler - handle processing by implementing interface
type Handler interface {
	Run(stop Gun) error // Stop - is signal to force stop
}

// Bird is a function which must run forver untill stop signal occrued.
// If bird stops whenever with or without error it will be restarted after
// specified timeout.
// External systems may request force stop by sending any value to stop channel
type Bird func(stop Gun) error

// The ErrorFunc function hould process error produced by Bird function
type ErrorFunc func(err error)

// StopFunc is a gun which may kill the Bird
type StopFunc func()

// Bird name generator (for unnamed birds)
var birdID uint64

// Fly forver with specified restart timeout. Returns gun which can
// interrupt bird flying
func Fly(bird Bird, restartTimeout time.Duration) StopFunc {
	id := atomic.AddUint64(&birdID, 1)
	return FlyNamed(bird, restartTimeout, "bird "+strconv.FormatUint(id, 10))
}

// FlyNamed is same as Fly but with custom bird name in log
func FlyNamed(bird Bird, restartTimeout time.Duration, name string) StopFunc {
	log.Println("[", name, "] [info] started")
	gun := FlyWithError(bird, func(err error) {
		log.Println("[", name, "] [error] (restart after", restartTimeout, ")", err)
	}, restartTimeout)
	return func() {
		log.Println("[", name, "] [info] killing..")
		gun()
		log.Println("[", name, "] [info] killed")
	}
}

// FlyHandle is simple wrapper for Fly which allows use interfaces as bird
func FlyHandle(bird Handler, restartTimeout time.Duration) StopFunc {
	return Fly(bird.Run, restartTimeout)
}

// FlyHandleNamed  is same as FlyHandle but with custom bird name in log
func FlyHandleNamed(bird Handler, restartTimeout time.Duration, name string) StopFunc {
	return FlyNamed(bird.Run, restartTimeout, name)
}

// FlyHandleWithError is simple wrapper for FlyWithError which allows use interfaces as bird
func FlyHandleWithError(bird Handler, errorHanlder ErrorFunc, restartTimeout time.Duration) StopFunc {
	return FlyWithError(bird.Run, errorHanlder, restartTimeout)
}

// FlyWithError is a function which monitor and restart bird flying untill gun invoked.
// Returns gun which can interrupt bird flying
func FlyWithError(bird Bird, errorHanlder ErrorFunc, restartTimeout time.Duration) StopFunc {
	return fly(bird, errorHanlder, restartTimeout)
}

func fly(bird Bird, errorHanlder ErrorFunc, restartTimeout time.Duration) StopFunc {
	kill := make(chan int, 1)
	wg := sync.WaitGroup{}
	restart := true
	//Start bird
	wg.Add(1)
	go func() {
		defer wg.Done()
	LOOP:
		for restart {
			err := bird(kill) // Bird can be killed by signal
			if err != nil && errorHanlder != nil {
				errorHanlder(err)
			}
			if restart {
				select {
				case <-time.After(restartTimeout):
				case <-kill:
					break LOOP
				}
			}
		}
	}()
	stopMutex := sync.Mutex{}
	return func() {
		if restart {
			stopMutex.Lock()
			defer stopMutex.Unlock()
			if restart {
				restart = false
				kill <- 1
				wg.Wait()
			}
		}
	}
}

// A SmartBird is smarter bird then usual: it can starts and stops many times
type SmartBird struct {
	name    string
	restart time.Duration
	bird    Bird
	gun     StopFunc
	access  sync.Mutex
}

// NewSmartBird constructs new Smart Bird based on usuall bird, but without auto start
func NewSmartBird(bird Bird, restart time.Duration, name string) *SmartBird {
	return &SmartBird{name: name, restart: restart, bird: bird}
}

// Name of bird
func (g *SmartBird) Name() string { return g.name }

// Interval between restart
func (g *SmartBird) Interval() time.Duration { return g.restart }

// Flying status
func (g *SmartBird) Flying() bool { return g.gun != nil }

// Start bird flying. Thread-safe
func (g *SmartBird) Start() {
	if g.gun != nil {
		return
	}
	g.access.Lock()
	defer g.access.Unlock()
	if g.gun == nil {
		g.gun = FlyNamed(g.bird, g.restart, g.name)
	}
}

// Stop bird flying but without bird's death. Thread-safe
func (g *SmartBird) Stop() {
	if g.gun == nil {
		return
	}
	g.access.Lock()
	defer g.access.Unlock()
	if g.gun != nil {
		g.gun()
		g.gun = nil
	}
}
