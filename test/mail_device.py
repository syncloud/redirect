import socket
import threading
import time


class MailDevice:

    def __init__(self):
        self.socket = socket.socket()
        self.socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
        self.socket.bind(('127.0.0.1', 0))
        self.socket.listen(5)
        self.port = self.socket.getsockname()[1]
        self.messages = []
        self.running = True
        threading.Thread(target=self.serve, daemon=True).start()

    def serve(self):
        while self.running:
            try:
                connection, _ = self.socket.accept()
            except OSError:
                return
            threading.Thread(target=self.session, args=(connection,), daemon=True).start()

    def session(self, connection):
        stream = connection.makefile('rwb')
        stream.write(b'220 device.syncloud.test ESMTP\r\n')
        stream.flush()
        recipients = []
        body = []
        reading = False
        for raw in stream:
            line = raw.decode('utf-8', 'replace')
            if reading:
                if line.strip() == '.':
                    reading = False
                    self.messages.append({'recipients': list(recipients), 'body': ''.join(body)})
                    stream.write(b'250 accepted\r\n')
                    stream.flush()
                else:
                    body.append(line)
                continue
            command = line.upper()
            if command.startswith('EHLO') or command.startswith('HELO'):
                stream.write(b'250-device.syncloud.test\r\n250 8BITMIME\r\n')
            elif command.startswith('DATA'):
                reading = True
                stream.write(b'354 go ahead\r\n')
            elif command.startswith('RCPT'):
                recipients.append(line.strip())
                stream.write(b'250 ok\r\n')
            elif command.startswith('RSET'):
                recipients, body = [], []
                stream.write(b'250 ok\r\n')
            elif command.startswith('QUIT'):
                stream.write(b'221 bye\r\n')
                stream.flush()
                break
            else:
                stream.write(b'250 ok\r\n')
            stream.flush()
        connection.close()

    def stop(self):
        self.running = False
        self.socket.close()

    def wait(self, attempts=30):
        for _ in range(attempts):
            if self.messages:
                return self.messages
            time.sleep(1)
        return self.messages
