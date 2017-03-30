package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"sync"
	"time"

	"encoding/json"

	"github.com/thomasf/internet"
)

func main() {

	for _, b := range []internet.BGPDump{
		// {Date: time.Date(2009, 03, 10, 0, 0, 0, 0, time.UTC)},
		{Date: time.Date(2014, 03, 10, 0, 0, 0, 0, time.UTC)},
		// {Date: time.Date(2015, 04, 18, 0, 0, 0, 0, time.UTC)},
		// {Date: time.Date(2015, 01, 01, 0, 0, 0, 0, time.UTC)},
		// {Date: time.Date(2015, 01, 02, 0, 0, 0, 0, time.UTC)},
		// {Date: time.Date(2012, 03, 10, 0, 0, 0, 0, time.UTC)},
		// {Date: time.Date(2010, 03, 04, 0, 0, 0, 0, time.UTC)},
		// {Date: time.Date(2009, 03, 11, 0, 0, 0, 0, time.UTC)},
	} {
		pp := b.ParsedPath()
		b, err := ioutil.ReadFile(pp)
		if err != nil {
			panic(err)
		}

		var entries map[string]uint32
		log.Printf("reading %s", pp)
		json.Unmarshal(b, &entries)
		log.Println("read!")
		db := IP2ASNDB{
			entries: entries,
		}
		var wg sync.WaitGroup
		start := time.Now()
		nq := 10000
		nw := 3
		for wn := 0; wn < nw; wn++ {
			wg.Add(1)
			go func(wn int) {
				log.Println("start", wn)
				defer wg.Done()
				for n := 0; n < nq; n++ {
				for _, i := range []string{"8.8.8.8", "5.150.255.150", "127.0.0.1"} {
					res, err := db.Current(i)
					if err != nil {
						panic(err)
					}
					_ = res
					// log.Printf("current   : %s: %s ", i, res)

				}}
				log.Println("done", wn)
			}(wn)

		}
		wg.Wait()
		end := time.Now().Sub(start)
		log.Printf("%d queries in %v, %f quries/s", nw*nq , end, float64(nw*nq)/end.Seconds())

	}

}

// RIPE-NCC-RIS BGP IPv6 Anchor Prefix @RRC00
// RIPE-NCC-RIS BGP Anchor Prefix @ rrc00 - RIPE NCC
var (
	asn12654blocks = map[string]bool{
		"2001:7fb:ff00::/48": true,
		"84.205.80.0/24":     true,
		"2001:7fb:fe00::/48": true,
		"84.205.64.0/24":     true,
	}
	netmasks []net.IPMask
)

func init() {
	for i := 0; i < 8*net.IPv4len; i++ {
		netmasks = append(netmasks, net.CIDRMask(8*net.IPv4len-i, 8*net.IPv4len))
	}
}

// ASNResult contains the lookup results from resolving an IP address to a ASN.
type ASNResult struct {
	Mask net.IPNet
	ASN  uint32
	Date time.Time // Recorded time for when
}

// IP2ASNDB is the query client.
type IP2ASNDB struct {
	sync.RWMutex
	entries map[string]uint32
}

// Current returns the latest known result for an IP2ASN lookup.
func (i *IP2ASNDB) Current(IP string) (*ASNResult, error) {
	ip, err := i.parseIP(IP)
	if err != nil {
		return &ASNResult{}, err
	}
	i.RLock()
	defer i.RUnlock()

	keys := i.keys(ip)
	for _, k := range keys {
		if v, ok := i.entries[k.String()]; ok {
			return &ASNResult{
				Mask: k,
				ASN:  v,
			}, nil

		}
	}
	return nil, nil
}

func (i *IP2ASNDB) parseIP(IP string) (net.IP, error) {
	I := net.ParseIP(IP)
	if I == nil {
		return nil, net.InvalidAddrError(IP)
	}
	return I, nil
}

// // dates resolves IP2ASN for all date entries, if available.
// func (i *IP2ASNDB) find(IP net.IP) (ASNResult, error) {
// 	keys := i.keys(IP)
// 	}
// 	for _, k := range keys {
// 		i.conn.Send("HGET", "i2a:"+k.String(), d)
// 	}

// 	i.conn.Flush()
// 	var results []ASNResult
// 	for _, date := range dates {

// 		var found bool

// 		for idx := 0; idx < len(keys); idx++ {
// 			if !found {
// 				r, err := redis.String(i.conn.Receive())
// 				if err != nil {
// 					if err == redis.ErrNil {
// 						continue
// 					}
// 					return []ASNResult{}, err
// 				}

// 				timedate, err := time.Parse("20060102", date)
// 				asn, err := strconv.Atoi(r)
// 				if err != nil {
// 					// redis data error
// 					return []ASNResult{}, err
// 				}
// 				results = append(results, ASNResult{
// 					Mask: keys[idx],
// 					ASN:  asn,
// 					Date: timedate,
// 				})
// 				found = true
// 			} else {
// 				_, _ = i.conn.Receive()
// 			}
// 		}

// 	}
// 	return results, nil
// }

func (i *IP2ASNDB) keys(IP net.IP) []net.IPNet {
	var keys []net.IPNet
	for _, n := range netmasks {
		ipn := net.IPNet{
			IP:   IP.Mask(n),
			Mask: n,
		}
		keys = append(keys, ipn)
	}
	return keys
}

// Stringer.
func (a *ASNResult) String() string {
	return fmt.Sprintf("%s %d %s",
		a.Date.Format("2006-01-02"),
		a.ASN,
		a.Mask.String())
}
