import json
import time
from playwright.sync_api import sync_playwright
import sys
import io

sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

def run():
    with sync_playwright() as p:
        # Подключаемся к уже запущенному Chrome
        browser = p.chromium.connect_over_cdp("http://localhost:9222")
        context = browser.contexts[0]  # Получаем первый (основной) контекст
        page = context.new_page()

        page.goto("https://livecenter.tiktok.com/live_monitor?lang=ru-RU")
        print("⏳ Ожидание загрузки сообщений...", flush=True)

        page.wait_for_selector("div[data-e2e='chat-message']", timeout=30000)
        print("✅ Страница загружена!", flush=True)

        seen = set()

        while True:
            comments = page.query_selector_all("div[data-e2e='chat-message']")
            for comment in comments:
                try:
                    user = comment.query_selector("span[data-e2e='message-owner-name']").evaluate("node => node.textContent")

                # Берём последний <div> в блоке — там сообщение
                    text_divs = comment.query_selector_all("div")
                    text = text_divs[-1].evaluate("node => node.textContent") if text_divs else ""

                    comment_id = user + text
                    if comment_id not in seen:
                        seen.add(comment_id)
                        print(json.dumps({"user": user, "comment": text}), flush=True)
                except Exception as e:
                    print("❌ Ошибка при извлечении комментария:", e)

            time.sleep(1.5)

if __name__ == "__main__":
    run()
