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
	// 每秒限流10个  (100ms)
	limitOpt := rate.Every(100 * time.Millisecond)
	limit := rate.NewLimiter(limitOpt, 10)
	// 修改 NewLimiter(r, b) 的 b 为 1 10 100 后观察运行结果, 设置为 0 会导致全部拒绝
	wg := sync.WaitGroup{}
	count := struct {
		sync.Mutex
		pass uint
		reject uint
	}{}
	for i:=0;i<100;i++ {
		wg.Add(1)
		// 每 10ms 发送一个请求 ,共100次, 一秒内发完请求
		time.Sleep(10 * time.Millisecond)
		go func(order int) {
			var i struct {};_=i // 不要获取i, i此时是100
			defer wg.Done()
			count.Lock()
			defer count.Unlock()
			if limit.Allow() {
				count.pass++
				log.Print("pass")
			} else {
				count.reject++
				log.Print("reject")
			}
		}(i)
	}
	wg.Wait()
	log.Print("pass count:", count.pass)
	log.Print("reject count:", count.reject)
}

// 先快速消耗10个,然后等2秒再次消耗15个,
func TestTimeRateAllow2(t *testing.T) {
	limitOpt := rate.Every(100 * time.Millisecond)
	limit := rate.NewLimiter(limitOpt, 10)
	// 修改 NewLimiter(r, b) 的 b 为 1 10 100 后观察运行结果, 设置为 0 会导致全部拒绝
	wg := sync.WaitGroup{}
	use := func (n int) {
		for i:=0;i<n;i++ {
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
	}
	// 因为初始化时 b = 10,所以第一次10个会快速消耗,这里消耗的是第一秒的10
	use(10)
	time.Sleep(time.Second*2)
	log.Print("sleep")
	// 等待2秒后桶中又 因为limitOpt 又有了10个, 第二次只会消耗10个 5个失败
	use(15)
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
