FROM golang:1.24.3-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o weatherapp .

FROM scratch
WORKDIR /
COPY --from=build /app/weatherapp /bin/weatherapp
ENV PORT=80
EXPOSE 80
CMD [ "/bin/weatherapp" ]