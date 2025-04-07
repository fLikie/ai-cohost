import asyncio
import sys
from TikTokLive import TikTokLiveClient
from TikTokLive.events import ConnectEvent, DisconnectEvent, CommentEvent, LiveEndEvent

# Укажи имя пользователя
username = "flikie4"
client = TikTokLiveClient(unique_id=username)

@client.on(ConnectEvent)
async def on_connect(_: ConnectEvent):
    print("✅ Подключено к TikTok Live!")

@client.on(DisconnectEvent)
async def on_disconnect(_: DisconnectEvent):
    print("❌ Отключено от TikTok!")

@client.on(LiveEndEvent)
async def on_end(_: LiveEndEvent):
    print("📴 Стрим завершён!")

@client.on(CommentEvent)
async def on_comment(event: CommentEvent):
    print(f"💬 {event.user.username}: {event.comment}")

async def main():
    try:
        await client.start()
    except Exception as e:
        print("❗ Ошибка:", e)

asyncio.run(main())
