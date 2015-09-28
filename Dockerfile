FROM concourse/busyboxplus:git

# make Go's SSL stdlib happy
RUN cat /etc/ssl/certs/*.pem > /etc/ssl/certs/ca-certificates.crt

ADD http://stedolan.github.io/jq/download/linux64/jq /usr/local/bin/jq
RUN chmod +x /usr/local/bin/jq

ADD built-check /opt/resource/check
ADD built-in /opt/resource/in
ADD built-out /opt/resource/out

ADD test/ /opt/resource-tests/
RUN /opt/resource-tests/all.sh && \
  rm -rf /tmp/*
