build:
	@bash scripts/linux/build.sh

generate:
	@./bin/main generate -folder=$(f) -config=$(c)

serve:
	@./bin/main serve -folder=$(f) -config=$(c)