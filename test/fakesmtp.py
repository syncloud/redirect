import shutil
import os


class FakeSmtp:

    def __init__(self, smtp_outbox_path):
        self.smtp_outbox_path = smtp_outbox_path

    def clear(self):
        if os.path.isdir(self.smtp_outbox_path):
            shutil.rmtree(self.smtp_outbox_path)

    def emails(self):
        emails = []
        for filename in os.listdir(self.smtp_outbox_path):
            with open(os.path.join(self.smtp_outbox_path, filename), 'r') as f:
                emails.append(f.read())
        return emails

    def empty(self):
        if not os.path.isdir(self.smtp_outbox_path):
            return True
        return len(os.listdir(self.smtp_outbox_path)) == 0