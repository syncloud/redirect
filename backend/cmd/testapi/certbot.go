package main

type TestCertbot struct {
}

func (c *TestCertbot) Present(token string, fqdn string, values []string) error {
	return nil
}
func (c *TestCertbot) CleanUp(token, fqdn string) error {
	return nil
}
