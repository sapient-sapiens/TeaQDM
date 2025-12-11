build:
	cd go && go build -o ../teaqdm.so -buildmode=c-shared main.go

run: build
	python3 python/test.py

clean:
	rm -f teaqdm.so teaqdm.h

.PHONY: build run clean