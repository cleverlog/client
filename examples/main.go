package main

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/cleverlog/client/log"
)

func main() {
	log.SetServiceName("New service")

	for i := 0; i < 10; i++ {
		ctx := context.WithValue(context.Background(), log.SpanKeyName, uuid.New())

		log.Error(ctx, "first error")
		log.Info(ctx, "sec error")
		log.Debug(ctx, "third error")
		log.Info(ctx, "fourth error")
		log.Error(ctx, "fifth error")
		log.Info(ctx, "sixth error")
	}

	time.Sleep(time.Minute)
}
