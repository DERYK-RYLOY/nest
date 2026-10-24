// Package classifier maps file extensions onto NEST categories and decides
// where each file belongs: either a real Windows known folder, or a
// self-created cluster folder inside the source directory.
package classifier

import (
	"path/filepath"
	"strings"

	"nest/internal/knownfolders"
)

// Category is a group of related file types.
type Category struct {
	Key     string   `json:"key"`
	Name    string   `json:"name"`
	Icon    string   `json:"icon"`
	Known   string   `json:"known"`   // known-folder key, empty => cluster in source
	Cluster string   `json:"cluster"` // cluster folder name when Known == ""
	Exts    []string `json:"exts"`
}

// Destination is the decision made for a single file.
type Destination struct {
	Kind     string `json:"kind"`     // "known" or "cluster"
	Category string `json:"category"` // display name, e.g. "Images"
	Icon     string `json:"icon"`     // icon hint for the UI
	KnownKey string `json:"knownKey"` // set when Kind == "known"
	Cluster  string `json:"cluster"`  // folder name when Kind == "cluster"
	Custom   bool   `json:"custom"`   // true for unknown-extension fallback clusters
}

var registry = []Category{
	{
		Key: "images", Name: "Images", Icon: "image", Known: "pictures",
		Exts: []string{".jpg", ".jpeg", ".png", ".gif", ".webp", ".bmp", ".tif", ".tiff", ".svg", ".heic", ".heif", ".raw", ".cr2", ".nef", ".psd", ".ai"},
	},
	{
		Key: "videos", Name: "Videos", Icon: "video", Known: "videos",
		Exts: []string{".mp4", ".mov", ".avi", ".mkv", ".wmv", ".flv", ".webm", ".m4v", ".mpeg", ".mpg", ".3gp", ".ts"},
	},
	{
		Key: "audio", Name: "Music", Icon: "music", Known: "music",
		Exts: []string{".mp3", ".wav", ".flac", ".aac", ".ogg", ".oga", ".wma", ".m4a", ".opus", ".aiff", ".mid", ".midi"},
	},
	{
		Key: "documents", Name: "Documents", Icon: "file-text", Known: "documents",
		Exts: []string{".pdf", ".doc", ".docx", ".xls", ".xlsx", ".xlsm", ".ppt", ".pptx", ".txt", ".rtf", ".odt", ".ods", ".odp", ".csv", ".md", ".epub", ".mobi"},
	},
	{
		Key: "installers", Name: "Installers", Icon: "package", Cluster: "_Installers",
		Exts: []string{".exe", ".msi", ".msix", ".appx", ".msp", ".deb", ".rpm"},
	},
	{
		Key: "archives", Name: "Archives", Icon: "archive", Cluster: "_Archives",
		Exts: []string{".zip", ".rar", ".7z", ".tar", ".gz", ".tgz", ".bz2", ".xz", ".cab", ".iso", ".img"},
	},
	{
		Key: "code", Name: "Code", Icon: "code", Cluster: "_Code",
		Exts: []string{".go", ".py", ".js", ".jsx", ".ts", ".tsx", ".json", ".html", ".css", ".scss", ".java", ".c", ".h", ".cpp", ".hpp", ".cs", ".rs", ".rb", ".php", ".sh", ".ps1", ".yml", ".yaml", ".toml", ".xml", ".sql", ".ipynb"},
	},
	{
		Key: "models", Name: "3D Models", Icon: "box", Cluster: "_3D",
		Exts: []string{".stl", ".obj", ".fbx", ".3mf", ".gltf", ".glb", ".blend", ".dae"},
	},
	{
		Key: "fonts", Name: "Fonts", Icon: "type", Cluster: "_Fonts",
		Exts: []string{".ttf", ".otf", ".woff", ".woff2", ".eot"},
	},
	{
		Key: "ebooks", Name: "Ebooks", Icon: "book", Cluster: "_Ebooks",
		Exts: []string{".azw", ".azw3", ".djvu", ".cbz", ".cbr"},
	},
}

var extIndex map[string]*Category

func init() {
	extIndex = make(map[string]*Category)
	for i := range registry {
		c := &registry[i]
		for _, e := range c.Exts {
			extIndex[strings.ToLower(e)] = c
		}
	}
}

// Categories returns a copy of the full registry for the UI. The extension
// slices are copied too, so a caller can never mutate the registry.
func Categories() []Category {
	out := make([]Category, len(registry))
	for i, c := range registry {
		out[i] = c
		out[i].Exts = append([]string(nil), c.Exts...)
	}
	return out
}

// Resolve decides where a file with the given extension belongs.
// Unknown extensions fall back to a per-extension cluster folder.
func Resolve(ext string) Destination {
	ext = strings.ToLower(ext)
	if c, ok := extIndex[ext]; ok {
		if c.Known != "" {
			return Destination{Kind: "known", Category: c.Name, Icon: c.Icon, KnownKey: c.Known}
		}
		return Destination{Kind: "cluster", Category: c.Name, Icon: c.Icon, Cluster: c.Cluster}
	}

	label := strings.ToUpper(strings.TrimPrefix(ext, "."))
	if label == "" {
		label = "Unknown"
	}
	return Destination{
		Kind:     "cluster",
		Category: label + " Files",
		Icon:     "file",
		Cluster:  "_" + label + "_Files",
		Custom:   true,
	}
}

// ResolvePath is a convenience wrapper that extracts the extension first.
func ResolvePath(path string) Destination {
	return Resolve(filepath.Ext(path))
}

// ResolveCustom applies a user rule: target is either a known-folder key or a
// cluster folder name created inside the source. Known-folder keys match
// case-insensitively ("Pictures" and "pictures" mean the same folder).
func ResolveCustom(ext, target string) Destination {
	target = strings.TrimSpace(target)
	if name := knownfolders.NameOf(strings.ToLower(target)); name != "" {
		return Destination{Kind: "known", Category: name, Icon: "file-text", KnownKey: strings.ToLower(target)}
	}
	name := strings.TrimPrefix(target, "_")
	if name == "" {
		name = strings.ToUpper(strings.TrimPrefix(ext, "."))
	}
	return Destination{Kind: "cluster", Category: name, Icon: "folder", Cluster: target}
}
