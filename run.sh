go build -o _build/fs cmd/fstest/fstest.go && \
go build -o _build/eng cmd/engine/engine.go && \
go build -o _build/arc cmd/arc/arc.go && \
_build/arc origin "copy 1" "copy 2"