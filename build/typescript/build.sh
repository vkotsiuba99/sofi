#!/bin/bash
docker build --no-cache -t sofi-typescript:v0.0.1 .
docker tag sofi-typescript:v0.0.1 vkotsiuba99/sofi-typescript