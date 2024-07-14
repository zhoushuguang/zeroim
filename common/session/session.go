package session

import (
	"strconv"
	"strings"
)

type Session string

func NewSession(name, token string, id uint64) Session {
	if len(name) == 0 || len(token) == 0 {
		panic("name or token is empty")
	}
	idstr := strconv.FormatUint(id, 10)
	return Session(name + ":" + token + ":" + idstr)
}

func FromString(str string) Session {
	return Session(str)
}

func (s Session) Name() string {
	arr := strings.Split(string(s), ":")
	if len(arr) != 3 {
		panic("invalid session")
	}
	return arr[0]
}

func (s Session) Token() string {
	arr := strings.Split(string(s), ":")
	if len(arr) != 3 {
		panic("invalid session")
	}
	return arr[1]
}

func (s Session) Id() uint64 {
	arr := strings.Split(string(s), ":")
	if len(arr) != 3 {
		panic("invalid session")
	}
	id, err := strconv.ParseUint(arr[2], 10, 64)
	if err != nil {
		panic("invalid id")
	}
	return id
}

func (s Session) Info() (string, string, uint64) {
	arr := strings.Split(string(s), ":")
	if len(arr) != 3 {
		panic("invalid session")
	}
	id, err := strconv.ParseUint(arr[2], 10, 64)
	if err != nil {
		panic("invalid id")
	}
	return arr[0], arr[1], id
}

func (s Session) String() string {
	return string(s)
}
