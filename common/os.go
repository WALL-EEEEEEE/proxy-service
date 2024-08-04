package common

import (
	"os"
	os_signal "os/signal"
)

func Signal(signals []os.Signal, callback func() error) {
	// graceful shutdown
	c := make(chan os.Signal, 2)
	os_signal.Notify(c, signals...)
	go func() {
		for range c {
			// sig is a ^C, handle it
		}
		err := callback()
		if err != nil {
			panic(err)
		}
	}()
}
