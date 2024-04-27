FROM node:current-alpine3.19 AS tailwind

WORKDIR /src
COPY ./public/input.css ./public/
COPY tailwind.config.js .
COPY ./view ./view

RUN npm install -g tailwindcss
RUN npx tailwindcss -i ./public/input.css -o ./public/output.css

FROM golang:1.22.2-alpine3.19 AS go_builder

WORKDIR /src
COPY . .
COPY --from=tailwind /src/public/output.css ./public/

RUN go install github.com/a-h/templ/cmd/templ@$(go list -m -f '{{ .Version }}' github.com/a-h/templ)
RUN templ generate
RUN go build -o ./tmp/main

FROM scratch

WORKDIR /src

COPY --from=go_builder /src /src

ENTRYPOINT ["./tmp/main"]
