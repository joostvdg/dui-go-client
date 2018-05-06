#!/usr/bin/env bash
docker run --rm -ti --init --name dui-go-1 --network dui-test -p 7777:7777 dui-go
