package redis

import (
	"encoding/json"
	"fmt"
	redisclient "github.com/garyburd/redigo/redis"
	"github.com/go-playground/log"
	assert "gopkg.in/go-playground/assert.v1"
	"net"
	"strings"
	"sync"
	"testing"
	"time"
)

func fetchLastListItemFromRedis() (string, error) {

	c, err := redisclient.DialTimeout("tcp", "127.0.0.1:6379", 0, 1*time.Second, 1*time.Second)
	if err != nil {
		fmt.Sprintf("[ERROR] Could not connect to Redis: %s\n", err.Error())
	} else {
		defer c.Close()
		_, err = c.Do("SELECT", "0")
		if err != nil {
			c.Close()
			fmt.Sprintf("[ERROR] Could not select Redis DB: %s\n", err.Error())
		}
		// Issue the command to push the entry onto the designated list
		res, err := c.Do("RPOP", "test-goplayground-log-redis")
		if err != nil {
			return "", err
		}

		str, err := redisclient.String(res, nil)
		if err != nil {
			return "", err
		}

		return str, nil
	}

}

func TestRedisLogger(t *testing.T) {

	rLog, err := redislogger.New(10, []string{"127.0.0.1:6379"})
	rLog.SetRedisList("test-goplayground-log-redis")

	assert.Equal(t, err, nil)

	log.RegisterHandler(rLog, log.AllLevels...)

	rLog.SetFormatter(func(e *log.Entry) string {
		return fmt.Sprintf("[%s] : %s", strings.ToUpper(e.Level.String()), e.Message)
	})

	e := &log.Entry{
		WG:        new(sync.WaitGroup),
		Level:     log.NoticeLevel,
		Message:   "This is a sample message",
		Timestamp: time.Now(),
	}

	log.HandleEntry(e)

	val, err := fetchLastListItemFromRedis()

	assert.Equal(t, err, nil)
	assert.Equal(t, hasString(conn, val, "[NOTICE] : This is a sample message")

}