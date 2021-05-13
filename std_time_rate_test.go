package xratelimit_test

import (
	"context"
	rate "golang.org/x/time/rate"
	"log"
	"sync"
	"testing"
	"time"
)

// 超过限制的则拒绝
func TestTimeRateAllow(t *testing.T) {
	limitOpt := rate.Every(100 * time.Millisecond)
	limit := rate.NewLimiter(limitOpt, 10)
	wg := sync.WaitGroup{}
	for i:=0;i<100;i++ {
		wg.Add(1)
		go func(order int) {
			var i struct {};_=i // 不要获取i, i此时是100
			defer wg.Done()
			if limit.Allow() {
				log.Print("pass")
			} else {
				log.Print("reject")
			}
		}(i)
	}
	wg.Wait()
}

// 超过选择则等待
func TestTimeRateWait(t *testing.T) {
	limitOpt := rate.Every(100 * time.Millisecond)
	limit := rate.NewLimiter(limitOpt, 10)
	wg := sync.WaitGroup{}
	for i:=0;i<100;i++ {
		wg.Add(1)
		go func(order int) {
			var i struct {};_=i // 不要获取i, i此时是100
			defer wg.Done()
			ctx := context.Background()
			err := limit.Wait(ctx) ; if err != nil {
			    log.Print(err)
			    return
			}
			log.Print("pass", order)
		}(i)
	}
	wg.Wait()
}
// 根据 rate 返回的延迟重试时间进行重试
func TestTimeRateReserve(t *testing.T) {
	limitOpt := rate.Every(100 * time.Millisecond)
	limit := rate.NewLimiter(limitOpt, 10)
	wg := sync.WaitGroup{}
	for i:=0;i<100;i++ {
		wg.Add(1)
		go func(order int) {
			var i struct {};_=i // 不要获取i, i此时是100
			defer wg.Done()
			rese := limit.Reserve()
			delay := rese.Delay()
			if delay == 0 {
				log.Print("pass", order)
			} else {
				log.Print("delay time: ",  delay.String(), " order: ", order)
				time.Sleep(delay) // 如果存在 ctx 则需使用 time.NewTimer() 配合 select 处理 ctx.Done()
				log.Print("delay pass", order)
			}
		}(i)
	}
	wg.Wait()
}
