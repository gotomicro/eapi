package analyzer

type Stack[T any] []T

func (s *Stack[T]) Push(v T) {
	*s = append(*s, v)
}

func (s *Stack[T]) Pop() *T {
	if len(*s) == 0 {
		return nil
	}

	res := (*s)[len(*s)-1]
	*s = (*s)[:len(*s)-1]
	return &res
}
