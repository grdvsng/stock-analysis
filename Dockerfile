FROM ubuntu:24.04

COPY . .

RUN apt update
RUN sudo apt install -y golang=2:1.26* python=3.14