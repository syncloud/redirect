package main

type TestMail struct {
}

func (m *TestMail) SendLogs(to string, data string, includeSupport bool) error {
	return nil
}
