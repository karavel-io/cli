package utils

func RunAsync[T any](function func() T) <-chan T {
	chn := make(chan T)
	go func() {
		chn <- function()
	}()
	return chn
}
