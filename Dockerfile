FROM 10.1.35.36:5000/devops/centos:7
# FROM scratch  
COPY cmd/server/server /opt/
COPY pkg/config/* /builds/digital-and-software/jt-video-bank/pkg/config/
COPY pkg/config/* /pkg/config/
WORKDIR /opt/
EXPOSE 9603
CMD ["/opt/server"]