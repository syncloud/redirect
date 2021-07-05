from subprocess import check_output


def premium_approve(email, artifact_dir):
    check_output("mysql --host=mysql --user=root --password=root redirect -e "
                 "\"update user set premium_status_id = 3 where email = '{0}';\""
                 " > {1}/db-user-premium.log".format(email, artifact_dir), shell=True)
