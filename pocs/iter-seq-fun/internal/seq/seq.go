package seq

import "iter"

func Map[A, B any](s iter.Seq[A], f func(A) B) iter.Seq[B] {
	return func(yield func(B) bool) {
		for v := range s {
			if !yield(f(v)) {
				return
			}
		}
	}
}

func Filter[T any](s iter.Seq[T], keep func(T) bool) iter.Seq[T] {
	return func(yield func(T) bool) {
		for v := range s {
			if keep(v) && !yield(v) {
				return
			}
		}
	}
}

func Take[T any](s iter.Seq[T], n int) iter.Seq[T] {
	return func(yield func(T) bool) {
		if n <= 0 {
			return
		}
		taken := 0
		for v := range s {
			if !yield(v) {
				return
			}
			taken++
			if taken == n {
				return
			}
		}
	}
}

func From(start int) iter.Seq[int] {
	return func(yield func(int) bool) {
		for i := start; ; i++ {
			if !yield(i) {
				return
			}
		}
	}
}

func Zip[A, B any](a iter.Seq[A], b iter.Seq[B]) iter.Seq2[A, B] {
	return func(yield func(A, B) bool) {
		nextA, stopA := iter.Pull(a)
		defer stopA()
		nextB, stopB := iter.Pull(b)
		defer stopB()
		for {
			va, ok := nextA()
			if !ok {
				return
			}
			vb, ok := nextB()
			if !ok {
				return
			}
			if !yield(va, vb) {
				return
			}
		}
	}
}
