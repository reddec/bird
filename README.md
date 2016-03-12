# bird
Simple restarter for Go rountings

# Example

```go

func SomethingThatCanFail(kill <-chan int) error {
  var err error
  // do something ...
  return err
}

func Test(){
  restart:=1*time.Second
  gun := Fly(SomethingThatCanFail, restart)
  // do something
  gun() // stop go-routing
}
```

# Smart bird

Smart bird can be started and stopped multiple times. Start and stop are thread-safe operations


```go
func SomethingThatCanFail(kill <-chan int) error {
  var err error
  // do something ...
  return err
}

func Test(){
  restart:=1*time.Second
  smartBird:= NewSmartBird(SomethingThatCanFail, restart, "Sokol")
  smartBird.Start() // Start when needs
  // do something
  smartBird.Stop()
  //....
  // Start again if needed
  smartBird.Start()
}
```

# Flocks

Smart birds can be grouped into flock with common operations: raise, land and others (see doc)

```go
flock := NewFlock()
bird := NewSmartBird(noop, 20*time.Second, "Betty")
flock.Include(bird) // Add bird to flock (dublicates are ignored)
flock.Raise() // Raise all birds in flock
// ... Do something
flock.Dissolve(true) // Land all birds and dissolve flock
```

# REST JSON API

Gazer provides JSON API for flock

```go
gz := NewGazer(bird.NewFlock(), nest)
http.Handle("/birds", gz)
http.ListenAndServe(":9090", nil)
```
