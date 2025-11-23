package middlewares

import (
	"net"
	"net/http"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var ipVisitors = make(map[string]*visitor)
var userVisitors = make(map[string]*visitor)
var mu sync.Mutex

func getLimiter(key string, isUser bool) *rate.Limiter {
	mu.Lock()
	defer mu.Unlock()

	var v *visitor
	var exists bool

	if isUser {
		v, exists = userVisitors[key]
	} else {
		v, exists = ipVisitors[key]
	}

	if !exists {
		limiter := rate.NewLimiter(3, 5)
		v = &visitor{limiter, time.Now()}
		if isUser {
			userVisitors[key] = v
		} else {
			ipVisitors[key] = v
		}
	}

	v.lastSeen = time.Now()

	return v.limiter
}

func CleanupVisitors() {
	for {
		time.Sleep(time.Minute)

		mu.Lock()
		for ip, v := range ipVisitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(ipVisitors, ip)
			}
		}
		for userID, v := range userVisitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(userVisitors, userID)
			}
		}
		mu.Unlock()
	}
}

func RateLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var limiter *rate.Limiter

		userID := r.Context().Value("userID")
		if userID != nil {
			limiter = getLimiter(userID.(string), true)
		} else {
			ip, _, err := net.SplitHostPort(r.RemoteAddr)
			if err != nil {
				http.Error(w, "Internal Server Error", http.StatusInternalServerError)
				return
			}
			limiter = getLimiter(ip, false)
		}

		if !limiter.Allow() {
			http.Error(w, http.StatusText(429), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
