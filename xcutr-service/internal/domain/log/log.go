package xcutrlog

type Log struct {
	msg string
}

func NewLog(msg string) *Log {
	return &Log{
		msg: msg,
	}
}

func (l *Log) Msg() string {
	return l.msg
}
