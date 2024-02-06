# Use a imagem oficial do Golang como imagem base
FROM golang:1.21-alpine as build

ENV TZ=America/Sao_Paulo

WORKDIR /code
COPY . /code/
RUN apk update && apk add tzdata
# Compile o código Go
RUN go build -o . ./main.go 

# Stage 2
FROM scratch
ENV TZ=America/Sao_Paulo
WORKDIR /

COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
# Copie o binário compilado para a imagem final
COPY --from=build /code/main /main
CMD [ "./main" ]

