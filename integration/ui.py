from selenium.webdriver.support import expected_conditions as EC
from selenium.webdriver.common.by import By

def domain_delete(screenshot, selenium):
    deactivate_xpath = "//div[contains(@class, 'panel')]//button[contains(text(), 'Deactivate')]"
    selenium.wait_or_screenshot(EC.presence_of_element_located((By.XPATH, deactivate_xpath)))
    selenium.find_by_xpath(deactivate_xpath).click()

    confirm_xpath = "//div[@id='delete_confirmation']//button[contains(., 'Confirm')]"
    selenium.wait_or_screenshot(EC.presence_of_element_located((By.XPATH, confirm_xpath)))
    selenium.find_by_xpath(confirm_xpath).click()

    device_label = "//h3[text()='Some Device']"
    selenium.wait_or_screenshot(EC.invisibility_of_element_located((By.XPATH, device_label)))
    found = selenium.exists_by(By.XPATH, device_label)
    selenium.screenshot(screenshot)
    assert not found
