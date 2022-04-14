FROM scratch
ADD /go-workflow //
EXPOSE 8080
ENTRYPOINT [ "/go-workflow" ]