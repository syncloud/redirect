# def enable(email, artifact_dir):
#     check_output("mysql --host=mysql --user=root --password=root redirect -e "
#                  "\"update user set subscription_id = '1' where email = '{0}';\""
#                  " > {1}/db-user-premium.log".format(email, artifact_dir), shell=True)

def subscribe_crypto(selenium):
    selenium.find_by_id('subscription_inactive')
    selenium.find_by_id('crypto_year').click()
    selenium.find_by_id('crypto_transaction_id').send_keys('12345678901')
    selenium.find_by_id('crypto_subscribe_btn').click()
    selenium.find_by_id('cancel')
    # subscription.enable(DEVICE_USER, artifact_dir)
    selenium.find_by_id('subscription_active')
    selenium.screenshot('account-subscription-active')
