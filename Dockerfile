FROM gcr.io/distroless/static:nonroot

COPY dependency-track-exporter /

USER nonroot
ENTRYPOINT ["/dependency-track-exporter"]
