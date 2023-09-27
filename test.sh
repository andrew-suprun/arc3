rm log-*.log

go build -o _build/log cmd/log/log.go && \
go build -o _build/arc cmd/arc/arc.go && \
go build -o _build/engine cmd/engine/engine.go && \
go build -o _build/fstest cmd/fstest/fstest.go && \
_build/arc origin "copy 1" "copy 2" -- \
_build/log "log-arc-engine.log" -- \
_build/engine -- \
_build/fstest -p=false


_build/log "log-engine-fstest.log" -- \