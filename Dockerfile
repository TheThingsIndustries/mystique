FROM alpine:3.15

ARG release_dir=./release

ARG bin_name=mystique-server
ENV bin_name ${bin_name}
ARG TARGETARCH=amd64

RUN apk --update --no-cache add ca-certificates
ADD ${release_dir}/${bin_name}-linux-${TARGETARCH} /usr/local/bin/${bin_name}
RUN chmod 755 /usr/local/bin/${bin_name}
CMD /usr/local/bin/${bin_name}
