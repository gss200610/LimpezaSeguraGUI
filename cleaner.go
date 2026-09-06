package main

import (
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type Options struct {
	UserTemp, WindowsTemp, SystemWebCache, RecycleBin, WindowsLogs, CrashDumps bool
	Chrome, Edge, Firefox, Brave, Opera, Vivaldi bool
	OneDrive, Spotify, Adobe, VMware, TeamViewer bool
}

func (o Options) Any() bool {
	return o.UserTemp || o.WindowsTemp || o.SystemWebCache || o.RecycleBin || o.WindowsLogs || o.CrashDumps ||
		o.Chrome || o.Edge || o.Firefox || o.Brave || o.Opera || o.Vivaldi ||
		o.OneDrive || o.Spotify || o.Adobe || o.VMware || o.TeamViewer
}

func (o Options) Aggressive() bool {
	return o.SystemWebCache || o.RecycleBin || o.WindowsLogs || o.CrashDumps ||
		o.OneDrive || o.Spotify || o.Adobe || o.VMware || o.TeamViewer
}

type Target struct { Label, Path string; Patterns []string }
type Result struct { Found, Freed int64; Errors int }

func (r Result) Summary(clean bool) string {
	if clean {
		return fmt.Sprintf("Concluído: %s liberados; %d aviso(s).", formatBytes(r.Freed), r.Errors)
	}
	return fmt.Sprintf("Análise concluída: %s podem ser removidos.", formatBytes(r.Found))
}

func Process(opts Options, clean bool, log func(string)) Result {
	targets := DiscoverTargets(opts)
	seen := make(map[string]bool)
	var result Result
	for _, target := range targets {
		key := strings.ToLower(filepath.Clean(target.Path))
		if key == "." || seen[key] { continue }
		seen[key] = true
		before := targetSize(target)
		if before == 0 { continue }
		result.Found += before
		if !clean {
			log(fmt.Sprintf("%s: %s", target.Label, formatBytes(before)))
			continue
		}
		errCount := clearTarget(target)
		after := targetSize(target)
		freed := before - after
		if freed < 0 { freed = 0 }
		result.Freed += freed
		result.Errors += errCount
		log(fmt.Sprintf("%s: %s liberados", target.Label, formatBytes(freed)))
	}
	if opts.RecycleBin {
		systemDrive := os.Getenv("SystemDrive")
		recyclePath := ""
		if len(systemDrive) == 2 && systemDrive[1] == ':' {
			recyclePath = filepath.Join(systemDrive+string(os.PathSeparator), "$Recycle.Bin")
		}
		before := folderSize(recyclePath)
		result.Found += before
		if clean {
			err := exec.Command("PowerShell.exe", "-NoProfile", "-Command", "Clear-RecycleBin -Force -Confirm:$false").Run()
			if err != nil { result.Errors++; log("Lixeira: não foi possível esvaziar") } else {
				result.Freed += before
				log(fmt.Sprintf("Lixeira: %s liberados", formatBytes(before)))
			}
		} else if before > 0 { log(fmt.Sprintf("Lixeira: %s", formatBytes(before))) }
	}
	if len(targets) == 0 && !opts.RecycleBin { log("Nenhuma pasta correspondente foi encontrada.") }
	return result
}

func DiscoverTargets(o Options) []Target {
	local := os.Getenv("LOCALAPPDATA")
	windir := os.Getenv("WINDIR")
	var out []Target
	add := func(label, path string, patterns ...string) {
		if path != "" && filepath.IsAbs(path) { out = append(out, Target{label, filepath.Clean(path), patterns}) }
	}
	if o.UserTemp { add("Temporários do usuário", os.TempDir()) }
	if o.WindowsTemp { add("Temporários do Windows", filepath.Join(windir, "Temp")) }
	if o.SystemWebCache {
		add("Cache da Internet do Windows", filepath.Join(local, "Microsoft", "Windows", "INetCache", "IE"))
		add("Logs do WebCache", filepath.Join(local, "Microsoft", "Windows", "WebCache"), "*.log")
	}
	if o.WindowsLogs {
		add("Logs do Windows", filepath.Join(windir, "Logs"), "*.log")
		add("Logs de instalação do Windows", filepath.Join(windir, "Panther"), "*.log")
		add("Logs de drivers", filepath.Join(windir, "inf"), "*.log")
		add("Logs do Windows Update", filepath.Join(windir, "SoftwareDistribution"), "*.log")
		add("Logs do .NET", filepath.Join(windir, "Microsoft.NET"), "*.log")
	}
	if o.CrashDumps { add("Dumps de falhas", filepath.Join(local, "CrashDumps"), "*.dmp") }
	if o.Chrome { out = append(out, chromiumTargets(filepath.Join(local, "Google", "Chrome", "User Data"), "Chrome")...) }
	if o.Edge { out = append(out, chromiumTargets(filepath.Join(local, "Microsoft", "Edge", "User Data"), "Edge")...) }
	if o.Brave { out = append(out, chromiumTargets(filepath.Join(local, "BraveSoftware", "Brave-Browser", "User Data"), "Brave")...) }
	if o.Vivaldi { out = append(out, chromiumTargets(filepath.Join(local, "Vivaldi", "User Data"), "Vivaldi")...) }
	if o.Firefox {
		root := filepath.Join(local, "Mozilla", "Firefox", "Profiles")
		profiles, _ := os.ReadDir(root)
		for _, p := range profiles { if p.IsDir() {
			add("Firefox - "+p.Name()+" - Cache", filepath.Join(root, p.Name(), "cache2"))
			add("Firefox - "+p.Name()+" - Inicialização", filepath.Join(root, p.Name(), "startupCache"))
		} }
	}
	if o.Opera {
		for _, name := range []string{"Opera Stable", "Opera GX Stable"} {
			root := filepath.Join(local, "Opera Software", name)
			for _, cache := range []string{"Cache", "Code Cache", "GPUCache"} { add(name+" - "+cache, filepath.Join(root, cache)) }
		}
	}
	if o.OneDrive {
		add("Logs do OneDrive", filepath.Join(local, "Microsoft", "OneDrive"), "*.log", "*.odl", "*.aodl", "*.otc", "*.qmlc")
	}
	if o.Spotify {
		add("Dados temporários do Spotify", filepath.Join(local, "Spotify", "Data"), "*.file")
		add("Cache web do Spotify", filepath.Join(local, "Spotify", "Browser", "Cache"))
		add("Cache gráfico do Spotify", filepath.Join(local, "Spotify", "Browser", "GPUCache"))
	}
	if o.Adobe {
		roaming := os.Getenv("APPDATA")
		add("Adobe Media Cache", filepath.Join(roaming, "Adobe", "Common", "Media Cache Files"))
		add("Logs da Adobe", filepath.Join(roaming, "Adobe"), "*.log")
	}
	if o.VMware { add("Logs do VMware", filepath.Join(os.Getenv("ProgramData"), "VMware", "logs"), "*.log") }
	if o.TeamViewer {
		add("Cache web do TeamViewer", filepath.Join(local, "TeamViewer", "EdgeBrowserControl"),
			"data_*", "f_*", "index.*", "*_0", "*_1", "*_2", "*_3", "*_4", "*_5")
	}
	return out
}

func chromiumTargets(root, browser string) []Target {
	entries, _ := os.ReadDir(root)
	var out []Target
	for _, e := range entries {
		if !e.IsDir() || !(e.Name() == "Default" || e.Name() == "Guest Profile" || strings.HasPrefix(e.Name(), "Profile ")) { continue }
		for _, cache := range []string{"Cache", "Code Cache", "GPUCache"} {
			out = append(out, Target{Label: browser+" - "+e.Name()+" - "+cache, Path: filepath.Join(root, e.Name(), cache)})
		}
	}
	return out
}

func folderSize(path string) int64 {
	if !isSafeTargetPath(path) { return 0 }
	var total int64
	_ = filepath.WalkDir(path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil { return nil }
		if !d.IsDir() { if info, e := d.Info(); e == nil { total += info.Size() } }
		return nil
	})
	return total
}

