# One million checkboxes (lite)

This is a non conforming replica of the original OMCB website. The major difference is that we do not send full state updates periodically. There are some other differences, but this is the primary one.

Because this project not going to get as much traffic as the original one (or any real traffic at all), we have a bot script that can mimic a large number of connected clients

## API

The API is written in go. To run the API you can

```bash
cd api/
go mod tidy
go run . localhost:8000
# Or
air localhost:8000
```

## Frontend

The frontend is written in React. It is not built using CRA. To run the frontend you can

```bash
cd frontend/
npm i
npx parcel index.html
```

## Bot

The bot is written in go. It create 1000 clients by default and starts sending 1 event per client per second.

```bash
cd bot/
go mod tidy
go run .
```
