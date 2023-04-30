FROM golang:1.20.3-alpine3.17
RUN mkdir app
WORKDIR /app
RUN mkdir backend
WORKDIR /app/backend
COPY . .
EXPOSE 5000
CMD ["go", "run", "main.go"]