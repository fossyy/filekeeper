run: generate tailwindcss
	@go build -o ./tmp/main.exe
	@./tmp/main.exe

generate:
	@templ generate

tailwindcss:
	@npx tailwindcss -i ./public/input.css -o ./public/output.css