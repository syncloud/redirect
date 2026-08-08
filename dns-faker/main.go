package main

import (
	"encoding/json"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/miekg/dns"
)

const zone = "test."

type Store struct {
	mu  sync.RWMutex
	txt map[string][]string
	mx  map[string][]string
}

func NewStore() *Store {
	return &Store{txt: map[string][]string{}, mx: map[string][]string{}}
}

func (s *Store) Set(name string, values []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.txt[strings.ToLower(name)] = values
}

func (s *Store) Delete(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.txt, strings.ToLower(name))
}

func (s *Store) Get(name string) ([]string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.txt[strings.ToLower(name)]
	return v, ok
}

func (s *Store) SetMX(name string, values []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.mx[strings.ToLower(name)] = values
}

func (s *Store) DeleteMX(name string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.mx, strings.ToLower(name))
}

func (s *Store) GetMX(name string) ([]string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	v, ok := s.mx[strings.ToLower(name)]
	return v, ok
}

func (s *Store) AllMX() map[string][]string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make(map[string][]string, len(s.mx))
	for k, v := range s.mx {
		out[k] = append([]string{}, v...)
	}
	return out
}

func soa() *dns.SOA {
	return &dns.SOA{
		Hdr:     dns.RR_Header{Name: zone, Rrtype: dns.TypeSOA, Class: dns.ClassINET, Ttl: 60},
		Ns:      "ns.test.",
		Mbox:    "admin.test.",
		Serial:  uint32(time.Now().Unix()),
		Refresh: 7200,
		Retry:   3600,
		Expire:  1209600,
		Minttl:  60,
	}
}

func unquote(v string) string {
	if len(v) >= 2 && v[0] == '"' && v[len(v)-1] == '"' {
		return v[1 : len(v)-1]
	}
	return v
}

func (s *Store) ServeDNS(w dns.ResponseWriter, req *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(req)
	m.Authoritative = true
	if len(req.Question) == 1 {
		q := req.Question[0]
		name := strings.ToLower(q.Name)
		switch {
		case !strings.HasSuffix(name, zone):
			m.Rcode = dns.RcodeRefused
		case q.Qtype == dns.TypeTXT:
			if vals, ok := s.Get(name); ok {
				for _, v := range vals {
					m.Answer = append(m.Answer, &dns.TXT{
						Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeTXT, Class: dns.ClassINET, Ttl: 60},
						Txt: []string{unquote(v)},
					})
				}
			} else {
				m.Ns = append(m.Ns, soa())
				m.Rcode = dns.RcodeNameError
			}
		case q.Qtype == dns.TypeMX:
			if vals, ok := s.GetMX(name); ok {
				for _, v := range vals {
					preference, host := splitMX(v)
					m.Answer = append(m.Answer, &dns.MX{
						Hdr:        dns.RR_Header{Name: q.Name, Rrtype: dns.TypeMX, Class: dns.ClassINET, Ttl: 60},
						Preference: preference,
						Mx:         dns.Fqdn(host),
					})
				}
			} else {
				m.Ns = append(m.Ns, soa())
				m.Rcode = dns.RcodeNameError
			}
		case q.Qtype == dns.TypeSOA && name == zone:
			m.Answer = append(m.Answer, soa())
		default:
			m.Ns = append(m.Ns, soa())
			if _, ok := s.Get(name); !ok {
				m.Rcode = dns.RcodeNameError
			}
		}
	}
	_ = w.WriteMsg(m)
}

type changeRequest struct {
	XMLName xml.Name `xml:"ChangeResourceRecordSetsRequest"`
	Changes []struct {
		Action            string `xml:"Action"`
		ResourceRecordSet struct {
			Name   string   `xml:"Name"`
			Type   string   `xml:"Type"`
			Values []string `xml:"ResourceRecords>ResourceRecord>Value"`
		} `xml:"ResourceRecordSet"`
	} `xml:"ChangeBatch>Changes>Change"`
}

const xmlns = "https://route53.amazonaws.com/doc/2013-04-01/"

type API struct {
	store *Store
}

func (a *API) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	log.Printf("route53 %s %s", r.Method, path)
	switch {
	case strings.Contains(path, "/change/"):
		a.getChange(w, r)
	case strings.Contains(path, "hostedzonesbyname"):
		a.listZonesByName(w)
	case strings.Contains(path, "/rrset"):
		if r.Method == http.MethodPost {
			a.change(w, r)
		} else {
			a.listRecords(w)
		}
	case strings.Contains(path, "hostedzone"):
		a.listZones(w)
	case path == "/faker/mx":
		a.mxRecords(w)
	case path == "/" || strings.Contains(path, "health"):
		w.WriteHeader(http.StatusOK)
		_, _ = io.WriteString(w, "ok")
	default:
		http.Error(w, "unsupported", http.StatusBadRequest)
	}
}

