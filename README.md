## Circuit breaker 

Thread-safe circuit breaker implementation in Go

*See more in https://docs.microsoft.com/en-us/azure/architecture/patterns/circuit-breaker*

### Install

```bash
go get github.com/ideahitme/circuit-breaker
```

### API 

Default circuitbreaker: 

```go
cb := circuitbreaker.New("twitter-api") // name the circuit-breaker, by the name of the service it guards 
...
```

Extensions: 

```go
// see example below on how to extend default circuitbreaker
WithFailureThreshold(x int) // how many consecutive requests should fail for the circuitbreaker to disallow any requests (default = 5)
WithSuccessThreshold(x int) // how many consucutive requests should succeed for the circuitbreaker to consider service recovered (default = 5)
WithOpenPeriod(x time.Duration) // how much time should circuitbreaker block access to the service, i.e. keep it in the open state (default = 1min)
WithCounterResetPeriod(x time.Duration) // interval after which consecutive requests counters are set to zero
WithLogger(Logger) // allows to enable logging, see circuitbreaker.Logger interface and example below (default no logging)
```

Invocation: 
```go
	resp, err := cb.Exec(HTTPGetter("https://www.google.com"))
	if err != nil {
		panic(err)
	}
	if httpResp, ok := resp.(*http.Response); ok {
		fmt.Printf("response status: %d\n", httpResp.StatusCode)
	}

  ...

  func HTTPGetter(url string) circuitbreaker.RequestFunc {
    return func() (interface{}, error) {
      return http.Get(url)
    }
  }

```

Additional methods: 

```go
cb.Reset() // resets the state and counter to the default state
cb.Block() // blocks all requests until unblock is called
cb.Unblock() // unblocks circuit breaker and returns it to normal operational mode
```

### Example

```go

package main

import (
	"errors"
	"time"
	"math/rand"

	"github.com/Sirupsen/logrus"
	"github.com/ideahitme/circuit-breaker"
)

func main() {
	rand.Seed(time.Now().Unix())

	cb := circuitbreaker.New( "twitter-api", 
		circuitbreaker.WithFailureThreshold(1),
		circuitbreaker.WithSuccessThreshold(10),
		circuitbreaker.WithOpenPeriod(2*time.Second),
		circuitbreaker.WithCounterResetPeriod(1*time.Minute),
		circuitbreaker.WithLogger(logrus.New()),
	)

	for i := 0; i < 10; i++ {
		res, err := cb.Exec(circuitbreaker.RequestFunc(RandFunc))
		time.Sleep(500 * time.Millisecond)
		if err != nil {
			logrus.Error(err)
			continue
		}
		d, ok := res.(int)
		if ok {
			logrus.Infof("got response: %d", d)
		}
	}
}

func RandFunc() (interface{}, error) {
	if rand.Intn(2) == 0 {
		return 0, errors.New("fail")
	}
	return rand.Intn(1000), nil
}

```

### To be implemented: 

- [ ] Write tests
- [ ] Be aware of the errors, for example if HTTP code indicates that service is overloaded, we should not wait for threshold but immediately switch to open state
- [ ] Adjust timeouts, for example timeout for service crash can be auto-adjusted as specified by the user
- [ ] Allow to dynamically reconfigure circuit breaker, i.e. change the default threshold or timeouts based on the returned error
- [ ] Allow to auto-ping service to enable faster switch to half-open state (instead of default timer), should be possible if service exposes a health check API
- [ ] Allow to record failed requests in the journal and replay them later 