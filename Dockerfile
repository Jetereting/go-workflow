FROM jetereting/alpine:1.0.1
ADD go-workflow /
EXPOSE 8080
ENTRYPOINT [ "/go-workflow" ]