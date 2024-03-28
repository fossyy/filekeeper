@echo off

REM Start the Go server using Air
start "" air

REM Watch for changes in Tailwind CSS
start "" npx tailwindcss -i ./public/input.css -o ./public/output.css --watch

REM Watch for changes in templates and proxy to Go server
start "" cmd /k "templ generate -watch -proxy=http://localhost:8000"