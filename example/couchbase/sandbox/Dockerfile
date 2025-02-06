FROM public.ecr.aws/docker/library/couchbase:7.6.5

COPY scripts/configure-node.sh /etc/service/config-couchbase/run
RUN chmod +x /etc/service/config-couchbase/run && chown couchbase:couchbase /etc/service/config-couchbase/run
COPY scripts/create-index.json /opt/couchbase
