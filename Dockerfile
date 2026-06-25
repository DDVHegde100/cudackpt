FROM nvidia/cuda:12.6.0-devel-ubuntu22.04
RUN apt-get update && DEBIAN_FRONTEND=noninteractive apt-get install -y \
  cmake build-essential golang-go criu ca-certificates \
  && rm -rf /var/lib/apt/lists/*
WORKDIR /src
COPY . .
RUN make
CMD ["./scripts/run_e2e_fast.sh"]
