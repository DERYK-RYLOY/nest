// Package knownfolders resolves the real Windows default user folders
// (Downloads, Pictures, Documents, ...) via the shell API so NEST never
// hardcodes C:\Users\... paths.
package knownfolders

import "golang.org/x/sys/windows"

// Folder is a resolved Windows known folder.
type Folder struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	Path string `json:"path"`
	Icon string `json:"icon"`
}

// Keys used across NEST to refer to a known folder.
const (
	KeyDownloads = "downloads"
	KeyDesktop   = "desktop"
	KeyDocuments = "documents"
	KeyPictures  = "pictures"
	KeyMusic     = "music"
	KeyVideos    = "videos"
)

func resolve(id *windows.KNOWNFOLDERID) string {
	p, err := windows.KnownFolderPath(id, windows.KF_FLAG_DEFAULT)
	if err != nil {
		return ""
	}
	return p
}

// Downloads returns the real Downloads path ("" when unavailable).
func Downloads() string { return resolve(windows.FOLDERID_Downloads) }

// Desktop returns the real Desktop path.
func Desktop() string { return resolve(windows.FOLDERID_Desktop) }

// All returns every known folder NEST is aware of.
func All() []Folder {
	defs := []struct {
		key, name, icon string
		id              *windows.KNOWNFOLDERID
	}{
		{KeyDownloads, "Downloads", "download", windows.FOLDERID_Downloads},
		{KeyDocuments, "Documents", "file-text", windows.FOLDERID_Documents},
		{KeyPictures, "Pictures", "image", windows.FOLDERID_Pictures},
		{KeyVideos, "Videos", "video", windows.FOLDERID_Videos},
		{KeyMusic, "Music", "music", windows.FOLDERID_Music},
		{KeyDesktop, "Desktop", "monitor", windows.FOLDERID_Desktop},
	}

	out := make([]Folder, 0, len(defs))
	for _, d := range defs {
		if p := resolve(d.id); p != "" {
			out = append(out, Folder{Key: d.key, Name: d.name, Path: p, Icon: d.icon})
		}
	}
	return out
}

// NameOf returns the display name for a known folder key, or "" when unknown.
func NameOf(key string) string {
	switch key {
	case KeyDownloads:
		return "Downloads"
	case KeyDesktop:
		return "Desktop"
	case KeyDocuments:
		return "Documents"
	case KeyPictures:
		return "Pictures"
	case KeyMusic:
		return "Music"
	case KeyVideos:
		return "Videos"
	}
	return ""
}

// Lookup returns the path for a known folder key, or "" when unknown.
func Lookup(key string) string {
	switch key {
	case KeyDownloads:
		return Downloads()
	case KeyDesktop:
		return Desktop()
	case KeyDocuments:
		return resolve(windows.FOLDERID_Documents)
	case KeyPictures:
		return resolve(windows.FOLDERID_Pictures)
	case KeyMusic:
		return resolve(windows.FOLDERID_Music)
	case KeyVideos:
		return resolve(windows.FOLDERID_Videos)
	}
	return ""
}
