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
