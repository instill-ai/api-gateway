ARG ELASTIC_STACK_VERSION

# https://github.com/elastic/beats/tree/master/filebeat
FROM docker.elastic.co/beats/filebeat:${ELASTIC_STACK_VERSION}

# Add healthcheck
HEALTHCHECK CMD filebeat test config && filebeat test output

# EXPOSE metrics port
EXPOSE 5066
