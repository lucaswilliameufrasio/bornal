package running

import "testing"

func TestWinePathCountsAsTheSameInstall(t *testing.T) {
	install := "/mnt/games/World of Warcraft/_forever_"
	if !refers(install, `C:\games\World of Warcraft\_forever_\Wow.exe`) {
		t.Fatal("caminho do Wine não contou")
	}
	if refers(install, `C:\games\World of Warcraft\_classic_\Wow.exe`) {
		t.Fatal("outro flavor contou")
	}
	if refers(install, "/mnt/games/World of Warcraft/_forever_/Wow.exe") == false {
		t.Fatal("caminho local não contou")
	}
}

func TestOpenUsesTheInjectedList(t *testing.T) {
	prev := List
	t.Cleanup(func() { List = prev })
	List = func() ([]Process, error) {
		return []Process{{Cmdline: `Z:\games\World of Warcraft\_forever_\Wow.exe`}}, nil
	}
	open, err := Open("/games/World of Warcraft/_forever_")
	if err != nil || !open {
		t.Fatalf("open %v %v", open, err)
	}
}
