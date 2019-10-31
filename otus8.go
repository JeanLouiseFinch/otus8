package otus8

import "sync"

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
