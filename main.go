//go:generate go-winres make
package main

import (
	"fmt"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

func main() {
	var mw *walk.MainWindow
	var userTemp, windowsTemp, systemWebCache, recycleBin, windowsLogs, crashDumps *walk.CheckBox
	var chrome, edge, firefox, brave, opera, vivaldi *walk.CheckBox
	var oneDrive, spotify, adobe, vmware, teamViewer *walk.CheckBox
	var analyzeButton, cleanButton *walk.PushButton
	var progress *walk.ProgressBar
	var logBox *walk.TextEdit
	var status *walk.Label

	selected := func() Options {
		return Options{
			UserTemp: userTemp.Checked(), WindowsTemp: windowsTemp.Checked(),
			SystemWebCache: systemWebCache.Checked(), RecycleBin: recycleBin.Checked(),
			WindowsLogs: windowsLogs.Checked(), CrashDumps: crashDumps.Checked(),
			Chrome: chrome.Checked(), Edge: edge.Checked(), Firefox: firefox.Checked(),
			Brave: brave.Checked(), Opera: opera.Checked(), Vivaldi: vivaldi.Checked(),
			OneDrive: oneDrive.Checked(), Spotify: spotify.Checked(), Adobe: adobe.Checked(),
			VMware: vmware.Checked(), TeamViewer: teamViewer.Checked(),
		}
	}

	setBusy := func(busy bool) {
		analyzeButton.SetEnabled(!busy)
		cleanButton.SetEnabled(!busy)
		progress.SetVisible(busy)
		if busy { progress.SetMarqueeMode(true) } else { progress.SetMarqueeMode(false) }
	}

	run := func(clean bool) {
		opts := selected()
		if !opts.Any() {
			walk.MsgBox(mw, "Nenhuma opção", "Marque pelo menos uma categoria.", walk.MsgBoxIconInformation)
			return
		}
		if clean {
			warning := "Os navegadores e aplicativos selecionados serão fechados quando necessário. Abas ou formulários não salvos poderão ser perdidos.\n\nOs caches e temporários selecionados serão apagados. No modo padrão, senhas, cookies, histórico, favoritos e extensões serão preservados."
			if opts.Aggressive() {
				warning += "\n\nATENÇÃO: você marcou uma opção avançada. Ela pode esvaziar a Lixeira, remover logs de diagnóstico, fechar o OneDrive ou provocar reconstrução de caches."
			}
			answer := walk.MsgBox(mw, "Confirmar limpeza",
				warning+"\n\nDeseja continuar?",
				walk.MsgBoxYesNo|walk.MsgBoxIconWarning)
			if answer != walk.DlgCmdYes { return }
		}

		setBusy(true)
		logBox.SetText("")
		if clean { status.SetText("Limpando...") } else { status.SetText("Analisando...") }

		go func() {
			if clean {
				closed, remaining := CloseBrowsers(opts)
				if len(closed) > 0 {
					mw.Synchronize(func() { logBox.SetText("Aplicativos fechados: " + strings.Join(closed, ", ")) })
				}
				if len(remaining) > 0 {
					mw.Synchronize(func() {
						setBusy(false)
						status.SetText("Limpeza cancelada: ainda há navegador aberto.")
						walk.MsgBox(mw, "Limpeza cancelada", "Não foi possível fechar: "+strings.Join(remaining, ", ")+". Feche manualmente e tente novamente.", walk.MsgBoxIconWarning)
					})
					return
				}
			}
			result := Process(opts, clean, func(line string) {
				mw.Synchronize(func() {
					old := logBox.Text()
					if old != "" { old += "\r\n" }
					logBox.SetText(old + line)
				})
			})
			mw.Synchronize(func() {
				setBusy(false)
				status.SetText(result.Summary(clean))
				if clean && result.Errors == 0 {
					walk.MsgBox(mw, "Limpeza concluída", result.Summary(true), walk.MsgBoxIconInformation)
				}
			})
		}()
	}

	err := (MainWindow{
		AssignTo: &mw,
		Title: "Limpeza Segura 2.0 - Windows 11",
		MinSize: Size{Width: 760, Height: 690},
		Size: Size{Width: 840, Height: 760},
		Layout: VBox{MarginsZero: false, Spacing: 8},
		Children: []Widget{
			Label{Text: "Limpeza segura de temporários e caches", Font: Font{PointSize: 14, Bold: true}},
			Label{Text: "O modo padrão preserva senhas, cookies, histórico, favoritos e extensões."},
			GroupBox{Title: "O que deseja limpar", Layout: Grid{Columns: 2}, Children: []Widget{
				CheckBox{AssignTo: &userTemp, Text: "Temporários do usuário", Checked: true},
				CheckBox{AssignTo: &windowsTemp, Text: "Temporários do Windows", Checked: true},
				CheckBox{AssignTo: &chrome, Text: "Google Chrome", Checked: true},
				CheckBox{AssignTo: &edge, Text: "Microsoft Edge", Checked: true},
				CheckBox{AssignTo: &firefox, Text: "Mozilla Firefox", Checked: true},
				CheckBox{AssignTo: &brave, Text: "Brave", Checked: true},
				CheckBox{AssignTo: &opera, Text: "Opera e Opera GX", Checked: true},
				CheckBox{AssignTo: &vivaldi, Text: "Vivaldi", Checked: true},
			}},
			GroupBox{Title: "Avançado — desmarcado por segurança", Layout: Grid{Columns: 2}, Children: []Widget{
				CheckBox{AssignTo: &recycleBin, Text: "Esvaziar Lixeira", Checked: false},
				CheckBox{AssignTo: &systemWebCache, Text: "INetCache/WebCache (pode encerrar sessões)", Checked: false},
				CheckBox{AssignTo: &windowsLogs, Text: "Logs de diagnóstico do Windows", Checked: false},
				CheckBox{AssignTo: &crashDumps, Text: "Dumps de falhas de programas", Checked: false},
				CheckBox{AssignTo: &oneDrive, Text: "Logs do OneDrive (fecha o OneDrive)", Checked: false},
				CheckBox{AssignTo: &spotify, Text: "Cache do Spotify", Checked: false},
				CheckBox{AssignTo: &adobe, Text: "Adobe Media Cache", Checked: false},
				CheckBox{AssignTo: &vmware, Text: "Logs do VMware", Checked: false},
				CheckBox{AssignTo: &teamViewer, Text: "Cache web do TeamViewer", Checked: false},
			}},
			Composite{Layout: HBox{}, Children: []Widget{
				PushButton{AssignTo: &analyzeButton, Text: "Analisar", OnClicked: func() { run(false) }},
				PushButton{AssignTo: &cleanButton, Text: "Limpar selecionados", OnClicked: func() { run(true) }},
				HSpacer{},
				PushButton{Text: "Fechar", OnClicked: func() { mw.Close() }},
			}},
			ProgressBar{AssignTo: &progress, Visible: false},
			Label{AssignTo: &status, Text: "Pronto. Feche os navegadores para liberar todos os arquivos."},
			TextEdit{AssignTo: &logBox, ReadOnly: true, VScroll: true},
			Label{Text: "Baseado no Baboo v3.2, com dados de sites, certificados e bancos internos protegidos."},
		},
	}).Create()
	if err != nil {
		walk.MsgBox(nil, "Erro", fmt.Sprintf("Não foi possível abrir a interface:\n%s", err), walk.MsgBoxIconError)
		return
	}

	if names := RunningBrowsers(); len(names) > 0 {
		status.SetText("Atenção: estes aplicativos estão abertos: " + strings.Join(names, ", "))
	}
	mw.Run()
}
