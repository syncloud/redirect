class Users:
    def __init__(self, create_storage):
        self.create_storage = create_storage

    def get_user(self, email):
        with self.create_storage() as storage:
            return storage.get_user_by_email(email)
