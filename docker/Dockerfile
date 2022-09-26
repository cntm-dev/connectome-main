
FROM tianon/ubuntu-core:14.04
ENV cntmOLOGY_PATH /var/cntmology
RUN mkdir -p $cntmOLOGY_PATH
COPY cntmology $cntmOLOGY_PATH
EXPOSE 20334 20335 20336 20337 20338 20339
WORKDIR $cntmOLOGY_PATH
ENTRYPOINT ["./cntmology"]

