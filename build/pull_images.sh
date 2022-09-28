#!/bin/bash
docker pull python:3.9.1-alpine
docker pull golang:1.18-alpine
docker pull gcc:latest
docker pull openjdk:8u232-jdk
docker pull node:lts-alpine
docker pull vkotsiuba99/sofi-typescript
docker pull julia:1.7.1-alpine
docker pull elixir:1.13.1-alpine
docker pull swift:5.5.2
docker-compose up -d