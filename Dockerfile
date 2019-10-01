FROM iron/go:dev

WORKDIR /app
ENV SRC_DIR=/go/src/jessica/
ADD . $SRC_DIR

RUN cd $SRC_DIR; go build -o jessica; cp jessica /app/;
ADD dist/. /app/

ENTRYPOINT ["./jessica"]
