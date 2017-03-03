// PT (Performance Testing), Copyright (c) 2017 Dmitry Doroginin
//
// This file is part of PT.
//
// PT is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// PT is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with pt. If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"crypto/tls"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
	"sync/atomic"
)

var (
	respTime      int64
	maxRespTime   int64
	minRespTime   int64
	respCount     uint64
	respCount200  uint64
	respSize      uint64
	reqCountTotal uint64
	lastDt        time.Time
	m             sync.Mutex
)

func main() {
	url := flag.String("url", "", "Fetching url")
	n := flag.Int("n", 50, "Number of multiple requests to make at a time")
	cert := flag.String("cert", "", "TLS client PEM encoded certificate file")
	ka := flag.Bool("ka", false, "Use keep alive")
	c := flag.Int("c", 0, "The total number of requests (0 - unlimit)")
	t := flag.Duration("t", 5 * time.Second, "Timeout per request")

	flag.Parse()

	if *url == "" {
		log.Fatal("Fetching url is not specified")
	}
	var cfg *tls.Config
	if *cert != "" {
		cfg = &tls.Config{}
		c, err := tls.LoadX509KeyPair(*cert, *cert)
		if err != nil {
			log.Fatalf("Failed to parse certificate: %s", err)
		}
		cfg.Certificates = append(cfg.Certificates, c)
		cfg.BuildNameToCertificate()
		log.Printf("Cert '%s' will be used", *cert)
	}

	log.Printf("Start load test for url '%s' with %d concurency (keep-alive: %v)", *url, *n, *ka)

	lastDt = time.Now()

	for i := 0; i < *n; i++ {
		go func() {
			client := &http.Client{
				Timeout: *t,
				Transport: &http.Transport{
					DisableKeepAlives:  !*ka,
					DisableCompression: true,
					TLSClientConfig:    cfg,
				},
			}

			for {
				start := time.Now()
				if *c > 0 {
					if atomic.LoadUint64(&reqCountTotal) >= uint64(*c) {
						break
					}
					atomic.AddUint64(&reqCountTotal, 1)
				}
				resp, err := client.Get(*url)

				code, length := 0, 0
				if err != nil {
					log.Printf("error: %s", err)
				} else {
					data, err := ioutil.ReadAll(resp.Body)
					resp.Body.Close()

					if err != nil {
						log.Printf("Error: %s", err)
					}

					code = resp.StatusCode
					length += len(data)
				}

				rt := int64(time.Now().Sub(start))

				m.Lock()
				respSize += uint64(length)
				if code == http.StatusOK {
					respCount200++
				}
				respCount++
				respTime += rt
				if rt > maxRespTime {
					maxRespTime = rt
				}
				if minRespTime > rt || minRespTime == 0 {
					minRespTime = rt
				}
				m.Unlock()
			}
		}()
	}

	for {
		time.Sleep(time.Second)

		m.Lock()
		if respCount == 0 {
			m.Unlock()
			continue
		}

		now := time.Now()

		log.Printf("Min RT: %s\t Max RT: %s\tAvg RT: %s\tRPS: %.3f\tMb/s: %.3f\tRC: %d\tSuccess(200): %.2f%%",
			time.Duration(minRespTime),
			time.Duration(maxRespTime),
			time.Duration(uint64(respTime)/respCount),
			float64(respCount)/now.Sub(lastDt).Seconds(),
			float64(respSize)/now.Sub(lastDt).Seconds()/1024/1024,
			respCount,
			100*float64(respCount200)/float64(respCount),
		)

		lastDt = now

		respTime = 0
		maxRespTime = 0
		minRespTime = 0
		respCount = 0
		respCount200 = 0
		respSize = 0

		m.Unlock()

		if *c > 0 && atomic.LoadUint64(&reqCountTotal) >= uint64(*c) {
			break
		}
	}
}
