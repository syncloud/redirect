package main

type TestPortProbe struct {
}

func (p *TestPortProbe) Probe(token string, port int, ip string) (*string, error) {
	result := "OK"
	return &result, nil
}
