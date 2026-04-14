package service

var fetchSemaphore chan struct{}

func InitSemaphore(maxConcurrent int) {
	fetchSemaphore = make(chan struct{}, maxConcurrent)
}
