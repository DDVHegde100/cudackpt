BUILD_DIR := build
GO_OUT := $(BUILD_DIR)/cudackpt
SHIM := $(BUILD_DIR)/libcudackpt.so
VECTORADD := $(BUILD_DIR)/vectoradd
CUBLAS := $(BUILD_DIR)/cublas_gemm

.PHONY: all clean test shim go vectoradd cublas install smoke checkpoint e2e e2e-fast e2e-cublas e2e-pipeline restore validate all-tests bench go-test install-systemd gc-images

all: shim go vectoradd

install: all
	install -d $(DESTDIR)/usr/lib $(DESTDIR)/usr/bin
	install -m 755 $(SHIM) $(DESTDIR)/usr/lib/libcudackpt.so
	install -m 755 $(GO_OUT) $(DESTDIR)/usr/bin/cudackpt

shim: $(SHIM) $(VECTORADD) $(CUBLAS)

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(SHIM) $(VECTORADD) $(CUBLAS): | $(BUILD_DIR)
	cmake -S . -B $(BUILD_DIR) -DCMAKE_BUILD_TYPE=Release
	cmake --build $(BUILD_DIR) -j

go: $(GO_OUT)

$(GO_OUT): $(shell find cmd pkg internal third_party -name '*.go' 2>/dev/null)
	go build -o $(GO_OUT) ./cmd/cudackpt

go-test:
	go test ./...

bench: go-test test
	./scripts/bench.sh

vectoradd: $(VECTORADD)

test: shim
	cmake --build $(BUILD_DIR) --target tracker_test
	$(BUILD_DIR)/tracker_test

clean:
	rm -rf $(BUILD_DIR)

smoke: all
	./scripts/run_shim_smoke.sh

checkpoint: all
	sudo -E ./scripts/run_checkpoint.sh

e2e: all
	sudo -E ./scripts/run_e2e.sh

e2e-fast: all
	sudo -E ./scripts/run_e2e_fast.sh

e2e-cublas: all
	sudo -E ./scripts/run_e2e_cublas.sh

e2e-pipeline: all
	sudo -E ./scripts/run_e2e_pipeline.sh

restore: all
	@test -n "$(IMAGE)" || (echo "usage: make restore IMAGE=/path/to/image" && exit 2)
	sudo -E ./scripts/run_restore_only.sh $(IMAGE)

all-tests: all
	sudo -E ./scripts/run_all.sh

validate: go
	@test -n "$(IMAGE)" || (echo "usage: make validate IMAGE=/path/to/image" && exit 2)
	$(GO_OUT) validate $(IMAGE)

install-systemd:
	sudo -E ./scripts/install-systemd.sh

gc-images: go
	./scripts/gc_images.sh

