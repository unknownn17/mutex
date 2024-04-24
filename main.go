package main

import (
	"fmt"
)

type SafeMap struct {
	requestChan chan mapRequest
}

type mapRequest struct {
	action string
	key    string
	value  string
	result chan<- mapResult
}

type mapResult struct {
	value string
	found bool
}

func NewSafeMap() *SafeMap {
	sm := &SafeMap{
		requestChan: make(chan mapRequest),
	}
	go sm.mapManager()
	return sm
}

func (sm *SafeMap) take(key string) (string, bool) {
	resultChan := make(chan mapResult)
	sm.requestChan <- mapRequest{action: "get", key: key, result: resultChan}
	result := <-resultChan
	return result.value, result.found
}

func (sm *SafeMap) S(key, value string) {
	sm.requestChan <- mapRequest{action: "set", key: key, value: value}
}

func (sm *SafeMap) remove(key string) {
	sm.requestChan <- mapRequest{action: "delete", key: key}
}

func (sm *SafeMap) mapManager() {
	items := make(map[string]string)
	for req := range sm.requestChan {
		switch req.action {
		case "get":
			value, found := items[req.key]
			req.result <- mapResult{value: value, found: found}
		case "set":
			items[req.key] = req.value
		case "delete":
			delete(items, req.key)
		}
	}
}

func main() {

	safeMap := NewSafeMap()

	for i := 0; i < 10; i++ {
		go func(i int) {
			key := fmt.Sprintf("key%d", i)
			value := fmt.Sprintf("value%d", i)

			safeMap.S(key, value)

			retrievedValue, found := safeMap.take(key)
			if found {
				fmt.Printf("Goroutine %d: Retrieved value: %s\n", i, retrievedValue)
			} else {
				fmt.Printf("Goroutine %d: Value not found\n", i)
			}

			safeMap.remove(key)
			fmt.Printf("Goroutine %d: Value deleted\n", i)
		}(i)
	}
	fmt.Scanln()
}
