package internet

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/garyburd/redigo/redis"
)

// RefreshBGPDump ensures that the latest dump available is the one which is installed.
func RefreshBGPDump(conn redis.Conn) error {
	for _, b := range []BGPDump{
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

// BGPDump encapuslates downloading and importing of BGP dumps.
type BGPDump struct {
	Date time.Time
}

// Import stores the contents of a downloaded BGP dump into a redis server.
func (b *BGPDump) Import(conn redis.Conn) error {

	alreadyImported, err := redis.Bool(conn.Do("SISMEMBER", "i2a:imported_dates", b.day()))
	if err != nil {
		panic(err)
	}
	if alreadyImported {
		return nil
	}
	c := exec.Command("bgpdump", "-m", b.Path())

	stdout, err := c.StdoutPipe()
	if err != nil {
		return err

	}
	var wg sync.WaitGroup
	wg.Add(1)
	go func(r io.Reader) {
		defer wg.Done()
		b.parseBGPCSV(r, conn)
	}(stdout)
	c.Run()
	wg.Wait()
	return nil
}

// IsDownloaded returns true if the BGPDump archive is downloaded locally.
func (b *BGPDump) IsDownloaded() bool {
	p := b.Path()
	if _, err := os.Stat(p); err == nil {
		return true
	}
	return false
}

// Download fetches an bgpdump archive from http://data.ris.ripe.net/rrc00.
// A http 404 status code does not generate an error, the isDownloaded() to check success after fetching.
// Download returns early with no error if the file already is downloaded to disk.
func (b *BGPDump) Download() error {
	dt := b.Date
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
	dlURL := fmt.Sprintf(
		"http://data.ris.ripe.net/rrc00/%s/bview.%s.%s.gz",
		dt.Format("2006.01"), b.day(), "0000")

	resp, err := http.Get(dlURL)
	if err != nil {
		return err
	}

	// Dumps from ??? to 2010-06-14 are named timestamped 2359 so do a check
	// for that if 0000 fails. For very early dumps the format is not static so those will fail.
	if resp.StatusCode == 404 && dt.Before(time.Date(2010, 06, 15, 0, 0, 0, 0, time.UTC)) {
		// log.Printf("trying different url, got 404 for %s", dlURL)
		dlURL = fmt.Sprintf(
			"http://data.ris.ripe.net/rrc00/%s/bview.%s.%s.gz",
			dt.Format("2006.01"), b.day(), "2359")
		resp, err = http.Get(dlURL)
		if err != nil {
			return err
		}
	}

	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == 404 {
			// log.Printf("Skipping download, got 404 for %s", dlURL)
			return nil
		}
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

func (b *BGPDump) parseBGPCSV(r io.Reader, conn redis.Conn) error {
	day := b.day()

	start := time.Now()
	s := bufio.NewScanner(r)
	n := 0
	var asn string
	for s.Scan() {
		cols := strings.Split(s.Text(), "|")
		if len(cols) < 7 {
			log.Printf("Too few columns in %s:%d: %s", filepath.Base(b.Path()), n, s.Text())
			continue
		}
		block := cols[5]

		if _, ok := asn12654blocks[block]; ok {
			asn = "12654"
		} else {
			asPath := cols[6]
			asns := strings.Split(asPath, " ")
			asn = asns[len(asns)-1]
			if asn == "" {
				log.Printf("No ASPATH data for %s:%d: %s", filepath.Base(b.Path()), n, s.Text())
				continue
			}
		}
		conn.Send("HSET", fmt.Sprintf("i2a:%s", block), day, asn)
		n++
		if n%10000 == 0 {
			err := conn.Flush()
			if err != nil {
				panic(err)
			}
		}
	}
	conn.Send("SADD", "i2a:imported_dates", day)
	err := conn.Flush()
	if err != nil {
		panic(err)
	}

	log.Printf("Imported %d rows from %s in %s",
		n, filepath.Base(b.Path()), time.Since(start))

	return nil
}

// Path returns the absolute path to the target archive dump download file.
func (b *BGPDump) Path() string {
	return filepath.Join(
		b.dir(), fmt.Sprintf("%s.gz", b.Date.Format("20060102")))
}

func (b *BGPDump) dir() string {
	return filepath.Join(
		dataDir, "cache", b.Date.Format("200601"))
}

func (b *BGPDump) day() string {
	return b.Date.Format("20060102")
}
