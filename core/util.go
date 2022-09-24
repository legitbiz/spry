package core

func Mapper[T any, U any](t []T, m func(T) U) []U {
	u := make([]U, len(t))
	for i, x := range t {
		u[i] = m(x)
	}
	return u
}

func Reducer[T any, A any](
	t []T,
	r func(A, T, int) A,
	s A) A {
	a := s
	for i, x := range t {
		a = r(a, x, i)
	}
	return a
}
