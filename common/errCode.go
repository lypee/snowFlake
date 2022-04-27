package common

import "encoding/json"

type Err struct {
	Code    int
	Msg     string
	TrueErr error
}

func NewErr(code int, msg string) Err {
	return Err{
		Code: code,
		Msg:  msg,
	}
}

func (e Err) Error() string {
	err, _ := json.Marshal(e)
	return string(err)
}

func (e *Err) WithTrueErr(err error) error {
	e.TrueErr = err
	return e
}

var (
	OpErr          Err = Err{Code: 10000, Msg: "OpErr"}
	ConnErr        Err = Err{Code: 10001, Msg: "ConnErr"}
	StartConnErr   Err = Err{Code: 10002, Msg: "StartConnErr"}
	ServersErr     Err = Err{Code: 10003, Msg: "ServersErr"}
	InvalidPathErr Err = Err{Code: 10003, Msg: "InvalidPathErr"}
	NodeNameErr    Err = Err{Code: 10004, Msg: "NodeNameErr"}
	PathLengthErr  Err = Err{Code: 10005, Msg: "PathLengthErr"}
)
