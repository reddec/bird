package bird

import (
	"errors"
	"fmt"
	"log"
	"os"
	"testing"
	"time"
)

func noop(kill Gun) error {
	select {
	case <-kill:
		break
	case <-time.After(1 * time.Second):
		log.Println("noop")
	}
	return nil
}

func noopErr(kill Gun) error {
	select {
	case <-kill:
		break
	case <-time.After(1 * time.Second):
		log.Println("noop 2")
	}
	return errors.New("NOOP err")
}

func TestFlySuccess(t *testing.T) {
	gun := Fly(noop, 1*time.Second)
	time.Sleep(1 * time.Second)
	gun()
}

func TestFlyErr(t *testing.T) {
	gun := Fly(noopErr, 200*time.Millisecond)
	time.Sleep(3 * time.Second)
	gun()
}

type eagle struct {
	counter int
}

func (ea *eagle) Run(kill Gun) error {
	ea.counter++
	return nil
}

func TestFlyHandle(t *testing.T) {
	orlando := &eagle{}
	gun := FlyHandleNamed(orlando, 1*time.Second, "Orlando")
	time.Sleep(3 * time.Second)
	gun()
	if orlando.counter < 3 || orlando.counter > 4 {
		t.Fatal("Bad synchronization: Orlando has", orlando.counter)
	}
	t.Log("Orlando has", orlando.counter)
}

func TestSmartFly(t *testing.T) {
	sokol := NewSmartBird(noop, 1*time.Second, "Sokol")
	time.Sleep(2 * time.Second)
	sokol.Stop()

	sokol.Start()
	sokol.Stop()
	sokol.Start()
	go sokol.Start()
	sokol.Start()
	sokol.Stop()
	go sokol.Stop()
	sokol.Stop()
}

func ExampleFly() {
	log.SetOutput(os.Stderr)
	gun := FlyNamed(func(kill Gun) error {
		fmt.Println("I'm flying!")
		return nil
	}, 1*time.Second, "Canary")
	time.Sleep(1900 * time.Millisecond)
	gun()
	fmt.Println("Bird is killed")
	//Output: I'm flying!
	//I'm flying!
	//Bird is killed
}
