# voice_listener.py
import queue
import random

import sounddevice as sd
import vosk
import json
import os
import sys
import io
from playsound import playsound

sound_dir = "C:/Users/fliki/GolandProjects/cohost/internal/audio/sounds"
sounds = [f for f in os.listdir(sound_dir) if f.endswith(".mp3")]

sys.stdout.reconfigure(line_buffering=True)
sys.stdout = io.TextIOWrapper(sys.stdout.detach(), encoding='utf-8')

wake_word = sys.argv[1] if len(sys.argv) > 1 else "пятница"

# Используй путь до модели
model_path = "C:/Users/fliki/GolandProjects/cohost/internal/audio/vosk-model-small-ru-0.22"
if not os.path.exists(model_path):
    print(f"[ERROR] Модель не найдена по пути: {model_path}")
    exit(1)

model = vosk.Model(model_path)
samplerate = 16000
q = queue.Queue()

def callback(indata, frames, time, status):
    if status:
        print(status, file=sys.stderr)
    q.put(bytes(indata))

rec = vosk.KaldiRecognizer(model, samplerate)
rec.SetWords(True)

print("🎙 Ожидание кодового слова...", flush=True)

with sd.RawInputStream(samplerate=samplerate, blocksize=8000, dtype='int16',
                       channels=1, callback=callback):
    while True:
        data = q.get()
        if rec.AcceptWaveform(data):
            result = json.loads(rec.Result())
            text = result.get("text", "").lower()

            if not text:
                continue

            # Если услышал кодовое слово
            if wake_word in text:
                sound_to_play = os.path.join(sound_dir, random.choice(sounds))
                playsound(sound_to_play)
                print("Кодовое слово услышано!",flush=True)
                print("Ожидание команды...",flush=True)

                # Слушаем команду (один раз)
                full_text = ""
                for _ in range(15):  # 15 блоков по ~0.5 сек
                    data = q.get()
                    if rec.AcceptWaveform(data):
                        res = json.loads(rec.Result())
                        full_text += " " + res.get("text", "")
                print(f"📤 Команда: {full_text.strip()}",flush=True)
                print(json.dumps({"command": full_text.strip()}))  # отправляем в stdout
                print("🎙 Снова жду кодовое слово...",flush=True)
