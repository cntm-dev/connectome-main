
FROM tianon/ubuntu-core:14.04
ENV CNTM_PATH /var/cntm
RUN mkdir -p $CNTM_PATH
COPY cntm $CNTM_PATH
EXPOSE 20334 20335 20336 20337 20338 20339
WORKDIR $CNTM_PATH
ENTRYPOINT ["./cntm"]

