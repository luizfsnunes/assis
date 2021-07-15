build:
	@bash scripts/bash/build.sh

generate:
	@./bin/main generate -config=$(c)

serve:
	@./bin/main serve -config=$(c)