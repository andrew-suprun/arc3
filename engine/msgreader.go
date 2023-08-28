package engine

import (
	"arc/message"
	"arc/parser"
	"fmt"
	"io"
	"strings"
)

func (m *model) readMessagesX(input io.Reader, events chan any) {
	msg := &parser.Message{}
	for !m.quit {
		switch msg.Type {
		case "folder-scanned":
			events <- message.FolderScanned{
				Root: msg.StringValue("root"),
				Path: msg.StringValue("path"),
			}

		case "file-scanned":
			events <- message.FileScanned{
				Root:    msg.StringValue("root"),
				Path:    msg.StringValue("path"),
				Size:    msg.Int("size"),
				ModTime: msg.Time("mod-time"),
			}

		case "archive-scanned":
			events <- message.ArchiveScanned{
				Root: msg.StringValue("root"),
			}

		case "hashing-progress":
			events <- message.HashingProgress{
				Root:   msg.StringValue("root"),
				Path:   msg.StringValue("path"),
				Hashed: msg.Int("size"),
			}

		case "file-hashed":
			events <- message.FileHashed{
				Root: msg.StringValue("root"),
				Path: msg.StringValue("path"),
				Hash: msg.StringValue("hash"),
			}

		case "archive-hashed":
			events <- message.ArchiveHashed{
				Root: msg.StringValue("root"),
			}

		case "copying-progress":
			events <- message.CopyingProgress{
				Root:   msg.StringValue("root"),
				Path:   msg.StringValue("path"),
				Copied: msg.Int("size"),
			}

		case "file-copied":
			events <- message.FileCopied{
				Root:    msg.StringValue("root"),
				Path:    msg.StringValue("path"),
				ToRoots: strings.Split(msg.StringValue("to"), ":"),
			}

		case "file-moved":
			events <- message.FileMoved{
				Root:     msg.StringValue("root"),
				FromPath: msg.StringValue("from-path"),
				ToPath:   msg.StringValue("to-path"),
			}

		case "file-deleted":
			events <- message.FileDeleted{
				Root: msg.StringValue("root"),
				Path: msg.StringValue("path"),
			}

		default:
			panic(fmt.Sprintf("UNKNOWN event type %q", msg.Type))
		}
	}
}
