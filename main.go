package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

//                               processor
//		producer -> consumer ->  processor -> terminator (выводит на экран результат, в наеш случае, суммы квадратов входящих наруальных чисел)
//                               ...
//                               processor
//
// При возникновении ошибки обработки, требуется отменить все последующие расчеты, и вернуть ошибку

const (
	limit           = 10000
	concurrencySize = 5
)

func main() {
	numsChan := producer(limit)
	res := make(chan int, limit)

	go func() {
		if err := consumer(numsChan, res, concurrencySize); err != nil {
			fmt.Println(err.Error())
		}
		close(res)
	}()

	terminator(res)
}

func producer(limit int) chan int {
	out := make(chan int, limit)
	for i := 1; i <= limit; i++ {
		out <- i
	}
	close(out)

	return out
}

func processor(i int) (int, error) {
	if i == 10 {
		return 0, errors.New("i hate 5")
	}
	time.Sleep(5 * time.Second)
	return i * i, nil
}

func consumer(numsChan, quadsChan chan int, threadsCount int) error {
	errChan := make(chan error, 1)
	defer close(errChan)

	semaphore := make(chan struct{}, threadsCount)
	defer close(semaphore)

	wg := sync.WaitGroup{}
	for i := range numsChan {
		if len(errChan) > 0 {
			return <-errChan
		}
		semaphore <- struct{}{}
		wg.Add(1)
		go func(x int) {
			defer wg.Done()
			quad, err := processor(x)
			if err != nil {
				errChan <- err
				return
			}
			quadsChan <- quad
			<-semaphore
		}(i)
	}

	wg.Wait()

	return nil
}

func terminator(results chan int) {
	for i := range results {
		fmt.Println(i)
	}
}