func (a *API) mxRecords(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(a.store.AllMX())
}

func splitMX(value string) (uint16, string) {
	fields := strings.Fields(value)
	if len(fields) != 2 {
		return 10, strings.TrimSpace(value)
	}
	preference, err := strconv.ParseUint(fields[0], 10, 16)
	if err != nil {
		return 10, fields[1]
	}
	return uint16(preference), fields[1]
}

func writeXML(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(http.StatusOK)
	_, _ = io.WriteString(w, `<?xml version="1.0" encoding="UTF-8"?>`+body)
}

const hostedZone = `<HostedZone><Id>/hostedzone/ZTEST</Id><Name>test.</Name>` +
	`<CallerReference>ci</CallerReference><Config><PrivateZone>false</PrivateZone></Config>` +
	`<ResourceRecordSetCount>1</ResourceRecordSetCount></HostedZone>`

func (a *API) listZonesByName(w http.ResponseWriter) {
	writeXML(w, `<ListHostedZonesByNameResponse xmlns="`+xmlns+`"><HostedZones>`+hostedZone+
		`</HostedZones><DNSName>test.</DNSName><IsTruncated>false</IsTruncated><MaxItems>100</MaxItems></ListHostedZonesByNameResponse>`)
}

func (a *API) listZones(w http.ResponseWriter) {
	writeXML(w, `<ListHostedZonesResponse xmlns="`+xmlns+`"><HostedZones>`+hostedZone+
		`</HostedZones><IsTruncated>false</IsTruncated><MaxItems>100</MaxItems></ListHostedZonesResponse>`)
}

func (a *API) getChange(w http.ResponseWriter, r *http.Request) {
	writeXML(w, `<GetChangeResponse xmlns="`+xmlns+`"><ChangeInfo><Id>/change/CSYNC</Id>`+
		`<Status>INSYNC</Status><SubmittedAt>2010-09-10T01:36:41.958Z</SubmittedAt></ChangeInfo></GetChangeResponse>`)
}

func (a *API) change(w http.ResponseWriter, r *http.Request) {
	body, _ := io.ReadAll(r.Body)
	var req changeRequest
	if err := xml.Unmarshal(body, &req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, c := range req.Changes {
		rr := c.ResourceRecordSet
		switch rr.Type {
		case "TXT":
			if strings.EqualFold(c.Action, "DELETE") {
				a.store.Delete(rr.Name)
			} else {
				a.store.Set(rr.Name, rr.Values)
			}
		case "MX":
			if strings.EqualFold(c.Action, "DELETE") {
				a.store.DeleteMX(rr.Name)
			} else {
				a.store.SetMX(rr.Name, rr.Values)
			}
		}
	}
	writeXML(w, `<ChangeResourceRecordSetsResponse xmlns="`+xmlns+`"><ChangeInfo><Id>/change/CSYNC</Id>`+
		`<Status>INSYNC</Status><SubmittedAt>2010-09-10T01:36:41.958Z</SubmittedAt></ChangeInfo></ChangeResourceRecordSetsResponse>`)
}

func (a *API) listRecords(w http.ResponseWriter) {
	var b strings.Builder
	b.WriteString(`<ListResourceRecordSetsResponse xmlns="` + xmlns + `"><ResourceRecordSets>`)
	a.store.mu.RLock()
	for name, vals := range a.store.txt {
		b.WriteString(`<ResourceRecordSet><Name>` + name + `</Name><Type>TXT</Type><TTL>60</TTL><ResourceRecords>`)
		for _, v := range vals {
			b.WriteString(`<ResourceRecord><Value>` + v + `</Value></ResourceRecord>`)
		}
		b.WriteString(`</ResourceRecords></ResourceRecordSet>`)
	}
	a.store.mu.RUnlock()
	b.WriteString(`</ResourceRecordSets><IsTruncated>false</IsTruncated><MaxItems>100</MaxItems></ListResourceRecordSetsResponse>`)
	writeXML(w, b.String())
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func main() {
	store := NewStore()

	dnsAddr := env("DNSSIM_DNS", ":53")
	dns.HandleFunc(".", store.ServeDNS)
	for _, net := range []string{"udp", "tcp"} {
		srv := &dns.Server{Addr: dnsAddr, Net: net}
		go func(s *dns.Server) {
			if err := s.ListenAndServe(); err != nil {
				log.Fatalf("dns %s: %v", s.Net, err)
			}
		}(srv)
	}

	apiAddr := env("DNSSIM_API", ":4566")
	log.Printf("dns-faker: route53 api on %s, dns on %s", apiAddr, dnsAddr)
	if err := http.ListenAndServe(apiAddr, &API{store: store}); err != nil {
		log.Fatal(err)
	}
}
