build:
	@bash scripts/linux/build.sh

generate:
	@./bin/main generate -config=$(c)

serve:
	@./bin/main serve -config=$(c)