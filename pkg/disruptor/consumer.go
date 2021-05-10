package disruptor

import (
	"errors"
	"sync"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/imdario/mergo"
	"github.com/ssgreg/repeat"
)

var defaultConsumerOptions = ConsumerOptions{
	ShardsCount:       10,                     // 分片队列的个数
	PrefetchCount:     100,                    // 每次从分片队列取消息的个数
	PendingBufferSize: 1000,                   // 本地消息缓存队列大小，需要 >= ShardsCount*PrefetchCount
	PipeBufferSize:    100,                    // 每个分片队列 ack pipeline 的个数
	PipePeriod:        100 * time.Millisecond, // ack分片队列等待的最长时间
}

type ConsumerOptions struct {
	QueueName         string
	Consumer          string
	ShardsCount       int8
	PrefetchCount     int64         // 每次从队列中读取的消息数量
	Block             time.Duration // 读取队列数据时阻塞的时长
	PendingBufferSize int64         // 本地缓冲队列长度
	PipeBufferSize    int64         // 每次批量ack的数量
	PipePeriod        time.Duration // 每次ack的时间间隔
	ErrorNotifier     ErrorNotifier
}

type consumer struct {
	*client
	*ConsumerOptions
	ackChan     chan Message
	msgChan     chan Message
	stopped     bool
	isConsuming bool
	emitter     ErrorNotifier
	wgAck       *sync.WaitGroup
	wgCons      *sync.WaitGroup
}

func NewConsumer(opt *ConsumerOptions, rdsCli redis.UniversalClient) (Consumer, error) {
	if err := mergo.Merge(opt, defaultConsumerOptions); err != nil {
		return nil, err
	}

	cli, err := newClient(opt.QueueName, opt.ShardsCount, rdsCli)
	if err != nil {
		return nil, err
	}

	msgChan := make(chan Message, opt.PendingBufferSize)
	ackChan := make(chan Message, opt.PendingBufferSize)
	cn := &consumer{
		ackChan:         ackChan,
		msgChan:         msgChan,
		client:          cli,
		emitter:         opt.ErrorNotifier,
		ConsumerOptions: opt,
		wgAck:           &sync.WaitGroup{},
		wgCons:          &sync.WaitGroup{},
		isConsuming:     true,
	}

	cn.wgAck.Add(1)
	go cn.processAck()

	cn.start()

	return cn, nil
}

func (c *consumer) start() {
	for i := 0; i < int(c.ShardsCount); i++ {
		c.wgCons.Add(1)
		go c.consume(i)
	}
}

func (c *consumer) Pop(data Marshaler, h Handler) (error, bool) {
	m, more := <-c.msgChan
	if !more {
		return nil, more
	}

	defer c.ack(m)

	err := data.Unmarshal(m.Body)
	if err != nil {
		return err, more
	}

	return h(m), more
}

func (c *consumer) Close() {
	c.isConsuming = false

	c.wgCons.Wait()
	close(c.msgChan)
	c.stopped = true

	close(c.ackChan)
	c.wgAck.Wait()
}

func (c *consumer) consume(shard int) {
	group := makeGroupName(c.QueueName)
	stream := makeStreamName(c.QueueName, shard)

	lastID := "0-0"
	checkBacklog := true

	// Millisecond is minimal for Redis
	block := time.Second * 1
	if c.Block != 0 {
		block = c.Block
	}

	for c.isConsuming {
		var id string
		if checkBacklog {
			id = lastID
		} else {
			id = ">"
		}

		var res []redis.XStream
		err := repeat.Repeat(
			repeat.Fn(func() error {
				var err error
				res, err = c.redisClient.XReadGroup(&redis.XReadGroupArgs{
					Block:    block,
					Consumer: c.Consumer,
					Count:    c.PrefetchCount,
					Group:    group,
					Streams:  []string{stream, id},
				}).Result()

				if err != nil && err != redis.Nil {
					if c.emitter != nil {
						c.emitter.EmitError(err)
					}

					return repeat.HintTemporary(err)
				}

				return nil
			}),
			repeat.StopOnSuccess(),
			repeat.WithDelay(repeat.FullJitterBackoff(500*time.Millisecond).Set()),
		)

		if err != nil {
			if c.emitter != nil {
				c.emitter.EmitError(err)
			}

			continue
		}

		if checkBacklog && (len(res) == 0 || len(res[0].Messages) == 0) {
			checkBacklog = false
			continue
		}

		for _, s := range res {
			for _, m := range s.Messages {
				lastID = m.ID

				msg := Message{
					Group:  group,
					ID:     m.ID,
					Stream: stream,
				}

				v, exist := m.Values[dataField]
				data, ok := v.(string)
				if !exist || !ok {
					if c.emitter != nil {
						c.emitter.EmitError(errors.New("Incorrect message format: no \"data\" field in message with id " + m.ID))
					}

					c.ack(msg)
					continue
				}

				msg.Body = []byte(data)
				c.msgChan <- msg
			}
		}
	}

	c.wgCons.Done()
}

func (c *consumer) ack(m Message) {
	c.ackChan <- m
}

func (c *consumer) processAck() {
	started := time.Now()
	tick := time.NewTicker(c.PipePeriod)

	toAck := make(map[string][]Message)
	cnt := make(map[string]int)

	for {
		var (
			doStop bool
			stream string
		)

		select {
		case m, more := <-c.ackChan:
			if !more {
				doStop = true
			} else {
				stream = m.Stream

				if toAck[stream] == nil {
					toAck[stream] = make([]Message, c.PipeBufferSize)
					cnt[stream] = 0
				}

				toAck[stream][cnt[stream]] = m
				cnt[stream]++
			}
		case <-tick.C:
		}

		if doStop {
			c.sendAckAllStreams(toAck, cnt)
			break
		}

		if cnt[stream] >= int(c.PipeBufferSize) {
			c.sendAckStreamWithLock(toAck[stream][0:cnt[stream]])
			cnt[stream] = 0
		} else if time.Since(started) >= c.PipePeriod && len(c.ackChan) == 0 {
			c.sendAckAllStreams(toAck, cnt)
			started = time.Now()
		}
	}

	c.wgAck.Done()
}

func (c *consumer) sendAckAllStreams(toAck map[string][]Message, cnt map[string]int) {
	for stream := range toAck {
		c.sendAckStreamWithLock(toAck[stream][0:cnt[stream]])
		cnt[stream] = 0
	}
}

func (c *consumer) sendAckStreamWithLock(lst []Message) {
	if len(lst) == 0 {
		return
	}

	ids := make([]string, len(lst))
	for i, m := range lst {
		ids[i] = m.ID
	}

	c.wgAck.Add(1)

	go func() {
		c.sendAckStream(lst[0].Stream, lst[0].Group, ids)
		c.wgAck.Done()
	}()
}

func (c *consumer) sendAckStream(stream string, group string, ids []string) {
	err := repeat.Repeat(
		repeat.Fn(func() error {
			pipe := c.redisClient.TxPipeline()

			pipe.XAck(stream, group, ids...)
			pipe.XDel(stream, ids...)

			_, err := pipe.Exec()

			if err != nil {
				if c.emitter != nil {
					c.emitter.EmitError(err)
				}

				return repeat.HintTemporary(err)
			}

			return nil
		}),
		repeat.StopOnSuccess(),
		repeat.WithDelay(repeat.FullJitterBackoff(500*time.Millisecond).Set()),
	)

	if err != nil && c.emitter != nil {
		c.emitter.EmitError(err)
	}
}
