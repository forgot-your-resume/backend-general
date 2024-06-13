# build
FROM --platform=amd64 golang:1.22.4-alpine AS build

WORKDIR /app
ADD . .

RUN ls

RUN go mod download
RUN GOARCH=amd64 go build .

# final
FROM --platform=amd64 alpine

WORKDIR /app
COPY --from=build /app/backend-general .

EXPOSE 8080

CMD ["/app/backend-general"]