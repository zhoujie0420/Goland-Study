package concurrency

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestConcurrencyGroup(t *testing.T) {
	cg := NewConcurrencyGroup()
	ctx := context.Background()
	cg.Go(quickOperate, ctx)
	cg.Go(errOperate, ctx)
	time.Sleep(time.Second)
	cg.Go(slowOperate, ctx)

	errs := cg.Wait()
	for _, err := range errs {
		fmt.Println(err)
	}
}

func quickOperate(ctx context.Context) error {
	// 模拟业务耗时操作
	time.Sleep(time.Second)
	return nil
}

func slowOperate(ctx context.Context) error {
	time.Sleep(5 * time.Second)
	return nil
}

func errOperate(ctx context.Context) error {
	return errors.New("test err")
}
