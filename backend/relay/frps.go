package relay

import (
	"bufio"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"
)

var trafficLine = regexp.MustCompile(`^frp_server_traffic_(?:in|out)\{[^}]*name="([^"]+)"[^}]*\}\s+([0-9.eE+-]+)$`)

type FrpsMetrics struct {
	url    string
	client *http.Client
}

func NewFrpsMetrics(url string) *FrpsMetrics {
	return &FrpsMetrics{
		url:    url,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (f *FrpsMetrics) Fetch() (map[string]int64, error) {
	response, err := f.client.Get(f.url)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("frps metrics status %d", response.StatusCode)
	}
	return parseTraffic(response.Body), nil
}

func parseTraffic(reader interface{ Read([]byte) (int, error) }) map[string]int64 {
	totals := map[string]int64{}
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		match := trafficLine.FindStringSubmatch(scanner.Text())
		if match == nil {
			continue
		}
		value, err := strconv.ParseFloat(match[2], 64)
		if err != nil {
			continue
		}
		totals[match[1]] += int64(value)
	}
	return totals
}