func clearContents(path string) int {
	if !isSafeTargetPath(path) { return 1 }
	entries, err := os.ReadDir(path)
	if err != nil { return 1 }
	errors := 0
	for _, entry := range entries {
		if err := os.RemoveAll(filepath.Join(path, entry.Name())); err != nil { errors++ }
	}
	return errors
}

func isSafeTargetPath(path string) bool {
	if path == "" || !filepath.IsAbs(path) { return false }
	clean := filepath.Clean(path)
	volume := filepath.VolumeName(clean)
	if clean == volume+string(os.PathSeparator) { return false }
	protected := []string{
		os.Getenv("USERPROFILE"), os.Getenv("LOCALAPPDATA"), os.Getenv("APPDATA"),
		os.Getenv("WINDIR"), os.Getenv("ProgramData"),
	}
	for _, root := range protected {
		if root != "" && strings.EqualFold(clean, filepath.Clean(root)) { return false }
	}
	return true
}

func matchesTarget(target Target, name string) bool {
	if len(target.Patterns) == 0 { return true }
	name = strings.ToLower(name)
	for _, pattern := range target.Patterns {
		if ok, _ := filepath.Match(strings.ToLower(pattern), name); ok { return true }
	}
	return false
}

func targetSize(target Target) int64 {
	if len(target.Patterns) == 0 { return folderSize(target.Path) }
	var total int64
	_ = filepath.WalkDir(target.Path, func(_ string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !matchesTarget(target, d.Name()) { return nil }
		if info, e := d.Info(); e == nil { total += info.Size() }
		return nil
	})
	return total
}

