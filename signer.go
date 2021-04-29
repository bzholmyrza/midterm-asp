package main

import (
	"sort"
	"strconv"
	"strings"
	"sync"
)

// сюда писать код
func ExecutePipeline(jobs ...job) {
	in := make(chan interface{})
	wg := &sync.WaitGroup{}

	for _, j := range jobs {
		out := make(chan interface{})
		wg.Add(1)
		go func(j job, in, out chan interface{}) {
			defer wg.Done()
			defer close(out)
			j(in, out)
		}(j, in, out)
		in = out
	}
	wg.Wait()
}

func SingleHash(in, out chan interface{}) {
	mu := &sync.Mutex{}
	wg := &sync.WaitGroup{}
	for data := range in {
		wg.Add(1)
		go singleHash(data, out, mu, wg)
	}
	wg.Wait()
}

func singleHash(data interface{}, out chan interface{}, mu *sync.Mutex, wg *sync.WaitGroup)  {
	defer wg.Done()
	mu.Lock()
	md5 := DataSignerMd5(strconv.Itoa(data.(int)))
	mu.Unlock()

	hashCh := make(chan string)
	go func(output chan string, data interface{}) {
		output <- DataSignerCrc32(strconv.Itoa(data.(int)))
	}(hashCh, data)
	md5crc32 := DataSignerCrc32(md5)
	crc32 := <-hashCh
	out <- crc32 + "~" + md5crc32
}

func MultiHash(in, out chan interface{}) {
	wg := &sync.WaitGroup{}
	for data := range in {
		wg.Add(1)
		go multiHash(data, out, wg)
	}
	wg.Wait()
}

func multiHash(data interface{}, out chan interface{}, wg *sync.WaitGroup) {
	defer wg.Done()
	var arr = make([]string, 6)
	localWG:= &sync.WaitGroup{}
	for i := 0; i < 6; i++ {
		localWG.Add(1)
		go func(i int, localWG *sync.WaitGroup, data string) {
			defer localWG.Done()
			arr[i] = DataSignerCrc32(strconv.Itoa(i) + data)
		}(i, localWG, data.(string))
	}
	localWG.Wait()
	out <- strings.Join(arr, "")
}

func CombineResults(in, out chan interface{}) {
	var arr []string
	for data := range in {
		arr = append(arr, data.(string))
	}
	sort.Strings(arr)
	out <- strings.Join(arr, "_")
}