package main

import (
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/ideahitme/circuit-breaker"
)

func main() {
	rand.Seed(time.Now().Unix())

	cb := circuitbreaker.New("twitter-api",
		circuitbreaker.WithFailureThreshold(1),
		circuitbreaker.WithSuccessThreshold(10),
		circuitbreaker.WithOpenPeriod(2*time.Second),
		circuitbreaker.WithCounterResetPeriod(1*time.Minute),
		circuitbreaker.WithLogger(logrus.New()),
	)

	resp, err := cb.Exec(HTTPGetter("https://www.google.com"))
	if err != nil {
		panic(err)
	}
	if httpResp, ok := resp.(*http.Response); ok {
		fmt.Printf("response status: %d\n", httpResp.StatusCode)
	}

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

func HTTPGetter(url string) circuitbreaker.RequestFunc {
	return func() (interface{}, error) {
		return http.Get(url)
	}
}
