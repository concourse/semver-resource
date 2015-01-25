FROM gliderlabs/alpine

RUN apk-install ca-certificates

ADD built-check /opt/resource/check
ADD built-in /opt/resource/in
ADD built-out /opt/resource/out
