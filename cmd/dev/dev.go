package main

// This file is just something I used while developing.. It's kind of an
// example of usage but currently not a very clean one.

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
	log.Printf("DOWNLOAD %v", cc)
	err := cc.Download()
	if err != nil {
		panic(err)
	}

	if cc.IsDownloaded() {
		conn := pool.Get()
		defer conn.Close()
		log.Printf("IMPORT %v", cc)
		start := time.Now()
		n, err := cc.Import(conn)
		if err != nil {
			panic(err)
		}
		log.Printf("Imported %d rows from %s in %s",
			n, filepath.Base(cc.Path()), time.Since(start))
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
		res, err := q.Current(i)
		if err != nil {
			panic(err)
		}
		log.Printf("current   : %s: %s ", i, res)
		if res != nil {
			cur, err := q2.Current(res.ASN)
			if err != nil {
				panic(err)
			}
			log.Printf("desc      : %s", cur)
		}
		log.Printf("allhistory: %s", i)
		allhist, err := q.AllHistory(i)
		if err != nil {
			panic(err)
		}

		for _, v := range allhist {
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
			log.Printf("DOWNLOAD %s", b.Path())
			err := b.Download()
			if err != nil {
				panic(err)
			}
			if b.IsDownloaded() {
				log.Printf("IMPORT %s", b.Path())
				conn := pool.Get()
				defer conn.Close()
				start := time.Now()
				n, err := b.Import(conn)
				if err != nil {
					panic(err)
				}
				log.Printf("Imported %d rows from %s in %s",
					n, filepath.Base(b.Path()), time.Since(start))

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
