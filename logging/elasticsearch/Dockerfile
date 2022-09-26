ARG ELASTIC_STACK_VERSION

# https://github.com/elastic/elasticsearch-docker
FROM docker.elastic.co/elasticsearch/elasticsearch:${ELASTIC_STACK_VERSION}

# Add healthcheck
COPY scripts/docker-healthcheck .
RUN chown elasticsearch:elasticsearch ./docker-healthcheck
HEALTHCHECK CMD sh ./docker-healthcheck

# Add elasticsearch plugins setup here
# Example: RUN elasticsearch-plugin install analysis-icu
#RUN elasticsearch-plugin install --batch repository-s3

USER elasticsearch
