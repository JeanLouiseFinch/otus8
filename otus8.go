package otus8

import (
	"sync"
)

func Starting(tasks []func() error, workersCnt, maxErrorCnt int) {
	var (
		taskChan                                               chan func() error
		errOut, finishAll, finishError, doneWorkers, doneError chan bool
	)
	taskChan = make(chan func() error)
	errOut = make(chan bool)
	finishAll = make(chan bool)

	var waitgroup sync.WaitGroup
	for i := 0; i < workersCnt; i++ {
		go worker(taskChan, errOut, finishAll, &waitgroup)
	}

	finishError = make(chan bool)
	doneError = make(chan bool)
	go func() {
		errorCnt := 0
		for {
			select {
			case <-errOut:
				errorCnt++
				if errorCnt == maxErrorCnt {
					close(doneError)
					return
				}
			case <-finishError:
				return
			}
		}
	}()

out:
	for _, task := range tasks {
		select {
		case taskChan <- task:
			waitgroup.Add(1)
		case <-doneError:
			break out
		}
	}

	doneWorkers = make(chan bool)
	go func() {
		waitgroup.Wait()
		doneWorkers <- true
	}()

	for {
		select {
		case <-doneError:
			close(finishAll)
			return
		case <-doneWorkers:
			close(finishError)
			close(finishAll)
			return
		}
	}

}

func worker(task <-chan func() error, errOut chan bool, finishAll <-chan bool, waitgroup *sync.WaitGroup) {
	for {
		select {
		case task := <-task:
			err := task()
			waitgroup.Done()
			if err != nil {
				errOut <- true
			}
		case <-finishAll:
			return
		}
	}

}