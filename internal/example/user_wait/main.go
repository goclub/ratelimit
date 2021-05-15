package main

import (
	"errors"
	xtime "github.com/goclub/time"
	"github.com/google/uuid"
	"golang.org/x/time/rate"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

func responseError(err error, writer http.ResponseWriter) {
	writer.WriteHeader(500)
	log.Print(err)
}


// 为了便于演示,在内存模拟数据库
type DB struct{
	sync.Mutex
	data map[string]int64
}
var mockdb DB = DB{data: map[string]int64{}}

func main () {
	// 每秒处理 1 ,最多1令牌,目的是便于测试,时间业务场景下不会 qps 1的
	rl := rate.NewLimiter(1, 1)
	http.HandleFunc("/app", func(writer http.ResponseWriter, request *http.Request) {
		allow, err := checkTicker(writer, request) ; if err != nil {
			// 优雅降级,不返回错误,只记录
			log.Print(err)
		}
		if allow == false {
			reservation := rl.Reserve()
			delay := reservation.Delay()
			if delay == 0 {
				allow = true
			} else {
				query := url.Values{}
				// 生成 ticker 和 retryUnixMilli
				ticket := uuid.New()
				log.Print("delay", delay)
				retryUnixMilli := xtime.UnixMilli(time.Now().Add(delay))
				// 存储数据库
				mockdb.Lock()
				mockdb.data[ticket.String()] = retryUnixMilli
				mockdb.Unlock()
				// 让用户跳转到等待页面,带上 ticker 和跳转时间
				query.Add("ticket", ticket.String())
				// 千万要传递 时间戳让客户端判断,因为客户端的时间可能与服务器不一致
				query.Add("delayMilli", strconv.FormatInt(delay.Milliseconds(), 10))
				query.Add("retryURL", request.URL.Path + "?ticket="+ticket.String())
				http.Redirect(writer, request, "/wait?" + query.Encode(), 302)
				return
			}
		}
		if allow {
			html := `<html>
				<button id="btn" >fetch app 5 times and refresh</button>
				<button id="refresh" >refresh</button>
				<script>
document.getElementById("btn").onclick= function () {
for(i=0;i<5;i++){ fetch('/app') }
location.href = "/app" 
}
document.getElementById("refresh").onclick= function () { location.href = "/app" }
</script>
				</html>
			`
			_, err := writer.Write([]byte(html)); if err != nil {
				responseError(err, writer)
			}
			return
		}
	})
	http.HandleFunc("/wait", func(writer http.ResponseWriter, request *http.Request) {
		query :=  request.URL.Query()
		// 前端根据query参数 定时跳转
		_, err := writer.Write([]byte(`
		
		<script>
		var nowUnixMilli = new Date().getTime()
		var delayMilli = ` +query.Get("delayMilli") +`
		setTimeout(function(){
			location.href = "`+ query.Get("retryURL") +`"
		}, delayMilli)
		var waitSec = Math.floor((delayMilli)/1000)
			if (waitSec < 0) {
				waitSec = 0
			}
		document.write("wait "waitSec + "s");
document.write('<a href="`+ query.Get("retryURL") +`">jump</a>')
		</script>
		`)) ; if err != nil {
			responseError(err, writer)
    		return
		}
	})
	addr := ":8353"
	log.Print("http://localhost" + addr+ "/app")
	log.Print(http.ListenAndServe(addr, nil))
}

func checkTicker(writer http.ResponseWriter, request *http.Request) (allow bool, err error) {
	query := request.URL.Query()
	ticket := query.Get("ticket")
	if len(ticket) == 0 {
		return false, nil
	}
	nowUnixMilli := xtime.UnixMilli(time.Now())
	mockdb.Lock()
	defer mockdb.Unlock()
	dbRetryUnixMilli, has := mockdb.data[ticket]
	if has == false {
		return false, errors.New("ticket invalid or overdue")
	}
	// 读到了立即删除
	delete(mockdb.data, ticket)
	if dbRetryUnixMilli > nowUnixMilli {
		return false, errors.New("retry too early")
	}
	if dbRetryUnixMilli - nowUnixMilli > time.Duration(time.Second*10).Milliseconds(){
		return false, errors.New("retry too late")
	}
	return true, nil
}