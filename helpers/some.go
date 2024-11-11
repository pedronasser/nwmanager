package helpers

func Some[T any](v T) *T {
	return &v
}
