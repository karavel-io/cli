package utils

// BindParam is a simple wrapper to bind parameters to functions.
// This is a weak implementation since it only works with single parameters!
func BindParam[T any, R any](fun func(T) R, value T) func() R {
	return func() R {
		return fun(value)
	}
}
