rm log-*.log

go build -o _build/log cmd/log/log.go && \
go build -o _build/arc cmd/arc/arc.go && \
go build -o _build/engine cmd/engine/engine.go && \
go build -o _build/fstest cmd/fstest/fstest.go && \

# export ARC_ENGINE="_build/log -o log-arc-engine.log -e _build/engine"
# export ARC_FS="_build/log -o log-engine-fs.log -e _build/fstest"

export ARC_ENGINE="_build/engine"
export ARC_FS="_build/fstest"

_build/arc origin "copy 1" "copy 2"
