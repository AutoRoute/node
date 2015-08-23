FROM base/archlinux
MAINTAINER Colin L. Rice
ENTRYPOINT ["/usr/bin/autoroute"]
EXPOSE 34321
CMD ["-listen=0.0.0.0:34321"]
ADD autoroute/autoroute /usr/bin/

