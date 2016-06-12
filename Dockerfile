FROM concourse/buildroot:git

ADD assets/ /opt/resource/

ADD test/ /opt/resource-tests/
RUN /opt/resource-tests/all.sh && \
  rm -rf /tmp/*
