package main

import (
	"fmt"
	"log"
	"sort"
	"strconv"
	"strings"
	"sync"
)

// сюда писать код

func main() {
	jobs := []job{
		job(func(in, out chan interface{}) {
			out <- 1
			out <- 2

			close(out)
		}),
		job(func(in, out chan interface{}) {
			for val := range in {
				num := val.(int) * 2

				out <- num
			}
			close(out)
		}),
		job(func(in, out chan interface{}) {
			for val := range in {
				num := val.(int) * 3
				out <- num
			}
			close(out)
		}),
		job(func(in, out chan interface{}) {
			for val := range in {
				log.Println(val)
			}
			close(out)
		}),
	}

	ExecutePipeline(jobs...)
}

func ExecutePipeline(jobs ...job) {
	if len(jobs) == 0 {
		return
	}

	var inCh chan any = nil
	outCh := make(chan any)
	for i, job := range jobs {
		// Run job
		go job(inCh, outCh)

		// Don't create new chans if there is last job
		if i == len(jobs)-1 {
			break
		}

		inCh = outCh
		outCh = make(chan any)
	}

	<-outCh
}

func SingleHash(in, out chan any) {
	wg := &sync.WaitGroup{}

	for data := range in {
		dataStr := strconv.Itoa(data.(int))
		md5 := DataSignerMd5(dataStr)

		wg.Add(1)
		go func(data, md5 string) {
			defer wg.Done()
			hash1, hash2 := "", ""
			hashWg := &sync.WaitGroup{}
			hashWg.Add(2)

			go func() {
				defer hashWg.Done()
				hash1 = DataSignerCrc32(data)
			}()
			go func() {
				defer hashWg.Done()
				hash2 = DataSignerCrc32(md5)
			}()

			hashWg.Wait()
			r := fmt.Sprintf("%s~%s", hash1, hash2)
			out <- r
		}(dataStr, md5)
	}

	wg.Wait()
	close(out)
}

func MultiHash(in, out chan any) {
	wg := &sync.WaitGroup{}

	for data := range in {
		wg.Add(1)
		go func(data any) {
			defer wg.Done()

			hashes := make([]string, 6)
			hashWg := &sync.WaitGroup{}

			for i := 0; i <= 5; i++ {
				hashWg.Add(1)
				go func(num int) {
					defer hashWg.Done()
					s := fmt.Sprintf("%d%s", num, data)
					r := DataSignerCrc32(s)

					hashes[num] = r
				}(i)
			}

			hashWg.Wait()
			r := strings.Join(hashes, "")
			out <- r
		}(data)

	}

	wg.Wait()
	close(out)
}

func CombineResults(in, out chan any) {
	results := make([]string, 0)

	for data := range in {
		results = append(results, data.(string))
	}

	if len(results) == 0 {
		return
	}
	sort.Strings(results)

	var sb strings.Builder
	for i := 0; i < len(results)-1; i++ {
		sb.WriteString(results[i])
		sb.WriteString("_")
	}
	sb.WriteString(results[len(results)-1])

	out <- sb.String()
	close(out)
}
