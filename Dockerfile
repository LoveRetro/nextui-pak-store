FROM ghcr.io/brandonkowalski/quasimodo:latest

WORKDIR /build

COPY go.mod go.sum* ./

RUN GOWORK=off go mod download

COPY . .

ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

RUN GOWORK=off go build -v \
    -tags nodefaultfont \
    -ldflags "-X github.com/LoveRetro/nextui-pak-store/version.Version=${VERSION} \
              -X github.com/LoveRetro/nextui-pak-store/version.GitCommit=${GIT_COMMIT} \
              -X github.com/LoveRetro/nextui-pak-store/version.BuildDate=${BUILD_DATE}" \
    -o pak-store ./app

CMD ["/bin/bash"]
