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

//Start  -
func Start(tasks []func() error, runCount, errorCount int) {
	var (
		taskC                                                 chan func() error
		errOut, signalExit, done, finishErrorCount, doneError chan bool
		wg                                                    *sync.WaitGroup
	)
	taskC = make(chan func() error)
	errOut = make(chan bool)
	signalExit = make(chan bool)
	done = make(chan bool)
	finishErrorCount = make(chan bool)
	doneError = make(chan bool)

	// zapuskaem funcs v rabotu
	for i := 0; i < runCount; i++ {
		go worker(taskC, errOut, signalExit, wg)
	}

	go func() {
		errorCnt := 0
		for {
			select {
			case <-errOut:
				errorCnt++
				if errorCnt == errorCount {
					close(doneError)
					return
				}
			case <-finishErrorCount:
				return
			}
		}
	}()

out:
	for _, task := range tasks {
		select {
		case taskC <- task:
			wg.Add(1)
		case <-doneError:
			break out
		}
	}
	// zdem
	go func() {
		wg.Wait()
		done <- true
	}()

	// zakruvaem
	for {
		select {
		case <-doneError:
			close(signalExit)
			return
		case <-done:
			close(finishErrorCount)
			close(signalExit)
			return
		}
	}

}

func worker(task <-chan func() error, errC chan bool, signal <-chan bool, waitgroup *sync.WaitGroup) {
	for {
		select {
		case task := <-task:
			err := task()
			waitgroup.Done()
			if err != nil {
				errC <- true
			}
		case <-signal:
			return
		}
	}

}
