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

	cb := circuitbreaker.New(
		circuitbreaker.WithFailureThreshold(1),
		circuitbreaker.WithSuccessThreshold(10),
		circuitbreaker.WithOpenPeriod(2*time.Second),
		circuitbreaker.WithCounterResetPeriod(1*time.Minute),
		circuitbreaker.WithLogger(logrus.New()),
	)

	for i := 0; i < 10; i++ {
		res, err := cb.Exec(circuitbreaker.RequestFunc(RandFunc))
		time.Sleep(500 * time.Millisecond)
		// seems okay
		// need tests
		// but first let's try some randomization
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

// THE MOST NAIVE IMPLEMENTATION SEEMS TO BE WORKING FINE
// WILL ITERATE ON IT LATER
// TIME TO SLEEP

func RandFunc() (interface{}, error) {
	if rand.Intn(2) == 0 {
		return 0, errors.New("fail")
	}
	return rand.Intn(1000), nil
}
