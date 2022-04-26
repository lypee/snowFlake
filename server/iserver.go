package server

type IServer interface {
	GetWorkerId() (id int, err error)
}
