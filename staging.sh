#!/bin/bash

air &
npx tailwindcss -i ./public/input.css -o ./public/output.css --watch &
templ generate -watch -proxy=http://localhost:8000
