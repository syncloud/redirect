from subprocess import check_output


def premium_buy(email, artifact_dir):
    check_output("mysql --host=mysql --user=root --password=root redirect -e "
                 "\"update user set subscription_id = '1' where email = '{0}';\""
                 " > {1}/db-user-premium.log".format(email, artifact_dir), shell=True)
