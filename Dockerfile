FROM golang:latest
ENV GIN_MODE=release

ARG TARGETPLATFORM
COPY artifacts/build/release/$TARGETPLATFORM/* /bin/

ENV GIT_ASKPASS="/bin/askpass"

ENTRYPOINT ["/bin/browser"]
