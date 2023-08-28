package engine

import (
	"cmp"
	"slices"
	"strings"
)

func (f *folder) entries() []*entry {
	if len(f.children)+len(f.files) == 0 {
		return nil
	}

	result := make([]*entry, 0, len(f.children)+len(f.files))
	for _, folder := range f.children {
		result = append(result, &entry{
			kind:     kindFolder,
			name:     folder.name,
			size:     folder.size,
			modTime:  folder.modTime,
			state:    folder.state,
			progress: folder.progress,
			selected: f.selectedName == folder.name,
		})
	}
	for _, file := range f.files {
		result = append(result, &entry{
			kind:     kindRegular,
			name:     file.name,
			size:     file.size,
			modTime:  file.modTime,
			state:    file.state,
			progress: file.progress,
			counts:   file.counts,
			selected: f.selectedName == file.name,
		})
	}

	var cmp func(a, b *entry) int
	switch f.sortColumn {
	case sortByName:
		cmp = cmpByAscendingName
	case sortByTime:
		cmp = cmpByAscendingTime
	case sortBySize:
		cmp = cmpByAscendingSize
	}
	slices.SortFunc(result, cmp)
	if !f.sortAscending[f.sortColumn] {
		slices.Reverse(result)
	}

	if f.selectedName == "" && len(result) > 0 {
		result[f.selectedIdx].selected = true
		f.selectedName = result[f.selectedIdx].name
	}
	return result
}

func cmpByName(a, b *entry) int {
	return cmp.Compare(strings.ToLower(a.name), strings.ToLower(b.name))
}

func cmpBySize(a, b *entry) int {
	return cmp.Compare(a.size, b.size)
}

func cmpByTime(a, b *entry) int {
	if a.modTime.Before(b.modTime) {
		return -1
	} else if b.modTime.Before(a.modTime) {
		return 1
	}
	return 0
}

func cmpByAscendingName(a, b *entry) int {
	result := cmpByName(a, b)
	if result != 0 {
		return result
	}

	result = cmpBySize(a, b)
	if result != 0 {
		return result
	}
	return cmpByTime(a, b)
}

func cmpByAscendingTime(a, b *entry) int {
	result := cmpByTime(a, b)
	if result != 0 {
		return result
	}

	result = cmpByName(a, b)
	if result != 0 {
		return result
	}

	return cmpBySize(a, b)
}

func cmpByAscendingSize(a, b *entry) int {
	result := cmpBySize(a, b)
	if result != 0 {
		return result
	}

	result = cmpByName(a, b)
	if result != 0 {
		return result
	}

	return cmpByTime(a, b)
}
