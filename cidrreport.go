package internet

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
	"golang.org/x/net/html"
)

type ASDescription struct {
	ASN         int
	Description string
	CountryCode string
}

type CIDRReport struct {
	Date time.Time
}

// Path returns the absolute path to the target archive dump download file.
func (b *CIDRReport) Path() string {
	return filepath.Join(
		b.dir(), fmt.Sprintf("cidr-report-%s.txt", b.Date.Format("20060102")))
}

func (b *CIDRReport) dir() string {
	return filepath.Join(
		dataDir, "cache", b.Date.Format("200601"))
}

func (b *CIDRReport) day() string {
	return b.Date.Format("20060102")
}

// IsDownloaded returns true if the CIDRReport archive is downloaded locally.
func (b *CIDRReport) IsDownloaded() bool {
	p := b.Path()
	if _, err := os.Stat(p); err == nil {
		return true
	}
	return false
}

// Import stores the contents of a downloaded BGP dump into a redis server.
func (b *CIDRReport) Import(conn redis.Conn) error {

	alreadyImported, err := redis.Bool(conn.Do("SISMEMBER", "asd:imported_dates", b.day()))
	if err != nil {
		return err
	}
	if alreadyImported {
		return nil
	}

	file, err := os.Open(b.Path())
	if err != nil {
		return err
	}
	n := 0
	day := b.day()
	err = parseReport(file, func(asd *ASDescription) {
		conn.Send("HSET", fmt.Sprintf("asd:%d", asd.ASN), day,
			fmt.Sprintf("%s, %s", asd.Description, asd.CountryCode))
		n++
		if n%10000 == 0 {
			err := conn.Flush()
			if err != nil {
				panic(err)
			}
		}
	})
	conn.Send("SADD", "asd:imported_dates", day)
	err = conn.Flush()
	if err != nil {
		panic(err)
	}
	return nil
}

// Download fetches http://www.cidr-report.org/as2.0/autnums.html and stores it
// in the cache. Currently one per day is fetched. Download returns early with
// no error if the file already is downloaded to disk.
func (b *CIDRReport) Download() error {
	dumpDir := b.dir()
	err := os.MkdirAll(dumpDir, 0777)
	if err != nil {
		return err
	}
	if b.IsDownloaded() {
		return nil
	}
	err = os.MkdirAll(filepath.Join(dataDir, "spool"), 0777)
	if err != nil {
		return err
	}
	tempFile, err := ioutil.TempFile(
		filepath.Join(dataDir, "spool"), b.day())
	if err != nil {
		return err
	}
	defer tempFile.Close()
	dlURL := "http://www.cidr-report.org/as2.0/autnums.html"

	resp, err := http.Get(dlURL)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("Got http status code %s response for %s", resp.Status, dlURL)
	}
	// log.Printf("Downloading %s\n", dlURL)
	defer resp.Body.Close()
	_, err = io.Copy(tempFile, resp.Body)
	if err != nil {
		return err
	}
	err = os.Rename(tempFile.Name(), b.Path())
	if err != nil {
		return err

	}
	return nil
}

// RefreshCIDRReport ensures that the latest dump available is the one which is installed.
func RefreshCIDRReport(conn redis.Conn) error {
	for _, b := range []CIDRReport{
		{Date: time.Now()},
		{Date: time.Now().Add(-time.Duration(time.Hour * 24))},
	} {
		err := b.Download()
		if err != nil {
			return err
		}
		if b.IsDownloaded() {
			if err := b.Import(conn); err != nil {
				return err
			}
			return nil
		}
	}
	return nil
}

func parseReport(r io.Reader, emitter func(*ASDescription)) error {
	z := html.NewTokenizer(r)
	n := 0
	depth := 0
	var asn *int
loop:
	for {
		tt := z.Next()
		switch tt {
		case html.ErrorToken:
			break loop
		case html.TextToken:
			if asn != nil {
				desc := strings.TrimSpace(string(z.Text()))
				ccpos := strings.LastIndex(desc, ",")
				if ccpos == -1 {
					return fmt.Errorf("Could not parse country code from %d %s", asn, desc)
				}
				emitter(&ASDescription{
					ASN:         *asn,
					Description: strings.TrimSpace(desc[:ccpos]),
					CountryCode: strings.TrimSpace(desc[ccpos+1:]),
				})
				n++
				asn = nil
			} else if depth > 0 {
				asnstr := strings.TrimPrefix(string(z.Text()), "AS")
				var err error
				i, err := strconv.Atoi(strings.TrimSpace(asnstr))
				if err != nil {
					return err
				}
				asn = &i
			}
		case html.StartTagToken, html.EndTagToken:
			tn, _ := z.TagName()
			if len(tn) == 1 && tn[0] == 'a' {
				if tt == html.StartTagToken {
					depth++
				} else {
					depth--
				}
			}
		}
	}
	if n == 0 {
		return errors.New("No entries found, the parsing failed")
	}
	return nil
}
