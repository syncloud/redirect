package caddysmtp

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strings"
)

const maxLineBytes = 4096

var ErrLineTooLong = errors.New("smtp line is too long")

func newReader(reader io.Reader) *bufio.Reader {
	return bufio.NewReaderSize(reader, maxLineBytes)
}

func readLine(reader *bufio.Reader) (string, error) {
	line, err := reader.ReadSlice('\n')
	if errors.Is(err, bufio.ErrBufferFull) {
		return "", ErrLineTooLong
	}
	if err != nil {
		return "", err
	}
	return strings.TrimRight(string(line), "\r\n"), nil
}

func readResponse(reader *bufio.Reader) (string, error) {
	for {
		line, err := readLine(reader)
		if err != nil {
			return "", err
		}
		if len(line) < 4 || line[3] != '-' {
			return line, nil
		}
	}
}

func writeLines(writer io.Writer, lines ...string) error {
	_, err := io.WriteString(writer, strings.Join(lines, "\r\n")+"\r\n")
	return err
}

func command(line string) string {
	name, _, _ := strings.Cut(line, " ")
	return strings.ToUpper(name)
}

func pipe(client net.Conn, clientReader io.Reader, server net.Conn, serverReader io.Reader) error {
	done := make(chan error, 2)
	go func() {
		_, err := io.Copy(server, clientReader)
		done <- err
	}()
	go func() {
		_, err := io.Copy(client, serverReader)
		done <- err
	}()
	err := <-done
	_ = client.Close()
	_ = server.Close()
	<-done
	return err
}
