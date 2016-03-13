package bird

import "sync"

// Operation type of manipulation
type Operation int

// Possible operation types
const (
	Land Operation = iota
	Raise
	Include
	Exclude
)

func (op Operation) String() string {
	switch op {
	case Land:
		return "Land"
	case Raise:
		return "Raise"
	case Include:
		return "Include"
	case Exclude:
		return "Exclude"
	default:
		return "Unknown"
	}
}

// Action details of single manipulation with birds group
type Action struct {
	Birds     []*SmartBird
	Operation Operation
}

// Flock is a collection of smart birds
type Flock struct {
	birds          map[*SmartBird]bool
	access         sync.RWMutex
	journalEnabled bool
	journal        chan Action
}

// NewFlock initializes new flock of smart birds
func NewFlock() *Flock {
	return &Flock{birds: make(map[*SmartBird]bool), journal: make(chan Action)}
}

// Journal of all manipulations with birds in the flock. If journal was invoked,
// reader must read it all otherwise all block
func (f *Flock) Journal() <-chan Action {
	f.journalEnabled = true
	return f.journal
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
		f.log(Include, smartBird)
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
		if landBird {
			birdToExclude.Stop()
			f.log(Land, birdToExclude)
		}
		delete(f.birds, birdToExclude)
		f.log(Exclude, birdToExclude)
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
	if landBird {
		f.log(Land, selected...)
	}
	f.log(Exclude, selected...)
	return ans
}

// Land some smart birds in a flock by their names.
// If names not specified - all birds are used
func (f *Flock) Land(names ...string) {
	f.access.RLock()
	defer f.access.RUnlock()
	wg := sync.WaitGroup{}
	selected := f.selectUnsafe(names...)
	wg.Add(len(selected))
	for _, bird := range selected {
		go func(bird *SmartBird) {
			defer wg.Done()
			bird.Stop()
		}(bird) // run stop in separate routing because of Stop() may be slow operation
	}
	wg.Wait()
	f.log(Land, selected...)
}

// Raise all smart birds in a flock.
// If names not specified - all birds are used
func (f *Flock) Raise(names ...string) {
	f.access.RLock()
	defer f.access.RUnlock()
	selected := f.selectUnsafe(names...)
	for _, bird := range selected {
		bird.Start()
	}
	f.log(Raise, selected...)
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
	f.Exclude(land)
}

func (f *Flock) log(operation Operation, birds ...*SmartBird) {
	if f.journalEnabled && len(birds) > 0 {
		f.journal <- Action{birds, operation}
	}
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
			if names[0] == bird.name {
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
