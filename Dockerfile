FROM alpine

ARG bin_name=mystique-server

RUN apk --update --no-cache add ca-certificates
ADD ./release/${bin_name} /usr/local/bin/${bin_name}
RUN chmod 755 /usr/local/bin/${bin_name}
CMD [/usr/local/bin/${bin_name}]