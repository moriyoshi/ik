package task

type NotCompletedStatus struct {}

var NotCompleted = &NotCompletedStatus {}

func (_ *NotCompletedStatus) Error() string { return "" }

type PanickedStatus struct {}

var Panicked = &PanickedStatus {}

func (_ *PanickedStatus) Error() string { return "" }

type TaskStatus interface {
	Status() error
	Result() interface {}
	Poll()
}

type TaskRunner interface {
	Run(func () (interface {}, error)) (TaskStatus, error)
}

