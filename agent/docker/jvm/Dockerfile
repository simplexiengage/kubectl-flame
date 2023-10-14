FROM golang:1.14-buster as agentbuild
WORKDIR /go/src/github.com/simplexiengage/kubectl-flame
ADD . /go/src/github.com/simplexiengage/kubectl-flame
RUN go get -d -v ./...
RUN cd agent && go build -o /go/bin/agent

FROM openjdk:8 as asyncprofiler

RUN sed -i -e 's/deb.debian.org/archive.debian.org/g' \
           -e 's|security.debian.org|archive.debian.org/|g' \
           -e '/stretch-updates/d' /etc/apt/sources.list

#RUN echo deb http://deb.debian.org/debian stretch-backports main > /etc/apt/sources.list.d/stretch-backports.list && \
#    apt-get update && \
#    apt-get -y install build-essential

RUN git clone https://github.com/edeNFed/async-profiler.git
RUN cd async-profiler && make

FROM bitnami/minideb:stretch
RUN mkdir -p /app/async-profiler/build
COPY --from=agentbuild /go/bin/agent /app
COPY --from=asyncprofiler /async-profiler/build /app/async-profiler/build
COPY --from=asyncprofiler /async-profiler/profiler.sh /app/async-profiler
CMD [ "/app/agent" ]