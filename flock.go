package bird

import "sync"

// Flock is a collection of smart birds
type Flock struct {
	birds  map[*SmartBird]bool
	access sync.RWMutex
}

// NewFlock initializes new flock of smart birds
func NewFlock() *Flock {
	return &Flock{birds: make(map[*SmartBird]bool)}
}

// Include new smart bird to a flock or do nothing if it is already in
func (f *Flock) Include(smartBird *SmartBird) {
	if f.birds[smartBird] {
		return
	}
	f.access.Lock()
	defer f.access.Unlock()
	if !f.birds[smartBird] {
		f.birds[smartBird] = true
	}
}

// Take out smart bird from a flock, land it (if required) and return it.
// If bird wasn't be in a flock then returns nil
func (f *Flock) Take(birdToExclude *SmartBird, landBird bool) *SmartBird {
	if !f.birds[birdToExclude] {
		return nil
	}
	f.access.Lock()
	defer f.access.Unlock()
	if f.birds[birdToExclude] {
		delete(f.birds, birdToExclude)
		if landBird {
			birdToExclude.Stop()
		}
		return birdToExclude
	}
	return nil
}

// Exclude some smart birds from a flock by their names and then returns them as list
func (f *Flock) Exclude(landBird bool, birdNames ...string) []*SmartBird {
	var ans []*SmartBird
	f.access.Lock()
	defer f.access.Unlock()
	selected := f.selectUnsafe(birdNames...)
	for _, bird := range selected {
		if landBird {
			bird.Stop()
		}
		delete(f.birds, bird)
	}
	return ans
}

// Land some smart birds in a flock by their names.
// If names not specified - all birds are used
func (f *Flock) Land(names ...string) {
	f.access.RLock()
	defer f.access.RUnlock()
	wg := sync.WaitGroup{}
	wg.Add(len(f.birds))
	for _, bird := range f.selectUnsafe(names...) {
		go func(bird *SmartBird) {
			defer wg.Done()
			bird.Stop()
		}(bird) // run stop in separate routing because of Stop() may be slow operation
	}
	wg.Wait()
}

// Raise all smart birds in a flock in the ai
// If names not specified - all birds are used
func (f *Flock) Raise(names ...string) {
	f.access.RLock()
	defer f.access.RUnlock()
	for _, bird := range f.selectUnsafe(names...) {
		bird.Start()
	}
}

// Select some smart birds from a flock by their names
// If names not specified - all birds are used
func (f *Flock) Select(names ...string) []*SmartBird {
	f.access.RLock()
	defer f.access.RUnlock()
	return f.selectUnsafe(names...)
}

// Dissolve all smart birds from a flock and optionally land them
func (f *Flock) Dissolve(land bool) {
	f.access.Lock()
	defer f.access.Unlock()
	if land {
		for bird := range f.birds {
			bird.Stop()
		}
	}
	f.birds = make(map[*SmartBird]bool)
}

func (f *Flock) selectUnsafe(names ...string) []*SmartBird {
	ans := make([]*SmartBird, 0, len(f.birds))
	switch { // Some nano-optimization
	case len(names) > 1:
		set := make(map[string]bool)
		for _, name := range names {
			set[name] = true
		}
		for bird := range f.birds {
			if set[bird.name] {
				ans = append(ans, bird)
			}
		}
	case len(names) == 1:
		for bird := range f.birds {
			if names[1] == bird.name {
				ans = append(ans, bird)
			}
		}
	default:
		for bird := range f.birds {
			ans = append(ans, bird)
		}

	}

	return ans
}
