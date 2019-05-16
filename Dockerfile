FROM alpine:3.9

ARG release_dir=./release

ARG bin_name=mystique-server
ENV bin_name ${bin_name}

RUN apk --update --no-cache add ca-certificates
ADD ${release_dir}/${bin_name}-linux-amd64 /usr/local/bin/${bin_name}
RUN chmod 755 /usr/local/bin/${bin_name}
CMD /usr/local/bin/${bin_name}
