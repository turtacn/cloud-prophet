FROM gcr.io/distroless/static:latest
MAINTAINER Krzysztof Grygiel "kgrygiel@google.com"

ARG ARCH
COPY recommender-$ARCH /recommender

ENTRYPOINT ["/recommender"]
CMD ["--v=4", "--stderrthreshold=info", "--prometheus-address=http://prometheus.monitoring.svc"]
