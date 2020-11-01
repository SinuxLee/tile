package disruptor

import (
	"github.com/go-redis/redis/v7"
)

type client struct {
	streamName  string                // 队列名称
	shardsCount int8                  // 队列分片数量
	redisClient redis.UniversalClient // 抽象客户端连接
}

func newClient(stream string, shard int8, cli redis.UniversalClient) (*client, error) {
	c := &client{
		streamName:  stream,
		shardsCount: shard,
		redisClient: cli,
	}

	err := c.init()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *client) init() error {
	group := makeGroupName(c.streamName)

	for i := 0; i < int(c.shardsCount); i++ {
		stream := makeStreamName(c.streamName, i)
		err := c.createShard(stream, group)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *client) createShard(stream string, group string) error {
	xinfo := redis.NewCmd("XINFO", "STREAM", stream)

	err := c.redisClient.Process(xinfo)
	if err != nil {
		xgroup := redis.NewCmd("XGROUP", "CREATE", stream, group, "$", "MKSTREAM")

		err := c.redisClient.Process(xgroup)
		if err != nil {
			return err
		}
	}

	// Check after creation
	err = c.redisClient.Process(xinfo)
	if err != nil {
		return err
	}

	return nil
}
