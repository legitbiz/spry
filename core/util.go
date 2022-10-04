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

func Contains[T any, I comparable](
	t []T,
	q T,
	identity func(item T) I) bool {
	target := identity(q)
	for _, x := range t {
		check := identity(x)
		if check == target {
			return true
		}
	}
	return false
}
