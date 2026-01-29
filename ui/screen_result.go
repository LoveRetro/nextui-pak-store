package ui

// ScreenResult wraps a screen's output value with an action.
type ScreenResult[T any] struct {
	Value  T
	Action Action
}

func success[T any](value T) ScreenResult[T] {
	return ScreenResult[T]{
		Value:  value,
		Action: ActionSelected,
	}
}

func back[T any](value T) ScreenResult[T] {
	return ScreenResult[T]{
		Value:  value,
		Action: ActionBack,
	}
}

func withAction[T any](value T, action Action) ScreenResult[T] {
	return ScreenResult[T]{
		Value:  value,
		Action: action,
	}
}
