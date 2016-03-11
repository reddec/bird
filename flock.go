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

// Exclude smart birds from a flock by their names and then returns them as list
func (f *Flock) Exclude(birdName string, landBird bool) []*SmartBird {
	var ans []*SmartBird
	f.access.Lock()
	defer f.access.Unlock()
	for bird := range f.birds {
		if bird.Name() == birdName {
			if landBird {
				bird.Stop()
			}
			ans = append(ans, bird)
		}
	}
	for _, bird := range ans {
		delete(f.birds, bird)
	}
	return ans
}

// Land all smart birds in a flock
func (f *Flock) Land() {
	f.access.RLock()
	defer f.access.RUnlock()
	wg := sync.WaitGroup{}
	wg.Add(len(f.birds))
	for bird := range f.birds {
		go func(bird *SmartBird) {
			defer wg.Done()
			bird.Stop()
		}(bird) // run stop in separate routing because of Stop() may be slow operation
	}
	wg.Wait()
}

// Raise all smart birds in a flock in the ai
func (f *Flock) Raise() {
	f.access.RLock()
	defer f.access.RUnlock()
	for bird := range f.birds {
		bird.Start()
	}
}

// Census all smart birds in a flock
func (f *Flock) Census() []*SmartBird {
	f.access.RLock()
	defer f.access.RUnlock()
	ans := make([]*SmartBird, 0, len(f.birds))
	for bird := range f.birds {
		ans = append(ans, bird)
	}
	return ans
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
