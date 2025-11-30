#!/bin/bash

for i in {1..10000};
do
	STR="$(tr -dc A-Za-z0-9 </dev/urandom | head -c 13; echo)"
	curl http://localhost:8080 \
	-X POST \
	-H "Content-Type: text/plain; charset=utf-8" \
	-i \
	-d "https://practicum.yandex.ru/${STR}"
done
