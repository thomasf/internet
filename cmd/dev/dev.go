package main

import (
	"flag"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
	"github.com/thomasf/internet"
)

func main() {
	flag.Parse()
	pool = newPool(*redisServer, "")

	mainCidrr()
	mainIP2ASN()
}

func mainCidrr() {
	cc := internet.CIDRReport{
		Date: time.Now(),
	}
	err := cc.Download()
	if err != nil {
		panic(err)
	}

	if cc.IsDownloaded() {
		conn := pool.Get()
		defer conn.Close()
		err := cc.Import(conn)
		if err != nil {
			panic(err)
		}

	}
}

func mainIP2ASN() {
	internet.SetDataDir(filepath.Join(os.TempDir(), "internet"))
	DoIndex()
	conn := pool.Get()
	defer conn.Close()
	internet.RefreshBGPDump(conn)
	DoLookup()
}

func DoLookup() {
	conn := pool.Get()
	defer conn.Close()
	q := internet.NewIP2ASNClient(conn)
	q2 := internet.NewASN2ASDescClient(conn)
	for _, i := range []string{"8.8.8.8", "5.150.255.150", "127.0.0.1"} {
		res := q.Current(i)
		log.Printf("current   : %s: %s ", i, res)
		if res != nil {
			log.Printf("desc      : %s", q2.Current(res.ASN))
		}
		log.Printf("allhistory: %s", i)
		for _, v := range q.AllHistory(i) {
			log.Println(v.String())
		}
	}
}

func DoIndex() {
	var wg sync.WaitGroup

	for _, b := range []internet.BGPDump{
		{Date: time.Date(2009, 03, 10, 0, 0, 0, 0, time.UTC)},
		{Date: time.Date(2014, 03, 10, 0, 0, 0, 0, time.UTC)},
		// {Date: time.Date(2015, 04, 18, 0, 0, 0, 0, time.UTC)},
		// {Date: time.Date(2015, 01, 01, 0, 0, 0, 0, time.UTC)},
		// {Date: time.Date(2015, 01, 02, 0, 0, 0, 0, time.UTC)},
		// {Date: time.Date(2012, 03, 10, 0, 0, 0, 0, time.UTC)},
		// {Date: time.Date(2010, 03, 04, 0, 0, 0, 0, time.UTC)},
		// {Date: time.Date(2009, 03, 11, 0, 0, 0, 0, time.UTC)},
	} {
		wg.Add(1)
		go func(b internet.BGPDump) {
			defer func() {
				wg.Done()
				if r := recover(); r != nil {
					log.Println("Recovered in f", r, b.Path())
				}
			}()
			// log.Printf("DOWNLOAD %s", b.Path())
			err := b.Download()
			if err != nil {
				panic(err)
			}
			if b.IsDownloaded() {
				// log.Printf("IMPORT %s", b.Path())
				conn := pool.Get()
				defer conn.Close()
				err := b.Import(conn)

				if err != nil {
					panic(err)
				}
			}
		}(b)
	}
	wg.Wait()
}

func newPool(server, password string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     10,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", server)
			if err != nil {
				return nil, err
			}
			// if _, err := c.Do("AUTH", password); err != nil {
			// c.Close()
			// return nil, err
			// }
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}

}

var (
	pool        *redis.Pool
	redisServer = flag.String("redisServer", ":28743", "")
	// redisPassword = flag.String("redisPassword", "", "")
)
