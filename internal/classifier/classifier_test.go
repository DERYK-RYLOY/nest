package classifier

import "testing"

func TestResolveKnownExtension(t *testing.T) {
	d := ResolvePath(`C:\Users\me\Downloads\holiday.JPG`)
	if d.Kind != "known" || d.KnownKey != "pictures" {
		t.Errorf("want pictures known folder, got %+v", d)
	}
}

func TestResolveUnknownExtensionClusters(t *testing.T) {
	d := Resolve(".xyz")
	if d.Kind != "cluster" || d.Cluster != "_XYZ_Files" || !d.Custom {
		t.Errorf("want a custom _XYZ_Files cluster, got %+v", d)
	}
}

func TestResolveCustomKnownKeyIsCaseInsensitive(t *testing.T) {
	d := ResolveCustom(".png", "Pictures")
	if d.Kind != "known" || d.KnownKey != "pictures" || d.Category != "Pictures" {
		t.Errorf("want the real Pictures folder, got %+v", d)
	}
}

func TestResolveCustomCluster(t *testing.T) {
	d := ResolveCustom(".blend", "_Blender")
	if d.Kind != "cluster" || d.Cluster != "_Blender" || d.Category != "Blender" {
		t.Errorf("want _Blender cluster, got %+v", d)
	}
}

func TestCategoriesDoesNotShareExtensionSlices(t *testing.T) {
	cats := Categories()
	if len(cats) == 0 || len(cats[0].Exts) == 0 {
		t.Fatal("empty registry")
	}
	cats[0].Exts[0] = ".mutated"

	if got := Categories()[0].Exts[0]; got == ".mutated" {
		t.Error("caller mutated the shared registry")
	}
}
