package bird

import (
	"testing"
	"time"
)

func TestFlock(t *testing.T) {
	flock := NewFlock()
	defer flock.Dissolve(true)
	bird := NewSmartBird(&testBird{}, 20*time.Second, "Betty")
	flock.Include(bird)
	match := flock.Select()[0]
	if match != bird {
		t.Fatal("Included bird not matched")
	}
	flock.Land()
	for _, bird := range flock.Select() {
		if bird.Flying() {
			t.Fatal("All birds must be landed")
		}
	}
	flock.Raise()
	for _, bird := range flock.Select() {
		if !bird.Flying() {
			t.Fatal("All birds must be raised")
		}
	}
	flock.Dissolve(true)
	if bird.Flying() {
		t.Fatal("Bird must be landed if flag specified in Dissolve")
	}
	if len(flock.Select()) != 0 {
		t.Fatal("Flock must be empty after dissolve")
	}
	flock.Include(bird)
	flock.Raise()
	flock.Dissolve(false)
	if !bird.Flying() {
		t.Fatal("Bird must be in air if flag NOT specified in Dissolve")
	}
	if len(flock.Select()) != 0 {
		t.Fatal("Flock must be empty after dissolve")
	}
	//TODO: Add tests for other functions
}

func TestFlockJournal(t *testing.T) {
	flock := NewFlock()
	defer flock.Dissolve(true)

	var removed, added, landed, raised bool
	journal := flock.Journal()
	go func() {
		for action := range journal {
			t.Log(action)
			switch action.Operation {
			case Land:
				landed = true
			case Exclude:
				removed = true
			case Raise:
				raised = true
			case Include:
				added = true
			}
		}
	}()

	bird := NewSmartBird(&testBird{}, 20*time.Second, "Betty")
	flock.Include(bird)
	time.Sleep(100 * time.Millisecond) // Allow go-routing do work
	if !(!removed && added && !landed && !raised) {
		t.Fatal("Include action not invoked")
	}
	added = false

	flock.Raise(bird.Name())
	time.Sleep(100 * time.Millisecond)
	if !(!removed && !added && !landed && raised) {
		t.Fatal("Raise action not invoked")
	}
	raised = false

	flock.Land(bird.Name())
	time.Sleep(100 * time.Millisecond)
	if !(!removed && !added && landed && !raised) {
		t.Fatal("Land action not invoked")
	}
	landed = false

	flock.Dissolve(false)
	time.Sleep(100 * time.Millisecond)
	if !(removed && !added && !landed && !raised) {
		t.Fatal("Exclude action not invoked")
	}
}
