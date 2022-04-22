package log

import (
	"context"
	"fmt"
	"sync"
	"time"

	proto "github.com/cleverlog/api/proto/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Client struct {
	grpcClient proto.LogServiceClient

	buf   []*Err
	bufMu sync.Mutex

	pollContext  context.Context
	pollCancel   func()
	pollGroup    sync.WaitGroup
	pollDeadline <-chan time.Time
}

func NewClient() *Client {
	viper.New()

	viper.AutomaticEnv()

	viper.SetDefault("SERVER_HOST", "cleverapi")
	viper.SetDefault("SERVER_PORT", "5555")

	conn, err := grpc.Dial(fmt.Sprintf("%s:%s",
		viper.GetString("SERVER_HOST"),
		viper.GetString("SERVER_PORT")), grpc.WithInsecure())
	if err != nil {
		logrus.Error(err)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	client := &Client{
		grpcClient:   proto.NewLogServiceClient(conn),
		bufMu:        sync.Mutex{},
		pollContext:  ctx,
		pollCancel:   cancel,
		pollGroup:    sync.WaitGroup{},
		pollDeadline: time.After(time.Second * 10),
	}

	client.pollGroup.Add(1)
	go client.pollBuf()

	return client
}

func (c *Client) Send(err *Err) {
	c.bufMu.Lock()

	c.buf = append(c.buf, err)

	c.bufMu.Unlock()
}

func (c *Client) sendBuf() error {
	c.bufMu.Lock()

	if len(c.buf) == 0 {
		c.bufMu.Unlock()
		return nil
	}

	protoLogs := c.toProto(c.buf)
	if _, err := c.grpcClient.SendLogs(context.Background(), protoLogs); err != nil {
		return err
	}

	c.buf = c.buf[:0]

	c.bufMu.Unlock()

	return nil
}

func (c *Client) toProto(logs []*Err) *proto.Logs {
	protoLogs := &proto.Logs{}

	for _, log := range logs {
		protoLogs.Logs = append(protoLogs.Logs, &proto.Log{
			Service:   log.ServiceName,
			Level:     proto.Log_LogLevel(log.Type),
			SpanId:    log.SpanID.String(),
			Timestamp: timestamppb.New(time.Now()),
			Source:    log.Source,
			Message:   log.Message,
		})
	}

	return protoLogs
}

func (c *Client) pollBuf() {
	for {
		select {
		case <-c.pollContext.Done():
			c.pollGroup.Done()
			break
		case <-c.pollDeadline:
			if err := c.sendBuf(); err != nil {
				logrus.Error(err)
			}

			c.pollDeadline = time.After(time.Second * 10)
		}
	}
}

func (c *Client) shutdown() {
	c.pollCancel()
	c.pollGroup.Wait()
}
