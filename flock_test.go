package bird

import (
	"testing"
	"time"
)

func TestFlock(t *testing.T) {
	flock := NewFlock()
	defer flock.Dissolve(true)
	bird := NewSmartBird(noop, 20*time.Second, "Betty")
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
