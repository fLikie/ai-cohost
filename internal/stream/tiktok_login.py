from playwright.sync_api import sync_playwright

with sync_playwright() as p:
    user_data_dir = "tiktok_profile"
    browser = p.chromium.launch_persistent_context(user_data_dir, headless=False)
    page = browser.new_page()
    page.goto("https://www.tiktok.com/login")
    print("🔐 Войди в аккаунт вручную и закрой окно")
    page.wait_for_timeout(120000)  # 2 минуты на вход вручную
    browser.close()