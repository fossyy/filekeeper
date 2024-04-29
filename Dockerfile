FROM node:current-alpine3.19 AS node_builder

WORKDIR /src
COPY /public /src/public
COPY tailwind.config.js .
COPY /view /src/view

RUN npm install -g tailwindcss
RUN npm install -g javascript-obfuscator
RUN npx tailwindcss -i ./public/input.css -o ./public/output.css
RUN javascript-obfuscator ./public/upload.js --compact true --self-defending true --output ./public/upload_obfuscated.js
RUN javascript-obfuscator ./public/validatePassword.js --compact true --self-defending true --output ./public/validatePassword_obfuscated.js

FROM golang:1.22.2-alpine3.19 AS go_builder

WORKDIR /src
COPY . .
COPY --from=node_builder /src/public /src/public
COPY --from=node_builder /src/public/upload_obfuscated.js /src/public/upload.js
COPY --from=node_builder /src/public/validatePassword_obfuscated.js /src/public/validatePassword.js

RUN apk update && apk upgrade && apk add --no-cache ca-certificates
RUN update-ca-certificates
RUN go install github.com/a-h/templ/cmd/templ@$(go list -m -f '{{ .Version }}' github.com/a-h/templ)
RUN templ generate
RUN go build -o ./tmp/main
RUN rm /src/public/validatePassword_obfuscated.js /src/public/upload_obfuscated.js

FROM scratch

WORKDIR /src

COPY --from=go_builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=go_builder /src/schema.sql /src
COPY --from=go_builder /src/public /src/public
COPY --from=go_builder /src/tmp/main /src

ENTRYPOINT ["./main"]
