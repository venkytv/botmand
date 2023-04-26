# syntax=docker/dockerfile:1

ARG ARCH=
FROM ${ARCH}python:3.9-slim-buster

WORKDIR /app

COPY requirements.txt requirements.txt

RUN pip install -r requirements.txt

COPY . .

ARG BOT_NAME=Botters
ENV BOT_NAME=${BOT_NAME}

CMD ["python3", "gptbot.py"]
