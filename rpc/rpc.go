package rpc 

type Task struct {
	TaskName string
	Done bool 
}


type Args struct {
	FinishedTask Task
}

type Reply struct {
	Task *Task
}