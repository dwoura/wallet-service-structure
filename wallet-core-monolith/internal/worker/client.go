package worker

import (
	"github.com/hibiken/asynq"
)

// Client 封装 Asynq Client
type Client struct {
	client *asynq.Client
}

// NewClient 初始化 Client
// addr: "localhost:6379"
func NewClient(addr string, password string, db int) *Client {
	c := asynq.NewClient(asynq.RedisClientOpt{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	return &Client{client: c}
}

// Enqueue 将任务推送到队列
func (c *Client) Enqueue(task *asynq.Task, opts ...asynq.Option) (*asynq.TaskInfo, error) {
	return c.client.Enqueue(task, opts...)
}

// Close 关闭客户端连接
func (c *Client) Close() error {
	return c.client.Close()
}
