package tests

type List struct {
	Owner string
	Name  string
	Tasks []Task
}

type Task struct {
	ListName    string
	Title       string
	Description string
	Complete    bool
}

type AddTask struct {
	ListName string
}
