import shutil
import os

import subprocess
import signal

import time
import smtplib


def wait_smtp(ip, port, timeout):
    available = False
    start = time.time()
    while not available:
        try:
            test_server = smtplib.SMTP()
            test_server.connect(ip, port)
            available = True
        except:
            pass
        elapsed = time.time() - start
        if elapsed > timeout:
            raise Exception('Timeout of {} seconds happened {} seconds elapsed'.format(timeout, elapsed))


class FakeSmtp:

    def __init__(self, ip, port):
        self.smtp_outbox_path = 'outbox'
        self.root_dir = '/home/test'
        self.smtp_outbox_fuul_path = os.path.join(self.root_dir, 'outbox')
        self.start(ip, port)

    def start(self, ip, port, timeout=1):
        smtp_sink_cmd = 'smtp-sink -u test -R {} -d "{}/%d.%H.%M.%S" {}:{} 1000'.format(
            self.root_dir, self.smtp_outbox_path, ip, port)
        self.server = subprocess.Popen(smtp_sink_cmd, shell=True, preexec_fn=os.setsid)
        wait_smtp(ip, port, timeout)

    def stop(self):
        os.killpg(self.server.pid, signal.SIGKILL)
        os.waitpid(self.server.pid, 0)
        self.clear()

    def clear(self):
        if os.path.isdir(self.smtp_outbox_fuul_path):
            shutil.rmtree(self.smtp_outbox_fuul_path)

    def emails(self):
        emails = []
        for filename in os.listdir(self.smtp_outbox_fuul_path):
            with open(os.path.join(self.smtp_outbox_fuul_path, filename), 'r') as f:
                emails.append(f.read())
        return emails

    def empty(self):
        if not os.path.isdir(self.smtp_outbox_fuul_path):
            return True
        return len(os.listdir(self.smtp_outbox_fuul_path)) == 0
