package engine

func (m *model) analyzeDiscrepancies() {
	for hash, files := range m.filesByHash {
		m.analyzeDiscrepancy(hash, files)
	}
}

func (m *model) analyzeDiscrepancy(hash string, files []*file) {
	discrepancy := false
	if len(files) != len(m.roots) {
		discrepancy = true
	}
	if !discrepancy {
		name := files[0].name
		for _, file := range files {
			if file.name != name {
				discrepancy = true
				break
			}
		}
	}

	if discrepancy {
		for _, file := range files {
			file.state = divergent
		}

		counts := make([]int, len(m.roots))

		for _, file := range files {
			for i, root := range m.roots {
				if root == file.root {
					counts[i]++
				}
			}
		}

		for _, file := range files {
			file.counts = counts
		}
	}
}
