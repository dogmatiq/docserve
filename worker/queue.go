package worker

type queue[T any] struct {
	elements []T
	empty    []T
}

func (q *queue[T]) Push(m T) {
	q.elements = append(q.elements, m)
	q.empty = q.elements[:0]
}

func (q *queue[T]) Pop() {
	switch len(q.elements) {
	case 0:
	case 1:
		q.elements = q.empty
	default:
		q.elements = q.elements[1:]
	}
}

func (q *queue[T]) Head() (T, bool) {
	if len(q.elements) == 0 {
		var zero T
		return zero, false
	}
	return q.elements[0], true
}
