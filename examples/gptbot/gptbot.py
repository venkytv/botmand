#!/usr/bin/env python3

from openai import OpenAI

from collections import deque
import logging
import os
import re
import sys
import tiktoken
import time
from typing import Deque, List

BOT_ID = os.environ["BOTMAND_USER_ID"]
BOT_NAME = os.environ["BOTMAND_USER_NAME"]

MODEL_NAME = os.environ.get("MODEL_NAME", "gpt-3.5-turbo")
TEMPERATURE = int(os.environ.get("MODEL_TEMPERATURE", 0.6))

MAX_TOKENS = int(os.environ.get("MAX_TOKENS", 1000))
MAX_HISTORY = int(os.environ.get("MAX_HISTORY", 100))   # 100 messages of history
MAX_AGE = int(os.environ.get("MAX_AGE", 60 * 30))       # 30 minutes of history
TONE = os.environ.get("TONE", "friendly")

VERBOSE = ( os.environ.get("VERBOSE", "") == "yes" )

logging.basicConfig(format="[%(levelname)s] %(message)s",
                    level=logging.DEBUG if VERBOSE else logging.INFO)
logging.debug("Enabled logging")

class Message:
    encoding = tiktoken.encoding_for_model(MODEL_NAME)

    def __init__(self, text: str, from_assistant=False):
        self.text = text
        self.from_assistant = from_assistant
        self.timestamp = time.time()
        self.num_tokens = len(self.encoding.encode(text))

    @property
    def age(self) -> float:
        return time.time() - self.timestamp

class MessageBuffer:
    def __init__(self, max_length: int, max_tokens: int):
        self._messages: Deque[Message] = deque(maxlen=max_length)
        self.max_tokens = max_tokens
        self.num_tokens = 0

    def append(self, message: Message):
        self._messages.append(message)
        self.num_tokens += message.num_tokens

        while self.num_tokens > self.max_tokens:
            self.num_tokens -= self._messages.popleft().num_tokens

        # Delete older messages
        while self._messages and self._messages[0].age > MAX_AGE:
            self.num_tokens -= self._messages.popleft().num_tokens

    @property
    def messages(self) -> List[str]:
        return [m for m in self._messages if m.age < MAX_AGE]

    def length(self) -> int:
        return len(self._messages)

    def __str__(self):
        return f"N={self.length()} T={self.num_tokens}"

class ChatEngine:

    def __init__(self, bot_user_id, model_name=MODEL_NAME, temperature=TEMPERATURE,
                 max_history=MAX_HISTORY, max_tokens=MAX_TOKENS, verbose=False):
        self.bot_user_id = f"<@{bot_user_id}>"
        self.model_name = model_name
        self.temperature = temperature
        self.verbose = verbose

        self.system_prompt = f"""
        The following is a conversation between a set of humans and an AI chatbot called {BOT_NAME}.
        {BOT_NAME} is talkitive and provides lots of specific details from its context.
        If {BOT_NAME} does not know the answer to a question, it truthfully says so.
        Tone for {BOT_NAME}'s responses: {TONE}
        Format everything in markdown."""

        self.buffer = MessageBuffer(max_length=max_history, max_tokens=max_tokens)
        self._client = None

    @property
    def client(self):
        if not self._client:
            api_key = os.environ.get("OPENAI_API_KEY", None)
            if api_key:
                self._client = OpenAI(api_key=api_key)
        return self._client

    # Record a message, and respond if necessary
    def record(self, raw_message):
        # Strip <> from usernames
        message = re.sub(r"<(U.*?)>", r"\1", raw_message.strip())

        # Replace bot user id with name
        message = message.replace(self.bot_user_id, BOT_NAME)

        # Add message to buffer
        self.buffer.append(Message(message))

        if self.bot_user_id in raw_message:
            # Directed message; respond
            self.respond("...")
            response = self.chat()

            if response:
                self.buffer.append(Message(response, from_assistant=True))

                # Replace all usernames with slack mentions
                response = re.sub(r"(U.*?)\b", r"<@\1>", response)

                self.respond(response)

    # Get the messages to send to the chat engine
    def get_messages(self) -> list[dict[str, str]]:
        messages = [
            { "role": "system", "content": self.system_prompt },
        ]
        for message in self.buffer.messages:
            messages.append({
                "role": "assistant" if message.from_assistant else "user",
                "content": message.text,
            })

        logging.debug(messages)
        return messages

    # Run the chat engine and return a response
    def chat(self) -> str:
        if not self.client:
            self.respond(f"No OpenAI API key found in environment.\n" +
                         "Please set the OPENAI_API_KEY environment variable in the config.")
            sys.exit("No OpenAI API key found in environment.")
            return None

        try:
            response = self.client.chat.completions.create(
                messages=self.get_messages(),
                model=self.model_name)
            return response.choices[0].message.content
        except Exception as e:
            return f"Error: {e}"

    # Print a response
    def respond(self, response):
        # Replace newlines with escaped newlines
        response = response.replace("\n", "\\n")
        print(response, flush=True)

    def shutdown(self):
        engine.respond("Shutting down...")

if __name__ == "__main__":
    logging.debug("Starting chat engine...")
    engine = ChatEngine(bot_user_id=BOT_ID, verbose=VERBOSE)
    try:
        for line in sys.stdin:
            engine.record(line)
    except KeyboardInterrupt:
        engine.shutdown()