func clearTarget(target Target) int {
	if !isSafeTargetPath(target.Path) { return 1 }
	if len(target.Patterns) == 0 { return clearContents(target.Path) }
	errors := 0
	_ = filepath.WalkDir(target.Path, func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !matchesTarget(target, d.Name()) { return nil }
		if err := os.Remove(path); err != nil { errors++ }
		return nil
	})
	return errors
}

func RunningBrowsers() []string {
	out, err := exec.Command("tasklist", "/FO", "CSV", "/NH").Output()
	if err != nil { return nil }
	wanted := map[string]string{"chrome.exe":"Chrome", "msedge.exe":"Edge", "firefox.exe":"Firefox", "brave.exe":"Brave", "opera.exe":"Opera", "vivaldi.exe":"Vivaldi", "onedrive.exe":"OneDrive", "spotify.exe":"Spotify"}
	found := map[string]bool{}
	text := strings.ToLower(string(out))
	for exe, name := range wanted { if strings.Contains(text, "\""+exe+"\"") { found[name] = true } }
	var names []string
	for name := range found { names = append(names, name) }
	sort.Strings(names)
	return names
}

func CloseBrowsers(o Options) (closed, remaining []string) {
	processes := []struct{ exe, name string; selected bool }{
		{"chrome.exe", "Chrome", o.Chrome}, {"msedge.exe", "Edge", o.Edge},
		{"firefox.exe", "Firefox", o.Firefox}, {"brave.exe", "Brave", o.Brave},
		{"opera.exe", "Opera", o.Opera}, {"vivaldi.exe", "Vivaldi", o.Vivaldi},
		{"onedrive.exe", "OneDrive", o.OneDrive}, {"spotify.exe", "Spotify", o.Spotify},
	}
	for _, process := range processes {
		if !process.selected { continue }
		cmd := exec.Command("taskkill", "/T", "/IM", process.exe)
		if err := cmd.Run(); err == nil { closed = append(closed, process.name) }
	}
	time.Sleep(1500 * time.Millisecond)
	stillOpen := RunningBrowsers()
	open := make(map[string]bool)
	for _, name := range stillOpen { open[name] = true }
	for _, process := range processes {
		if !process.selected || !open[process.name] { continue }
		_ = exec.Command("taskkill", "/F", "/T", "/IM", process.exe).Run()
	}
	time.Sleep(750 * time.Millisecond)
	stillOpen = RunningBrowsers()
	for _, name := range stillOpen {
		for _, process := range processes {
			if process.selected && process.name == name { remaining = append(remaining, name) }
		}
	}
	return closed, remaining
}

func formatBytes(n int64) string {
	const (kb=1024; mb=1024*kb; gb=1024*mb)
	switch { case n >= gb: return fmt.Sprintf("%.2f GB", float64(n)/gb); case n >= mb: return fmt.Sprintf("%.2f MB", float64(n)/mb); case n >= kb: return fmt.Sprintf("%.2f KB", float64(n)/kb); default: return fmt.Sprintf("%d bytes", n) }
}
