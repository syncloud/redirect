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
Url: https://www.{domain}/path?token={token}
`
	subject, body, err := ParseBody(template, map[string]string{"domain": "example.com", "token": "123"})

	assert.Nil(t, err, "should parse template")
	assert.Equal(t, "Reset password", subject)
	assert.Equal(t, "Url: https://www.example.com/path?token=123\n", body)
}
