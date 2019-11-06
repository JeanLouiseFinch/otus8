package otus8

import "sync"

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
				*runnig++
			} else {
				*runnig--
			}
		}
	}(runningChan, &runnig)

	select {
	case <-signalStop:
		wg.Wait()
		return
	case <-signalError:
		errCount++
		if errCount == errorCount {
			return
		}
	default:
		if runnig < runCount {
			wg.Add(1)
			go func(signalError chan bool, f func() error, runningChan chan<- bool) {
				runningChan <- true
				defer func(runningChan chan<- bool) {
					runningChan <- false
					wg.Done()
				}(runningChan)
				err := f()
				if err != nil {
					signalError <- true
				}
			}(signalError, tasks[0], runningChan)
			if len(tasks) == 1 {
				signalStop <- true
			} else {
				tasks = tasks[1:]
			}
		}
	}
}
