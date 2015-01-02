from subprocess import check_output
from os.path import join

class Runner:
    def __init__(self, host, user, passwd):
        self.host = host
        self.user = user
        self.passwd = passwd

    def __mysql_command(self, executable, command):
        cmd = '{0} --host={1} --user={2}'.format(executable, self.host, self.user)
        if self.passwd:
            cmd += ' --password={0}'.format(self.passwd)
        cmd += ' '+command
        return cmd

    def run(self, database, script):
        print "running {} on database {}".format(script, database)
        command = self.__mysql_command('mysql', '{0} < {1}'.format(database, script))
        print command
        print check_output(command, shell=True)

    def execute(self, query):
        print 'running "{}"'.format(query)
        command = self.__mysql_command('mysql', '--execute="{0}"'.format(query))
        print command
        print check_output(command, shell=True)

    def create(self, database):
        self.execute('create database {0};'.format(database))

    def drop(self, database):
        self.execute('drop database if exists {0};'.format(database))

    def backup(self, database, filename):
        print 'creating backup for database {0} to file {1}'.format(database, filename)
        command = self.__mysql_command('mysqldump', '--compact {0} > {1}'.format(database, filename))
        print command
        print check_output(command, shell=True)

    def restore(self, database, filename):
        print 'restoring backup file {0} to database {1}'.format(filename, database)
        self.drop(database)
        self.create(database)
        self.execute('set global foreign_key_checks=0;')
        self.run(database, filename)
        self.execute('set global foreign_key_checks=1;')