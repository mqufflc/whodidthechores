FROM gcr.io/distroless/static
COPY whodidthechores /

ENTRYPOINT ["/whodidthechores"]