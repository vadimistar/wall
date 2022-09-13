package wall

import "fmt"

type Error struct {
	pos Pos
	msg string
}

func NewError(pos Pos, format string, args ...interface{}) Error {
	return Error{
		pos: pos,
		msg: fmt.Sprintf(format, args...),
	}
}

func (e Error) Error() string {
	return fmt.Sprintf("%s: error: %s", e.pos, e.msg)
}
