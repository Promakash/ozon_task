FROM ubuntu:latest
LABEL authors="promakash"

ENTRYPOINT ["top", "-b"]