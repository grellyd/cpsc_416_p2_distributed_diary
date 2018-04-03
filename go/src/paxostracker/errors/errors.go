package errors

type BadTransition string

func (e BadTransition) Error() string {
	return "bad attempting transition"
}

type UnknownTransition string

func (e UnknownTransition) Error() string {
	return "unknown transition"
}
