@echo off
setlocal
chcp 65001 >nul
cd /d "%~dp0"

set "GOROOT=C:\Program Files\Go"
set "PATH=%GOROOT%\bin;%USERPROFILE%\go\bin;%PATH%"

if not exist "%GOROOT%\bin\go.exe" (
  echo ERRO: Go nao encontrado em C:\Program Files\Go
  pause
  exit /b 1
)

echo [1/4] Baixando dependencias...
go mod tidy || goto :error

echo [2/4] Instalando gerador de recursos compativel...
go install github.com/tc-hib/go-winres@latest || goto :error

echo [3/4] Incorporando manifesto de administrador...
if exist rsrc.syso del /f /q rsrc.syso
if exist rsrc_windows_amd64.syso del /f /q rsrc_windows_amd64.syso
if exist rsrc_windows_386.syso del /f /q rsrc_windows_386.syso
"%USERPROFILE%\go\bin\go-winres.exe" make || goto :error

echo [4/4] Compilando EXE...
if not exist dist mkdir dist
go build -trimpath -ldflags="-H windowsgui -s -w" -o "dist\LimpezaSeguraGUI.exe" . || goto :error

echo.
echo SUCESSO: dist\LimpezaSeguraGUI.exe
pause
exit /b 0

:error
echo.
echo A compilacao falhou. Veja a mensagem acima.
pause
exit /b 1
