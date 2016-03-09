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
