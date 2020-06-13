from os.path import dirname
from subprocess import check_output


DIR = dirname(__file__)


def recreate():
    check_output('mysql --host=mysql --user=root --password=root -e "drop DATABASE redirect"', shell=True)
    check_output('mysql --host=mysql --user=root --password=root -e "CREATE DATABASE redirect"', shell=True)
    check_output('mysql --host=mysql --user=root --password=root redirect < {0}/../db/init.sql'.format(DIR), shell=True)
