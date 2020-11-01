package disruptor

import "fmt"

// 100W条445字节的数据，大概占用550M内存，gzip可以减少30%的内存。

const (
	dataField = "data"
)

type Marshaler interface {
	Marshal() ([]byte, error)
	Unmarshal(data []byte) error
}

type Handler func(m Message) error

// Message from queue
type Message struct {
	ID     string
	Stream string
	Group  string
	Body   []byte
}

type ErrorNotifier interface {
	EmitError(error)
}

type Consumer interface {
	Pop(data Marshaler, h Handler) (error, bool)
	Close()
}

type Producer interface {
	Close()
	Push(data Marshaler) error
}

func makeStreamName(name string, shard int) string {
	return fmt.Sprintf("disruptor:%v:{%v}", name, shard)
}

func makeGroupName(name string) string {
	return fmt.Sprintf("disruptor_%v_group", name)
}
