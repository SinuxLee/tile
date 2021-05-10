package disruptor

import (
	"sync"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/imdario/mergo"
	"github.com/ssgreg/repeat"
)

var defaultProducerOptions = ProducerOptions{
	ShardsCount:       10,                     // 分片队列的个数
	PendingBufferSize: 50000,                  // 本地消息缓存队列
	PipeBufferSize:    100,                    // 每个分片队列 pipeline 个数
	PipePeriod:        100 * time.Millisecond, // 分片队列每次发送前等待的最长时间
}

type ProducerOptions struct {
	QueueName         string
	ShardsCount       int8
	PendingBufferSize int64         // 本地消息缓冲的大小
	PipeBufferSize    int64         // 每次批量发送的数量
	PipePeriod        time.Duration // 批量发送数据的时间间隔
	ErrorNotifier     ErrorNotifier
}

type producer struct {
	*client
	*ProducerOptions
	msgChan chan []byte // 本地的消息缓冲
	emitter ErrorNotifier
	wg      *sync.WaitGroup
}

func NewProducer(opt *ProducerOptions, rdsCli redis.UniversalClient) (Producer, error) {
	if err := mergo.Merge(opt, defaultProducerOptions); err != nil {
		return nil, err
	}

	cli, err := newClient(opt.QueueName, opt.ShardsCount, rdsCli)
	if err != nil {
		return nil, err
	}

	cache := make(chan []byte, opt.PendingBufferSize)
	pr := &producer{
		msgChan:         cache,
		client:          cli,
		emitter:         opt.ErrorNotifier,
		ProducerOptions: opt,
		wg:              &sync.WaitGroup{},
	}

	pr.wg.Add(1)
	go pr.produce()

	return pr, nil
}

func (p *producer) Close() {
	close(p.msgChan)
	p.wg.Wait()
}

func (p *producer) Push(data Marshaler) error {
	d, err := data.Marshal()
	if err != nil {
		return err
	}

	p.msgChan <- d
	return nil
}

func (p *producer) produce() {
	shard := 0
	idx := 0
	isRunning := true
	doSend := false

	buf := make([][]byte, p.PipeBufferSize)
	tick := time.NewTicker(p.PipePeriod)
	started := time.Now()
	for isRunning {
		select {
		case msg, more := <-p.msgChan:
			if !more {
				isRunning = false
			} else {
				buf[idx] = msg
				idx++
			}
		case <-tick.C:
		}

		if isRunning {
			switch {
			case idx == int(p.PipeBufferSize):
				doSend = true
			case time.Since(started) >= p.PipePeriod && len(p.msgChan) == 0:
				doSend = true
			default:
				doSend = false
			}

			if !doSend {
				continue
			}
		}

		p.sendWithLock(shard, buf[0:idx])

		idx = 0
		shard++
		shard %= int(p.ShardsCount)
		started = time.Now()
	}

	p.wg.Done()
}

func (p *producer) sendWithLock(shard int, buf [][]byte) {
	if len(buf) == 0 {
		return
	}

	args := make([]redis.XAddArgs, len(buf))
	stream := makeStreamName(p.QueueName, shard)

	for i, m := range buf {
		args[i] = redis.XAddArgs{
			ID:     "*",
			Stream: stream,
			Values: map[string]interface{}{dataField: m},
		}
	}

	p.wg.Add(1)
	go func() {
		p.pipelineTransfer(args)
		p.wg.Done()
	}()
}

func (p *producer) pipelineTransfer(args []redis.XAddArgs) {
	err := repeat.Repeat(
		repeat.Fn(func() error {
			pipe := p.redisClient.TxPipeline()

			for _, m := range args {
				pipe.XAdd(&m)
			}

			_, err := pipe.Exec()
			if err != nil {
				if p.emitter != nil {
					p.emitter.EmitError(err)
				}

				return repeat.HintTemporary(err)
			}

			return nil
		}),
		repeat.StopOnSuccess(),
		repeat.WithDelay(repeat.FullJitterBackoff(500*time.Millisecond).Set()),
	)

	if err != nil && p.emitter != nil {
		p.emitter.EmitError(err)
	}
}
