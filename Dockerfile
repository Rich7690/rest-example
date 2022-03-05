FROM gcr.io/distroless/static
COPY rest-example /
ENTRYPOINT ["/rest-example"]
