build-win:
	@go build -ldflags "-s -w" -o assis cmd/main.go
	@start cmd scripts/windows/build.bat

generate-win:
	assis.exe generate -config=_site/config.json

serve-win:
	assis.exe serve -config=_site/config.json

build-linux:
	@go build -ldflags "-s -w" -o main cmd/main.go
	@bash scripts/linux/build.sh

generate-linux:
	@cd bin && ./main generate -config=../_site/config.json

serve-linux:
	@cd bin && ./main serve -config=../_site/config.json