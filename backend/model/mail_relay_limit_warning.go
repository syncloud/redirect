package model

import "strconv"

type MailRelayLimitWarning struct {
	Domain string `sub:"domain"`
	Used   int64  `sub:"used"`
	Limit  int64  `sub:"limit"`
}

func (w MailRelayLimitWarning) Subs() map[string]string {
	return map[string]string{
		"domain": w.Domain,
		"used":   strconv.FormatInt(w.Used, 10),
		"limit":  strconv.FormatInt(w.Limit, 10),
	}
}
