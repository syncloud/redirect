package relay

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os/exec"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

type fakerMessage struct {
	Source       string   `json:"source"`
	Destinations []string `json:"destinations"`
	Body         string   `json:"body"`
}

func startSesFaker(t *testing.T) string {
	t.Helper()
	binary := t.TempDir() + "/ses-faker"
	build := exec.Command("go", "build", "-o", binary, ".")
	build.Dir = "../../ses-faker"
	if out, err := build.CombinedOutput(); err != nil {
		t.Skipf("cannot build ses faker: %v %s", err, out)
	}

	address := "127.0.0.1:14579"
	faker := exec.Command(binary)
	faker.Env = append(faker.Environ(), "SESSIM_ADDR="+address)
	assert.NoError(t, faker.Start())
	t.Cleanup(func() { _ = faker.Process.Kill() })

	url := "http://" + address
	for i := 0; i < 100; i++ {
		if _, err := http.Get(url + "/faker/messages"); err == nil {
			return url
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("ses faker did not start")
	return ""
}

func sesSenderFor(url string) *SesSender {
	awsSession := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewStaticCredentials("key", "secret", ""),
	}))
	return NewSesSender(awsSession, "us-west-2", url, "", zap.NewNop())
}

func fakerMessages(t *testing.T, url string) []fakerMessage {
	t.Helper()
	response, err := http.Get(url + "/faker/messages")
	assert.NoError(t, err)
	defer response.Body.Close()
	var messages []fakerMessage
	assert.NoError(t, json.NewDecoder(response.Body).Decode(&messages))
	return messages
}

func setBehaviour(t *testing.T, url string, behaviour Behaviour) {
	t.Helper()
	body, err := json.Marshal(behaviour)
	assert.NoError(t, err)
	response, err := http.Post(url+"/faker/behaviour", "application/json", bytes.NewReader(body))
	assert.NoError(t, err)
	defer response.Body.Close()
}

type Behaviour struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

func TestSesSender_SendsThroughTheFaker(t *testing.T) {
	url := startSesFaker(t)

	err := sesSenderFor(url).Send("user@device.syncloud.it",
		[]string{"someone@example.com"}, []byte("Subject: hi\r\n\r\nbody\r\n"))
	assert.NoError(t, err)

	messages := fakerMessages(t, url)
	assert.Len(t, messages, 1)
	assert.Equal(t, "user@device.syncloud.it", messages[0].Source)
	assert.Equal(t, []string{"someone@example.com"}, messages[0].Destinations)
	assert.Contains(t, messages[0].Body, "Subject: hi")
}

func TestSesSender_ReportsWhatSesRefuses(t *testing.T) {
	url := startSesFaker(t)
	setBehaviour(t, url, Behaviour{Status: 454, Code: "Throttling", Message: "Maximum sending rate exceeded"})

	err := sesSenderFor(url).Send("user@device.syncloud.it",
		[]string{"someone@example.com"}, []byte("Subject: hi\r\n\r\nbody\r\n"))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "Throttling")
	assert.Empty(t, fakerMessages(t, url))
}
