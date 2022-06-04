FROM golang:1.18-alpine AS base
WORKDIR /app

FROM base AS build
COPY . .
RUN go build -v -o notdeadyet

FROM base AS final
COPY --from=build /app/notdeadyet .
EXPOSE 80
CMD [ "/app/notdeadyet", "--config-file=/config/config.yml" ]