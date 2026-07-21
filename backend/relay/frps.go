package relay

import (
	"bufio"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var trafficLine = regexp.MustCompile(`^frp_server_traffic_(?:in|out)\{[^}]*name="([^"]+)"[^}]*\}\s+([0-9.eE+-]+)$`)

type FrpsMetrics struct {
	url          string
	user         string
	passwordFile string
	client       *http.Client
}

func NewFrpsMetrics(url string, user string, passwordFile string) *FrpsMetrics {
	return &FrpsMetrics{
		url:          url,
		user:         user,
		passwordFile: passwordFile,
		client:       &http.Client{Timeout: 5 * time.Second},
	}
}

func (f *FrpsMetrics) Fetch() (map[string]int64, error) {
	request, err := http.NewRequest("GET", f.url, nil)
	if err != nil {
		return nil, err
	}
	if password, err := os.ReadFile(f.passwordFile); err == nil {
		request.SetBasicAuth(f.user, strings.TrimSpace(string(password)))
	}
	response, err := f.client.Do(request)
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
