import subprocess
from os.path import dirname
from subprocess import check_output


DIR = dirname(__file__)


def recreate():
     check_output('{0}/../ci/recreatedb'.format(DIR), shell=True, stderr=subprocess.STDOUT)
