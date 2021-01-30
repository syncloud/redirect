package service

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestParseBody(t *testing.T) {
	template := `Subject: Reset password
Hello
new line
`
	subject, body, err := ParseBody(template, map[string]string{})

	assert.Nil(t, err, "should parse template")
	assert.Equal(t, "Reset password", subject)
	assert.Equal(t, "Hello\nnew line\n", body)
}

func TestParseBodySubstitute(t *testing.T) {
	template := `Subject: Reset password
Url: {url}
`
	subject, body, err := ParseBody(template, map[string]string{"url": "http://url"})

	assert.Nil(t, err, "should parse template")
	assert.Equal(t, "Reset password", subject)
	assert.Equal(t, "Url: http://url\n", body)
}

func TestParseUrl(t *testing.T) {
	url := ParseUrl("http://url?={0}", "token")
	assert.Equal(t, "http://url?=token", url)
}
