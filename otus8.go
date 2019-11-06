package otus8

import (
	"fmt"
	"sync"
)

func Init(tasks []func() error, runCount, errorCount int) {
	var (
		signalStop, signalError, runningChan chan bool
		errCount                             int
		wg                                   sync.WaitGroup
		runnig                               int
	)

	signalStop = make(chan bool)
	runningChan = make(chan bool)
	signalError = make(chan bool, runCount)

	go func(runningChan <-chan bool, runnig *int) {
		select {
		case val := <-runningChan:
			if val {
				fmt.Println("+1")
				*runnig++
			} else {
				fmt.Println("-1")
				*runnig--
			}
		}
	}(runningChan, &runnig)

	select {
	case <-signalStop:
		fmt.Println("Waiting...")
		wg.Wait()
		return
	case <-signalError:
		errCount++
		if errCount == errorCount {
			fmt.Printf("Error count:%v\n", errCount)
			return
		}
	default:
		if runnig < runCount {
			wg.Add(1)
			go func(signalError chan bool, f func() error, runningChan chan<- bool) {
				fmt.Println("Starting func")
				runningChan <- true
				defer func(runningChan chan<- bool) {
					runningChan <- false
					wg.Done()
				}(runningChan)
				err := f()
				if err != nil {
					fmt.Println("Signal error func")
					signalError <- true
				}
			}(signalError, tasks[0], runningChan)
			if len(tasks) == 1 {
				fmt.Println("Signal stop")
				signalStop <- true
			} else {
				tasks = tasks[1:]
			}
		}
	}
}
