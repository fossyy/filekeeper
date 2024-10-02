#!/bin/bash

air &
templ generate -watch &
npx tailwindcss -i ./public/input.css -o ./public/output.css --watch
